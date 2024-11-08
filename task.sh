#!/bin/sh
set -o nounset -o errexit

test "${guard_6ee3caf+set}" = set && return 0; guard_6ee3caf=x

# Update the script by replacing itself with the latest version.
if test "${1+SET}" = SET && test "$1" = "update-me"
then
  temp_dir_path_5de91af="$(mktemp -d)"
  # shellcheck disable=SC2317
  cleanup_7f0c4de() { rm -fr "$temp_dir_path_5de91af"; }
  trap cleanup_7f0c4de EXIT
  curl --fail --location --output "$temp_dir_path_5de91af"/task_sh https://raw.githubusercontent.com/knaka/src/main/task.sh
  cat "$temp_dir_path_5de91af"/task_sh > "$0"
  exit 0
fi

# --------------------------------------------------------------------------

unset shell_flags_c225b8f

# Backup the current shell flags.
backup_shell_flags() {
  if test "${shell_flags_c225b8f+set}" = set
  then
    # Fails to save the state if it was already saved. Does not nest.
    return 1
  fi
  shell_flags_c225b8f="$(set +o)"
}

# Restore the shell flags saved by `backup_shell_flags`.
restore_shell_flags() {
  if ! test "${shell_flags_c225b8f+set}" = set
  then
    # Fails to restore the state if it was not saved.
    return 1
  fi
  eval "$shell_flags_c225b8f"
  unset shell_flags_c225b8f
}

is_windows() {
  case "$(uname -s)" in
    Windows_NT|CYGWIN*|MINGW*|MSYS*) return 0 ;;
    *) return 1 ;;
  esac
}

# Executable file extension.
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
}

is_darwin() {
  if test "$(uname -s)" = "Darwin"
  then
    return 0
  fi
  return 1
}

is_mac() {
  is_darwin
}

# Set the extra attributes of the file/directory.
set_path_attr() (
  path="$1"
  attribute="$2"
  value="$3"
  if which xattr > /dev/null 2>&1
  then
    xattr -w "$attribute" "$value" "$path"
  elif which PowerShell > /dev/null 2>&1
  then
    # Run in the background because it takes much time to run.
    PowerShell -Command "Set-Content -Path '$path' -Stream '$attribute' -Value '$value'" &
  elif which attr > /dev/null 2>&1
  then
    attr -s "$attribute" -V "$value" "$path"
  else
    echo "No command to set attribute: $attribute" >&2
    exit 1
  fi
)

readonly psv_file_sharing_ignorance_attributes="com.dropbox.ignored|com.apple.fileprovider.ignore#P"

# Set the file/directory to be ignored by file sharing such as Dropbox.
set_sync_ignored() (
  for path in "$@"
  do
    if ! test -e "$path"
    then
      continue
    fi
    set_ifs_pipe
    for file_sharing_ignorance_attribute in $psv_file_sharing_ignorance_attributes
    do
      set_path_attr "$path" "$file_sharing_ignorance_attribute" 1
    done
    restore_ifs
  done
)

# Create a directory and set it to be ignored by file sharing such as Dropbox.
mkdir_sync_ignored() (
  for path in "$@"
  do
    if test -d "$path"
    then
      continue
    fi
    mkdir -p "$path"
    set_ifs_pipe
    for attribute in $psv_file_sharing_ignorance_attributes
    do
      set_path_attr "$path" "$attribute" 1
    done
    restore_ifs
  done
)

# Set the file/directory to be ignored by file sharing such as Dropbox.
force_sync_ignored() (
  for path in "$@"
  do
    set_ifs_pipe
    for attribute in $psv_file_sharing_ignorance_attributes
    do
      set_path_attr "$path" "$attribute" 1
    done
    restore_ifs
  done
)

# Check if the file(s)/directory(s) is/are newer than the destination.
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
    echo "Destination does not exist: $dest" >&2
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
    echo "No destination file" >&2
    return 0
  fi
  test -n "$(find "$@" -newer "$dest" 2> /dev/null)"
)

# Busybox sh seems to fail to detect proper executable if a file without executable extension exists in the same directory.
cross_exec() {
  cleanup
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

# Run a command preferring the Windows version if available.
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

# Ensure an argument for an option.
ensure_opt_arg() (
  if test -z "$2"
  then
    echo "No argument for option --$1." >&2
    usage
    exit 2
  fi
  echo "$2"
)

# Open a URL in a browser.
open_browser() (
  case "$(uname -s)" in
    Linux)
      xdg-open "$1" ;;
    Darwin)
      open "$1" ;;
    Windows_NT)
      PowerShell -Command "Start-Process '$1'" ;;
    *)
      echo "Unsupported OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
)

