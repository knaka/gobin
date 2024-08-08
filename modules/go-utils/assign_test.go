package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssign(t *testing.T) {
	n := 123
	for i := 0; i < 10; i++ {
		nSave := n
		assert.Equal(t, Assign(&n, n+1), nSave+1)
	}
	assert.Equal(t, n, 123+10)
}
