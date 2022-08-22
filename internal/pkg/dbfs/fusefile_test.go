package dbfs

import "testing"

func TestAttr(t *testing.T) {
	a := Attr{
		MTime: 1,
		CTime: 2,
		Size:  3,
		Mode:  4,
	}
	bs := a.toByte()

	var b Attr
	b.fromByte(bs)
	t.Logf("%s", bs)
	t.Logf("%+v", b)
}
