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

func BindInterceptor[T any](binder *Binder, fn func(inj Injector, value T) T) {

	binder.BindInterceptor[T](fn)

}

// BindInterceptor binds interceptor
func (b *Binder) AddRegister[T any](
	registerFunc func(injector Injector, instance T),
) {
	b.AddDecoratorOf[T](func(ij Injector) {
		ins := GetInstance[T](ij)
		registerFunc(ij, ins)
	})
}

// BindInterceptor binds interceptor
func (b *Binder) BindInterceptor[T any](
	interceptorProvider func(injector Injector, instance T) T,
) {
	binding := b.Bind[T]()
	t := binding.tpe
	b.interceptors[t] = append(b.interceptors[t], &Binding{
		binder:        b,
		tpe:           t,
		isInterceptor: true,
		interceptor: func(inj Injector, value interface{}) interface{} {
			return interceptorProvider(inj, value.(T))
		},
	})
	//return b.Bind(ptrToType).ToInstance(instance)
}

// IfNotBinded returns Binding that will used if there are no other binding for tpe type
func (b *Binder) IfNotBinded[T any]() BindingTP[T] {
	binding := b.Bind[T]()
	t := binding.tpe
	return BindingTP[T]{&Binding{
		binder:      b,
		tpe:         t,
		isSingleton: true,
		isFallback:  true,
	}}
}

func IfNotBinded[T any](binder *Binder) BindingTP[T] {
	return binder.IfNotBinded[T]()
}
