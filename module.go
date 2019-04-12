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
	current := len(binder.bindRecords)
	for _, m := range r.overrides {
		m.Configure(binder)
	}
	delta := binder.bindRecords[current:]

	for _, binding := range delta {
		binder.setIgnoreDuplicate(binding.tpe)
	}

	binder.ignoreDuplicate = true
	for _, m := range r.modules {
		m.Configure(binder)
	}
	binder.ignoreDuplicate = false
	binder.clearIgnoreSet()
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
