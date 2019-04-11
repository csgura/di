package di

import (
	"fmt"
	"reflect"
	"time"
)

// Injector used to get instance
type Injector interface {
	GetInstance(intf interface{}) interface{}
	GetProperty(propName string) string
	SetProperty(propName string, value string)

	GetInstancesOf(intf interface{}) []interface{}
	InjectMembers(pointer interface{})
	InjectAndCall(function interface{}) interface{}
}

type injectorImpl struct {
	binder *Binder
	props  map[string]string
}

type injectorContext struct {
	injector  *injectorImpl
	loopCheck map[reflect.Type]bool
	stack     []reflect.Type
}

func (r *injectorImpl) GetInstance(intf interface{}) interface{} {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil}
	return context.GetInstance(intf)
}

func (r *injectorImpl) GetInstancesOf(intf interface{}) []interface{} {
	//fmt.Println("impl getIns")
	return r.binder.getInstancesOf(intf)
}

func (r *injectorImpl) getInstanceByType(t reflect.Type) interface{} {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil}
	return context.getInstanceByType(t)
}

func (r *injectorImpl) InjectMembers(pointer interface{}) {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil}
	context.InjectMembers(pointer)
}

func (r *injectorImpl) InjectAndCall(function interface{}) interface{} {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil}
	return context.InjectAndCall(function)
}

func (r *injectorImpl) GetProperty(propName string) string {
	return r.props[propName]
}

func (r *injectorImpl) SetProperty(propName string, value string) {
	r.props[propName] = value
}

func (r *injectorContext) GetProperty(propName string) string {
	return r.injector.GetProperty(propName)
}

func (r *injectorContext) SetProperty(propName string, value string) {
	r.injector.SetProperty(propName, value)
}

func (r *injectorContext) createInstance(t reflect.Type, p *Binding) interface{} {
	r.stack = append(r.stack, t)
	if r.loopCheck[t] == true {
		loopStr := ""
		for _, k := range r.stack {
			if loopStr == "" {
				loopStr = k.String()
			} else {
				loopStr = loopStr + "\n  -> " + k.String()
			}

		}
		panic("dependency cycle : \n" + loopStr + "\n")
	}
	r.loopCheck[t] = true

	defer func() {
		r.stack = r.stack[0 : len(r.stack)-1]
		delete(r.loopCheck, t)
	}()

	if list := r.injector.binder.creatingBefore[t]; list != nil {
		for _, shouldBefore := range list {
			r.getInstanceByType(shouldBefore.tpe)
		}
	}
	return p.provider(r)
}

func (r *injectorContext) getInstanceByType(t reflect.Type) interface{} {
	p := r.injector.binder.providers[t]
	if p != nil {
		if p.isSingleton {
			ret := p.instance
			if ret == nil {
				if p.provider != nil {
					ret = r.createInstance(t, p)
					p.instance = ret
				}
			}
			return ret
		}
		ret := r.createInstance(t, p)
		return ret

	}

	p = r.injector.binder.providersFallback[t]
	if p != nil {
		if p.isSingleton {
			ret := p.instance
			if ret == nil {
				if p.provider != nil {
					ret = r.createInstance(t, p)
					p.instance = ret
				}
			}
			return ret
		}
		ret := r.createInstance(t, p)
		return ret

	}
	return nil
}

func (r *injectorContext) GetInstance(intf interface{}) interface{} {
	//fmt.Println("context getIns")

	t := reflect.TypeOf(intf)

	return r.getInstanceByType(t)
	//fmt.Printf("type = %s\n", t.String())

}

func (r *injectorContext) GetInstancesOf(intf interface{}) []interface{} {
	//fmt.Println("impl getIns")
	return r.injector.binder.getInstancesOf(intf)
}

func hasInjectTag(tag reflect.StructTag) bool {
	value, ok := tag.Lookup("di")
	if ok && value == "inject" {
		return true
	}
	return false
}

func (r *injectorContext) InjectAndCall(function interface{}) interface{} {
	ftype := reflect.TypeOf(function)
	if ftype == nil {
		panic("function type is nil")
	}
	if ftype.Kind() != reflect.Func {
		panic(fmt.Sprintf("can't inject non-function %v (type %v)", function, ftype))
	}

	if ftype.IsVariadic() {
		panic(fmt.Sprintf("can't inject variadic function %v (type %v)", function, ftype))
	}

	args := []reflect.Value{}
	for i := 0; i < ftype.NumIn(); i++ {
		argtype := ftype.In(i)
		bindtype := argtype
		if bindtype.Kind() != reflect.Ptr {
			bindtype = reflect.PtrTo(argtype)
		}
		instance := r.getInstanceByType(bindtype)
		if instance == nil {
			if argtype.Kind() == reflect.Ptr && bindtype.Elem().Kind() == reflect.Struct && r.injector.binder.providers[bindtype] == nil {
				nv := reflect.New(bindtype.Elem())

				r.InjectMembers(nv.Interface())
				args = append(args, nv)
			} else {
				panic(fmt.Sprintf("%s is Not Binded. So Can't Inject argument at index %d", argtype.String(), i))
			}
		} else {
			args = append(args, reflect.ValueOf(instance).Convert(argtype))
		}
	}

	resultValue := reflect.ValueOf(function).Call(args)

	if len(resultValue) == 0 {
		return nil
	}

	if len(resultValue) == 1 {
		return resultValue[0].Interface()
	}

	ret := make([]interface{}, len(resultValue))
	for i, v := range resultValue {
		ret[i] = v.Interface()
	}
	return ret
}