# Ensure a package is installed and return the command and arguments separated by tabs.
install_pkg_cmd_tabsep_args() (
  cmd_name=
  winget_id=
  win_cmd_path=
  scoop_id=
  brew_id=
  brew_cmd_path=
  while getopts nc:p:b:P:w:s:-: OPT
  do
    if test "$OPT" = "-"
    then
      OPT="${OPTARG%%=*}"
      # shellcheck disable=SC2030
      OPTARG="${OPTARG#"$OPT"}"
      OPTARG="${OPTARG#=}"
    fi
    case "$OPT" in
      b|brew-id) brew_id="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      B|brew-cmd-path) brew_cmd_path="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      c|cmd) cmd_name="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      w|winget-id) winget_id="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      p|winget-cmd-path) win_cmd_path="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      s|scoop-id) scoop_id="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      \?) exit 2;;
      *) echo "Unexpected option: $OPT" >&2; exit 2;;
    esac
  done
  shift $((OPTIND-1))
  unset OPTIND

  cmd_path="$cmd_name"
  if is_windows
  then
    if test -n "$scoop_id"
    then
      if ! type scoop > /dev/null 2>&1
      then
        powershell -Command "Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser; Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression" 1>&2
      fi
      cmd_path="$HOME"/scoop/shims/ghcup
      if ! type "$cmd_path" > /dev/null 2>&1
      then
        scoop install "$scoop_id" 1>&2
      fi
    elif test -n "$winget_id"
    then
      cmd_path="$win_cmd_path"
      if ! type "$cmd_path" > /dev/null 2>&1
      then
        winget install --accept-package-agreements --accept-source-agreements --exact --id "$winget_id" 2>&1
      fi
    else
      echo "No package ID for Windows specified." >&2
      exit 1
    fi
  elif is_darwin
  then
    if test -n "$brew_id"
    then
      if test -n "$brew_cmd_path"
      then
        cmd_path="$brew_cmd_path"
      fi
      if ! type "$cmd_path" > /dev/null 2>&1
      then
        brew install "$brew_id" 1>&2
      fi
    else
      echo "No package ID for macOS specified." >&2
      exit 1
    fi
  fi
  if which "$cmd_path" > /dev/null 2>&1
  then
    printf "%s\t" "$(which "$cmd_path")"
  else
    echo "Command not installed: $cmd_path" >&2
    exit 1
  fi
  for arg in "$@"
  do
    printf "%s\t" "$arg"
  done
)

unset ifs_a8fded1

# Save the current IFS and set it to the specified value.
set_ifs() {
  if test -z "${ifs_a8fded1+set}" && test -n "${IFS+set}"
  then
    ifs_a8fded1="$IFS"
  fi
  IFS="$1"
}

# For CSV.
set_ifs_comma() {
  set_ifs "$(printf ',')"
}

# For TSV.
set_ifs_tab() {
  set_ifs "$(printf '\t')"
}

# For PSV.
set_ifs_pipe() {
  set_ifs "$(printf '|')"
}

# Mianly for paths, files, and directories.
set_ifs_colon() {
  set_ifs "$(printf ':')"
}

set_ifs_path_list_sepaprator() {
  set_ifs_colon
}

# To split path.
set_ifs_slashes() {
  set_ifs "$(printf "/\\")"
}

set_ifs_path_sepaprator() {
  set_ifs_slashes
}

set_ifs_newline() {
  set_ifs "$(printf '\n\r')"
}

set_ifs_default() {
  set_ifs "$(printf ' \t\n\r')"
}

# Restore the saved IFS.
restore_ifs() {
  if test -n "${ifs_a8fded1+set}"
  then
    IFS="$ifs_a8fded1"
    unset ifs_a8fded1
  else
    IFS="$(printf ' \t\n\r')"
  fi
}

install_pkg_cmd() {
  set_ifs_tab
  # shellcheck disable=SC2046
  set -- $(install_pkg_cmd_tabsep_args "$@")
  restore_ifs
  echo "$1"
}

run_pkg_cmd() { # Run a command after ensuring it is installed.
  set_ifs_tab
  # shellcheck disable=SC2046
  set -- $(install_pkg_cmd_tabsep_args "$@")
  restore_ifs
  # echo 01d637b "$@" >&2
  cross_run "$@"
}

