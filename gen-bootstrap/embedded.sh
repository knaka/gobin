#!/bin/sh
set -o nounset -o errexit

temp_dir=$(mktemp -d)
trap 'rm -rf $temp_dir' EXIT

if test "${1+SET}" = SET && test "$1" = "update-me"
then
  # shellcheck disable=SC2064
  curl --fail --location --output "$temp_dir"/27bf4b5 url_a8e2423
  cat "$temp_dir"/27bf4b5 > "$0"
  exit 0
fi

name=$(realpath "$0" | tr / _)
file_path="$temp_dir/main.go"
cat <<'EOF' > "$file_path"
embed_fce761e
EOF

# gobin returns the path to the Go bin directory.
gobin() (
  # All releases - The Go Programming Language https://go.dev/dl/
  ver=1.22.7

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

gopath="${GOPATH:-$HOME/go}"
mkdir -p "$gopath"/bin
cmd_path="$gopath"/bin/embedded-"$name"
if test -z "$(find "$cmd_path" -newer "$0" 2>/dev/null)"
then
  "$(gobin)"/go build -o "$cmd_path" "$file_path"
fi

"$cmd_path" "$@"
