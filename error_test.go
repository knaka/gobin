package utils

import (
	"github.com/stretchr/testify/assert"
	"io"
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

func TestExpect(t *testing.T) {
	Expect((func() error {
		return nil
	})(), nil, io.EOF)
	Expect((func() error {
		return io.EOF
	})(), nil, io.EOF)
	//Expect((func() error {
	//	return io.EOF
	//})(), nil, io.ErrClosedPipe)
}
