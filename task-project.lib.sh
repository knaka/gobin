#!/bin/sh
set -o nounset -o errexit

test "${guard_0416bc0+set}" = set && return 0; guard_0416bc0=-

. task.sh
. task-go.lib.sh

task_gen() (
  cross_run ./cmd-gobin run go-generate-fast .
)
