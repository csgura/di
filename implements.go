package di

import (
	"fmt"
	"reflect"
	"time"
)

// BindFunc can act as AbstractModule
type BindFunc func(binder *Binder)

// Configure is implements of AbstractModule
func (bf BindFunc) Configure(binder *Binder) {
	bf(binder)
}

// Implements is registry of AbstractModule
type Implements struct {
	implements      map[string]AbstractModule
	anonymousModule []AbstractModule
}

// AddImplement adds named abstract module to Implements
func (r *Implements) AddImplement(name string, impl AbstractModule) {
	r.implements[name] = impl
}

// HasImplement returns whether named module is registered
func (r *Implements) HasImplement(name string) bool {
	_, exists := r.implements[name]
	return exists
}

// AddImplements adds all named AbstraceModule of impl to this
func (r *Implements) AddImplements(impl *Implements) {
	for k, v := range impl.implements {
		if _, exists := r.implements[k]; exists == false {
			r.implements[k] = v
		}
	}
}

// AddBind adds no named abstrace module
func (r *Implements) AddBind(bindF func(binder *Binder)) {
	r.anonymousModule = append(r.anonymousModule, BindFunc(bindF))
}

// Clone returns clone of this
func (r *Implements) Clone() *Implements {
	ret := NewImplements()
	ret.AddImplements(r)
	for _, nonamed := range r.anonymousModule {
		ret.anonymousModule = append(ret.anonymousModule, nonamed)
	}
	return ret
}

// NewInjector returns new Injector from implements with enabled modulenames
func (r *Implements) NewInjector(moduleNames []string) Injector {

	binder := newBinder()

	binder.ignoreDuplicate = true
	for i := len(r.anonymousModule) - 1; i >= 0; i-- {
		r.anonymousModule[i].Configure(binder)
	}

	binder.ignoreDuplicate = false

	hasOverride := false
	for _, m := range moduleNames {
		module := r.implements[m]
		if module != nil {
			if overriden, ok := module.(*orverriden); ok {
				if overriden.overrides == nil {
					hasOverride = true
					continue
				}
			}
			r.implements[m].Configure(binder)
		} else {
			panic(fmt.Sprintf("module %s is not implemented", m))
		}
	}
	if hasOverride {

		overBinder := newBinder()

		for _, name := range moduleNames {
			module := r.implements[name]
			if module != nil {
				if overriden, ok := module.(*orverriden); ok {
					for _, m := range overriden.modules {
						m.Configure(overBinder)
					}
				}
			}
		}

		binder.merge(overBinder, false)
	}

	injector := &injectorImpl{binder, make(map[string]string)}

	var injectorIntf *Injector
	injectorType := reflect.TypeOf(injectorIntf)

	binder.providers[injectorType] = &Binding{
		binder:      binder,
		tpe:         injectorType,
		instance:    injector,
		isSingleton: true,
	}

	for t := range binder.providers {
		if binder.providers[t].isEager {
			injector.getInstanceByType(t)
			//fmt.Printf("eager singleton %v -> %v\n", t, ret)
		}
	}
	return injector
}

// NewInjectorWithTimeout returns new Injector from implements with enabled modulenames
// and it checks timeout
func (r *Implements) NewInjectorWithTimeout(moduleNames []string, timeout time.Duration) Injector {
	ch := make(chan Injector)

	go func() {
		ret := r.NewInjector(moduleNames)
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

// NewImplements returns new empty Implements
func NewImplements() *Implements {
	ret := Implements{make(map[string]AbstractModule), nil}
	return &ret
}
