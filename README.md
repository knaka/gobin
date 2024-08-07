# gobin

The command-line tool `gobin` installs the binaries in `~/go/bin` or project-local directories according to the versions specified in the package list file. It is useful for managing Go tools that are not installed globally in `$GOBIN` .

## Features

- Installs and executes Go program packages of the versions specified in `go.mod` file.
- Installs and executes Go packages of the versions specified in package manifest file.
- Caches executed binaries for reuse.
- Cached binaries can be executed even if the environment is offline.

## Usage

To install:

```bash
go install github.com/knaka/cmd/gobin@latest
```

Executed packages are cached for later reuse. You can omit the `@...` version suffix for packages listed in the `go.mod` file. In that case, the binaries are cached in the same directory as the `go.mod` file.

```console
$ cat Gobinfile
golang.org/x/tools/cmd/goyacc@latest
github.com/hairyhenderson/gomplate/v4/cmd/gomplate@latest
$ gobin run goyacc --help # Binaries are stored in the same directory as the Gobinfile 
Usage of ...
$
```

You can use commands in Go generate without installing them globally in `$GOBIN` as follows:

```console
$ curl https://raw.githubusercontent.com/knaka/gobin/main/gobin-run.go -o gobin-run.go
$ echo golang.org/x/tools/cmd/stringer@v0.11 >> Gobinfile
$ echo github.com/sqlc-dev/sqlc/cmd/sqlc@latest >> Gobinfile
```

You can write additional information to `Gobinfile` entries as follows:

```text
github.com/sqlc-dev/sqlc/cmd/sqlc@latest tags=foo,bar requires=command1,command2 # comment. “tags” for build tags, “requires” for the commands required to run the command.
```

Or record the module of the program package to `go.mod` file as described in “[Go Wiki: Go Modules - The Go Programming Language](https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)”:

```go
//go:build tools
package main

import _ "golang.org/x/tools/cmd/stringer"
```

then, add the following to the source code:

```go
package foo

//go:generate -command stringer go run gobin-run.go golang.org/x/tools/cmd/stringer
//go:generate stringer -type Fruit .

//go:generate -command sqlc go run gobin-run.go sqlc
//go:generate sqlc generate
```

For managing "Go generate" dependencies between source and generated files, it is beneficial to use in combination with [go-generate-fast](https://github.com/oNaiPs/go-generate-fast).

```go
package foo

//go:generate -command sqlc go run gobin-run.go sqlc

//go:generate_input ./sqlc.yaml ./schema*.sql ./migrations/*.sql
//go:generate_output ./sqlcgen/models.go
//go:generate sqlc generate
```

## Usage as library

You can use `gobin` as a library mainly in task-runner written in Go.

```go
TBD
```
