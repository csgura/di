package di

import (
	"reflect"
)

type Binding struct {
	provider    func(injector Injector) interface{}
	instance    interface{}
	isSingleton bool
	isEager     bool
}

func (this *Binding) AsEagerSingleton() *Binding {
	this.isEager = true
	return this
}

func (this *Binding) AsNonSingleton() *Binding {
	if this.provider != nil {
		this.isSingleton = false
	} else {
		panic("call BindProvider to make non-singleton")
	}

	return this
}

type Binder struct {
	providers map[reflect.Type]*Binding
}

func (b *Binder) BindProvider(intf interface{}, constructor func(injector Injector) interface{}) *Binding {
	t := reflect.TypeOf(intf)
	if b.providers[t] == nil {
		b.providers[t] = &Binding{constructor, nil, true, false}
		return b.providers[t]
	} else {
		panic("duplicated bind for " + t.String())
	}

}

func (b *Binder) BindSingleton(intf interface{}, instance interface{}) *Binding {
	t := reflect.TypeOf(intf)
	if b.providers[t] == nil {
		b.providers[t] = &Binding{nil, instance, true, false}
		return b.providers[t]
	} else {
		panic("duplicated bind for " + t.String())
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
	var ret []interface{} = nil
	interfaceType := reflect.TypeOf(intf).Elem()

	for k := range b.providers {
		p := b.providers[k]
		if p.instance != nil {
			realType := reflect.TypeOf(p.instance)
			//fmt.Printf("interfaceType = %v , realType = %v\n", interfaceType, realType)
			//fmt.Printf("implements = %t \n", realType.Implements(interfaceType))
			//fmt.Printf("assign = %t \n", realType.AssignableTo(interfaceType))
			//fmt.Printf("assign = %t \n", interfaceType.AssignableTo(realType))
			if isImplements(realType, interfaceType) {
				ret = append(ret, p.instance)
			}
		}
	}
	return ret
}

type AbstractModule interface {
	Configure(binder *Binder)
}

func newBinder() *Binder {
	ret := new(Binder)
	ret.providers = make(map[reflect.Type]*Binding)
	return ret
}
