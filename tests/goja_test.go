package tests

import (
	"github.com/dop251/goja"
	"testing"
)

type Student struct {
	Name string
	Age  int
}

func TestGetSet(t *testing.T) {
	vm := goja.New()

	//o := vm.ToValue(Student{Name: "1"}).ToObject(vm)

	o := vm.NewObject()

	err := o.DefineAccessorProperty("Age", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(1)
	}), nil, goja.FLAG_TRUE, goja.FLAG_TRUE)
	if err != nil {
		t.Fatal(err)
	}

	vm2 := goja.New()
	err = vm2.Set("aaaa", o)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", o.Export())
	t.Logf("%+v", vm2.Get("aaa"))
}
