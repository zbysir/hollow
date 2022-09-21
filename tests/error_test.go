package tests

import (
	"errors"
	"fmt"
	"testing"
)

type LoginError struct {
	Count int
	i     error
}

func (l LoginError) Error() string {
	return "login error"
}

func (l LoginError) Unwrap() error {
	return l.i
}

func (l LoginError) Is(err error) bool {
	return err == ErrLogin
}

func (l LoginError) WithCount(count int) error {
	return LoginError{
		i:     l,
		Count: count,
	}
}

var ErrLogin = LoginError{}

var ErrLoginE = LoginError{i: ErrLogin}

func TestIsError(t *testing.T) {
	e1 := ErrLogin
	e2 := fmt.Errorf("%w count: %v", ErrLogin, 1)
	e3 := LoginError{Count: 3}

	t.Logf("%+v", errors.Is(e1, ErrLogin))
	t.Logf("%+v", errors.Is(e2, ErrLogin))
	t.Logf("%+v", errors.Is(e3, ErrLogin))
	{
		var e LoginError
		t.Logf("%+v %v", errors.As(e2, &e), e.Count)
	}

	{
		var e LoginError
		t.Logf("%+v %v", errors.As(e3, &e), e.Count)
	}

	t.Run("with", func(t *testing.T) {
		e1 := ErrLoginE.WithCount(1)
		t.Logf("%+v", errors.Is(e1, ErrLogin))

		{
			var e LoginError
			t.Logf("%+v %v", errors.As(e1, &e), e.Count)
		}
	})
}
