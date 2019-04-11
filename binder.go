package di

import (
	"reflect"
)

// Binding has provider function and created singlton instances
// and it has some configurtions about binding
type Binding struct {
	binder      *Binder
	tpe         reflect.Type
	provider    func(injector Injector) interface{}
	instance    interface{}
	isSingleton bool
	isEager     bool
	isFallback  bool
}

// ToInstance binds type to singleton instance
func (b *Binding) ToInstance(instance interface{}) *Binding {
	b.instance = instance
	b.binder.bind(b)
	return b
}

// ToProvider binds type to the provider
func (b *Binding) ToProvider(provider func(injector Injector) interface{}) *Binding {
	b.provider = provider
	b.binder.bind(b)
	return b
}

// ToConstructor binds type to the constructor
func (b *Binding) ToConstructor(function interface{}) *Binding {
	return b.ToProvider(func(injector Injector) interface{} {
		return injector.InjectAndCall(function)
	})
}

// AsEagerSingleton set binding as eager singleton
func (b *Binding) AsEagerSingleton() *Binding {
	b.isEager = true
	return b
}

// AsNonSingleton set binding as non singleton
func (b *Binding) AsNonSingleton() *Binding {
	if b.provider != nil {
		b.isSingleton = false
	} else {
		panic("call BindProvider to make non-singleton")
	}

	return b
}

// ShouldCreateBefore set creating order. this creation of instance should be performed before instance creation of the tpe type
func (b *Binding) ShouldCreateBefore(ptrToType interface{}) *Binding {

	b.binder.shouldCreateBefore(ptrToType, b)
	return b
}

// Binder has bindings
type Binder struct {
	providers         map[reflect.Type]*Binding
	providersFallback map[reflect.Type]*Binding
	creatingBefore    map[reflect.Type][]*Binding
	ignoreDuplicate   bool
}

func (b *Binder) shouldCreateBefore(ptrToType interface{}, binding *Binding) {
	t := reflect.TypeOf(ptrToType)
	list := b.creatingBefore[t]

	list = append(list, binding)
	b.creatingBefore[t] = list
}

// Bind returns Binding that it is not binded anything
func (b *Binder) Bind(ptrToType interface{}) *Binding {
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

func (b *Binder) bind(binding *Binding) {
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
					ret = append(ret, p.instance)
				}
			} else {
				if realType.Kind() == reflect.Ptr {
					if interfaceType == realType.Elem() {
						ret = append(ret, p.instance)
					}
				} else if interfaceType == realType {
					ret = append(ret, p.instance)
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

	ret.creatingBefore = make(map[reflect.Type][]*Binding)
	return ret
}
