package di

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"maps"
	"slices"

	"github.com/csgura/fp/reflectfp"
)

// TraceType is trace event type
type TraceType int

func (r TraceType) String() string {
	switch r {
	case InstanceRequest:
		return "Request Instance"
	case InstanceWillBeCreated:
		return "Creating Instance"
	case InstanceCreated:
		return "Create Instance"
	case InstanceReturned:
		return "Instance Returned"
	}
	return ""
}

const (
	// InstanceRequest is trace event
	InstanceRequest TraceType = iota

	// InstanceWillBeCreated is trace event
	InstanceWillBeCreated

	// InstanceCreated is trace event
	InstanceCreated

	// InstanceReturned is trace event
	InstanceReturned
)

// TraceInfo is trace message
type TraceInfo struct {
	TraceType        TraceType
	RequestedType    reflect.Type
	Referer          reflect.Type
	ReturnedInstance interface{}
	IsCreatedNow     bool
	ElapsedTime      time.Duration
	IsSingleton      bool
	IsBinded         bool
	IsEager          bool
}

func (r *TraceInfo) String() string {
	if r == nil {
		return ""
	}
	if r.TraceType == InstanceCreated {
		if r.Referer != nil {
			return fmt.Sprintf("Complete Instance : %s -> %s , ElapsedTime : %s", r.Referer, r.RequestedType, r.ElapsedTime)
		}
		return fmt.Sprintf("Complete Instance : %s , ElapsedTime : %s", r.RequestedType, r.ElapsedTime)
	}

	if r.Referer != nil {
		return fmt.Sprintf("%s : %s -> %s", r.TraceType, r.Referer, r.RequestedType)
	}
	return fmt.Sprintf("%s : %s", r.TraceType, r.RequestedType)
}

// TraceCallback is trace call back function
type TraceCallback func(info *TraceInfo)

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
	binder        *Binder
	props         map[string]string
	traceCallback TraceCallback
}

type injectorContext struct {
	injector *injectorImpl
	// 다음 세개 변수가 multi thread safe 하지 않음.
	loopCheck     map[reflect.Type]bool
	stack         []reflect.Type
	refererStack  []reflect.Type
	traceCallback TraceCallback
	lock          sync.Mutex
}

func (r *injectorImpl) GetInstance(ptrToType interface{}) interface{} {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil, nil, r.traceCallback, sync.Mutex{}}
	return context.GetInstance(ptrToType)
}

func (r *injectorImpl) GetInstancesOf(ptrToType interface{}) []interface{} {
	//fmt.Println("impl getIns")
	return r.binder.getInstancesOf(ptrToType)
}

func (r *injectorImpl) getInstanceByType(t reflect.Type) interface{} {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil, nil, r.traceCallback, sync.Mutex{}}
	return context.getInstanceByType(t)
}

func (r *injectorImpl) InjectMembers(ptrToStruct interface{}) {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil, nil, r.traceCallback, sync.Mutex{}}
	context.InjectMembers(ptrToStruct)
}

func (r *injectorImpl) InjectAndCall(function interface{}) interface{} {
	//fmt.Println("impl getIns")
	context := injectorContext{r, make(map[reflect.Type]bool), nil, nil, r.traceCallback, sync.Mutex{}}
	return context.InjectAndCall(function)
}

