package di

type JustInModule struct {
	bindF func(binder *Binder)
}

func (this *JustInModule) Configure(binder *Binder) {
	this.bindF(binder)
}

type Implements struct {
	implements map[string]AbstractModule
	just       []AbstractModule
}

func (this *Implements) AddImplement(name string, impl AbstractModule) {
	this.implements[name] = impl
}

func (this *Implements) AddImplements(impl *Implements) {
	for k, v := range impl.implements {
		if _, exists := this.implements[k]; exists == false {
			this.implements[k] = v
		}
	}
}

func (this *Implements) AddBind(bindF func(binder *Binder)) {
	this.just = append(this.just, &JustInModule{bindF})
}

func NewImplements() *Implements {
	ret := Implements{make(map[string]AbstractModule), nil}
	return &ret
}
