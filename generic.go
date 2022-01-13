//go:build go1.18
// +build go1.18

package di

import (
	"reflect"

	"github.com/csgura/fp"
	"github.com/csgura/fp/option"
)

func GetInstance[T any](injector Injector) T {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		ret := injector.GetInstance(t)
		if ret != nil {
			return ret.(T)
		}
		return t
	}

	ret := injector.GetInstance(&t)
	if ret != nil {
		return ret.(T)
	}
	return t
}

func GetInstanceOpt[T any](injector Injector) fp.Option[T] {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		ret := injector.GetInstance(t)
		if ret != nil {
			return option.Some(ret.(T))
		}
		return option.None[T]()
	}

	ret := injector.GetInstance(&t)
	if ret != nil {
		return option.Some(ret.(T))
	}
	return option.None[T]()
}

func GetInstancesOf[T any](injector Injector) []T {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		ret := []T{}
		list := injector.GetInstancesOf(t)
		for _, v := range list {
			ret = append(ret, v.(T))
		}
		return ret
	} else {
		ret := []T{}
		list := injector.GetInstancesOf(&t)
		for _, v := range list {
			ret = append(ret, v.(T))
		}
		return ret
	}
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

func BindInterceptor[T any](binder *Binder, fn func(inj Injector, value T) T) {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		binder.BindInterceptor(t, func(inj Injector, value interface{}) interface{} {
			return fn(inj, value.(T))
		})
	} else {
		binder.BindInterceptor(&t, func(inj Injector, value interface{}) interface{} {
			return fn(inj, value.(T))
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

type BindingTP[T any] struct {
	binding *Binding
}

func (b BindingTP[T]) ToProvider(provider func(injector Injector) T) BindingTP[T] {
	b.binding.ToProvider(func(injector Injector) interface{} {
		return provider(injector)
	})
	return b
}

func (b BindingTP[T]) ToInstance(singleton T) BindingTP[T] {
	b.binding.ToInstance(singleton)
	return b
}

func (b BindingTP[T]) ToConstructor(constructor interface{}) BindingTP[T] {
	b.binding.ToConstructor(constructor)
	return b
}

func (b BindingTP[T]) ShouldCreateBefore(tpe interface{}) BindingTP[T] {
	b.binding.ShouldCreateBefore(tpe)
	return b
}

func (b BindingTP[T]) AsEagerSingleton() BindingTP[T] {
	b.binding.AsEagerSingleton()
	return b
}

func Bind[T any](binder *Binder) BindingTP[T] {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		return BindingTP[T]{binder.Bind(t)}
	} else {
		return BindingTP[T]{binder.Bind(&t)}
	}
}

func IfNotBinded[T any](binder *Binder) BindingTP[T] {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		return BindingTP[T]{binder.IfNotBinded(t)}
	} else {
		return BindingTP[T]{binder.IfNotBinded(&t)}
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

func TypeOf[T any]() interface{} {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		return t
	} else {
		return &t
	}
}
