//go:build go1.18
// +build go1.18

package di_test

import (
	"fmt"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/csgura/di"
	"github.com/csgura/fp"
)

type Nothing any

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
	di.BindProvider[Hello](binder, func(inj di.Injector) Hello {
		configfile := inj.GetProperty("config.file")
		fmt.Printf("configfile = %s\n", configfile)
		db := inj.GetInstance((*ConfigDB)(nil)).(ConfigDB)
		return &HelloWorld{db}
	})
}

func (*MyModuleDup) Configure(binder *di.Binder) {
	di.BindSingleton[Hello](binder, &HelloGura{})
}

// func (*MyModule2) Configure(binder *di.Binder) {
// 	binder.BindSingleton((*ConfigDB)(nil), &MemoryConfigDB{})
// }

func (*MyModule2) Configure(binder *di.Binder) {
	di.BindProvider[ConfigDB](binder, func(inj di.Injector) ConfigDB {
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
	di.BindProvider[ConfigDB](binder, func(inj di.Injector) ConfigDB {
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
	di.BindProvider[EagerRun](binder, func(inj di.Injector) EagerRun {

		fmt.Printf("EagerModule configured\n")
		//h := inj.GetInstance((*Hello)(nil)).(Hello)
		//h.Hello()
		ret := &EagerunImpl{}
		ret.Run()
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

	ins = di.GetInstance[Hello](injector)
	ins.Hello()

	db := injector.GetInstance((*ConfigDB)(nil)).(ConfigDB)
	fmt.Printf("config = %s\n", db.get())
}

type TypeA struct {
	B *TypeB `di:"inject"`
}

type TypeB struct {
	C *TypeC `di:"inject"`
}

type TypeC struct {
	A *TypeA `di:"inject"`
}

func TestCycle(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	implements := di.NewImplements()

	implements.AddBind(func(binder *di.Binder) {
		di.BindSingleton[*TypeA](binder, &TypeA{})
		di.BindSingleton[*TypeB](binder, &TypeB{})
		di.BindSingleton[*TypeC](binder, &TypeC{})
	})

	injector := di.NewInjector(implements, nil)
	di.GetInstance[*TypeA](injector)

	//injector.GetInstance((*TypeA)(nil))
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

	ins := di.GetInstance[Hello](injector)
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

	ins := di.GetInstance[Hello](injector)
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

	ins := di.GetInstance[Hello](injector)
	ins.Hello()

	fmt.Printf("this code should not execute\n")
}

func TestAnonymous(t *testing.T) {

	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.BindSingleton[Hello](binder, &HelloGura{})
	})
	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{}

	injector := di.NewInjector(implements, loadingModuleList)

	ins := di.GetInstance[Hello](injector)
	ins.Hello()
}

func TestNonSingleton(t *testing.T) {
	implements := di.NewImplements()
	implements.AddImplement("MyModule2", &MyModuleNonSingleton{})

	loadingModuleList := []string{"MyModule2"}

	injector := di.NewInjector(implements, loadingModuleList)

	db := di.GetInstance[ConfigDB](injector)
	fmt.Printf("config = %p, %s\n", db, db.get())

	db2 := di.GetInstance[ConfigDB](injector)
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
	world string
}

type Closeable2 struct {
	world string
}

type NotCloseable struct {
}

func (r *Closeable1) Close() error {
	fmt.Println("Close Closeable1 : ", r.world)
	return nil
}

func (*Closeable2) Close() error {
	fmt.Println("Close Closeable2")
	return nil
}

type Closeable3 Closeable1
type Closeable4 io.Closer

func (*CloserModule) Configure(binder *di.Binder) {
	c1 := &Closeable1{"c1"}
	di.BindSingleton[*Closeable1](binder, c1)
	di.BindSingleton[*Closeable2](binder, &Closeable2{})
	di.BindSingleton[*Closeable3](binder, (*Closeable3)(c1))

	di.IfNotBinded[Closeable4](binder).ToInstance(&Closeable2{"c4"})
	di.BindSingleton[*NotCloseable](binder, &NotCloseable{})

}

func TestGetInstancesOf(t *testing.T) {

	implements := di.NewImplements()
	implements.AddImplement("MyModule", &CloserModule{})
	// implements.AddImplement("MyModule2", &MyModule2{})
	// implements.AddImplement("MyModuleDup", &MyModuleDup{})

	loadingModuleList := []string{"MyModule"}

	injector := di.NewInjector(implements, loadingModuleList)

	list := di.GetInstancesOf[io.Closer](injector)

	count := 0
	for _, c := range list {
		count++
		c.Close()
	}

	if count != 3 {
		t.Errorf("GetInstanceOf( io.Close ) failed. close count = %d", count)
	}
	list2 := di.GetInstancesOf[*Closeable1](injector)
	count = 0
	for _, c := range list2 {
		count++
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
	di.BindProvider[Encoder](binder, func(injector di.Injector) Encoder {
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

	list := di.GetInstancesOf[Encoder](injector)

	count := 0
	for _, c := range list {
		count++
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
	di.BindProvider[*First](binder, func(injector di.Injector) *First {
		createdOrder = append(createdOrder, "FirstModule")
		f := First(1)
		return &f
	}).ShouldCreateBefore((*Second)(nil))
}

func (r *SecondModule) Configure(binder *di.Binder) {
	di.BindProvider[*Second](binder, func(injector di.Injector) *Second {
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
	di.IfNotBinded[*First](binder).ToProvider(func(injector di.Injector) *First {
		createdOrder = append(createdOrder, "FirstModuleFallback")
		f := First(1)
		return &f
	}).ShouldCreateBefore((*Second)(nil))
}

func TestBindFallback(t *testing.T) {
	createdOrder = nil

	di.CreateInjector(&FirstModule{}, &SecondModule{}, &FirstModuleFallback{})

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
type Value3 ValueInterface

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
	Value          fp.Option[ValueInterface]
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
	Injector       di.Injector
	ShouldNone     fp.Option[Nothing]
}

func TestInjectMembers(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.Bind[ValueInterface](binder).ToInstance(&ValueImpl{"Value"})
		di.Bind[Value1](binder).ToInstance(&ValueImpl{"Value1"})
		di.Bind[Value2](binder).ToInstance(&ValueImpl{"Value2"})
		di.Bind[GetValueFunc](binder).ToProvider(func(injector di.Injector) GetValueFunc {
			return GetValueFunc(func() string {
				return "GetValueFunc"
			})
		})

		di.Bind[*SubStructPointer](binder).ToProvider(func(injector di.Injector) *SubStructPointer {
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

	if target.Value.IsEmpty() {
		t.Errorf("value not injected")
		return
	}

	if target.Value.Get().Value() != "Value" {
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

	if target.Port != 0 {
		t.Errorf("target.Port != 0")
		return
	}

	if target.Injector == nil {
		t.Errorf("target.Injector == nil")
		return
	}

	if target.ShouldNone.IsDefined() {
		t.Errorf("target.ShouldNone is defined")
		return
	}
}

type TargetExplicit struct {
	ValueInject ValueInterface `di:"inject,nilable"`
	notExported ValueInterface `di:"inject"`

	ValueNotInject ValueInterface
	Sub            SubStruct `di:"inject"`
	SubNotInject   SubStruct
	Address        PrometheusAddress `di:"inject"`
	Port           PrometheusPort    `di:"inject"`
	Injector       di.Injector       `di:"inject"`
	Nilable        Hello             `di:"inject,nilable"`
}

func TestInjectMembersExplicit(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.Bind[ValueInterface](binder).ToInstance(&ValueImpl{"Value"})
		di.Bind[Value1](binder).ToInstance(&ValueImpl{"Value1"})
		di.Bind[Value2](binder).ToInstance(&ValueImpl{"Value2"})
		di.Bind[GetValueFunc](binder).ToProvider(func(injector di.Injector) GetValueFunc {
			return GetValueFunc(func() string {
				return "GetValueFunc"
			})
		})

		di.Bind[*SubStructPointer](binder).ToProvider(func(injector di.Injector) *SubStructPointer {
			ret := SubStructPointer{}
			injector.InjectMembers(&ret)
			return &ret
		})

		//binder.Bind((*HttpPort)(nil)).ToInstance(HttpPort(8080))
		di.Bind[PrometheusPort](binder).ToInstance(8080)
		di.Bind[PrometheusAddress](binder).ToInstance("google.com")

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

	if target.Injector == nil {
		t.Errorf("target.Injector == nil")
		return
	}

	if target.notExported != nil {
		t.Errorf("target.notExported != nil")
		return
	}

	if target.Nilable != nil {
		t.Errorf("target.Nilable != nil")
		return
	}
}

func TestPrimitiveBinding(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {

		//binder.Bind((*HttpPort)(nil)).ToInstance(HttpPort(8080))
		di.Bind[PrometheusPort](binder).ToInstance(PrometheusPort(8080))
		di.Bind[PrometheusAddress](binder).ToInstance("google.com")

	})

	injector := di.NewInjector(implements, []string{})
	port := di.GetInstance[PrometheusPort](injector)
	addr := di.GetInstance[PrometheusAddress](injector)

	fmt.Printf("addr = %s , port= %d\n", addr, port)

	if port != 8080 {
		t.Errorf("target.Port != 8080")
		return
	}

	if addr != "google.com" {
		t.Errorf("target.Address != google.com")
		return
	}
}

func TestInjectCall(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.Bind[ValueInterface](binder).ToInstance(&ValueImpl{"Value"})
		di.Bind[Value1](binder).ToInstance(&ValueImpl{"Value1"})
		di.Bind[Value2](binder).ToInstance(&ValueImpl{"Value2"})
		di.Bind[GetValueFunc](binder).ToProvider(func(injector di.Injector) GetValueFunc {
			return GetValueFunc(func() string {
				return "GetValueFunc"
			})
		})

		di.Bind[*SubStructPointer](binder).ToProvider(func(injector di.Injector) *SubStructPointer {
			ret := SubStructPointer{}
			injector.InjectMembers(&ret)
			return &ret
		})

		//di.Bind[*HttpPort](binder).ToInstance(HttpPort(8080))
		di.Bind[PrometheusPort](binder).ToInstance(8080)
		di.Bind[PrometheusAddress](binder).ToInstance("google.com")

	})

	injector := di.NewInjector(implements, []string{})
	ret := injector.InjectAndCall(func(v1 fp.Option[ValueInterface], v2 Value1, ptr *SubStructPointer, port PrometheusPort, shouldNone fp.Option[Nothing]) ValueInterface {
		if v1.IsEmpty() {
			t.Fatal("v1 is empty")
			return nil
		}
		if v2 == nil {
			t.Fatal("v2 == nil")
			return nil
		}

		if ptr == nil {
			t.Fatal("ptr == nil")
			return nil
		}
		if port != 8080 {
			t.Fatal("port != 8080")
			return nil
		}

		if shouldNone.IsDefined() {
			t.Fatal("shouldNone is defined")
			return nil
		}

		return &ValueImpl{"ValueRet"}
	}).(ValueInterface)

	if ret.Value() != "ValueRet" {
		t.Errorf("ret.Value() != ValueRet")
	}

	ret2 := injector.InjectAndCall(func(v1 ValueInterface, v2 Value1, ptr *SubStructPointer, port PrometheusPort) (ValueInterface, error) {
		return &ValueImpl{"ValueRet"}, nil
	}).([]interface{})

	if ret2[0].(ValueInterface).Value() != "ValueRet" {
		t.Errorf("ret.Value() != ValueRet")
	}

	injector.InjectAndCall(func() int {
		return 10
	})
}

type constructorResult struct {
	v1     ValueInterface
	v2     Value1
	ptr    *SubStructPointer
	server *PrometheusServerInfo
}

type PrometheusServerInfo struct {
	Address PrometheusAddress `di:"inject"`
	Port    PrometheusPort    `di:"inject"`
}

func constructor(v1 ValueInterface, v2 Value1, ptr *SubStructPointer, server *PrometheusServerInfo) *constructorResult {
	return &constructorResult{v1, v2, ptr, server}
}
func TestBindConstructor(t *testing.T) {

	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.Bind[ValueInterface](binder).ToInstance(&ValueImpl{"Value"})
		di.Bind[Value1](binder).ToInstance(&ValueImpl{"Value1"})
		di.Bind[Value2](binder).ToInstance(&ValueImpl{"Value2"})
		di.Bind[GetValueFunc](binder).ToProvider(func(injector di.Injector) GetValueFunc {
			return GetValueFunc(func() string {
				return "GetValueFunc"
			})
		})

		di.Bind[*SubStructPointer](binder).ToProvider(func(injector di.Injector) *SubStructPointer {
			ret := SubStructPointer{}
			injector.InjectMembers(&ret)
			return &ret
		})

		//di.Bind[*HttpPort](binder).ToInstance(HttpPort(8080))
		di.Bind[PrometheusPort](binder).ToInstance(8080)
		di.Bind[PrometheusAddress](binder).ToInstance("google.com")

		di.Bind[*constructorResult](binder).ToConstructor(constructor)

	})
	injector := di.NewInjector(implements, []string{})
	res := di.GetInstance[*constructorResult](injector)
	if res == nil {
		t.Errorf("res == nil")
	}

	if res.v1 == nil {
		t.Errorf("res.v1 == nil")
	}

	if res.v2 == nil {
		t.Errorf("res.v2 == nil")
	}

	if res.ptr == nil {
		t.Errorf("res.ptr == nil")
	}

	if res.server == nil {
		t.Errorf("res.server == nil")
	}

	if res.server.Address != "google.com" {
		t.Errorf("res.server.Address != google.com")
	}

	if res.server.Port != 8080 {
		t.Errorf("res.server.Port != 8080")
	}

}

type Registry interface {
	Register(name string, value interface{})
	Get(name string) interface{}
}

type MapRegistry map[string]interface{}

func (m MapRegistry) Register(name string, value interface{}) {

	m[name] = value
}

func (m MapRegistry) Get(name string) interface{} {
	return m[name]
}
func TestDecorator(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.Bind[Registry](binder).ToInstance(MapRegistry{})

		di.AddDecoratorOf[Registry](binder, func(injector di.Injector) {
			regi := injector.GetInstance((*Registry)(nil)).(Registry)
			regi.Register("hello", "world")
		})
	})

	injector := di.NewInjector(implements, []string{})
	regi := injector.GetInstance((*Registry)(nil)).(Registry)
	if regi.Get("hello") != "world" {
		t.Errorf("decorator not called")
	}
}

type ClassA struct {
	s *ClassS
	t string
}
type ClassB struct {
	s *ClassS
	t string
}
type ClassS struct {
	a *ClassA
	t string
}

func TestDecoratorCycle(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.Bind[*ClassS](binder).ToInstance(&ClassS{t: "hello"})

		di.AddDecoratorOf[*ClassS](binder, func(injector di.Injector) {
			s := di.GetInstance[*ClassS](injector)
			s.a = di.GetInstance[*ClassA](injector)
			s.t = s.a.t
		})

		di.BindConstructor[*ClassA](binder, func(s *ClassS) interface{} {
			return &ClassA{s, "a"}
		})

		di.BindConstructor[*ClassB](binder, func(s *ClassS) interface{} {
			return &ClassB{s, "b"}
		})

	})

	injector := di.NewInjector(implements, []string{})
	s := di.GetInstance[*ClassS](injector)
	if s.t != "a" {
		t.Errorf("decorator not called")
	}
}

func TestInjectValue(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.Bind[ValueInterface](binder).ToInstance(&ValueImpl{"Value"})
		di.Bind[Value1](binder).ToInstance(&ValueImpl{"Value1"})
		di.Bind[Value2](binder).ToInstance(&ValueImpl{"Value2"})
		di.Bind[GetValueFunc](binder).ToProvider(func(injector di.Injector) GetValueFunc {
			return GetValueFunc(func() string {
				return "GetValueFunc"
			})
		})

		di.Bind[*SubStructPointer](binder).ToProvider(func(injector di.Injector) *SubStructPointer {
			ret := SubStructPointer{}
			injector.InjectMembers(&ret)
			return &ret
		})

		//di.Bind[HttpPort](binder).ToInstance(HttpPort(8080))
		di.Bind[PrometheusPort](binder).ToInstance(8080)
		di.Bind[PrometheusAddress](binder).ToInstance("google.com")

		di.Bind[*constructorResult](binder).ToConstructor(constructor)

	})
	injector := di.NewInjector(implements, []string{})

	var v ValueInterface
	injector.InjectValue(&v)
	if v.Value() != "Value" {
		t.Errorf("t.Value() != Value")
	}

	var v1 Value1
	injector.InjectValue(&v1)
	if v1.Value() != "Value1" {
		t.Errorf("v1.Value() != Value1")
	}

	var f GetValueFunc
	injector.InjectValue(&f)
	if f() != "GetValueFunc" {
		t.Error("f() != GetValueFunc")
	}

	var s *SubStructPointer

	injector.InjectValue(&s)
	if s.Value.Value() != "Value2" {
		t.Error(`s.Value.Value() != "Value2"`)
	}

	var p PrometheusPort
	injector.InjectValue(&p)
	if p != 8080 {
		t.Error(`p != 8080`)
	}

	var a PrometheusAddress
	injector.InjectValue(&a)
	if a != "google.com" {
		t.Error(`a != "google.com"`)
	}

	var res *constructorResult
	injector.InjectValue(&res)
	if res.v1 == nil {
		t.Errorf("res.v1 == nil")
	}

	if res.v2 == nil {
		t.Errorf("res.v2 == nil")
	}

	if res.ptr == nil {
		t.Errorf("res.ptr == nil")
	}

	if res.server == nil {
		t.Errorf("res.server == nil")
	}

	if res.server.Address != "google.com" {
		t.Errorf("res.server.Address != google.com")
	}

	if res.server.Port != 8080 {
		t.Errorf("res.server.Port != 8080")
	}

}

func TestInjectDecorator(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.AddDecoratorOf[di.Injector](binder, func(injector di.Injector) {
			injector.SetProperty("hello", "world")
		})

		di.Bind[ValueInterface](binder).ToProvider(func(injector di.Injector) ValueInterface {
			return &ValueImpl{injector.GetProperty("hello")}
		}).AsEagerSingleton()
	})

	injector := implements.NewInjector(nil)
	var v ValueInterface
	injector.InjectValue(&v)
	if v.Value() != "world" {
		t.Errorf("t.Value != world")
	}
}

