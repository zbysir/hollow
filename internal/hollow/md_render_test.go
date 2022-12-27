package hollow

import (
	"github.com/go-git/go-billy/v5/memfs"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"testing"
)

func TestMdRender(t *testing.T) {
	m := newMdRender(func(s string) string {
		return s
	})
	a := m.Render([]byte(`## h2
![在 Golang 中尝试“干净架构”](../../static/img/在%20Golang%20中尝试干净架构_1.png)
![在 Golang 中尝试“干净架构”](/static/img/在%20Golang%20中尝试干净架构_1.png)

在[文中](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)，他提出的干净架构是这样的：

<HelloJSX></HelloJSX>
`))
	t.Logf("%s", a)
}

func TestMDXRender(t *testing.T) {
	m := newMdRender(func(s string) string {
		return s
	})
	a := m.Render([]byte(`
<div>
  <h1></h1>
</div>
<HelloJSX id={123} style={{top: 2}}> 
  <div></div>
</HelloJSX>
`))
	t.Logf("%s", a)

	f := memfs.New()
	indexFile, err := f.Create("index.jsx")
	if err != nil {
		t.Fatal(err)
	}
	_, err = indexFile.Write(a)
	if err != nil {
		t.Fatal(err)
	}
	jx, err := jsx.NewJsx(jsx.Option{
		SourceCache: nil,
		Debug:       false,
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	n, err := jx.Render("./index", nil, jsx.WithRenderFs(gobilly.NewStdFs(f)))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%v", n)
}
