package hollow

import (
	"fmt"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/zbysir/hollow/internal/pkg/db"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"sync"
	"testing"
)

func TestSource(t *testing.T) {
	b, err := NewHollow(Option{
		SourceFs: osfs.New("/Users/bysir/goproj/bysir/zbysir.github.io"),
	})
	if err != nil {
		t.Fatal(err)
	}

	as := b.getContents(NewRenderContext())("./contents", getBlogOption{
		Tree: true,
	})

	for _, b := range as.List {
		t.Logf("%+v %v", b.Name, b.IsDir)
		for _, b := range b.Children {
			t.Logf("\t\t%+v %v", b.Name, b.IsDir)
		}
	}

	//bs, err := json.MarshalIndent(as.List, " ", " ")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//t.Logf("%s", bs)
}

func TestSourceParallel(t *testing.T) {
	b, err := NewHollow(Option{
		SourceFs: osfs.New("/Users/bysir/goproj/bysir/zbysir.github.io"),
	})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	context := NewRenderContext()
	contents := b.getContents(context)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			contents("./contents", getBlogOption{
				Tree: false,
			})
		}()
	}
	wg.Wait()
}

func TestBuiltin(t *testing.T) {
	b, err := NewHollow(Option{
		SourceFs: osfs.New("./testdata"),
	})
	if err != nil {
		t.Fatal(err)
	}

	context := NewRenderContext()
	contents := b.builtinAssert(context)

	s := contents("./b/b1.mdx")
	t.Logf("%s", s)
}

func TestMd(t *testing.T) {
	b, err := NewHollow(Option{
		SourceFs: osfs.New("./testdata"),
	})
	if err != nil {
		t.Fatal(err)
	}

	as := b.md(NewRenderContext())("# h1", MdOptions{})
	t.Logf("%+v", as)
	as = b.md(NewRenderContext())("123", MdOptions{})
	t.Logf("%+v", as)
}

func TestBuildFromFs(t *testing.T) {
	d, err := db.NewKvDb("./editor/database")
	if err != nil {
		t.Fatal(err)
	}

	st2, err := d.Open(fmt.Sprintf("project_1"), "project")
	if err != nil {
		t.Fatal(err)
	}
	fs := gobilly.NewDbFs(st2)

	b, err := NewHollow(Option{
		SourceFs: fs,
	})

	err = b.Build(NewRenderContext(), "docs", ExecOption{
		//Out: &WsSink{hub: hub},
	})
	if err != nil {
		t.Fatal(err)
		return
	}
}
