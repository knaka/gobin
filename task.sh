#!/bin/sh
set -o nounset -o errexit

test "${guard_6ee3caf+set}" = set && return 0; guard_6ee3caf=-

if test "${1+SET}" = SET && test "$1" = "update-me"
then
  temp_dir_path="$(mktemp -d)"
  # shellcheck disable=SC2064
  trap "rm -fr \"$temp_dir_path\"" EXIT
  curl --fail --location --output "$temp_dir_path"/task_sh https://raw.githubusercontent.com/knaka/src/main/task.sh
  cat "$temp_dir_path"/task_sh > "$0"
  rm -fr "$temp_dir_path"
  exit 0
fi

script_dir_path="$(realpath "$(dirname "$0")")"

# --------------------------------------------------------------------------

is_windows() {
  case "$(uname -s)" in
    Windows_NT|CYGWIN*|MINGW*|MSYS*) return 0 ;;
    *) return 1 ;;
  esac
}

exe_ext() {
  if is_windows
  then
    echo ".exe"
  fi
}

is_bsd() {
  if stat -f "%z" . > /dev/null 2>&1
  then
    return 0
  fi
  return 1
  # case "$(uname -s)" in
  #   Darwin|FreeBSD|NetBSD|OpenBSD) return 0 ;;
  #   *) return 1 ;;
  # esac
}

set_path_attr() (
  path="$1"
  attribute="$2"
  value="$3"
  if which xattr > /dev/null 2>&1
  then
    xattr -w "$attribute" "$value" "$path"
  elif which PowerShell > /dev/null 2>&1
  then
    PowerShell -Command "Set-Content -Path '$path' -Stream '$attribute' -Value '$value'"
  elif which attr > /dev/null 2>&1
  then
    attr -s "$attribute" -V "$value" "$path"
  fi
)

set_dir_sync_ignored() (
  for path in "$@"
  do
    if test -d "$path"
    then
      continue
    fi
    mkdir -p "$path"
    for attribute in "com.dropbox.ignored" "com.apple.fileprovider.ignore#P"
    do
      set_path_attr "$path" "$attribute" 1
    done
  done
)

is_newer_than() {
  test -n "$(find "$1" -newer "$2"  2>/dev/null)" || return 1
}

newer() (
  found_than=false
  dest=
  for arg in "$@"
  do
    shift
    if test "$arg" = "--than"
    then
      found_than=true
    elif $found_than
    then
      dest="$arg"
    else
      set -- "$@" "$arg"
    fi
  done
  if test -z "$dest"
  then
    echo "No --than option" >&2
    exit 1
  fi
  if test "$#" -eq 0
  then
    echo "No source files" >&2
    exit 1
  fi
  # If the destination does not exist, it is considered newer than the destination.
  if ! test -e "$dest"
  then
    return 0
  fi
  # If the destination is a directory, the newest file in the directory is used.
  if test -d "$dest"
  then
    if is_bsd
    then
      dest="$(find "$dest" -type f -exec stat -l -t "%F %T" {} \+ | cut -d' ' -f6- | sort -n | tail -1 | cut -d' ' -f3)"
    else
      dest="$(find "$dest" -type f -exec stat -Lc '%Y %n' {} \+ | sort -n | tail -1 | cut -d' ' -f2)"
    fi
  fi
  if test -z "$dest"
  then
    return 0
  fi
  test -n "$(find "$@" -newer "$dest" 2> /dev/null)"
)

# Busybox sh seems to fail to detect proper executable if POSIX style one exists in the same directory.
cross_exec() {
  if type cleanup > /dev/null 2>&1
  then
    cleanup
  fi
  if ! is_windows
  then
    exec "$@"
  fi
  if ! test "${1+set}" = set
  then
    exit 1
  fi
  cmd_path="$1"
  shift
  if type "$cmd_path.exe" >/dev/null 2>&1
  then
    exec "$cmd_path.exe" "$@"
  fi
  if type "$cmd_path.cmd" >/dev/null 2>&1
  then
    exec "$cmd_path.cmd" "$@"
  fi
  if type "$cmd_path.bat" >/dev/null 2>&1
  then
    exec "$cmd_path.bat" "$@"
  fi
  exec "$cmd_path" "$@"
}

cross_run() (
  if ! is_windows
  then
    "$@"
    return $?
  fi
  cmd="$1"
  shift
  for ext in .exe .cmd .bat
  do
    if type "$cmd$ext" > /dev/null 2>&1
    then
      "$cmd$ext" "$@"
      return $?
    fi
  done
  "$cmd" "$@"
)

ensure_opt_arg() (
  if test -z "$2"
  then
    echo "No argument for option --$1." >&2
    usage
    exit 2
  fi
  echo "$2"
)

run_installed() (
  name=
  id=
  path=
  while getopts n:i:p:-: OPT
  do
    if test "$OPT" = "-"
    then
      OPT="${OPTARG%%=*}"
      # shellcheck disable=SC2030
      OPTARG="${OPTARG#"$OPT"}"
      OPTARG="${OPTARG#=}"
    fi
    case "$OPT" in
      n|name) name="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      i|winget-id) id="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      p|winget-path) path="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      \?) usage; exit 2;;
      *) echo "Unexpected option: $OPT" >&2; exit 2;;
    esac
  done
  shift $((OPTIND-1))

  if is_windows
  then
    if ! type "$path" > /dev/null 2>&1
    then
      winget install -e --id "$id"
    fi
    "$path" "$@"
    return $?
  fi
  if ! type "$name" > /dev/null 2>&1
  then
    echo "Install $name and try again." >&2
    exit 1
  fi
  "$name" "$@"
)

# --------------------------------------------------------------------------

