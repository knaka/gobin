# go-run-cache

The command-line tool `go-run-cache` caches binaries when running a specified "main" package or local, non-module-aware `.go` files, thereby speeding up subsequent launches.

## Features

- Installs and executes Go packages of specified versions.
- Caches executed binaries for reuse. Packages without versions specification are cached in the same directory as `go.mod` file.
- Records not only the binaries but also the build options used which include the compiler version, the source hashes, the build tags and other build arguments. This ensures that the appropriate binary is reused even when the same package is built with different build options.
- Automatically rebuilds and does not use the cache for old binaries or binaries built with the `@latest` version specification.

## Usage

To install:

```bash
go install github.com/knaka/go-run-cache@latest
```

Executed packages are cached for later reuse. You can omit the `@...` version suffix for packages listed in the `go.mod` file. In that case, the binaries are cached in the same directory as the `go.mod` file.

```console
$ go-run-cache golang.org/x/tools/cmd/stringer@v0.15.0 -help
Usage of stringer:
stringer [flags] -type T [directory]
stringer [flags] -type T files... # Must be a single package
...
$ 
```

It is also possible to use `go-run-cache` as a shebang line in a `.go` file to cache the binary. Inserting `--` is desirable to distinguish build targets from arguments for the command.

```console
$ cat <<EOF > hello.go
#!/usr/bin/env go-run-cache -- $0 $@; exit
package main

import "fmt"

func main() {
    fmt.Println("Hello World!")
}
EOF
chmod +x hello.go
$ ./hello.go
Hello World!
$
```

You can use commands in Go generate without installing them globally in `$GOBIN` as follows:

```go
package foo

//go:generate -command stringer go run github.com/knaka/gobin@latest golang.org/x/tools/cmd/stringer@v0.15.0
//go:generate stringer -type Fruit .

//go:generate -command sqlc go run github.com/knaka/gobin@latest github.com/sqlc-dev/sqlc/cmd/sqlc@v1.22.0
//go:generate sqlc generate
//go:generate sqlc vet
```

For managing "Go generate" dependencies between source and generated files, it is beneficial to use in combination with [go-generate-fast](https://github.com/oNaiPs/go-generate-fast).

```go
package foo

//go:generate -command sqlc go run github.com/knaka/go-run-cache@latest github.com/sqlc-dev/sqlc/cmd/sqlc@v1.22.0

//go:generate_input ./sqlc.yaml ./schema*.sql ./migrations/*.sql
//go:generate_output ./sqlcgen/models.go
//go:generate sqlc generate
```

# Todo

* [ ] Add tests // [Add tests · Issue #1 · knaka/gobin](https://github.com/knaka/gobin/issues/1)
* [ ] Add `install` subcommand // [Add `install` subcommand · Issue #3 · knaka/gobin](https://github.com/knaka/gobin/issues/3)
* [ ] Add `run` subcommand // [Add `run` subcommand · Issue #4 · knaka/gobin](https://github.com/knaka/gobin/issues/4)
* [ ] Add local installation feature // [Add local installation feature · Issue #5 · knaka/gobin](https://github.com/knaka/gobin/issues/5)