load() {
  if ! test -r "$1"
  then
    return 0
  fi
  IFS=
  while read -r line_0e9c96b
  do
    key_07bde23="$(echo "$line_0e9c96b" | sed -E -n -e 's/^([a-zA-Z_][[:alnum:]_]*)=.*$/\1/p')"
    if test -z "$key_07bde23"
    then
      continue
    fi
    val_5d77cea="$(eval "echo \"\${$key_07bde23:=}\"")"
    # echo bcff3d2 "$val" >&2
    if test -n "$val_5d77cea"
    then
      continue
    fi
    eval "$line_0e9c96b"
  done < "$1"
  unset IFS
}

load_env() { # Load environment variables.
  if test "${APP_ENV+set}" = set
  then
    load "$SCRIPT_DIR"/.env."$APP_ENV".dynamic
    load "$SCRIPT_DIR"/.env."$APP_ENV".local
  fi
  if test "${APP_SENV+set}" != set || test "${APP_SENV}" != "test"
  then
    load "$SCRIPT_DIR"/.env.dynamic
    load "$SCRIPT_DIR"/.env.local
  fi
  if test "${APP_ENV+set}" = set
  then
    load "$SCRIPT_DIR"/.env."$APP_ENV"
  fi
  # shellcheck disable=SC1091
  load "$SCRIPT_DIR"/.env
}

get_key() (
  # Some "/bin/sh" provides `-s` option.
  # shellcheck disable=SC3045
  read -rsn1 key
  echo "$key"
)

memoize() {
  cache_file_name_5f7dc05="$1"
  shift
  if ! test -r "$(temp_dir_path)/$cache_file_name_5f7dc05"
  then
    "$@" > "$(temp_dir_path)/$cache_file_name_5f7dc05"
  fi
  cat "$(temp_dir_path)/$cache_file_name_5f7dc05"
}

memoize_silent() (
  cache_file_name="$1"
  shift
  if ! test -r "$(temp_dir_path)/$cache_file_name"
  then
    "$@" > "$(temp_dir_path)/$cache_file_name" 2> /dev/null
  fi
  cat "$(temp_dir_path)/$cache_file_name"
)

first_call() {
  if eval "test \"\${guard_$1+set}\" = set"
  then
    return 1
  fi
  eval "guard_$1=x"
}

underline() {
  printf '\033[4m%s\033[0m' "$1"
}

bold() {
  printf '\033[1m%s\033[0m' "$1"
}

enclose_with_brackets() {
  printf '[%s]' "$1"
}

# Emphasize text.
emph() {
  if test -z "$1"
  then
    return
  fi
  if is_windows
  then
    enclose_with_brackets "$(bold "$(underline "$1")")"
  else
    bold "$(underline "$1")"
  fi
}

# Print a menu item.
menu_item() (
  if ! echo "$1" | grep -q -E -e '&'
  then
    echo "$1"
    return
  fi
  pre="$(printf "%s" "$1" | sed -E -e 's/&.*//')"
  ch="$(printf "%s" "$1" | sed -E -e 's/.*&(.).*/\1/')"
  post="$(printf "%s" "$1" | sed -E -e 's/.*&.//')"
  echo "$pre$(emph "$ch")$post"
)

# Sort version strings.
# shellcheck disable=SC2120
sort_version() {
  # Version strings which are composed of three parts are sorted considering the third part as a patch version.
  # `--version-sort` is specific to BSD sort.
  sed -E -e '/-/! { s/^([^.]+(\.[^.]+){2})$/\1_/; }' -e 's/-patch/_patch/' | sort -V "$@" | sed -e 's/_$//' -e 's/_patch/-patch/'
}

# Check if the version is greater than the specified version.
version_gt() {
  test "$(printf '%s\n' "$@" | sort_version | head -n 1)" != "$1"
}

version_ge() {
  test "$(printf '%s\n' "$@" | sort_version -r | head -n 1)" = "$1"
}

# --------------------------------------------------------------------------
# Array functions. "array string + delimiter" is used for the array representation.

# Join an array with a delimiter.
array_join() (
  arr="$1"
  IFS="$2"
  delim="$3"
  # shellcheck disable=SC2086
  printf "%s\n" $arr | paste -sd "$delim" -
)

