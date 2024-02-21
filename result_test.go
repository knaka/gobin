package utils

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type foo struct{}

func doSomething() (*foo, error) {
	return &foo{}, io.EOF
}

func TestNewResult(t *testing.T) {
	foo1, err := R((func() (*foo, error) {
		return &foo{}, io.EOF
	})()).NilIf(io.EOF)
	assert.Nil(t, err)
	assert.Nil(t, foo1)

	foo2, err := R((func() (*foo, error) {
		return &foo{}, nil
	})()).NilIf(io.EOF)
	assert.Nil(t, err)
	assert.NotNil(t, foo2)

	foo3, err := R((func() (*foo, error) {
		return &foo{}, io.EOF
	})()).NilIf(errors.New("bar"))
	assert.NotNil(t, err)
	assert.NotNil(t, foo3)
}
