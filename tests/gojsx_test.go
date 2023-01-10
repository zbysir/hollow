package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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

	jsCode := "const c2 = 'bysir';const A = ({name}) => <>Hello {name}</>\nconst B = ({name}) => <p>HH</p>;"
	jsxCode := "<><A name={<B/>}></A> {c2}</>"

	code := fmt.Sprintf("%s;module.exports = %s", jsCode, jsxCode)

	v, err := jx.ExecCode([]byte(code), jsx.WithFileName("index.tsx"))
	if err != nil {
		t.Fatal(err)
	}

	vd := jsx.VDom(v.Exports)

	assert.Equal(t, "Hello <p>HH</p> bysir", vd.Render())
}