array_head() (
  arr="$1"
  test -z "$arr" && return 1
  IFS="$2"
  # shellcheck disable=SC2086
  set -- $arr
  echo "$1"
)

array_first() {
  array_head "$@"
}

array_tail() (
  arr="$1"
  test -z "$arr" && return 1
  IFS="$2"
  # shellcheck disable=SC2086
  set -- $arr
  shift
  echo "$*"
)

array_rest() {
  array_tail "$@"
}

array_length() (
  arr="$1"
  IFS="$2"
  # shellcheck disable=SC2086
  set -- $arr
  echo "$#"
)

array_append() (
  arr="$1"
  shift
  IFS="$1"
  shift
  (
    test -n "$arr" && echo "$arr"
    printf "%s\n" "$@" | while read -r arr2
    do
      # shellcheck disable=SC2086
      printf "%s\n" $arr2
    done
  ) | paste -sd "$IFS" -
)

array_prepend() (
  arr="$1"
  shift
  IFS="$1"
  shift
  (
    printf "%s\n" "$@" | while read -r arr2
    do
      # shellcheck disable=SC2086
      printf "%s\n" $arr2
    done
    test -n "$arr" && echo "$arr"
  ) | paste -sd "$IFS" -
)

array_at() (
  arr="$1"
  shift
  IFS="$1"
  shift
  i=0
  for item in $arr
  do
    if test "$i" -eq "$1"
    then
      echo "$item"
      return
    fi
    i=$((i + 1))
  done
  return 1
)

# Map an array a command. If the command contains "_", then it is replaced with the item. If the array is "-", then the items are read from stdin.
array_map() (
  reads_stdin=false
  arr="$1"
  shift
  if test "$arr" = "-"
  then
    reads_stdin=true
    arr="$(cat)"
  fi
  IFS="$1"
  shift
  should_replace=false
  for arg in "$@"
  do
    if test "$arg" = "_" || test "$arg" = "it"
    then
      should_replace=true
    fi
  done
  delim=
  for i in $arr
  do
    printf "%s" "$delim"
    if $reads_stdin
    then
      printf "%s" "$(printf "%s" "$i" | "$@")"
    elif $should_replace
    then
      (
        for arg in "$@"
        do
          if test "$arg" = "_" || test "$arg" = "it"
          then
            # shellcheck disable=SC2016
            arg='$i'
          fi
          set -- "$@" "$arg"
          shift
        done
        printf "%s" "$("$@")"
      )
    else
      printf "%s" "$("$@" "$i")"
    fi
    delim="$IFS"
  done
)

# Filter an array with a command. If the command contains "_", then it is replaced with the item.
array_filter() (
  arr="$1"
  shift
  IFS="$1"
  shift
  should_replace=false
  for arg in "$@"
  do
    if test "$arg" = "_"
    then
      should_replace=true
    fi
  done
  delim=
  for i in $arr
  do
    if $should_replace
    then
      if ! (
        for arg in "$@"
        do
          if test "$arg" = "_"
          then
            # shellcheck disable=SC2016
            arg="$i"
          fi
          set -- "$@" "$arg"
          shift
        done
        "$@" 
      )
      then
        continue
      fi
    elif ! "$@" "$i"
    then
      continue
    fi
    printf "%s%s" "$delim" "$i"
    delim="$IFS"
  done
)

# Reduce an array with a function. If the function contains "_"s, then they are replaced with the accumulator and the item.
array_reduce() (
  arr="$1"
  shift
  IFS="$1"
  shift
  acc="$1"
  shift
  has_place_holder=false
  for arg in "$@"
  do
    if test "$arg" = "_"
    then
      has_place_holder=true
    fi
  done
  for i in $arr
  do
    if $has_place_holder
    then
      acc="$(
        first=true
        for arg in "$@"
        do
          if test "$arg" = "_"
          then
            if $first
            then
              arg="$acc"
              first=false
            else
              arg="$i"
            fi
          fi
          set -- "$@" "$arg"
          shift
        done
        "$@"
      )"
    else
      acc="$("$@" "$acc" "$i")"
    fi
  done
  echo "$acc"
)

array_reverse() (
  arr="$1"
  IFS="$2"
  # shellcheck disable=SC2086
  set -- $arr
  i=$#
  delim=
  while test "$i" -gt 0
  do
    eval printf "%s%s" "$delim" "\$$i"
    delim="$IFS"
    i=$((i - 1))
  done
)

