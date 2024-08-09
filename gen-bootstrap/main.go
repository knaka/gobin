package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	. "github.com/knaka/go-utils"
)

var rePackage = sync.OnceValue(func() (re *regexp.Regexp) {
	return regexp.MustCompile(`^package (\w+)`)
})

func main() {
	for _, subCmd := range []string{"run", "install"} {
		(func() {
			reader := V(os.Open(filepath.Join("minlib", "minlib.go")))
			defer (func() { V0(reader.Close()) })()
			scanner := bufio.NewScanner(reader)
			writer := V(os.Create(filepath.Join("gobin-" + subCmd + ".go")))
			defer (func() { V0(writer.Close()) })()
			V0(writer.WriteString(fmt.Sprintf(`// Code generated by gen-bootstrap; DO NOT EDIT.

// Latest version is available from https://raw.githubusercontent.com/knaka/gobin/main/gobin-%s.go

//go:build ignore

`, subCmd)))
			first := true
			for scanner.Scan() {
				line := scanner.Text()
				if first && rePackage().MatchString(line) {
					line = "package main"
					first = false
				}
				V0(writer.WriteString(line + "\n"))
			}
			V0(writer.WriteString(fmt.Sprintf(`
func main() {
	%s()
}
`, subCmd)))
		})()
	}
}