func TestNilPointerBinding(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {

		di.Bind[*ValueImpl](binder).ToProvider(func(injector di.Injector) *ValueImpl {

			var ret *ValueImpl

			return ret
		})
	})

	injector := implements.NewInjector(nil)
	var v *ValueImpl
	injector.InjectValue(&v)

	if v == nil {
		t.Errorf("t == nil")
	}

}

func constructorTimeout(v1 ValueInterface, v2 Value1, ptr *SubStructPointer, server *PrometheusServerInfo) *constructorResult {
	time.Sleep(50 * time.Millisecond)
	return &constructorResult{v1, v2, ptr, server}
}

func TestInjectorTimeout(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		} else {
			fmt.Println(r)
		}
	}()

	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.Bind[ValueInterface](binder).ToInstance(&ValueImpl{"Value"})
		di.Bind[Value1](binder).ToInstance(&ValueImpl{"Value1"})
		di.Bind[Value2](binder).ToInstance(&ValueImpl{"Value2"})
		di.Bind[GetValueFunc](binder).ToProvider(func(injector di.Injector) GetValueFunc {
			return GetValueFunc(func() string {
				return "GetValueFunc"
			})
		})

		di.Bind[*SubStructPointer](binder).ToProvider(func(injector di.Injector) *SubStructPointer {
			ret := SubStructPointer{}
			injector.InjectMembers(&ret)
			return &ret
		})

		//di.Bind[HttpPort](binder).ToInstance(HttpPort(8080))
		di.Bind[PrometheusPort](binder).ToInstance(8080)
		di.Bind[PrometheusAddress](binder).ToInstance("google.com")

		di.Bind[*constructorResult](binder).ToConstructor(constructorTimeout).AsEagerSingleton()

	})
	implements.NewInjectorWithTimeout(nil, 20*time.Millisecond)
}