array_contains() (
  arr="$1"
  shift
  IFS="$1"
  shift
  item="$1"
  shift
  for i in $arr
  do
    if test "$i" = "$item"
    then
      return 0
    fi
  done
  return 1
)

array_each() (
  reads_stdin=false
  arr="$1"
  shift
  if test "$arr" = "-"
  then
    reads_stdin=true
    arr="$(cat)"
  fi
  IFS="$1"
  shift
  should_replace=false
  for arg in "$@"
  do
    if test "$arg" = "_" || test "$arg" = "it"
    then
      should_replace=true
    fi
  done
  # shellcheck disable=SC2086
  printf "%s\n" $arr | while read -r i
  do
    if $reads_stdin
    then
      echo "$i" | "$@"
    elif $should_replace
    then
      (
        for arg in "$@"
        do
          if test "$arg" = "_" || test "$arg" = "it"
          then
            arg="$i"
          fi
          set -- "$@" "$arg"
          shift
        done
        "$@"
      )
    else
      "$@" "$i"
    fi
  done
)

array_sort() (
  arr="$1"
  shift
  IFS="$1"
  shift
  # shellcheck disable=SC2086
  printf "%s\n" $arr | if test "$#" -eq 0; then sort; else "$@"; fi | paste -sd "$IFS" -
)

if ! type shuf > /dev/null 2>&1
then
  alias shuf='sort -R'
fi

array_shuffle() (
  arr="$1"
  shift
  IFS="$1"
  shift
  # shellcheck disable=SC2086
  printf "%s\n" $arr | shuf | paste -sd "$IFS" -
)

# --------------------------------------------------------------------------

# Original directory when the script is invoked.

ORIGINAL_DIR="$PWD"
export ORIGINAL_DIR

chdir_original() {
  cd "$ORIGINAL_DIR" || exit 1
}

# Directory in which the main script is located.

SCRIPT_DIR="$(realpath "$(dirname "$0")")"
export SCRIPT_DIR

chdir_script() {
  cd "$SCRIPT_DIR" || exit 1
}

# Directory specified by the user with the `-d` option.

user_specified_directory=

chdir_user() {
  if test -n "$user_specified_directory"
  then
    cd "$user_specified_directory" || exit 1
  else
    chdir_original
  fi
}

# Check if the working directory is in the script directory.
in_script_dir() {
  echo "$PWD" | grep -q "^$SCRIPT_DIR"
}

# Create a temporary directory if required. BusyBox sh not supports -t.

_temp_dir_path_d4a4197="$(mktemp -d --dry-run)"

temp_dir_path() {
  if ! test -d "$_temp_dir_path_d4a4197"
  then
    mkdir -p "$_temp_dir_path_d4a4197"
  fi
  echo "$_temp_dir_path_d4a4197"
}

if ! type tac > /dev/null 2>&1
then
  tac() {
    tail -r
  }
fi

kill_children() {
  for i_519fa93 in $(seq 10)
  do
    kill "%$i_519fa93" > /dev/null 2>&1 || :
    wait "%$i_519fa93" > /dev/null 2>&1 || :
  done
}

csv_cleanup_handlers=

# Main cleanup handler.
cleanup() {
  rc=$?
  # On some systems, `kill` cannot detect the process if `jobs` is not called before it.
  if is_windows 
  then
    kill_children
  else 
    for i_519fa93 in $(jobs | tac | sed -E -e 's/^\[([0-9]+).*/\1/')
    do
      kill "%$i_519fa93"
      wait "%$i_519fa93" || :
    done
  fi
  # echo "Killed children." >&2

  rm -fr "$_temp_dir_path_d4a4197"
  # echo "Cleaned up temporary files." >&2

  if test "$rc" -ne 0
  then
    echo "Exiting with status $rc" >&2
    if type on_error > /dev/null 2>&1
    then
      on_error
    fi
  fi

  set_ifs_comma
  for cleanup_handler in $csv_cleanup_handlers
  do
    "$cleanup_handler"
  done
  restore_ifs

  exit "$rc"
}

# Add a cleanup handler, not replacing the existing ones.
add_cleanup_handler() {
  csv_cleanup_handlers="$(array_prepend "$csv_cleanup_handlers" , "$1")"
}

# Verbose flag.

verbose_f26120b=false 

verbose() {
  "$verbose_f26120b"
}

# -------------------------------------------------------------------------- 

psv_task_file_paths=

