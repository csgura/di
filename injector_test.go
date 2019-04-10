package di_test

import (
	"fmt"
	"io"
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

var eagerResult string = ""

func (this *EagerunImpl) Run() {
	eagerResult = "done"
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

	time.Sleep(100 * time.Millisecond)
	if eagerResult != "done" {
		t.Errorf("EagerSingleton not created")
	}

}

type CloserModule struct {
}

type Closeable1 struct {
}

type Closeable2 struct {
}

type NotCloseable struct {
}

func (*Closeable1) Close() error {
	fmt.Println("Close Closeable1")
	return nil
}

func (*Closeable2) Close() error {
	fmt.Println("Close Closeable2")
	return nil
}

func (*CloserModule) Configure(binder *di.Binder) {
	binder.BindSingleton((*Closeable1)(nil), &Closeable1{})
	binder.BindSingleton((*Closeable2)(nil), &Closeable2{})
	binder.BindSingleton((*NotCloseable)(nil), &NotCloseable{})
}

func TestGetInstancesOf(t *testing.T) {

	implements := di.NewImplements()
	implements.AddImplement("MyModule", &CloserModule{})
	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{"MyModule"}

	injector := di.NewInjector(implements, loadingModuleList)

	list := injector.GetInstancesOf((*io.Closer)(nil))

	count := 0
	for _, ins := range list {
		count++
		c := ins.(io.Closer)
		c.Close()
	}

	if count != 2 {
		t.Errorf("GetInstanceOf( io.Close ) failed. close count = %d", count)
	}
	list = injector.GetInstancesOf((*Closeable1)(nil))
	count = 0
	for _, ins := range list {
		count++
		c := ins.(io.Closer)
		c.Close()
	}

	if count != 1 {
		t.Errorf("GetInstanceOf( Closeable1 ) failed. close count = %d", count)
	}

}

type Encoder func(data string) string

type EncoderModule struct {
}

func (*EncoderModule) Configure(binder *di.Binder) {
	binder.BindProvider((*Encoder)(nil), func(injector di.Injector) interface{} {
		ret := func(data string) string {
			return "hello : " + data
		}

		return (Encoder)(ret)
	}).AsEagerSingleton()
}

func TestFunctionBind(t *testing.T) {

	implements := di.NewImplements()
	implements.AddImplement("MyModule", &CloserModule{})
	implements.AddImplement("EncoderModule", &EncoderModule{})

	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{"MyModule", "EncoderModule"}

	injector := di.NewInjector(implements, loadingModuleList)

	list := injector.GetInstancesOf((*Encoder)(nil))

	count := 0
	for _, ins := range list {
		count++
		c := ins.(Encoder)
		fmt.Printf("after encode = %s\n", c("world"))
	}

	if count != 1 {
		t.Errorf("GetInstanceOf( func() {} ) failed. return count = %d", count)
	}

}

type First int
type Second int

var createdOrder = []string{}

type FirstModule struct{}
type SecondModule struct{}

func (r *FirstModule) Configure(binder *di.Binder) {
	binder.BindProvider((*First)(nil), func(injector di.Injector) interface{} {
		createdOrder = append(createdOrder, "FirstModule")
		f := First(1)
		return &f
	}).ShouldCreateBefore((*Second)(nil))
}

func (r *SecondModule) Configure(binder *di.Binder) {
	binder.BindProvider((*Second)(nil), func(injector di.Injector) interface{} {
		createdOrder = append(createdOrder, "SecondModule")
		f := Second(2)
		return &f
	}).AsEagerSingleton()
}

func TestBindOrder(t *testing.T) {
	createdOrder = nil
	implements := di.NewImplements()
	implements.AddImplement("FirstModule", &FirstModule{})
	implements.AddImplement("SecondModule", &SecondModule{})

	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{"FirstModule", "SecondModule"}

	di.NewInjector(implements, loadingModuleList)

	if createdOrder[0] != "FirstModule" {
		t.Errorf("FirstModule not binded first")
	}

	if createdOrder[1] != "SecondModule" {
		t.Errorf("SecondModule not binded second")
	}
}

type FirstModuleFallback struct{}

func (r *FirstModuleFallback) Configure(binder *di.Binder) {
	binder.IfNotBinded((*First)(nil)).ToProvider(func(injector di.Injector) interface{} {
		createdOrder = append(createdOrder, "FirstModuleFallback")
		f := First(1)
		return &f
	}).ShouldCreateBefore((*Second)(nil))
}

func TestBindFallback(t *testing.T) {
	createdOrder = nil
	implements := di.NewImplements()
	implements.AddImplement("FirstModule", &FirstModule{})
	implements.AddImplement("SecondModule", &SecondModule{})
	implements.AddImplement("FirstModuleFallback", &FirstModuleFallback{})

	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{"FirstModule", "SecondModule", "FirstModuleFallback"}

	di.NewInjector(implements, loadingModuleList)

	if len(createdOrder) != 2 {
		t.Errorf("more then 2 module binded")
	}

	if createdOrder[0] != "FirstModule" {
		t.Errorf("FirstModule not binded first")
	}

	if createdOrder[1] != "SecondModule" {
		t.Errorf("SecondModule not binded second")
	}
}

func TestBindFallback2(t *testing.T) {
	createdOrder = nil
	implements := di.NewImplements()
	implements.AddImplement("FirstModule", &FirstModule{})
	implements.AddImplement("SecondModule", &SecondModule{})
	implements.AddImplement("FirstModuleFallback", &FirstModuleFallback{})

	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{"SecondModule", "FirstModuleFallback"}

	di.NewInjector(implements, loadingModuleList)

	if len(createdOrder) != 2 {
		t.Errorf("more then 2 module binded")
	}

	if createdOrder[0] != "FirstModuleFallback" {
		t.Errorf("FirstModuleFallback not binded first")
	}

	if createdOrder[1] != "SecondModule" {
		t.Errorf("SecondModule not binded second")
	}
}

type ValueInterface interface {
	Value() string
}

type Value1 ValueInterface
type Value2 ValueInterface

type ValueImpl struct {
	value string
}

func (r *ValueImpl) Value() string {
	return r.value
}

type SubStruct struct {
	Value Value1
	str   string
}

type SubStructPointer struct {
	Value Value2
	iv    int
}

type GetValueFunc func() string

type PrometheusPort int
type PrometheusAddress string

type Target struct {
	Value          ValueInterface
	value          ValueInterface
	NotNilValue    ValueInterface
	Sub            SubStruct
	sub            SubStruct
	SubPtr         *SubStructPointer
	NotNilSubPtr   *SubStructPointer
	subPtr         *SubStructPointer
	str            string
	iv             int
	GetValue       GetValueFunc
	getValue       GetValueFunc
	NotNilGetValue GetValueFunc
	Port           PrometheusPort
	Address        PrometheusAddress
}

func TestInjectMembers(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		binder.Bind((*ValueInterface)(nil)).ToInstance(&ValueImpl{"Value"})
		binder.Bind((*Value1)(nil)).ToInstance(&ValueImpl{"Value1"})
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2"})
		binder.Bind((*GetValueFunc)(nil)).ToProvider(func(injector di.Injector) interface{} {
			return GetValueFunc(func() string {
				return "GetValueFunc"
			})
		})

		binder.Bind((*SubStructPointer)(nil)).ToProvider(func(injector di.Injector) interface{} {
			ret := SubStructPointer{}
			injector.InjectMembers(&ret)
			return &ret
		})
	})

	injector := di.NewInjector(implements, []string{})
	target := Target{
		NotNilValue:  &ValueImpl{"NotNilValue"},
		NotNilSubPtr: &SubStructPointer{Value: &ValueImpl{"NotNilSubPtr"}},
		NotNilGetValue: func() string {
			return "NotNilGetValue"
		},
	}

	injector.InjectMembers(&target)

	if target.value != nil {
		t.Errorf("target.value != nil")
		return
	}

	if target.Value == nil {
		t.Errorf("value not injected")
		return
	}

	if target.Value.Value() != "Value" {
		t.Errorf("target.Value.Value() != Value")
		return
	}

	if target.NotNilValue.Value() != "NotNilValue" {
		t.Errorf("target.NotNilValue.Value() != NotNilValue")
		return
	}

	if target.sub.Value != nil {
		t.Errorf("target.sub.Value != nil")
		return
	}

	if target.Sub.Value == nil {
		t.Errorf("target.Sub.Value == nil")
		return
	}
	if target.Sub.Value.Value() != "Value1" {
		t.Errorf("target.Sub.Value.Value() != Value1")
		return
	}

	if target.subPtr != nil {
		t.Errorf("target.subPtr != nil")
		return
	}

	if target.SubPtr == nil {
		t.Errorf("target.SubPtr == nil")
		return
	}

	if target.SubPtr.Value == nil {
		t.Errorf("target.SubPtr.Value == nil")
		return
	}

	if target.SubPtr.Value.Value() != "Value2" {
		t.Errorf("target.SubPtr.Value.Value() != Value2")
		return
	}

	if target.NotNilSubPtr.Value.Value() != "NotNilSubPtr" {
		t.Errorf("target.NotNilSubPtr.Value.Value() != NotNilSubPtr")
		return
	}

	if target.getValue != nil {
		t.Errorf("target.getValue != nil")
		return
	}

	if target.GetValue == nil {
		t.Errorf("target.GetValue == nil")
		return
	}

	if target.GetValue() != "GetValueFunc" {
		t.Errorf("target.GetValue() != GetValueFunc")
		return
	}

	if target.NotNilGetValue() != "NotNilGetValue" {
		t.Errorf("target.NotNilGetValue() != NotNilGetValue")
		return
	}
}

