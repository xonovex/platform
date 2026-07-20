#!/usr/bin/env bash
set -euo pipefail

die() {
  printf 'ERROR: %s\n' "$1" >&2
  exit "${2:-1}"
}

check_package() {
  local package_path="$1"
  local minimum="$2"
  local package_output
  local coverage

  if ! package_output="$(go test -cover "$package_path" 2>&1)"; then
    printf '%s\n' "$package_output" >&2
    die "tests failed for $package_path"
  fi
  printf '%s\n' "$package_output"
  coverage="$(printf '%s\n' "$package_output" | sed -n 's/.*coverage: \([0-9][0-9.]*\)%.*/\1/p' | tail -n 1)"
  [ -n "$coverage" ] || die "coverage result missing for $package_path"
  awk -v package_path="$package_path" -v coverage="$coverage" -v minimum="$minimum" 'BEGIN {
        if (coverage + 0 < minimum + 0) {
            printf "%s coverage %.1f%% is below %.1f%%\n", package_path, coverage, minimum > "/dev/stderr"
            exit 1
        }
    }'
}

[ "$#" -gt 0 ] || die "usage: check-go-package-coverage.sh PACKAGE=MINIMUM [...]" 2

for requirement in "$@"; do
  case "$requirement" in
  *=*)
    package_path="${requirement%%=*}"
    minimum="${requirement#*=}"
    [ -n "$package_path" ] || die "package path is empty in $requirement" 2
    [ -n "$minimum" ] || die "coverage minimum is empty in $requirement" 2
    check_package "$package_path" "$minimum"
    ;;
  *)
    die "expected PACKAGE=MINIMUM, got $requirement" 2
    ;;
  esac
done