func (r *injectorContext) InjectMembers(pointer interface{}) {
	//fmt.Println("context getIns")

	ptrvalue := reflect.ValueOf(pointer)

	if ptrvalue.Kind() != reflect.Ptr || ptrvalue.IsNil() {
		return
	}

	if ptrvalue.Elem().Kind() != reflect.Struct {
		return
	}

	t := reflect.TypeOf(pointer).Elem()

	rv := ptrvalue.Elem()

	explicitInject := false
	for i := 0; i < rv.NumField(); i++ {
		fieldType := t.Field(i)
		if hasInjectTag(fieldType.Tag) {
			explicitInject = true
			break
		}
	}

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := t.Field(i)

		switch field.Kind() {
		case reflect.Func:
			if field.IsNil() && field.CanSet() {
				if explicitInject == false || hasInjectTag(fieldType.Tag) {
					res := r.getInstanceByType(reflect.PtrTo(fieldType.Type))
					if res != nil {
						//field.Elem().Set(reflect.ValueOf(res))
						field.Set(reflect.ValueOf(res))
					} else if explicitInject {
						panic(fmt.Sprintf("%s is Not Binded. So Can't Inject to %s.%s", fieldType.Type.String(), t.String(), fieldType.Name))
					}
				}

			}

		case reflect.Struct:
			if field.CanSet() {
				if explicitInject == false || hasInjectTag(fieldType.Tag) {
					r.InjectMembers(field.Addr().Interface())
				}
			}
		case reflect.Ptr:
			if field.IsNil() && field.CanSet() {
				if explicitInject == false || hasInjectTag(fieldType.Tag) {
					res := r.getInstanceByType(fieldType.Type)
					if res != nil {
						//field.Elem().Set(reflect.ValueOf(res))
						field.Set(reflect.ValueOf(res))
					} else if explicitInject {
						panic(fmt.Sprintf("%s is Not Binded. So Can't Inject to %s.%s", fieldType.Type.String(), t.String(), fieldType.Name))
					}

				}
			}
		case reflect.Interface:
			if field.IsNil() && field.CanSet() {
				if explicitInject == false || hasInjectTag(fieldType.Tag) {
					res := r.getInstanceByType(reflect.PtrTo(fieldType.Type))
					if res != nil {
						//field.Elem().Set(reflect.ValueOf(res))
						field.Set(reflect.ValueOf(res))
					} else if explicitInject {
						panic(fmt.Sprintf("%s is Not Binded. So Can't Inject to %s.%s", fieldType.Type.String(), t.String(), fieldType.Name))
					}
				}
			}
		default:
			if field.CanSet() && hasInjectTag(fieldType.Tag) {
				res := r.getInstanceByType(reflect.PtrTo(fieldType.Type))
				if res != nil {
					field.Set(reflect.ValueOf(res).Convert(fieldType.Type))
				} else if explicitInject {
					panic(fmt.Sprintf("%s is Not Binded. So Can't Inject to %s.%s", fieldType.Type.String(), t.String(), fieldType.Name))
				}
			}
		}
	}
}

// NewInjector returns new Injector from implements with enabled modulenames
func NewInjector(implements *Implements, moduleNames []string) Injector {

	binder := newBinder()

	binder.ignoreDuplicate = true
	for i := len(implements.anonymousModule) - 1; i >= 0; i-- {
		implements.anonymousModule[i].Configure(binder)
	}

	binder.ignoreDuplicate = false

	for _, m := range moduleNames {
		module := implements.implements[m]
		if module != nil {
			implements.implements[m].Configure(binder)
		} else {
			panic(fmt.Sprintf("module %s is not implemented", m))
		}
	}
	injector := &injectorImpl{binder, make(map[string]string)}

	for t := range binder.providers {
		if binder.providers[t].isEager {
			injector.getInstanceByType(t)
			//fmt.Printf("eager singleton %v -> %v\n", t, ret)
		}
	}
	return injector
}

// NewInjectorWithTimeout returns new Injector from implements with enabled modulenames
// and it checks timeout
func NewInjectorWithTimeout(implements *Implements, moduleNames []string, timeout time.Duration) Injector {
	ch := make(chan Injector)

	go func() {
		ret := NewInjector(implements, moduleNames)
		ch <- ret
	}()

	timer := time.NewTimer(timeout)
	select {
	case injector := <-ch:
		return injector
	case <-timer.C:
		panic(fmt.Sprintf("Creation failed within the time limit : %d", timeout))
	}
}
