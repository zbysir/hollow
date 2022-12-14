package hollow

import (
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMdx(t *testing.T) {

	f := memfs.New()
	file, _ := f.Create("index.tsx")
	file.Write([]byte(`export default function HelloJSX({name}){ return <>{name}</> }`))

	m := NewGoldMdRender(GoldMdRenderOptions{
		AssetsUrlProcess: nil,
	})

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
			Name: "Empty",
			In: []byte(`
import Name from './index.tsx'

Hello

<>
</>

<Name name={'bysir'}>
</Name>
`),
			Get: []byte("<p>Hello</p>\nbysir"),
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
		}, {
			Name: "A",

			In: []byte(`---
a: 1
---

import A from './index.tsx'

Hello

<A name={'bysir'}>
</A>

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
		{
			Name: "Empty",

			In: []byte(`
<>
	{[1,2].map((i)=><p> {i} </p>)}
</>
`),
			Get: []byte(`<p> 1 </p><p> 2 </p>`),
		},
		{
			Name: "Inline",

			In: []byte(`
const A = ({name}) => <>Hello {name}</>
const name = 1

<A name={name}></A>
`),
			Get: []byte(`Hello 1`),
		},
		{
			Name: "Inline2",

			In: []byte(`
const A = ({name}) => <>Hello {name}</>
const name = 1

<A name={name}></A>
`),
			Get: []byte(`Hello 1`),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			s, err := m.Render(c.In)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, string(c.Get), string(s.Body))
		})
	}

}

func TestAssert(t *testing.T) {

	f := memfs.New()
	file, _ := f.Create("index.tsx")
	file.Write([]byte(`export default function HelloJSX({name}){ return <>{name}</> }`))

	m := NewGoldMdRender(GoldMdRenderOptions{
		AssetsUrlProcess: func(s string) string {
			return "1"
		},
	})

	a, _ := m.Render([]byte(`
import HelloJSX from "./index"

## h2
![??? Golang ???????????????????????????](../../static/img/???%20Golang%20?????????????????????_1.png)
![??? Golang ???????????????????????????](/static/img/???%20Golang%20?????????????????????_1.png)

???[??????](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)??????????????????????????????????????????

<HelloJSX name="hi"></HelloJSX>
`))
	t.Logf("%s", a.Body)
}

func TestMdToJsx(t *testing.T) {
	f := memfs.New()
	file, _ := f.Create("index.mdx")
	file.Write([]byte(`
const A = ()=> <> AAA </>

## h1 <A></A>
`))

	m := NewGoldMdRender(GoldMdRenderOptions{
		AssetsUrlProcess: func(s string) string {
			return "1"
		},
	})

	a, err := m.Render([]byte(`
import HelloJSX from "./index"

<HelloJSX>
  <p></p>

## 123

</HelloJSX>
`))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", a.Body)

}
