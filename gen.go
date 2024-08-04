package main

// go:generate go run ./gen-bootstrap/

//go:generate echo foo

// Generate from module definition.
//go:generate_input gen-from-mod/*.go go.mod **/*.gen-from-mod.go.tmpl
//go:generate_output *.gen-from-mod.go */*.gen-from-mod.go
//go:generate cd gen-from-mod && echo doing && go run .
