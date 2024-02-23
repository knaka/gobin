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
	foo1 := PR((func() (*foo, error) {
		return &foo{}, io.EOF
	})()).NilIf(io.EOF)
	assert.Nil(t, foo1)

	foo2 := PR((func() (*foo, error) { return &foo{}, nil })()).NilIf(io.EOF)
	assert.NotNil(t, foo2)
}

func testCatch() (err error) {
	defer Catch(&err)
	Ignore(PR((func() (*foo, error) { return &foo{}, io.EOF })()).NilIf(errors.New("bar")))
	return nil
}

func TestCatch(t *testing.T) {
	err := testCatch()
	assert.NotNil(t, err)
}
