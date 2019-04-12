package di

type combineModule struct {
	modules []AbstractModule
}

func (r *combineModule) Configure(binder *Binder) {
	for _, m := range r.modules {
		m.Configure(binder)
	}
}

// OverridableModule is overridable module
type OverridableModule interface {
	AbstractModule
	With(overrides ...AbstractModule) AbstractModule
}

type orverriden struct {
	modules   []AbstractModule
	overrides []AbstractModule
}

func (r *orverriden) Configure(binder *Binder) {

	tempBinder := newBinder()
	for _, m := range r.overrides {
		m.Configure(tempBinder)
	}

	tempBinder.ignoreDuplicate = true

	for _, m := range r.modules {
		m.Configure(tempBinder)
	}

	binder.merge(tempBinder, true)

}

func (r *orverriden) With(overrides ...AbstractModule) AbstractModule {
	r.overrides = overrides
	return r
}

// CombineModule Returns a new module that installs all of modules
func CombineModule(modules ...AbstractModule) AbstractModule {
	return &combineModule{modules: modules}
}

//OverrideModule Returns a builder that creates a module that overlays override modules over the given modules.
func OverrideModule(modules ...AbstractModule) OverridableModule {
	return &orverriden{modules: modules}
}
