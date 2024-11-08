#!/bin/sh

test "${guard_1e6bc22+set}" = set && return 0; guard_1e6bc22=-

. ./task.sh

# Returns the path to the Go root directory.
goroot_path() (
  # All releases - The Go Programming Language https://go.dev/dl/
  required_min_ver=go1.23.0

  # Search for Go SDK which fulfills the minimum version requirement.
  psv_go_dir_paths=
  # $GOROOT
  if test "${GOROOT+set}" = set
  then
    psv_go_dir_paths="$(array_append "$psv_go_dir_paths" "|" "$(realpath "$GOROOT")")"
  fi
  # `go` command
  if type go > /dev/null 2>&1
  then
    psv_go_dir_paths="$(array_append "$psv_go_dir_paths" "|" "$(realpath "$(go env GOROOT)")")"
  fi
  # System-wide installation
  if is_windows
  then
    psv_go_dir_paths="$(array_append "$psv_go_dir_paths" "|" "C:/Program Files/Go")"
  else
    psv_go_dir_paths="$(array_append "$psv_go_dir_paths" "|" "/usr/local/go")"
  fi
  # Automatically installed SDKs
  psv_sdk_go_dir_paths=
  for dir_path in "$HOME"/sdk/go*
  do
    if ! test -d "$dir_path"
    then
      continue
    fi
    psv_sdk_go_dir_paths="$(array_append "$psv_sdk_go_dir_paths" "|" "$dir_path")"
  done
  # Sort the SDKs in descending order.
  psv_sdk_go_dir_paths="$(array_sort "$psv_sdk_go_dir_paths" "|" sort_version -r)"
  psv_go_dir_paths="$(array_append "$psv_go_dir_paths" "|" "$psv_sdk_go_dir_paths")"
  # Find the first Go SDK which fulfills the minimum version requirement.
  set_ifs_pipe
  for go_dir_path in $psv_go_dir_paths
  do
    if type "$go_dir_path"/bin/go > /dev/null 2>&1 && version_ge "$("$go_dir_path"/bin/go env GOVERSION)" "$required_min_ver"
    then
      echo "$go_dir_path"
      return
    fi
  done
  restore_ifs

  # If Go SDK is not found, download and install it.
  sdk_dir_path="$HOME"/sdk
  goroot="$sdk_dir_path"/${required_min_ver}
  case "$(uname -s)" in
    Linux) goos=linux;;
    Darwin) goos=darwin;;
    Windows_NT) goos=windows;;
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
    zip_path="$(temp_dir_path)"/temp.zip
    curl --location -o "$zip_path" "https://go.dev/dl/$required_min_ver.$goos-$goarch.zip"
    (
      cd "$sdk_dir_path" || exit 1
      unzip -q "$zip_path" >&2
    )
  else
    curl --location -o - "https://go.dev/dl/$required_min_ver.$goos-$goarch.tar.gz" | (cd "$sdk_dir_path" || exit 1; tar -xzf -)
  fi
  mv "$sdk_dir_path"/go "$goroot"
  echo "$goroot"
)

# Sets the Go environment. If CGO is required, call `set_unixy_dev_env` also.
set_go_env() {
  first_call 1dc30dd || return 0
  unset GOROOT
  echo Using Go toolchain in "$(goroot_path)" >&2
  PATH="$(array_prepend "$PATH" : "$(goroot_path)"/bin)"
  export PATH
}

subcmd_go() { # Run go command.
  set_go_env
  go "$@"
}

subcmd_gofmt() { # Run gofmt command.
  set_go_env
  gofmt "$@"
}
