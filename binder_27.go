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

// AddDecoratorOf add customizing function which will be applied to the created singleton instance
// if the type is not singleton, then the decorator callback will not be called
func (b *Binder) AddDecoratorOf[T any](decorator func(ij Injector)) {
	binding := b.Bind[T]()
	binding.isDecoratorOf = true
	binding.provider = func(ij Injector) interface{} {
		decorator(ij)
		return nil
	}
	b.bind(binding)
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

func AddDecoratorOf[T any](binder *Binder, fn func(injector Injector)) {
	binder.AddDecoratorOf[T](fn)
}
