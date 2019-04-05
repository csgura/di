package di

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

// NewImplements returns new empty Implements
func NewImplements() *Implements {
	ret := Implements{make(map[string]AbstractModule), nil}
	return &ret
}