type client interface {
	Do(req string) string
}

type clientImpl struct{}

func (r *clientImpl) Do(req string) string {
	return "hello " + req
}

type clientLogger struct {
	client
}

func (r *clientLogger) Do(req string) string {
	res := r.client.Do(req)
	fmt.Printf("req : %s, res : %s\n", req, res)
	return "logged " + res
}

type Counter struct {
	count int
}

type clientCounter struct {
	counter *Counter
	client
}

func (r *clientCounter) Do(req string) string {
	r.counter.count++
	return r.client.Do(req)
}

func TestInterceptor(t *testing.T) {
	implements := di.NewImplements()

	implements.AddBind(func(binder *di.Binder) {
		di.BindConstructor[client](binder, func() client {
			return &clientImpl{}
		})
	})

	implements.AddBind(func(binder *di.Binder) {
		di.BindInterceptor[client](binder, func(injector di.Injector, value client) client {
			return &clientLogger{value}
		})
	})

	implements.AddBind(func(binder *di.Binder) {
		di.BindSingleton[*Counter](binder, &Counter{})
		di.BindInterceptor[client](binder, func(injector di.Injector, value client) client {
			counter := di.GetInstance[*Counter](injector)
			return &clientCounter{counter, value}
		})
	})

	injector := implements.NewInjector(nil)

	var cli client
	injector.InjectValue(&cli)

	var counter *Counter
	injector.InjectValue(&counter)

	res := cli.Do("world")
	if res != "logged hello world" {
		t.Errorf("res = %s", res)
	}

	if counter.count != 1 {
		t.Errorf("counter.count != 1 , %d", counter.count)
	}

}

