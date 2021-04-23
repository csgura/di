package di

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
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
func (r *Implements) AddImplement(name string, impl AbstractModule) *Implements {
	r.implements[name] = impl
	return r
}

// HasImplement returns whether named module is registered
func (r *Implements) HasImplement(name string) bool {
	_, exists := r.implements[name]
	return exists
}

// GetImplement returns AbstractModule
func (r *Implements) GetImplement(name string) AbstractModule {
	return r.implements[name]
}

// AddImplements adds all named AbstraceModule of impl to this
func (r *Implements) AddImplements(impl *Implements) *Implements {
	for k, v := range impl.implements {
		if _, exists := r.implements[k]; exists == false {
			r.implements[k] = v
		}
	}
	return r
}

// AddBind adds no named abstrace module
func (r *Implements) AddBind(bindF func(binder *Binder)) *Implements {
	r.anonymousModule = append(r.anonymousModule, BindFunc(bindF))
	return r
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

// NewInjectorWithTrace creates injector and call callback function when instances are created
func (r *Implements) NewInjectorWithTrace(moduleNames []string, traceCallback TraceCallback) Injector {
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

		for _, name := range moduleNames {
			module := r.implements[name]
			if module != nil {
				overBinder := newBinder()

				if overriden, ok := module.(*orverriden); ok {
					for _, m := range overriden.modules {
						m.Configure(overBinder)
					}
				}

				binder.merge(overBinder, false)
			}
		}

	}

	binder.mergeFallbacks()

	injector := &injectorImpl{binder, make(map[string]string), traceCallback}

	var injectorIntf *Injector
	injectorType := reflect.TypeOf(injectorIntf)

	binder.providers[injectorType] = &Binding{
		binder:      binder,
		tpe:         injectorType,
		instance:    injector,
		isSingleton: true,
	}

	context := injectorContext{injector, make(map[reflect.Type]bool), nil, nil, traceCallback}
	context.callDecorators(injectorType)

	for t := range binder.providers {
		if binder.providers[t].isEager {
			injector.getInstanceByType(t)
			//fmt.Printf("eager singleton %v -> %v\n", t, ret)
		}
	}
	return injector
}

// NewInjector returns new Injector from implements with enabled modulenames
func (r *Implements) NewInjector(moduleNames []string) Injector {
	return r.Clone().NewInjectorWithTrace(moduleNames, func(info *TraceInfo) {

	})
}

// NewInjectorWithTimeout returns new Injector from implements with enabled modulenames
// and it checks timeout
func (r *Implements) NewInjectorWithTimeout(moduleNames []string, timeout time.Duration) Injector {
	ch := make(chan Injector)

	var lastRequested *TraceInfo
	var lastCreated *TraceInfo
	var longest *TraceInfo
	requested := map[reflect.Type]*TraceInfo{}
	lock := sync.Mutex{}

	go func() {
		ret := r.NewInjectorWithTrace(moduleNames, func(info *TraceInfo) {
			lock.Lock()
			defer lock.Unlock()
			if info.TraceType == InstanceWillBeCreated {
				lastRequested = info
				requested[info.RequestedType] = info
			} else if info.TraceType == InstanceCreated {
				delete(requested, info.RequestedType)
				lastCreated = info
				if longest == nil || info.ElapsedTime > longest.ElapsedTime {
					longest = info
				}
			}
		})
		ch <- ret
	}()

	timer := time.NewTimer(timeout)
	select {
	case injector := <-ch:
		return injector
	case <-timer.C:
		builder := strings.Builder{}
		builder.WriteString("Not completed : \n\t\t")
		lock.Lock()
		defer lock.Unlock()
		for _, v := range requested {
			builder.WriteString(v.String())
			builder.WriteString("\n\t\t")
		}
		panic(fmt.Sprintf("Creation failed within the time limit : %d\n\tLast %s\n\tLast %s\n\tlongest time to %s\n\t%s", timeout, lastRequested, lastCreated, longest, builder.String()))
	}
}

// NewImplements returns new empty Implements
func NewImplements() *Implements {
	ret := Implements{make(map[string]AbstractModule), nil}
	return &ret
}
