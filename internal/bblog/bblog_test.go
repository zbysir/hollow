package bblog

import "testing"

func TestSource(t *testing.T) {
	b, err := NewBblog(Option{})
	if err != nil {
		t.Fatal(err)
	}

	x := b.getSource("../../blogs")
	bs := x.([]*Blog)
	for _, b := range bs {
		t.Logf("%+v", b)
	}
}

func TestLoad(t *testing.T) {
	b, err := NewBblog(Option{})
	if err != nil {
		t.Fatal(err)
	}

	c, err := b.Load("./src/config.ts", ExecOption{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", c)
}