task_subcmds() ( # List subcommands.
  chdir_script
  delim=" delim_2ed1065 "
  set_ifs_pipe
  # shellcheck disable=SC2086
  set -- $psv_task_file_paths
  restore_ifs
  cnt="$(grep -E -h -e "^subcmd_[_[:alnum:]]+\(" "$@" | sed -r -e 's/^subcmd_//' -e 's/([[:alnum:]]+)__/\1:/g' -e "s/\(\) *[{(] *(# *)?/$delim/")"
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

task_tasks() ( # List tasks.
  delim=" delim_d3984dd "
  set_ifs_pipe
  # shellcheck disable=SC2086
  set -- $psv_task_file_paths
  restore_ifs
  cnt="$(grep -E -h -e "^task_[_[:alnum:]]+\(" "$@" | sed -E -e 's/^task_//' -e 's/([[:alnum:]]+)__/\1:/g' -e "s/\(\) *[{(] *(# *)?/$delim/")"
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

usage() ( # Show help message.
  chdir_script
  cat <<EOF
Usage:
  $ARG0BASE [options] <subcommand> [args...]
  $ARG0BASE [options] <task[arg1,arg2,...]> [tasks...]

Options:
  -d, --directory=<dir>  Change directory before running tasks.
  -h, --help             Display this help and exit.
  -v, --verbose          Verbose mode.

Subcommands:
EOF
  task_subcmds | sed -r -e 's/^/  /'
  cat <<EOF

Tasks:
EOF
  task_tasks | sed -r -e 's/^/  /'
)

main() {
  trap cleanup EXIT

  chdir_script

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
  else
    # shellcheck disable=SC2034
    ARG0="$0"
    ARG0BASE="$(basename "$0")"
  fi

  psv_task_file_paths="$(realpath "$0")"
  for task_file_path in "$SCRIPT_DIR"/task_*.sh "$SCRIPT_DIR"/task-*.sh
  do
    if ! test -r "$task_file_path"
    then
      continue
    fi
    case "$(basename "$task_file_path")" in
      task-dev.sh|task-prd.sh)
      continue
      ;;
    esac
    psv_task_file_paths="$psv_task_file_paths|$task_file_path"
    # shellcheck disable=SC1090
    . "$task_file_path"
  done

  shows_help=false
  while getopts d:hv-: OPT
  do
    if test "$OPT" = "-"
    then
      # shellcheck disable=SC2031
      OPT="${OPTARG%%=*}"
      # shellcheck disable=SC2031
      OPTARG="${OPTARG#"$OPT"}"
      OPTARG="${OPTARG#=}"
    fi
    case "$OPT" in
      d|directory) user_specified_directory="$(ensure_opt_arg "$OPT" "$OPTARG")";;
      h|help) shows_help=true;;
      v|verbose) verbose_f26120b=true;;
      \?) usage; exit 2;;
      *) echo "Unexpected option: $OPT" >&2; exit 2;;
    esac
  done
  shift $((OPTIND-1))
  unset OPTIND

  if $shows_help || test "${1+set}" != "set"
  then
    usage
    exit 0
  fi

  subcmd="$(echo "$1" | sed -r -e 's/:/__/g')"
  if type subcmd_"$subcmd" > /dev/null 2>&1
  then
    shift
    if alias subcmd_"$subcmd" > /dev/null 2>&1
    then
      # shellcheck disable=SC2294
      eval subcmd_"$subcmd" "$@"
      exit $?
    fi
    subcmd_"$subcmd" "$@"
    exit $?
  fi

  for task_with_args in "$@"
  do
    task_name="$task_with_args"
    args=""
    case "$task_with_args" in
      *\[*)
        task_name="${task_with_args%%\[*}"
        args="$(echo "$task_with_args" | sed -r -e 's/^.*\[//' -e 's/\]$//' -e 's/,/ /')"
        ;;
    esac
    task_name="$(echo "$task_name" | sed -r -e 's/:/__/g')"
    if ! type task_"$task_name" > /dev/null 2>&1
    then
      if type delegate_tasks > /dev/null 2>&1
      then
        verbose && echo "Delegating to delegate_tasks: $task_with_args" >&2
        delegate_tasks "$@"
        continue
      fi
      echo "Unknown task: $task_name" >&2
      exit 1
    fi
    # shellcheck disable=SC2086
    task_"$task_name" $args
  done
}

# To make this file can be sourced to provide functions.
if test "$(basename "$0")" = "task.sh"
then
  main "$@"
fi
