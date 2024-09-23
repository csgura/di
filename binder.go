package di

import (
	"reflect"
	"sync"
)

// Binding has provider function and created singlton instances
// and it has some configurtions about binding
type Binding struct {
	binder        *Binder
	tpe           reflect.Type
	provider      func(injector Injector) interface{}
	instance      interface{}
	isSingleton   bool
	isEager       bool
	isFallback    bool
	isDecoratorOf bool
	isInterceptor bool
	interceptor   interceptorProvider
	singletonOnce sync.Once
}

// ToInstance binds type to singleton instance
func (b *Binding) ToInstance(instance interface{}) *Binding {

	if b.isDecoratorOf {
		panic("Decorator can't bind to instance")
	}

	// b.instance = instance
	// b.binder.bind(b)

	if b.provider == nil {
		return b.ToProvider(func(injector Injector) interface{} {
			injector.InjectMembers(instance)
			return instance
		}).AsEagerSingleton()
	}

	return b
}

// ToProvider binds type to the provider
func (b *Binding) ToProvider(provider func(injector Injector) interface{}) *Binding {
	if b.isDecoratorOf {
		panic("Decorator can't bind to provider")
	}

	b.provider = provider
	b.binder.bind(b)
	return b
}

// ToConstructor binds type to the constructor
func (b *Binding) ToConstructor(function interface{}) *Binding {
	if b.isDecoratorOf {
		panic("Decorator can't bind to constructor")
	}

	return b.ToProvider(func(injector Injector) interface{} {
		return injector.InjectAndCall(function)
	})
}

// AsEagerSingleton set binding as eager singleton
func (b *Binding) AsEagerSingleton() *Binding {
	if b.isDecoratorOf {
		panic("Decorator can't not be singleton")
	}

	b.isEager = true
	return b
}

// AsNonSingleton set binding as non singleton
func (b *Binding) AsNonSingleton() *Binding {

	b.isSingleton = false

	return b
}

type interceptorProvider func(injector Injector, instance interface{}) interface{}

// Binder has bindings
type Binder struct {
	providers         map[reflect.Type]*Binding
	providersFallback map[reflect.Type]*Binding
	decorators        map[reflect.Type][]*Binding
	interceptors      map[reflect.Type][]*Binding
	ignoreDuplicate   bool
}

func safeAppend(list []*Binding, b *Binding) []*Binding {
	for _, v := range list {
		if v == b {
			return list
		}
	}
	return append(list, b)
}

func (b *Binder) addDecorator(binding *Binding) {
	list := b.decorators[binding.tpe]

	list = safeAppend(list, binding)
	b.decorators[binding.tpe] = list
}

func (b *Binder) addInterceptor(binding *Binding) {
	list := b.interceptors[binding.tpe]

	list = safeAppend(list, binding)
	b.interceptors[binding.tpe] = list
}

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

// IfNotBinded returns Binding that will used if there are no other binding for tpe type
func (b *Binder) IfNotBinded(ptrToType interface{}) *Binding {
	t := reflect.TypeOf(ptrToType)
	return &Binding{
		binder:      b,
		tpe:         t,
		isSingleton: true,
		isFallback:  true,
	}
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

func (b *Binder) bind(binding *Binding) {
	if binding.isDecoratorOf {
		b.addDecorator(binding)
	} else {
		t := binding.tpe
		if binding.isFallback {
			if b.providersFallback[t] == nil {
				b.providersFallback[t] = binding
			}
		} else {
			if b.providers[t] == nil {
				b.providers[t] = binding
			} else {
				if !b.ignoreDuplicate {
					panic("duplicated bind for " + t.String())
				}
			}
		}
	}

}

func (b *Binder) merge(other *Binder, panicOnDup bool) {
	for k, v := range other.providers {
		if b.providers[k] == nil {
			b.providers[k] = v
		} else if panicOnDup {
			panic("duplicated bind for " + v.tpe.String())
		}
	}
	for k, v := range other.providersFallback {
		if b.providersFallback[k] == nil {
			b.providersFallback[k] = v
		}
	}

	for _, list := range other.decorators {
		for _, v := range list {
			b.addDecorator(v)
		}
	}

	for _, list := range other.interceptors {
		for _, v := range list {
			b.addInterceptor(v)
		}
	}

}

func (b *Binder) mergeFallbacks() {
	for k, v := range b.providersFallback {
		if b.providers[k] == nil {
			b.providers[k] = v
		}
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

// BindInterceptor binds interceptor
func (b *Binder) BindInterceptor(
	ptrToType interface{},
	interceptorProvider func(injector Injector, instance interface{}) interface{},
) {
	t := reflect.TypeOf(ptrToType)
	b.interceptors[t] = append(b.interceptors[t], &Binding{
		binder:        b,
		tpe:           t,
		isInterceptor: true,
		interceptor:   interceptorProvider,
	})
	//return b.Bind(ptrToType).ToInstance(instance)
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

func (b *Binder) getInstancesOf(ptrToType interface{}) []interface{} {
	var ret []interface{}
	dupcheck := map[interface{}]bool{}

	checkDup := func(realType reflect.Type, instance interface{}) bool {
		if realType.Comparable() {
			if _, ok := dupcheck[instance]; !ok {
				dupcheck[instance] = true
				return true
			}
			return false

		}
		return true
	}

	interfaceType := reflect.TypeOf(ptrToType).Elem()

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
					if checkDup(realType, p.instance) {
						ret = append(ret, p.instance)
					}
				}
			} else {
				if realType.Kind() == reflect.Ptr {
					if interfaceType == realType.Elem() {
						if checkDup(realType, p.instance) {
							ret = append(ret, p.instance)
						}
					}
				} else if interfaceType == realType {
					if checkDup(realType, p.instance) {
						ret = append(ret, p.instance)
					}
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
	ret.providersFallback = make(map[reflect.Type]*Binding)

	ret.decorators = make(map[reflect.Type][]*Binding)
	ret.interceptors = make(map[reflect.Type][]*Binding)

	return ret
}
