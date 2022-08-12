package log

import "testing"

func TestColor(t *testing.T) {
	SetDev(true)
	Infof("%v", 1)
}
