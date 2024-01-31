# gobin

The command-line tool `gobin` caches binaries when executing specified packages, thereby speeding up subsequent launches. It is primarily intended for use with Go generate (`//go:generate ...`).

## Features

- Installs and executes Go packages of specified versions.
- Saves executed binaries in the `.gobin` directory, which is located in the same directory as `go.mod` file.
- Uses symbolic links to keep track of which versions have been installed.

## Usage

Options before `--` are for building, and options after `--` are for running the package. You can omit `@...` version suffixes of the package if the module is listed in `go.mod` file.

```bash
$ go run github.com/knaka/gobin@latest golang.org/x/tools/cmd/stringer@v0.15.0 -- -help
Usage of stringer:
stringer [flags] -type T [directory]
stringer [flags] -type T files... # Must be a single package
...
```

You can use commands in Go generate without installing them globally in `$GOBIN` as follows:

```go
//go:generate -command stringer go run github.com/knaka/gobin@latest golang.org/x/tools/cmd/stringer@v0.15.0 --
//go:generate stringer -type Fruit .

//go:generate -command sqlc go run github.com/knaka/gobin@latest github.com/sqlc-dev/sqlc/cmd/sqlc@v1.22.0 --
//go:generate sqlc generate
//go:generate sqlc vet`
```

For “Go generate” dependencies between source and generated files, it is beneficial to use in combination with [go-generate-fast](https://github.com/oNaiPs/go-generate-fast).

```go
//go:generate -command sqlc go run github.com/knaka/gobin@latest github.com/sqlc-dev/sqlc/cmd/sqlc@v1.22.0 --

//go:generate_input ./sqlc.yaml ./schema*.sql ./migrations/*.sql
//go:generate_output ./sqlcgen/models.go
//go:generate sqlc generate
```