func TestBindToInstanceInjection(t *testing.T) {
	implements := di.NewImplements()

	implements.AddBind(func(binder *di.Binder) {
		di.BindConstructor[client](binder, func() client {
			return &clientImpl{}
		})
	})

	type proxy struct {
		Cli client `di:"inject"`
	}

	implements.AddBind(func(binder *di.Binder) {
		di.BindSingleton[*proxy](binder, &proxy{})
	})

	injector := implements.NewInjector(nil)

	var p *proxy
	injector.InjectValue(&p)

	if p.Cli == nil {
		t.Errorf("p.Cli is nil")
	}
}

func TestBindToInstanceInjectionError(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	implements := di.NewImplements()

	type proxy struct {
		Cli client `di:"inject"`
	}

	implements.AddBind(func(binder *di.Binder) {
		di.BindSingleton[*proxy](binder, &proxy{})
	})

	injector := implements.NewInjector(nil)

	var p *proxy
	injector.InjectValue(&p)

	if p.Cli == nil {
		t.Errorf("p.Cli is nil")
	}

}

func TestInjectorRace(t *testing.T) {
	implements := di.NewImplements()

	implements.AddBind(func(binder *di.Binder) {
		di.BindConstructor[client](binder, func() client {
			return &clientImpl{}
		})

		di.AddDecoratorOf[*client](binder, func(injector di.Injector) {
			injector.GetInstance((*client)(nil))
		})
	})

	injector := implements.NewInjector(nil)

	go func() {
		di.GetInstance[client](injector)
	}()

	go func() {
		di.GetInstance[client](injector)
	}()

	time.Sleep(10 * time.Millisecond)
}

