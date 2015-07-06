package gonr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseMemValue(t *testing.T) {
	assert.Equal(t, int64(0), parseMemValue("foo bar")) // bad number
	assert.Equal(t, int64(0), parseMemValue("12"))      // no units
	assert.Equal(t, int64(13312), parseMemValue("13 kB"))
	assert.Equal(t, int64(14680064), parseMemValue("14 mB"))
	assert.Equal(t, int64(16106127360), parseMemValue("15 gB"))
}
