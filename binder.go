package di

import (
	"reflect"
)

// Binding has provider function and created singlton instances
// and it has some configurtions about binding
type Binding struct {
	binder      *Binder
	tpe         reflect.Type
	provider    func(injector Injector) interface{}
	instance    interface{}
	isSingleton bool
	isEager     bool
}

// AsEagerSingleton set binding as eager singleton
func (b *Binding) AsEagerSingleton() *Binding {
	b.isEager = true
	return b
}

// AsNonSingleton set binding as non singleton
func (b *Binding) AsNonSingleton() *Binding {
	if b.provider != nil {
		b.isSingleton = false
	} else {
		panic("call BindProvider to make non-singleton")
	}

	return b
}

// ShouldBindBefore set binding order. this binding should be performed before tpe type binding
func (b *Binding) ShouldBindBefore(tpe interface{}) *Binding {

	b.binder.shouldBindBefore(tpe, b)
	return b
}

// Binder has bindings
type Binder struct {
	providers       map[reflect.Type]*Binding
	bindBefore      map[reflect.Type][]*Binding
	ignoreDuplicate bool
}

func (b *Binder) shouldBindBefore(intf interface{}, binding *Binding) {
	t := reflect.TypeOf(intf)
	list := b.bindBefore[t]

	list = append(list, binding)
	b.bindBefore[t] = list
}

// BindProvider binds intf type to provider function
func (b *Binder) BindProvider(intf interface{}, constructor func(injector Injector) interface{}) *Binding {
	t := reflect.TypeOf(intf)
	if b.providers[t] == nil {
		b.providers[t] = &Binding{b, t, constructor, nil, true, false}
		return b.providers[t]
	} else {
		if b.ignoreDuplicate {
			return b.providers[t]
		} else {
			panic("duplicated bind for " + t.String())
		}
	}

}

// BindSingleton binds intf type to singleton instance
func (b *Binder) BindSingleton(intf interface{}, instance interface{}) *Binding {
	t := reflect.TypeOf(intf)
	if b.providers[t] == nil {
		b.providers[t] = &Binding{b, t, nil, instance, true, false}
		return b.providers[t]
	} else {
		if b.ignoreDuplicate {
			return b.providers[t]
		} else {
			panic("duplicated bind for " + t.String())
		}
	}
}

func isImplements(realType reflect.Type, interfaceType reflect.Type) (eq bool) {
	defer func() {
		if r := recover(); r != nil {
			eq = false
		}
	}()
	eq = realType.Implements(interfaceType)
	return eq
}
func (b *Binder) getInstancesOf(intf interface{}) []interface{} {
	var ret []interface{}
	interfaceType := reflect.TypeOf(intf).Elem()

	for k := range b.providers {
		p := b.providers[k]
		if p.instance != nil {
			realType := reflect.TypeOf(p.instance)
			//fmt.Printf("interfaceType = %v,%d , realType = %v,%d\n", interfaceType, interfaceType.Kind(), realType, realType.Kind())
			//fmt.Printf("implements = %t \n", realType.Implements(interfaceType))
			//fmt.Printf("assign = %t \n", realType.AssignableTo(interfaceType))
			//fmt.Printf("assign = %t \n", interfaceType.AssignableTo(realType))
			if interfaceType.Kind() == reflect.Interface {
				if isImplements(realType, interfaceType) {
					ret = append(ret, p.instance)
				}
			} else {
				if realType.Kind() == reflect.Ptr {
					if interfaceType == realType.Elem() {
						ret = append(ret, p.instance)
					}
				} else if interfaceType == realType {
					ret = append(ret, p.instance)
				}
			}
		}
	}
	return ret
}

// AbstractModule is used to create bindings
type AbstractModule interface {
	Configure(binder *Binder)
}

func newBinder() *Binder {
	ret := new(Binder)
	ret.providers = make(map[reflect.Type]*Binding)
	ret.bindBefore = make(map[reflect.Type][]*Binding)
	return ret
}
