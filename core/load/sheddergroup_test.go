package load

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	group := NewShedderGroup()
	t.Run("get", func(t *testing.T) {
		limiter := group.GetShedder("test")
		assert.NotNil(t, limiter)
	})
}

func TestShedderClose(t *testing.T) {
	var nop nopCloser
	assert.Nil(t, nop.Close())
}
