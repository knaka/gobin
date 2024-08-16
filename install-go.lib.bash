#!/bin/bash

ver=${1:-1.23.0}

while true
do
  if type go >/dev/null 2>&1
  then
    break
  fi
  for path in /usr/local/go/bin \
    /usr/local/bin \
    /usr/bin \
    "$HOME"/go/bin
  do
    if test -x "$path/go"
    then
      PATH="$path:$PATH"
      break 2
    fi
  done
  case "$(uname -s)" in
    Linux*) goos=linux;;
    Darwin*) goos=darwin;;
    *) exit 1;;
  esac
  case "$(uname -m)" in
    arm64*) goarch=arm64;;
    x86_64*) goarch=amd64;;
    *) exit 1;;
  esac
  curl --location - "https://go.dev/dl/go$ver.$goos-$goarch.tar.gz" | tar -C "$PWD"/.sdk -xzf -
  PATH="$(pwd)/.sdk/go/bin:$PATH"
done
export PATH
