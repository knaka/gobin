# gobin

コマンドラインツール `gobin` は、指定されたパッケージを実行すると同時に、そのバイナリをキャッシュすることで、次回以降の起動を高速化します。主に、Go generate (`//go:generate ...`) での利用を想定しています。

## Features

- 指定されたバージョンの Go パッケージをインストールし、実行します。
- 実際されたバイナリを、ローカルの `.gobin` ディレクトリに保存します。
- シンボリックリンクを利用して、どのバージョンがインストールされたかを記録します。

## Usage

`--` より前は、ビルドオプションです。 `--` より後は、パッケージを実行する際のオプションです。もしモジュールが `go.mod` ファイルにリストされているならば、バージョンサフィックス `@...` を省略することもできます。

```
$ go run github.com/knaka/gobin@latest golang.org/x/tools/cmd/stringer@v0.15.0 -- -help
Usage of stringer:
        stringer [flags] -type T [directory]
        stringer [flags] -type T files... # Must be a single package
...
```

グローバルの `$GOBIN` にコマンドをインストールしなくても、以下のように Go generate においてコマンドを利用できます。

```go
//go:generate -command stringer go run github.com/knaka/gobin@latest golang.org/x/tools/cmd/stringer@v0.15.0 --
//go:generate stringer -type Fruit .

//go:generate -command sqlc go run github.com/knaka/gobin@latest github.com/sqlc-dev/sqlc/cmd/sqlc@v1.22.0 --
//go:generate sqlc generate
//go:generate sqlc vet
```

Go generate のファイル間の依存関係については、 [go-generate-fast](https://github.com/oNaiPs/go-generate-fast) と組み合わせて利用すると有用です。

```go
//go:generate -command sqlc go run github.com/knaka/gobin@latest github.com/sqlc-dev/sqlc/cmd/sqlc@v1.22.0 --

//go:generate_input ./sqlc.yaml ./schema*.sql ./migrations/*.sql
//go:generate_output ./sqlcgen/models.go
//go:generate sqlc generate
```
