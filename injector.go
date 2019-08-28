package di

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// Injector used to get instance
type Injector interface {
	GetInstance(ptrToType interface{}) interface{}
	GetProperty(propName string) string
	SetProperty(propName string, value string)

	GetInstancesOf(ptrToType interface{}) []interface{}
	InjectMembers(ptrToStruct interface{})
	InjectAndCall(function interface{}) interface{}
	InjectValue(ptrToInterface interface{})
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

func (r *injectorImpl) GetInstance(ptrToType interface{}) interface{} {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil}
	return context.GetInstance(ptrToType)
}

func (r *injectorImpl) GetInstancesOf(ptrToType interface{}) []interface{} {
	//fmt.Println("impl getIns")
	return r.binder.getInstancesOf(ptrToType)
}

func (r *injectorImpl) getInstanceByType(t reflect.Type) interface{} {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil}
	return context.getInstanceByType(t)
}

func (r *injectorImpl) InjectMembers(ptrToStruct interface{}) {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil}
	context.InjectMembers(ptrToStruct)
}

func (r *injectorImpl) InjectAndCall(function interface{}) interface{} {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil}
	return context.InjectAndCall(function)
}

func (r *injectorImpl) InjectValue(ptrToInterface interface{}) {
	context := injectorContext{r, make(map[reflect.Type]bool), nil}
	context.InjectValue(ptrToInterface)
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
	//fmt.Printf("create instance of type %s\n", t)
	return p.provider(r)
}

func (r *injectorContext) callDecorators(t reflect.Type) {
	if list := r.injector.binder.decorators[t]; list != nil {
		for _, decorator := range list {
			//fmt.Printf("call decorator of %s\n", t)
			decorator.provider(r)
		}
	}
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
					if p.instance != nil {
						r.callDecorators(t)
					}
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
					r.callDecorators(t)
				}
			}
			return ret
		}
		ret := r.createInstance(t, p)
		return ret

	}
	return nil
}

func (r *injectorContext) GetInstance(ptrToType interface{}) interface{} {
	//fmt.Println("context getIns")

	t := reflect.TypeOf(ptrToType)

	return r.getInstanceByType(t)
	//fmt.Printf("type = %s\n", t.String())

}

func (r *injectorContext) GetInstancesOf(ptrToType interface{}) []interface{} {
	//fmt.Println("impl getIns")
	return r.injector.binder.getInstancesOf(ptrToType)
}

