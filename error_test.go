package utils

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestV(t *testing.T) {
	reader := strings.NewReader("Hello, Reader!")
	bytes := make([]byte, 8)
	for {
		n := V(reader.Read(bytes))
		assert.True(t, n >= 0)
	}
}