type TargetExplicit struct {
	ValueInject    ValueInterface `di:"inject"`
	ValueNotInject ValueInterface
	Sub            SubStruct `di:"inject"`
	SubNotInject   SubStruct
	Address        PrometheusAddress `di:"inject"`
	Port           PrometheusPort    `di:"inject"`
}

func TestInjectMembersExplicit(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		binder.Bind((*ValueInterface)(nil)).ToInstance(&ValueImpl{"Value"})
		binder.Bind((*Value1)(nil)).ToInstance(&ValueImpl{"Value1"})
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2"})
		binder.Bind((*GetValueFunc)(nil)).ToProvider(func(injector di.Injector) interface{} {
			return GetValueFunc(func() string {
				return "GetValueFunc"
			})
		})

		binder.Bind((*SubStructPointer)(nil)).ToProvider(func(injector di.Injector) interface{} {
			ret := SubStructPointer{}
			injector.InjectMembers(&ret)
			return &ret
		})

		//binder.Bind((*HttpPort)(nil)).ToInstance(HttpPort(8080))
		binder.Bind((*PrometheusPort)(nil)).ToInstance(8080)
		binder.Bind((*PrometheusAddress)(nil)).ToInstance("google.com")

	})

	injector := di.NewInjector(implements, []string{})
	target := TargetExplicit{}

	injector.InjectMembers(&target)

	if target.ValueInject == nil {
		t.Errorf("value not injected")
		return
	}

	if target.ValueInject.Value() != "Value" {
		t.Errorf("target.Value.Value() != Value")
		return
	}

	if target.ValueNotInject != nil {
		t.Errorf("target.ValueNotInject != nil")
		return
	}

	if target.Sub.Value == nil {
		t.Errorf("target.Sub.Value == nil")
		return
	}
	if target.Sub.Value.Value() != "Value1" {
		t.Errorf("target.Sub.Value.Value() != Value1")
		return
	}

	if target.SubNotInject.Value != nil {
		t.Errorf("target.Sub.Value != nil")
		return
	}

	if target.Port != 8080 {
		t.Errorf("target.Port != 8080")
		return
	}

	if target.Address != "google.com" {
		t.Errorf("target.Address != google.com")
		return
	}
}