type injectTag struct {
	inject  bool
	nilable bool
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func hasInjectTag(tag reflect.StructTag) injectTag {
	value, ok := tag.Lookup("di")
	if ok {
		if value == "inject" {
			return injectTag{true, false}
		}
		sp := strings.Split(value, ",")
		if contains(sp, "inject") {
			if contains(sp, "nilable") {
				return injectTag{true, true}
			}
			return injectTag{true, false}
		}
	}
	return injectTag{false, true}
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
				fname := filepath.Base(runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name())
				panic(fmt.Sprintf("%s is Not Binded. So Can't Inject argument of function %s at index %d", argtype.String(), fname, i))
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

func (r *injectorContext) InjectMembers(ptrToStruct interface{}) {
	//fmt.Println("context getIns")

	ptrvalue := reflect.ValueOf(ptrToStruct)

	if ptrvalue.Kind() != reflect.Ptr || ptrvalue.IsNil() {
		return
	}

	if ptrvalue.Elem().Kind() != reflect.Struct {
		return
	}

	t := reflect.TypeOf(ptrToStruct).Elem()

	rv := ptrvalue.Elem()

	explicitInject := false
	for i := 0; i < rv.NumField(); i++ {
		fieldType := t.Field(i)
		if hasInjectTag(fieldType.Tag).inject {
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
				if tag := hasInjectTag(fieldType.Tag); explicitInject == false || tag.inject {
					res := r.getInstanceByType(reflect.PtrTo(fieldType.Type))
					if res != nil {
						//field.Elem().Set(reflect.ValueOf(res))
						field.Set(reflect.ValueOf(res))
					} else if explicitInject && tag.nilable == false {
						panic(fmt.Sprintf("%s is Not Binded. So Can't Inject to %s.%s", fieldType.Type.String(), t.String(), fieldType.Name))
					}
				}

			}

		case reflect.Struct:
			if field.CanSet() {
				if explicitInject == false || hasInjectTag(fieldType.Tag).inject {
					r.InjectMembers(field.Addr().Interface())
				}
			}
		case reflect.Ptr:
			if field.IsNil() && field.CanSet() {
				if tag := hasInjectTag(fieldType.Tag); explicitInject == false || tag.inject {
					res := r.getInstanceByType(fieldType.Type)
					if res != nil {
						//field.Elem().Set(reflect.ValueOf(res))
						field.Set(reflect.ValueOf(res))
					} else if explicitInject && tag.nilable == false {
						panic(fmt.Sprintf("%s is Not Binded. So Can't Inject to %s.%s", fieldType.Type.String(), t.String(), fieldType.Name))
					}

				}
			}
		case reflect.Interface:
			if field.IsNil() && field.CanSet() {
				if tag := hasInjectTag(fieldType.Tag); explicitInject == false || tag.inject {
					res := r.getInstanceByType(reflect.PtrTo(fieldType.Type))
					if res != nil {
						//field.Elem().Set(reflect.ValueOf(res))
						field.Set(reflect.ValueOf(res))
					} else if explicitInject && tag.nilable == false {
						panic(fmt.Sprintf("%s is Not Binded. So Can't Inject to %s.%s", fieldType.Type.String(), t.String(), fieldType.Name))
					}
				}
			}
		default:
			if field.CanSet() {
				if tag := hasInjectTag(fieldType.Tag); tag.inject {
					res := r.getInstanceByType(reflect.PtrTo(fieldType.Type))
					if res != nil {
						field.Set(reflect.ValueOf(res).Convert(fieldType.Type))
					} else if explicitInject && tag.nilable == false {
						panic(fmt.Sprintf("%s is Not Binded. So Can't Inject to %s.%s", fieldType.Type.String(), t.String(), fieldType.Name))
					}
				}
			}
		}
	}
}

func (r *injectorContext) InjectValue(ptrToInterface interface{}) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("panic : %v ", r)
	// 	}
	// }()

	rv := reflect.ValueOf(ptrToInterface)
	if rv.Kind() != reflect.Ptr {
		panic(fmt.Errorf("argument is not pointer"))
	}

	if rv.Type().Elem().Kind() == reflect.Ptr {
		v := r.getInstanceByType(rv.Type().Elem())

		if v == nil {
			panic(fmt.Sprintf("%s is Not Binded.", rv.Type().Elem()))
		}

		rv.Elem().Set(reflect.ValueOf(v))

	} else {

		v := r.GetInstance(ptrToInterface)
		if v == nil {
			panic(fmt.Sprintf("%s is Not Binded.", rv.Type().Elem()))
		}

		rv.Elem().Set(reflect.ValueOf(v).Convert(rv.Type().Elem()))
	}
}

// NewInjector returns new Injector from implements with enabled modulenames
func NewInjector(implements *Implements, moduleNames []string) Injector {

	return implements.NewInjector(moduleNames)
}

// NewInjectorWithTimeout returns new Injector from implements with enabled modulenames
// and it checks timeout
func NewInjectorWithTimeout(implements *Implements, moduleNames []string, timeout time.Duration) Injector {
	return implements.NewInjectorWithTimeout(moduleNames, timeout)
}

// CreateInjector creates new Injector with provided modules
func CreateInjector(modules ...AbstractModule) Injector {
	impls := NewImplements()
	for _, m := range modules {
		impls.AddBind(m.Configure)
	}
	return impls.NewInjector(nil)
}
