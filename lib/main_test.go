package lib

import "testing"

func TestRun(t *testing.T) {
	err := Run("golang.org/x/tools/cmd/stringer@v0.15.0", "-help")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
