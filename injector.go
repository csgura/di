package di

import (
	"fmt"
	"reflect"
	"time"
)

type Injector interface {
	GetInstance(intf interface{}) interface{}
	GetProperty(propName string) string
	SetProperty(propName string, value string)

	GetInstancesOf(intf interface{}) []interface{}
}

type InjectorImpl struct {
	binder *Binder
	props  map[string]string
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

func (this *InjectorImpl) GetInstancesOf(intf interface{}) []interface{} {
	//fmt.Println("impl getIns")
	return this.binder.getInstancesOf(intf)
}

func (this *InjectorImpl) getInstanceByType(t reflect.Type) interface{} {
	//fmt.Println("impl getIns")
	context := InjectorContext{this, make(map[reflect.Type]bool), nil}
	return context.getInstanceByType(t)
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

func (this *InjectorContext) getInstanceByType(t reflect.Type) interface{} {
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

			ret := p.instance
			if ret == nil {
				if p.provider != nil {
					ret = p.provider(this)
					p.instance = ret
				}
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
	return nil
}
func (this *InjectorContext) GetInstance(intf interface{}) interface{} {
	//fmt.Println("context getIns")

	t := reflect.TypeOf(intf)

	return this.getInstanceByType(t)
	//fmt.Printf("type = %s\n", t.String())

}

func (this *InjectorContext) GetInstancesOf(intf interface{}) []interface{} {
	//fmt.Println("impl getIns")
	return this.injector.binder.getInstancesOf(intf)
}

func NewInjector(implements *Implements, moduleNames []string) Injector {

	binder := newBinder()

	binder.ignoreDuplicate = true
	for i := len(implements.just) - 1; i >= 0; i-- {
		implements.just[i].Configure(binder)
	}

	binder.ignoreDuplicate = false

	for _, m := range moduleNames {
		module := implements.implements[m]
		if module != nil {
			implements.implements[m].Configure(binder)
		} else {
			panic(fmt.Sprintf("module %s is not implemented", m))
		}
	}
	injector := &InjectorImpl{binder, make(map[string]string)}

	for t := range binder.providers {
		if binder.providers[t].isEager {
			injector.getInstanceByType(t)
			//fmt.Printf("eager singleton %v -> %v\n", t, ret)
		}
	}
	return injector
}

func NewInjectorWithTimeout(implements *Implements, moduleNames []string, timeout time.Duration) Injector {
	ch := make(chan Injector)

	go func() {
		ret := NewInjector(implements, moduleNames)
		ch <- ret
	}()

	timer := time.NewTimer(timeout)
	select {
	case injector := <-ch:
		return injector
	case <-timer.C:
		panic(fmt.Sprintf("Creation failed within the time limit : %d", timeout))
	}
}
