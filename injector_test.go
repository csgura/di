package di_test

import (
	"fmt"
	"testing"

	"github.com/csgura/di"
)

type MyModule struct {
}

type MyModule2 struct {
}

type Hello interface {
	Hello()
}

type ConfigDB interface {
	get() string
}

type MemoryConfigDB struct {
}

func (m *MemoryConfigDB) get() string {
	return "cool"
}

type HelloWorld struct {
	db ConfigDB
}

func (h *HelloWorld) Hello() {
	fmt.Printf("%s : this is Hello World. say hello world\n", h.db.get())
}

func (*MyModule) Configure(binder *di.Binder) {
	binder.BindProvider((*Hello)(nil), func(inj di.Injector) interface{} {
		configfile := inj.GetProperty("config.file")
		fmt.Printf("configfile = %s\n", configfile)
		db := inj.GetInstance((*ConfigDB)(nil)).(ConfigDB)
		return &HelloWorld{db}
	})
}

// func (*MyModule2) Configure(binder *di.Binder) {
// 	binder.BindSingleton((*ConfigDB)(nil), &MemoryConfigDB{})
// }

func (*MyModule2) Configure(binder *di.Binder) {
	binder.BindProvider((*ConfigDB)(nil), func(inj di.Injector) interface{} {
		//h := inj.GetInstance((*Hello)(nil)).(Hello)
		//h.Hello()
		return &MemoryConfigDB{}
	})
}

func TestInjector(t *testing.T) {
	implements := di.NewImplements()
	implements.AddImplement("MyModule", &MyModule{})
	implements.AddImplement("MyModule2", &MyModule2{})

	loadingModuleList := []string{"MyModule", "MyModule2"}

	injector := di.NewInjector(implements, loadingModuleList)
	injector.SetProperty("config.file", "application.conf")

	ins := injector.GetInstance((*Hello)(nil)).(Hello)
	ins.Hello()

	db := injector.GetInstance((*ConfigDB)(nil)).(ConfigDB)
	fmt.Printf("config = %s\n", db.get())
}
