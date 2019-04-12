package di_test

import (
	"testing"

	"github.com/csgura/di"
)

func TestCombine(t *testing.T) {

	v1 := func(binder *di.Binder) {
		binder.Bind((*Value1)(nil)).ToInstance(&ValueImpl{"Value1"})
	}

	v2 := func(binder *di.Binder) {
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2"})

	}

	newm := di.CombineModule(di.BindFunc(v1), di.BindFunc(v2))

	injector := di.CreateInjector(newm)

	if injector.GetInstance((*Value1)(nil)).(Value1).Value() != "Value1" {
		t.Errorf("Value1 not binded")
	}

	if injector.GetInstance((*Value2)(nil)).(Value1).Value() != "Value2" {
		t.Errorf("Value2 not binded")
	}
}
func TestOverride(t *testing.T) {

	v0 := func(binder *di.Binder) {
		binder.Bind((*ValueInterface)(nil)).ToInstance(&ValueImpl{"Value"})
	}

	v1 := func(binder *di.Binder) {
		binder.Bind((*Value1)(nil)).ToInstance(&ValueImpl{"Value1"})
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2"})

	}

	v2 := func(binder *di.Binder) {
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2Over"})

	}

	newm := di.OverrideModule(di.BindFunc(v1)).With(di.BindFunc(v2))

	injector := di.CreateInjector(di.BindFunc(v0), newm)

	if injector.GetInstance((*Value1)(nil)).(Value1).Value() != "Value1" {
		t.Errorf("Value1 not binded")
	}

	if injector.GetInstance((*Value2)(nil)).(Value1).Value() != "Value2Over" {
		t.Errorf("Value2 not binded")
	}
}

func TestNotOverride(t *testing.T) {

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	v0 := func(binder *di.Binder) {
		binder.Bind((*ValueInterface)(nil)).ToInstance(&ValueImpl{"Value"})
	}

	v1 := func(binder *di.Binder) {
		binder.Bind((*Value1)(nil)).ToInstance(&ValueImpl{"Value1"})
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2"})

	}

	v2 := func(binder *di.Binder) {
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2Over"})

	}

	v3 := func(binder *di.Binder) {
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2Over2"})

	}

	newm := di.OverrideModule(di.BindFunc(v1)).With(di.BindFunc(v2))

	injector := di.CreateInjector(di.BindFunc(v0), newm, di.BindFunc(v3))

	if injector.GetInstance((*Value1)(nil)).(Value1).Value() != "Value1" {
		t.Errorf("Value1 not binded")
	}

	if injector.GetInstance((*Value2)(nil)).(Value1).Value() != "Value2Over" {
		t.Errorf("Value2 not binded")
	}
}

func TestImplementsOverride(t *testing.T) {
	impls := di.NewImplements()
	impls.AddImplement("All", di.OverrideModule(di.BindFunc(func(binder *di.Binder) {
		binder.Bind((*Value1)(nil)).ToInstance(&ValueImpl{"Value1All"})
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2All"})
		binder.Bind((*Value3)(nil)).ToInstance(&ValueImpl{"Value3All"})
	})))

	impls.AddImplement("V1", di.BindFunc(func(binder *di.Binder) {
		binder.Bind((*Value1)(nil)).ToInstance(&ValueImpl{"Value1"})
	}))

	impls.AddImplement("V2", di.BindFunc(func(binder *di.Binder) {
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2"})
	}))

	injector := impls.NewInjector([]string{"All", "V1", "V2"})
	if injector.GetInstance((*Value1)(nil)).(Value1).Value() != "Value1" {
		t.Errorf("Value1 not binded")
	}

	if injector.GetInstance((*Value2)(nil)).(Value1).Value() != "Value2" {
		t.Errorf("Value2 not binded")
	}

	if injector.GetInstance((*Value3)(nil)).(Value1).Value() != "Value3All" {
		t.Errorf("Value3 not binded")
	}
}

func TestImplementsNotEmptyOverride(t *testing.T) {

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	impls := di.NewImplements()
	impls.AddImplement("All", di.OverrideModule(di.BindFunc(func(binder *di.Binder) {
		binder.Bind((*Value1)(nil)).ToInstance(&ValueImpl{"Value1All"})
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2All"})
		binder.Bind((*Value3)(nil)).ToInstance(&ValueImpl{"Value3All"})
	})).With(di.BindFunc(func(binder *di.Binder) {
		binder.Bind((*Value1)(nil)).ToInstance(&ValueImpl{"Value1"})
	})))

	impls.AddImplement("V2", di.BindFunc(func(binder *di.Binder) {
		binder.Bind((*Value2)(nil)).ToInstance(&ValueImpl{"Value2"})
	}))

	injector := impls.NewInjector([]string{"All", "V2"})
	if injector.GetInstance((*Value1)(nil)).(Value1).Value() != "Value1" {
		t.Errorf("Value1 not binded")
	}

	if injector.GetInstance((*Value2)(nil)).(Value1).Value() != "Value2" {
		t.Errorf("Value2 not binded")
	}

	if injector.GetInstance((*Value3)(nil)).(Value1).Value() != "Value3All" {
		t.Errorf("Value3 not binded")
	}
}
