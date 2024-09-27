package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	. "github.com/knaka/go-utils"
)

//go:embed embedded.sh
var embeddedSh string

//go:embed embedded.cmd
var embeddedCmd string

var rePackage = sync.OnceValue(func() (re *regexp.Regexp) {
	return regexp.MustCompile(`^package (\w+)`)
})

func main() {
	reader := V(os.Open(filepath.Join("minlib", "minlib.go")))
	defer (func() { V0(reader.Close()) })()
	scanner := bufio.NewScanner(reader)
	writer := strings.Builder{}
	//defer (func() { Ignore(writer.Close()) })()
	V0(writer.WriteString(fmt.Sprintf(`// Code generated by gen-bootstrap; DO NOT EDIT.

// Latest version is available by running:
//
//   curl --remote-name https://raw.githubusercontent.com/knaka/gobin/main/bootstrap/gobin.go

//go:build ignore

`)))
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
	bootstrapMain()
}
`)))
	//V0(writer.Close)

	bootstrapGoPath := filepath.Join("bootstrap", "cmd-gobin.go")
	bootstrapGo := V(os.Create(bootstrapGoPath))
	defer (func() { V0(bootstrapGo.Close()) })()
	V0(bootstrapGo.WriteString(writer.String()))

	embeddedSh = strings.Replace(embeddedSh, "embed_fce761e", writer.String(), 1)
	embeddedSh = strings.Replace(embeddedSh, "url_a8e2423", "https://raw.githubusercontent.com/knaka/gobin/refs/heads/main/bootstrap/cmd-gobin", 1)
	bootstrapShPath := filepath.Join("bootstrap", "cmd-gobin")
	bootstrapSh := V(os.Create(bootstrapShPath))
	defer (func() { V0(bootstrapSh.Close()) })()
	V0(bootstrapSh.WriteString(embeddedSh))
	V0(os.Chmod(bootstrapShPath, 0755))

	embeddedCmd = strings.Replace(embeddedCmd, "embed_fce761e", writer.String(), 1)
	embeddedCmd = strings.Replace(embeddedCmd, "url_935916d", "https://raw.githubusercontent.com/knaka/gobin/refs/heads/main/bootstrap/cmd-gobin.cmd", 1)
	bootstrapCmdPath := filepath.Join("bootstrap", "cmd-gobin.cmd")
	bootstrapCmd := V(os.Create(bootstrapCmdPath))
	defer (func() { V0(bootstrapCmd.Close()) })()
	V0(bootstrapCmd.WriteString(embeddedCmd))
}
