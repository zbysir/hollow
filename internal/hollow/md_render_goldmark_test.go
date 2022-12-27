package hollow

import (
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"path/filepath"
	"testing"
)

func TestGoldmark(t *testing.T) {
	jx, err := jsx.NewJsx(jsx.Option{
		SourceCache: nil,
		Debug:       false,
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	f := memfs.New()
	file, _ := f.Create("index.tsx")
	file.Write([]byte(`export default function HelloJSX({name}){ return <>{name}</> }`))

	m := newGoldMdRender(jx, gobilly.NewStdFs(f))

	cases := []struct {
		Name string
		In   []byte
		Get  []byte
	}{
		{
			Name: "Jsx",
			In: []byte(`
import Name from './index.tsx'

Hello

<Name name={'bysir'}>
</Name>
`),
			Get: []byte(`<p>Hello</p>
bysir`),
		},
		{
			Name: "Empty(Currently not supported)",
			In: []byte(`
import Name from './index.tsx'

Hello

<>
</>

<Name name={'bysir'}>
</Name>
`),
			Get: []byte("<p>Hello</p>\n<p>&lt;&gt;\n&lt;/&gt;</p>\nbysir"),
		},
		{
			Name: "Jsx Inline",
			In: []byte(`
import Name from './index.tsx'

Hello

<Name name={'bysir'}></Name>

!`),
			Get: []byte(`<p>Hello</p>
bysir<p>!</p>
`),
		},
		{
			Name: "Jsx SelfClose",
			In: []byte(`
import Name from './index.tsx'

Hello

<Name name={'bysir'} /> 333

!`),
			Get: []byte(`<p>Hello</p>
bysir<p>!</p>
`),
		},
		{
			Name: "Jsx with md",
			In: []byte(`
import Name from './index.tsx'

Hello

<Name name={'bysir'}>

**strong**

</Name>
`),
			Get: []byte(`<p>Hello</p>
bysir`),
		},
		{
			Name: "Base",
			In: []byte(`
Hello

<Warning>

**strong**

</Warning>
`),
			Get: []byte(`<p>Hello</p>
<Warning>
<p><strong>strong</strong></p>
</Warning>
`),
		},
		{
			Name: "WithMeta",

			In: []byte(`---
a: 1
---

import Name from './index.tsx'

Hello

<Name name={'bysir'}>
</Name>

!
`),
			Get: []byte(`<p>Hello</p>
bysir<p>!</p>
`),
		},
		{
			Name: "WithOutImport",

			In: []byte(`---
a: 1
---
Hello

<Name name={'bysir'}>
</Name>
`),
			Get: []byte(`<p>Hello</p>
<p>&lt;Name name={'bysir'}&gt;
</Name></p>
`),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			assert.Equal(t, string(c.Get), string(m.Render(c.In)))
		})
	}

}

func TestX(t *testing.T) {
	r := filepath.Join("contents/vue", "../../component/SearchBtn")
	t.Logf("%s", r)
}
