#!/bin/sh

test "${guard_1e6bc22+set}" = set && return 0; guard_1e6bc22=-

. task.sh

# gobin returns the path to the Go bin directory.
gobin() (
  # All releases - The Go Programming Language https://go.dev/dl/
  ver=1.23.1

  if test "${GOROOT+SET}" = "SET"
  then
    echo "$GOROOT"/bin
    return
  fi
  if which go > /dev/null 2>&1
  then
    echo "$(go env GOROOT)"/bin
    return
  fi
  for dir_path in \
    "$HOME"/sdk/go${ver} \
    /usr/local/go \
    /usr/local/Cellar/go/* \
    /"Program Files"/Go \
    "$HOME"/go
  do
    if type "$dir_path"/bin/go > /dev/null 2>&1
    then
      echo "$dir_path"/bin
      return
    fi
  done
  sdk_dir_path="$HOME"/sdk
  goroot="$sdk_dir_path"/go${ver}
  case "$(uname -s)" in
    Linux) _goos=linux;;
    Darwin) _goos=darwin;;
    Windows_NT) _goos=windows;;
    *)
      echo "Unsupported OS: $(uname -s)" >&2
      exit 1;;
  esac
  case "$(uname -m)" in
    arm64) goarch=arm64;;
    x86_64) goarch=amd64;;
    *)
      echo "Unsupported architecture: $(uname -m)" >&2
      exit 1;;
  esac
  mkdir -p "$sdk_dir_path"
  if is_windows
  then
    _temp_dir_path=$(mktemp -d)
    zip_path="$_temp_dir_path"/temp.zip
    curl --location -o "$zip_path" "https://go.dev/dl/go$ver.$_goos-$goarch.zip"
    (cd "$sdk_dir_path" || exit 1; unzip -q "$zip_path" >&2)
    rm -fr "$_temp_dir_path"
  else
    curl --location -o - "https://go.dev/dl/go$ver.$_goos-$goarch.tar.gz" | (cd "$sdk_dir_path" || exit 1; tar -xzf -)
  fi
  mv "$sdk_dir_path"/go "$goroot"
  echo "$goroot"/bin
)

subcmd_go() { # Run go command.
  go_cmd_path="$(gobin)"/go
  "$go_cmd_path" "$@"
}
