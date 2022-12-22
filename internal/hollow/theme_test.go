package hollow

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolveGitUrl(t *testing.T) {
	{
		r, b, s, err := resolveGitUrl("https://github.com/zbysir/hollow-theme/tree/master/hollow")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "https://github.com/zbysir/hollow-theme", r)
		assert.Equal(t, "master", b)
		assert.Equal(t, "hollow", s)
	}
	{
		r, b, s, err := resolveGitUrl("https://github.com/zbysir/hollow-theme/tree/master")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "https://github.com/zbysir/hollow-theme", r)
		assert.Equal(t, "master", b)
		assert.Equal(t, "", s)
	}
	{
		_, _, _, err := resolveGitUrl("https://github.com/zbysir/hollow-theme")
		assert.Error(t, err)
	}
}
