package tests

import (
	jsx "github.com/zbysir/gojsx"
	"testing"
)

func TestGoJsx(t *testing.T) {
	jx, err := jsx.NewJsx(jsx.Option{
		SourceCache: nil,
		Debug:       false,
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	v, err := jx.RunJs([]byte(`import {Chart} from './snowfall.js'; const HelloJSX = ({a}) => <>Hello {a}</>; <HelloJSX a={1}></HelloJSX>;<Chart/>;`), jsx.WithTransform(true), jsx.WithRunFileName("index.tsx"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", v.Export())
	vd := jsx.VDom(v.Export().(map[string]interface{}))
	t.Logf("%+v", vd.Render())

}
