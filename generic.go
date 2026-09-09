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

func (b BindingTP[T]) AsEagerSingleton() BindingTP[T] {
	b.binding.AsEagerSingleton()
	return b
}

func TypeOf[T any]() interface{} {
	var t T
	if reflect.ValueOf(t).Kind() == reflect.Ptr {
		return t
	} else {
		return &t
	}
}
