package tests

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/zbysir/hollow/jslib/mdx"
	"io"
	"testing"
)

type Student struct {
	Name string
	Age  int
}

func TestGetterSetter(t *testing.T) {
	vm := goja.New()

	//o := vm.ToValue(Student{Name: "1"}).ToObject(vm)
	// https://github.com/dop251/goja/issues/279
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

func TestRunMdx(t *testing.T) {
	vm := goja.New()
	require.NewRegistry().Enable(vm)
	console.Enable(vm)
	f, err := mdx.Dist.Open("dist/index.js")
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	//t.Logf("%s", body)
	vm.RunString("module = {}")
	v, err := vm.RunScript("x", string(body))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", v.Export())
}