func (r *injectorImpl) InjectValue(ptrToInterface interface{}) {
	context := injectorContext{r, make(map[reflect.Type]bool), nil, nil, r.traceCallback, sync.Mutex{}}
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

func (r *injectorContext) createJitBinding(binder *Binder, bindType reflect.Type, actualType reflect.Type) *Binding {
	return &Binding{
		binder: binder,
		tpe:    bindType,
		provider: func(injector Injector) interface{} {
			nv := reflect.New(actualType.Elem())

			iv := nv.Interface()
			r.InjectMembers(iv)
			return iv
		},
		instance:      nil,
		isSingleton:   false,
		isEager:       false,
		isFallback:    false,
		isDecoratorOf: false,
	}
}

func (r *injectorContext) withLock(f func()) {
	r.lock.Lock()
	defer r.lock.Unlock()
	f()
}

func (r *injectorContext) paninOnLoop(t reflect.Type) {

	r.withLock(func() {
		if r.loopCheck[t] == true {
			loopStr := ""
			for _, k := range r.stack {
				if loopStr == "" {
					loopStr = k.String()
				} else {
					loopStr = loopStr + "\n  -> " + k.String()
				}

			}
			panic("dependency cycle : \n" + loopStr + "\n  -> " + t.String())
		}
	})
}

func (r *injectorContext) createInstance(t reflect.Type, p *Binding) interface{} {
	var referer reflect.Type

	r.paninOnLoop(t)

	r.withLock(func() {
		if len(r.stack) > 0 {
			referer = r.stack[len(r.stack)-1]
		}

		r.stack = append(r.stack, t)

		r.loopCheck[t] = true
	})

	defer func() {
		r.withLock(func() {
			r.stack = r.stack[0 : len(r.stack)-1]
			delete(r.loopCheck, t)
		})
	}()

	before := time.Now()
	//fmt.Printf("create instance of type %s\n", t)
	if r.traceCallback != nil {
		r.traceCallback(&TraceInfo{
			TraceType:     InstanceWillBeCreated,
			RequestedType: t,
			Referer:       referer,
		})
	}
	ret := p.provider(r)
	after := time.Now()
	if r.traceCallback != nil {
		r.traceCallback(&TraceInfo{
			TraceType:     InstanceCreated,
			RequestedType: t,
			Referer:       referer,
			IsCreatedNow:  true,
			ElapsedTime:   after.Sub(before),
		})
	}
	return ret
}

func (r *injectorContext) wrapInterceptor(t reflect.Type, instance interface{}) interface{} {
	ret := instance
	if list := r.injector.binder.interceptors[t]; list != nil {
		for _, interceptor := range list {
			if w := interceptor.interceptor(r, ret); w != nil {
				ret = w
			}
		}
	}
	return ret
}

func (r *injectorContext) callDecorators(t reflect.Type) {
	if list := r.injector.binder.decorators[t]; list != nil {

		for _, decorator := range list {
			//fmt.Printf("call decorator of %s\n", t)
			decorator.provider(r)
		}
	}
}

func (r *injectorContext) getBinding(t reflect.Type) *Binding {
	p := r.injector.binder.providers[t]
	if p != nil {
		return p
	}

	// p = r.injector.binder.providersFallback[t]
	// if p != nil {
	// 	return p
	// }
	return nil
}

func (r *injectorContext) getInstanceByBinding(p *Binding) interface{} {

	if p == nil {
		return nil
	}

	var referer reflect.Type

	r.withLock(func() {
		if len(r.refererStack) > 0 {
			referer = r.refererStack[len(r.refererStack)-1]
		}

		r.refererStack = append(r.refererStack, p.tpe)
	})

	defer func() {
		r.withLock(func() {
			r.refererStack = r.refererStack[0 : len(r.refererStack)-1]
		})
	}()

	if r.traceCallback != nil && referer != p.tpe {

		r.traceCallback(&TraceInfo{
			TraceType:     InstanceRequest,
			RequestedType: p.tpe,
			Referer:       referer,
			IsBinded:      true,
			IsSingleton:   p.isSingleton,
			IsEager:       p.isEager,
		})

	}

	ret := func() interface{} {
		if p != nil {
			if p.isSingleton {
				created := false
				r.paninOnLoop(p.tpe)
				p.singletonOnce.Do(func() {
					if p.provider != nil {
						ins := r.createInstance(p.tpe, p)
						if ins != nil {
							ins = r.wrapInterceptor(p.tpe, ins)
						}

						p.instance = ins
						created = true
					}
				})
				if created && p.instance != nil {
					r.callDecorators(p.tpe)
				}
				ret := p.instance
				return ret
			}
			ret := r.createInstance(p.tpe, p)
			if ret != nil {
				ret = r.wrapInterceptor(p.tpe, ret)
			}
			return ret
		}
		return nil
	}()

	if r.traceCallback != nil && referer != p.tpe {

		r.traceCallback(&TraceInfo{
			TraceType:        InstanceReturned,
			RequestedType:    p.tpe,
			Referer:          referer,
			IsBinded:         true,
			IsSingleton:      p.isSingleton,
			IsEager:          p.isEager,
			ReturnedInstance: ret,
		})
	}
	return ret
}

func (r *injectorContext) getInstanceByType(t reflect.Type) interface{} {

	p := r.getBinding(t)

	if p == nil {

		var referer reflect.Type
		if len(r.refererStack) > 0 {
			referer = r.refererStack[len(r.refererStack)-1]
		}

		r.traceCallback(&TraceInfo{
			TraceType:     InstanceRequest,
			RequestedType: t,
			Referer:       referer,
			IsBinded:      false,
			IsSingleton:   false,
			IsEager:       false,
		})

		r.traceCallback(&TraceInfo{
			TraceType:        InstanceReturned,
			RequestedType:    t,
			Referer:          referer,
			IsBinded:         false,
			IsSingleton:      false,
			IsEager:          false,
			ReturnedInstance: nil,
		})

		return nil
	}
	ret := r.getInstanceByBinding(p)
	return ret

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

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Ptr,
		reflect.UnsafePointer,
		reflect.Interface,
		reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
func (r *injectorContext) clone() *injectorContext {
	r.lock.Lock()
	defer r.lock.Unlock()

	ret := injectorContext{r.injector, maps.Clone(r.loopCheck), slices.Clone(r.stack), slices.Clone(r.refererStack), r.traceCallback, sync.Mutex{}}
	return &ret
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

		lazyType := reflectfp.MatchLazyEval(argtype)

		optType := reflectfp.MatchOption(argtype)

		bindtype := argtype
		if optType.IsDefined() {
			bindtype = reflect.PtrTo(optType.Get())
		} else if lazyType.IsDefined() {
			bindtype = reflect.PtrTo(lazyType.Get())
		}

		if bindtype.Kind() != reflect.Ptr {
			bindtype = reflect.PtrTo(argtype)
		}

		binding := r.getBinding(bindtype)
		if binding != nil && optType.IsEmpty() && lazyType.IsDefined() {
			lazyv := reflectfp.LazyCall(argtype, func() reflect.Value {

				lazyCtx := r.clone()

				instance := lazyCtx.getInstanceByBinding(binding)
				return reflect.ValueOf(instance)
			})
			args = append(args, lazyv.Get())
		} else {
			if optType.IsEmpty() && binding == nil && argtype.Kind() == reflect.Ptr && bindtype.Elem().Kind() == reflect.Struct && r.injector.binder.providers[bindtype] == nil {
				binding = r.createJitBinding(r.injector.binder, argtype, bindtype)
			}
			instance := r.getInstanceByBinding(binding)

			if optType.IsDefined() {
				if instance == nil {
					args = append(args, reflectfp.None(argtype).Get())
				} else {
					somev := reflectfp.Some(argtype, reflect.ValueOf(instance))
					args = append(args, somev.Get())
				}

			} else {
				if instance == nil {
					fname := filepath.Base(runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name())
					panic(fmt.Sprintf("%s is Not Binded. So Can't Inject argument of function %s at index %d", argtype.String(), fname, i))
				} else {
					args = append(args, reflect.ValueOf(instance).Convert(argtype))
				}
			}
		}

	}

	resultValue := reflect.ValueOf(function).Call(args)

	if len(resultValue) == 0 {
		return nil
	}

	if len(resultValue) == 1 {
		if isNil(resultValue[0]) {
			return nil
		}
		return resultValue[0].Interface()
	}

	ret := make([]interface{}, len(resultValue))
	for i, v := range resultValue {
		if isNil(v) {
			ret[i] = nil
		}
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
					if valType, ok := reflectfp.MatchOption(fieldType.Type).Unapply(); ok {
						res := r.getInstanceByType(reflect.PtrTo(valType))
						if res != nil {
							field.Set(reflectfp.Some(fieldType.Type, reflect.ValueOf(res)).Get())
						} else {
							field.Set(reflectfp.None(fieldType.Type).Get())
						}

					} else if valType, ok := reflectfp.MatchLazyEval(fieldType.Type).Unapply(); ok {
						res := reflectfp.LazyCall(fieldType.Type, func() reflect.Value {
							lazyCtx := r.clone()
							return reflect.ValueOf(lazyCtx.getInstanceByType(reflect.PtrTo(valType)))
						})
						field.Set(res.Get())
					} else {
						r.InjectMembers(field.Addr().Interface())
					}
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

		retval := reflect.ValueOf(v)
		if retval.IsNil() {
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
