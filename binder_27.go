//go:build go1.27

package di

import "reflect"

// Bind returns Binding that it is not binded anything
func (b *Binder) Bind[T any]() *Binding {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		tpe := reflect.TypeOf(t)

		return &Binding{
			binder:      b,
			tpe:         tpe,
			isSingleton: true,
		}
	} else {
		tpe := reflect.TypeOf(&t)

		return &Binding{
			binder:      b,
			tpe:         tpe,
			isSingleton: true,
		}
	}

}

func Bind[T any](binder *Binder) BindingTP[T] {
	return BindingTP[T]{binder.Bind[T]()}
}

// BindProvider binds intf type to provider function
func (b *Binder) BindProvider[T any](provider func(injector Injector) interface{}) *Binding {
	return b.Bind[T]().ToProvider(provider)
}

// BindConstructor binds intf type to constructor function
func (b *Binder) BindConstructor[T any](constructor interface{}) *Binding {
	return b.Bind[T]().ToConstructor(constructor)
}

// BindSingleton binds intf type to singleton instance
func (b *Binder) BindSingleton[T any](instance interface{}) *Binding {
	return b.Bind[T]().ToInstance(instance)

}

func BindProvider[T any](binder *Binder, fn func(inj Injector) T) *Binding {

	return binder.BindProvider[T](func(inj Injector) interface{} {
		return fn(inj)
	})

}

func BindSingleton[T any](binder *Binder, singleton T) *Binding {

	return binder.BindSingleton[T](singleton)

}

func BindConstructor[T any](binder *Binder, constructor interface{}) *Binding {

	return binder.BindConstructor[T](constructor)

}
