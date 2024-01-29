# gobin

`gobin` is a command-line tool written in Go for easily installing and managing specific versions of Go language binaries. This tool is particularly useful for versioning dependencies in Go projects.

## Features

- Install Go packages with specified versions.
- Save binaries in a local `.gobin` directory.
- Utilize symbolic links for easy access to specific versioned binaries.

## Usage

```go
//go:generate go run github.com/knaka/gobin golang.org/x/tools/cmd/stringer@v0.15.0
//go:generate .gobin/stringer

//go:generate go run github.com/knaka/gobin github.com/sqlc-dev/sqlc/cmd/sqlc@v1.22.0
//go:generate .gobin/sqlc generate
```
