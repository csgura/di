//go:build !go1.27

package di

import "reflect"

// Bind returns Binding that it is not binded anything
func (b *Binder) Bind(ptrToType interface{}) *Binding {
	if ptrToType == nil {
		panic("Bind : invalid type ( nil ). ")
	}

	t := reflect.TypeOf(ptrToType)
	return &Binding{
		binder:      b,
		tpe:         t,
		isSingleton: true,
	}
}

func Bind[T any](binder *Binder) BindingTP[T] {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		return BindingTP[T]{binder.Bind(t)}
	} else {
		return BindingTP[T]{binder.Bind(&t)}
	}
}

// BindProvider binds intf type to provider function
func (b *Binder) BindProvider(ptrToType interface{}, provider func(injector Injector) interface{}) *Binding {
	return b.Bind(ptrToType).ToProvider(provider)
}

// BindConstructor binds intf type to constructor function
func (b *Binder) BindConstructor(ptrToType interface{}, constructor interface{}) *Binding {
	return b.Bind(ptrToType).ToConstructor(constructor)
}

// BindSingleton binds intf type to singleton instance
func (b *Binder) BindSingleton(ptrToType interface{}, instance interface{}) *Binding {
	return b.Bind(ptrToType).ToInstance(instance)

}

// AddDecoratorOf add customizing function which will be applied to the created singleton instance
// if the type is not singleton, then the decorator callback will not be called
func (b *Binder) AddDecoratorOf(ptrToType interface{}, decorator func(ij Injector)) {
	t := reflect.TypeOf(ptrToType)
	b.bind(&Binding{
		binder:        b,
		tpe:           t,
		isDecoratorOf: true,
		provider: func(ij Injector) interface{} {
			decorator(ij)
			return nil
		},
	})
}

func BindProvider[T any](binder *Binder, fn func(inj Injector) T) *Binding {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		return binder.BindProvider(t, func(inj Injector) interface{} {
			return fn(inj)
		})
	} else {
		return binder.BindProvider(&t, func(inj Injector) interface{} {
			return fn(inj)
		})
	}

}

func BindSingleton[T any](binder *Binder, singleton T) *Binding {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		return binder.BindSingleton(t, singleton)
	} else {
		return binder.BindSingleton(&t, singleton)
	}
}

func BindConstructor[T any](binder *Binder, constructor interface{}) *Binding {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		return binder.BindConstructor(t, constructor)
	} else {
		return binder.BindConstructor(&t, constructor)
	}
}

func AddDecoratorOf[T any](binder *Binder, fn func(injector Injector)) {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		binder.AddDecoratorOf(t, fn)
	} else {
		binder.AddDecoratorOf(&t, fn)
	}
}
