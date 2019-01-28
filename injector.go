package di

import (
	"reflect"
)

type Injector interface {
	GetInstance(intf interface{}) interface{}
	GetProperty(propName string) string
	SetProperty(propName string, value string)
}

type InjectorImpl struct {
	binder     *Binder
	singletons map[reflect.Type]interface{}
	props      map[string]string
}

type InjectorContext struct {
	injector  *InjectorImpl
	loopCheck map[reflect.Type]bool
	stack     []reflect.Type
}

func (this *InjectorImpl) GetInstance(intf interface{}) interface{} {
	//fmt.Println("impl getIns")
	context := InjectorContext{this, make(map[reflect.Type]bool), nil}
	return context.GetInstance(intf)
}

func (this *InjectorImpl) GetProperty(propName string) string {
	return this.props[propName]
}

func (this *InjectorImpl) SetProperty(propName string, value string) {
	this.props[propName] = value
}

func (this *InjectorContext) GetProperty(propName string) string {
	return this.injector.GetProperty(propName)
}

func (this *InjectorContext) SetProperty(propName string, value string) {
	this.injector.SetProperty(propName, value)
}

func (this *InjectorContext) GetInstance(intf interface{}) interface{} {
	//fmt.Println("context getIns")

	t := reflect.TypeOf(intf)

	this.stack = append(this.stack, t)
	if this.loopCheck[t] == true {
		loopStr := ""
		for _, k := range this.stack {
			if loopStr == "" {
				loopStr = k.String()
			} else {
				loopStr = loopStr + "\n  -> " + k.String()
			}

		}
		panic("dependency cycle : \n" + loopStr + "\n")
	}
	this.loopCheck[t] = true

	p := this.injector.binder.providers[t]
	if p != nil {
		if p.isSingleton {
			ret := this.injector.singletons[t]
			if ret == nil {
				ret = p.provider(this)
				this.injector.singletons[t] = ret
			}
			this.stack = this.stack[0 : len(this.stack)-1]
			delete(this.loopCheck, t)
			return ret
		} else {
			ret := p.provider(this)
			this.stack = this.stack[0 : len(this.stack)-1]
			delete(this.loopCheck, t)
			return ret
		}
	}
	this.stack = this.stack[0 : len(this.stack)-1]
	delete(this.loopCheck, t)
	return this.injector.binder.singletons[t]

	//fmt.Printf("type = %s\n", t.String())

}

func NewInjector(implements *Implements, moduleNames []string) Injector {

	binder := newBinder()
	for _, m := range moduleNames {
		implements.implements[m].Configure(binder)
	}

	return &InjectorImpl{binder, make(map[reflect.Type]interface{}), make(map[string]string)}
}
