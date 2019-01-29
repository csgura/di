package di_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/csgura/di"
)

type MyModule struct {
}

type MyModule2 struct {
}

type MyModuleDup struct {
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

type HelloGura struct {
}

func (h *HelloGura) Hello() {
	fmt.Printf("Hello Gura\n")
}

func (*MyModule) Configure(binder *di.Binder) {
	binder.BindProvider((*Hello)(nil), func(inj di.Injector) interface{} {
		configfile := inj.GetProperty("config.file")
		fmt.Printf("configfile = %s\n", configfile)
		db := inj.GetInstance((*ConfigDB)(nil)).(ConfigDB)
		return &HelloWorld{db}
	})
}

func (*MyModuleDup) Configure(binder *di.Binder) {
	binder.BindSingleton((*Hello)(nil), &HelloGura{})
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

type MultipleInstance struct {
	sequence int
}

func (m *MultipleInstance) get() string {
	return strconv.Itoa(m.sequence)
}

type MyModuleNonSingleton struct {
	sequence int
}

func (r *MyModuleNonSingleton) Configure(binder *di.Binder) {
	binder.BindProvider((*ConfigDB)(nil), func(inj di.Injector) interface{} {
		//h := inj.GetInstance((*Hello)(nil)).(Hello)
		//h.Hello()
		r.sequence++
		return &MultipleInstance{r.sequence}
	}).AsNonSingleton()
}

type EagerModule struct {
}

type EagerRun interface {
	Run()
}

type EagerunImpl struct {
}

func (this *EagerunImpl) Run() {

	for {
		time.Sleep(1 * time.Second)
		fmt.Println("eager work")
	}
}
func (*EagerModule) Configure(binder *di.Binder) {
	binder.BindProvider((*EagerRun)(nil), func(inj di.Injector) interface{} {

		fmt.Printf("EagerModule configured\n")
		//h := inj.GetInstance((*Hello)(nil)).(Hello)
		//h.Hello()
		ret := &EagerunImpl{}
		go func() { ret.Run() }()
		return ret
	}).AsEagerSingleton()
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

func TestDuplicatedBind(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	implements := di.NewImplements()
	implements.AddImplement("MyModule", &MyModule{})
	implements.AddImplement("MyModule2", &MyModule2{})
	implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{"MyModule", "MyModule2", "MyModuleDup"}

	injector := di.NewInjector(implements, loadingModuleList)

	ins := injector.GetInstance((*Hello)(nil)).(Hello)
	ins.Hello()

	fmt.Printf("this code should not execute\n")
}

func TestNotBind(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	implements := di.NewImplements()
	// implements.AddImplement("MyModule", &MyModule{})
	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{}

	injector := di.NewInjector(implements, loadingModuleList)

	ins := injector.GetInstance((*Hello)(nil)).(Hello)
	ins.Hello()

	fmt.Printf("this code should not execute\n")
}

func TestNotImplemented(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	implements := di.NewImplements()
	// implements.AddImplement("MyModule", &MyModule{})
	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{"MyModule"}

	injector := di.NewInjector(implements, loadingModuleList)

	ins := injector.GetInstance((*Hello)(nil)).(Hello)
	ins.Hello()

	fmt.Printf("this code should not execute\n")
}

func TestJustIn(t *testing.T) {

	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		binder.BindSingleton((*Hello)(nil), &HelloGura{})
	})
	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{}

	injector := di.NewInjector(implements, loadingModuleList)

	ins := injector.GetInstance((*Hello)(nil)).(Hello)
	ins.Hello()
}

func TestNonSingleton(t *testing.T) {
	implements := di.NewImplements()
	implements.AddImplement("MyModule2", &MyModuleNonSingleton{})

	loadingModuleList := []string{"MyModule2"}

	injector := di.NewInjector(implements, loadingModuleList)

	db := injector.GetInstance((*ConfigDB)(nil)).(ConfigDB)
	fmt.Printf("config = %p, %s\n", db, db.get())

	db2 := injector.GetInstance((*ConfigDB)(nil)).(ConfigDB)
	fmt.Printf("config = %p, %s\n", db2, db2.get())

	if db.get() == db2.get() {
		t.Errorf("new instance not generated")
	}
}
func TestEager(t *testing.T) {

	implements := di.NewImplements()
	implements.AddImplement("MyModule", &EagerModule{})
	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{"MyModule"}

	di.NewInjector(implements, loadingModuleList)

	time.Sleep(3 * time.Second)

}
