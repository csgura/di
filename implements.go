package di

type Implements struct {
	implements map[string]AbstractModule
}

func (this *Implements) AddImplement(name string, impl AbstractModule) {
	this.implements[name] = impl
}

func NewImplements() *Implements {
	ret := Implements{make(map[string]AbstractModule)}
	return &ret
}
