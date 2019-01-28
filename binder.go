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
	b.providers[t] = &provider{constructor, true}
}

func (b *Binder) BindSingleton(intf interface{}, instance interface{}) {
	t := reflect.TypeOf(intf)
	b.singletons[t] = instance
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
