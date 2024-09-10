#!/bin/sh
set -o nounset -o errexit

temp_dir=$(mktemp -d)
trap 'rm -rf $temp_dir' EXIT

name=$(basename "$0")
file_path="$temp_dir/main.go"
cat <<'EOF' > "$file_path"
embed_fce761e
EOF

go_cmd_path() {
  # Command in $PATH
  if type go >/dev/null 2>&1
  then
    which go
    return
  fi
  # Official installation paths
  for _dir_path in \
    /usr/local/go \
    "/Program Files/Go"
  do
    if type "$_dir_path"/bin/go > /dev/null 2>&1
    then
      echo "$_dir_path"/bin/go
      return
    fi
  done
  # User installation paths. Use the latest in case of multiple installations.
  _latest_dir_path=
  for _dir_path in \
    "$HOME"/go/go* \
    "$HOME"/sdk/go*
  do
    if ! test -d "$_dir_path"
    then
      continue
    fi
    if test -z "$_latest_dir_path"
    then
      _latest_dir_path=$_dir_path
      continue
    fi
    if expr "$(basename "$_latest_dir_path")" \> "$(basename "$_dir_path")" >/dev/null
    then
      _latest_dir_path=$_dir_path
      continue
    fi
  done
  if test -n "$_latest_dir_path"
  then
    echo "$_latest_dir_path"/bin/go
    return
  fi
  # Command in user $GOPATH
  if type "$HOME"/go/bin/go >/dev/null 2>&1
  then
    echo "$HOME"/go/bin/go
    return
  fi
  # If not found, download // All releases - The Go Programming Language https://go.dev/dl/
  _ver=1.23.1
  case "$(uname -s)" in
    Linux) _goos=linux;;
    Darwin) _goos=darwin;;
    *) exit 1;;
  esac
  case "$(uname -m)" in
    arm64) _goarch=arm64;;
    x86_64) _goarch=amd64;;
    *) exit 1;;
  esac
  _installation_dir_path="$HOME/sdk"
  mkdir -p "$_installation_dir_path"
  curl --location -o - "https://go.dev/dl/go$_ver.$_goos-$_goarch.tar.gz" | (cd "$_installation_dir_path" || exit 1; tar -xzf -; mv go go"$_ver")
  echo "$_installation_dir_path/go$_ver/bin/go"
}

gopath="${GOPATH:-$HOME/go}"
mkdir -p "$gopath"/bin
cmd_path="$gopath"/bin/embedded-"$name"
if test -z "$(find "$cmd_path" -newer "$0" 2>/dev/null)"
then
  "$(go_cmd_path)" build -o "$cmd_path" "$file_path"
fi

"$cmd_path" "$@"
