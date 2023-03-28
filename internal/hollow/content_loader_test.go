package hollow

import (
	"encoding/json"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"testing"
)

func TestNewMDLoader(t *testing.T) {
	f := memfs.New()
	file, _ := f.Create("index.mdx")
	file.Write([]byte(`---
title: "title"
---

const A = ()=> <> AAA </>

# h1 <A></A>

![](../../statics/img/img.jpg)

![](img.jpg)

`))

	jx, err := jsx.NewJsx(jsx.Option{
		SourceCache: nil,
		Debug:       false,
		Fs:          gobilly.NewStdFs(f),
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	m := NewMDLoader(Assets{"statics"}, jx, nil)

	c, err := m.Load("index.mdx", false)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", c.Assets)
	t.Logf("%+v", c.Content)
}

func TestToc(t *testing.T) {
	f := memfs.New()
	file, _ := f.Create("index.mdx")
	file.Write([]byte(`---
title: "title"
---

const A = ()=> <> AAA </>
const contents = [{meta: {title:"h2"}}]

# h1 <A></A>


<>
  <div x-data="{ selected: ''}">
    {1}
    {contents.map(i => {
      return <div
      >
        <h2 id={i.meta?.hash || i.meta?.title}>{i.meta?.title}</h2>
        <div dangerouslySetInnerHTML={{__html: i.content}}></div>
      </div>
    })}
  </div>
</>
`))

	jx, err := jsx.NewJsx(jsx.Option{
		SourceCache: nil,
		Debug:       false,
		Fs:          gobilly.NewStdFs(f),
	})

	if err != nil {
		t.Fatal(err)
		return
	}

	m := NewMDLoader(Assets{"statics"}, jx, nil)

	c, err := m.Load("index.mdx", false)
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.Marshal(c.Toc)
	assert.Equal(t, `[{"title":"h1  AAA","items":[{"title":"h2","id":"h2"}],"id":"h1-aa"}]`, string(bs))
}

func TestToc2(t *testing.T) {
	d := jsx.VDom{
		"nodeName": "div",
		"attributes": map[string]interface{}{
			"children": []interface{}{

				map[string]interface{}{
					"nodeName": "h3",
					"attributes": map[string]interface{}{
						"children": "h3",
					},
				},
				map[string]interface{}{
					"nodeName": "h3",
					"attributes": map[string]interface{}{
						"children": "h3",
					},
				},
				map[string]interface{}{
					"nodeName": "h1",
					"attributes": map[string]interface{}{
						"children": "h1",
					},
				},
				map[string]interface{}{
					"nodeName": "h2",
					"attributes": map[string]interface{}{
						"children": "h2",
					},
				},

				map[string]interface{}{
					"nodeName": "h4",
					"attributes": map[string]interface{}{
						"children": "h4",
					},
				},
				map[string]interface{}{
					"nodeName": "h3",
					"attributes": map[string]interface{}{
						"children": "h3",
					},
				},
				map[string]interface{}{
					"nodeName": "h3",
					"attributes": map[string]interface{}{
						"children": "h3",
					},
				},
				map[string]interface{}{
					"nodeName": "h4",
					"attributes": map[string]interface{}{
						"children": "h4",
					},
				},
				map[string]interface{}{
					"nodeName": "h2",
					"attributes": map[string]interface{}{
						"children": "h2",
					},
				},
				map[string]interface{}{
					"nodeName": "h1",
					"attributes": map[string]interface{}{
						"children": "h1",
					},
				},
			},
		},
	}
	toc := generateTOC(d)
	str := toc.Dump(0)
	assert.Equal(t, `
  h3
  h3
  h1
    h2
      h4
      h3
      h3
        h4
    h2
  h1
`, str)
}
