package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type S struct {
	Foo string
	Bar time.Time
}

func TestElvis(t *testing.T) {
	type args[T comparable] struct {
		t T
		f T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want T
	}
	testsString := []testCase[string]{
		{
			name: "Test Str1",
			args: args[string]{
				t: "a",
				f: "b",
			},
			want: "a",
		},
	}
	for _, tt := range testsString {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Elvis(tt.args.t, tt.args.f), "Elvis(%v, %v)", tt.args.t, tt.args.f)
		})
	}
	value := S{Foo: "a", Bar: time.Now()}
	testsPointer := []testCase[*S]{
		{
			name: "Test Ptr1",
			args: args[*S]{
				t: &value,
				f: &S{Foo: "c", Bar: time.Time{}},
			},
			want: &value,
		},
	}
	zeroValue := S{}
	value2 := S{Foo: "c", Bar: V(time.Parse(time.RFC3339, "2021-01-01T00:00:00Z"))}
	for _, tt := range testsPointer {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Elvis(tt.args.t, tt.args.f), "Elvis(%v, %v)", tt.args.t, tt.args.f)
		})
	}
	testsValue := []testCase[S]{
		{
			name: "Test Val1",
			args: args[S]{
				t: value,
				f: value2,
			},
			want: value,
		},
		{
			name: "Test Val2",
			args: args[S]{
				t: zeroValue,
				f: value2,
			},
			want: value2,
		},
	}
	for _, tt := range testsValue {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Elvis(tt.args.t, tt.args.f), "Elvis(%v, %v)", tt.args.t, tt.args.f)
		})
	}
}