task_subcmds() ( # List subcommands.
  cd "$script_dir_path" || exit 1
  delim=" delim_2ed1065 "
  # shellcheck disable=SC2086
  cnt="$(grep -E -h -e "^subcmd_[_[:alnum:]]+\(" $task_file_paths | sed -r -e 's/^subcmd_//' -e 's/^([^ ()]+)__/\1:/g' -e "s/\(\) *[{(] *(# *)?/$delim/")"
  if type delegate_tasks > /dev/null 2>&1
  then
    if delegate_tasks subcmds > /dev/null 2>&1
    then
      cnt="$(printf "%s\n%s" "$cnt" "$(delegate_tasks subcmds | sed -r -e "s/(^[^ ]+) +/\1$delim/")")"
    fi
  fi
  max_len="$(echo "$cnt" | awk '{ if (length($1) > max_len) max_len = length($1) } END { print max_len }')"
  echo "$cnt" | sort | awk -F"$delim" "{ printf \"%-${max_len}s  %s\n\", \$1, \$2 }"
)

task_tasks() { # List tasks.
  (
    delim=" delim_d3984dd "
    # shellcheck disable=SC2086
    cnt="$(grep -E -h -e "^task_[_[:alnum:]]+\(" $task_file_paths | sed -r -e 's/^task_//' -e 's/^([^ ()]+)__/\1:/g' -e "s/\(\) *[{(] *(# *)?/$delim/")"
    if type delegate_tasks > /dev/null 2>&1
    then
      if delegate_tasks tasks > /dev/null 2>&1
      then
        cnt="$(printf "%s\n%s" "$cnt" "$(delegate_tasks tasks | sed -r -e "s/(^[^ ]+) +/\1$delim/")")"
      fi
    fi
    max_len="$(echo "$cnt" | awk '{ if (length($1) > max_len) max_len = length($1) } END { print max_len }')"
    echo "$cnt" | sort | awk -F"$delim" "{ printf \"%-${max_len}s  %s\n\", \$1, \$2 }"
  )
}

usage() ( # Show help message.
  cd "$script_dir_path" || exit 1
  cat <<EOF
Usage:
  $0 [options] <subcommand> [args...]
  $0 [opttions] <task[arg1,arg2,...]> [tasks...]

Options:
  -d, --directory  Change directory before running tasks.
  -h, --help       Display this help and exit.
  -v, --verbose    Verbose mode.

Subcommands:
EOF
  task_subcmds | sed -r -e 's/^/  /'
  cat <<EOF

Tasks:
EOF
  task_tasks | sed -r -e 's/^/  /'
)

subcmd_pwd() {
  pwd "$@"
}

subcmd_false() { # Always fail.
  false "$@"
}

task_nop() { # Do nothing.
  echo NOP
}

main() {
  if test "${ARG0BASE+set}" = "set"
  then
    case "$ARG0BASE" in
      task-*)
        env="${ARG0BASE#task-}"
        case "$env" in
          dev|development)
            APP_ENV=development
            APP_SENV=dev
            ;;
          prd|production)
            APP_ENV=production
            APP_SENV=prd
            ;;
          *) echo "Unknown environment: $env" >&2; exit 1;;
        esac
        export APP_ENV APP_SENV
        ;;
      *)
        ;;
    esac
  fi

  task_file_paths="$(realpath "$0")"
  set -- "$PWD" "$@"; cd "$script_dir_path" || exit 1
  for file_path in "$script_dir_path"/task_*.sh "$script_dir_path"/task-*.sh
  do
    if ! test -r "$file_path"
    then
      continue
    fi
    case "$(basename "$file_path")" in
      task-dev.sh|task-prd.sh) continue;;
    esac
    task_file_paths="$task_file_paths $file_path"
    # shellcheck disable=SC1090
    . "$file_path"
  done
  cd "$1"; shift

  verbose=false
  shows_help=false
  directory=""
  while getopts d:hv-: OPT
  do
    if test "$OPT" = "-"
    then
      OPT="${OPTARG%%=*}"
      OPTARG="${OPTARG#"$OPT"}"
      OPTARG="${OPTARG#=}"
    fi
    case "$OPT" in
      d|directory) directory="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      h|help) shows_help=true;;
      v|verbose) verbose=true;;
      \?) usage; exit 2;;
      *) echo "Unexpected option: $OPT" >&2; exit 2;;
    esac
  done
  shift $((OPTIND-1))

  if test -n "$directory"
  then
    if $verbose; then echo "cd $directory" >&2; fi
    cd "$directory"
  fi

  if $shows_help || test "${1+set}" != "set"
  then
    usage
    exit 0
  fi

  subcmd="$1"
  subcmd="$(echo "$subcmd" | sed -r -e 's/:/__/g')"
  if type subcmd_"$subcmd" > /dev/null 2>&1
  then
    shift
    subcmd_"$subcmd" "$@" || exit $?
    exit 0
  fi

  for task_with_args in "$@"
  do
    task="$task_with_args"
    args=""
    case "$task_with_args" in
      *\[*)
        task="${task_with_args%%\[*}"
        args="$(echo "$task_with_args" | sed -r -e 's/^.*\[//' -e 's/\]$//' -e 's/,/ /')"
        ;;
    esac
    task="$(echo "$task" | sed -r -e 's/:/__/g')"
    if ! type task_"$task" > /dev/null 2>&1
    then
      if type delegate_tasks > /dev/null 2>&1
      then  
        delegate_tasks "$@"
        exit 0
      fi
      echo "Unknown task: $task" >&2
      exit 1
    fi
    # shellcheck disable=SC2086
    task_"$task" $args
  done
}

# To make this file can be sourced to provide functions.
if test "$(basename "$0")" = "task.sh"
then
  main "$@"
fi