func TestNilBinding(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	implements := di.NewImplements()

	implements.AddBind(func(binder *di.Binder) {
		binder.BindConstructor((client)(nil), func() client {
			return &clientImpl{}
		}).AsEagerSingleton()
	})

	injector := implements.NewInjector(nil)
	injector.GetInstance((client)(nil))

	t.Error("not panic")
}

func TestInjectorNilConstructor(t *testing.T) {
	implements := di.NewImplements()

	implements.AddBind(func(binder *di.Binder) {
		di.BindConstructor[client](binder, func() *clientImpl {
			return nil
		})

	})

	injector := implements.NewInjector(nil)

	//	ret := injector.GetInstance((*client)(nil))

	ret := di.GetInstance[client](injector)
	if ret != nil {
		t.Error("ret is not nil")
	}
}

type server struct {
}

type mapper struct {
}

type starter interface{}

func NotTestDecoratorPanic(t *testing.T) {
	implements := di.NewImplements()
	implements.AddBind(func(binder *di.Binder) {
		di.BindConstructor[*server](binder, func() *server {
			fmt.Println("server created")
			return &server{}
		})

		di.BindConstructor[*mapper](binder, func(server *server) *mapper {
			fmt.Println("mapper created")

			return &mapper{}
		})

		di.AddDecoratorOf[*server](binder, func(ij di.Injector) {
			fmt.Println("server decorated")

			ij.GetInstance((*mapper)(nil))
		})

		di.BindConstructor[*starter](binder, func(server *server) interface{} {
			fmt.Println("started")

			return "started"
		})

		di.BindConstructor[client](binder, func(mapper *mapper) interface{} {
			fmt.Println("client created")

			return &clientImpl{}
		}).AsEagerSingleton()

	})
	implements.NewInjectorWithTrace(nil, func(info *di.TraceInfo) {
		fmt.Println(info)
	})
}

func TestOptional(t *testing.T) {
	implements := di.NewImplements()
	implements.AddImplement("MyModule", &MyModule{})
	implements.AddImplement("MyModule2", &MyModule2{})

	loadingModuleList := []string{"MyModule", "MyModule2"}

	injector := di.NewInjector(implements, loadingModuleList)
	injector.SetProperty("config.file", "application.conf")

	ins := di.GetInstanceOpt[Hello](injector)
	if ins.IsDefined() == false {
		t.Error("ins is empty")
	}

	ins2 := di.GetInstanceOpt[client](injector)
	if ins2.IsDefined() != false {
		t.Error("ins is not empty")
	}

}
