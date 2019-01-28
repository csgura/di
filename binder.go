package di

import (
	"reflect"
)

type provider struct {
	provider    func(injector Injector) interface{}
	isSingleton bool
}
type Binder struct {
	providers  map[reflect.Type]*provider
	singletons map[reflect.Type]interface{}
}

func (b *Binder) BindProvider(intf interface{}, constructor func(injector Injector) interface{}) {
	t := reflect.TypeOf(intf)
	if b.providers[t] == nil && b.singletons[t] == nil {
		b.providers[t] = &provider{constructor, true}
	} else {
		panic("duplicated bind for " + t.String())
	}

}

func (b *Binder) BindSingleton(intf interface{}, instance interface{}) {
	t := reflect.TypeOf(intf)
	if b.providers[t] == nil && b.singletons[t] == nil {
		b.singletons[t] = instance
	} else {
		panic("duplicated bind for " + t.String())
	}

}

type AbstractModule interface {
	Configure(binder *Binder)
}

func newBinder() *Binder {
	ret := new(Binder)
	ret.providers = make(map[reflect.Type]*provider)
	ret.singletons = make(map[reflect.Type]interface{})
	return ret
}
