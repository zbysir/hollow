package log

import "testing"

func TestColor(t *testing.T) {
	//SetDev(true)
	buf := BuffSink{}
	l := New(Options{
		IsDev:         false,
		To:            &buf,
		DisableCaller: true,
		CallerSkip:    0,
		Name:          "",
	})
	l.Infof("%v", 1)

	t.Logf("buf: %s", buf.buf.String())
}
