#!/usr/bin/env bash
# Enforces a per-package Go coverage floor and asserts that the floors cover
# every package in the module. Generated files are excluded, so a floor states
# what the hand-written code is held to.
#
# Each argument is PACKAGE=MINIMUM, or PACKAGE=exempt for a package that is
# deliberately ungated. A package the module builds but no argument names is an
# error: a new package cannot enter the module without declaring its posture.
set -euo pipefail

die() {
  printf 'ERROR: %s\n' "$1" >&2
  exit "${2:-1}"
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

check_package() {
  local package_path="$1"
  local minimum="$2"
  local profile="$3"
  local coverage
  local test_output

  # go test reports its own coverage figure, which still counts generated code.
  # Only the post-strip figure below is reported, so the two cannot be confused.
  if ! test_output="$(go test -coverprofile="$profile" "$package_path" 2>&1)"; then
    printf '%s\n' "$test_output" >&2
    die "tests failed for $package_path"
  fi
  bash "$script_dir/strip-generated-coverage.sh" "$profile"

  # A profile holding only its "mode:" header has nothing to measure, and
  # go tool cover reports that as 0.0% rather than as no result at all.
  if [ "$(wc -l <"$profile")" -le 1 ]; then
    printf '%s has no statements outside generated code\n' "$package_path"
    return 0
  fi

  coverage="$(go tool cover -func="$profile" |
    awk '/^total:/ { gsub(/%/, "", $3); print $3 }')"
  [ -n "$coverage" ] || die "coverage result missing for $package_path"

  printf '%s coverage %s%%\n' "$package_path" "$coverage"
  awk -v package_path="$package_path" -v coverage="$coverage" -v minimum="$minimum" 'BEGIN {
        if (coverage + 0 < minimum + 0) {
            printf "%s coverage %.1f%% is below %.1f%%\n", package_path, coverage, minimum > "/dev/stderr"
            exit 1
        }
    }'
}

[ "$#" -gt 0 ] || die "usage: check-go-package-coverage.sh PACKAGE=MINIMUM|PACKAGE=exempt [...]" 2

module="$(go list -m)"
[ -n "$module" ] || die "cannot resolve the Go module path"

profile="$(mktemp)"
declared="$(mktemp)"
trap 'rm -f "$profile" "$declared"' EXIT

for requirement in "$@"; do
  case "$requirement" in
  *=*)
    package_path="${requirement%%=*}"
    minimum="${requirement#*=}"
    [ -n "$package_path" ] || die "package path is empty in $requirement" 2
    [ -n "$minimum" ] || die "coverage minimum is empty in $requirement" 2
    printf '%s\n' "$package_path" >>"$declared"
    case "$minimum" in
    exempt) printf '%s is exempt from a coverage floor\n' "$package_path" ;;
    *[!0-9.]* | "") die "expected a number or 'exempt', got '$minimum' in $requirement" 2 ;;
    *) check_package "$package_path" "$minimum" "$profile" ;;
    esac
    ;;
  *)
    die "expected PACKAGE=MINIMUM, got $requirement" 2
    ;;
  esac
done

# go list emits absolute import paths; the arguments name packages relative to
# the module root, so the module prefix comes off before comparing.
undeclared="$(
  go list ./... |
    sed -e "s|^${module}\$|.|" -e "s|^${module}/|./|" |
    sort -u |
    comm -23 - <(sort -u "$declared")
)"

if [ -n "$undeclared" ]; then
  printf 'ERROR: these packages declare no coverage floor:\n' >&2
  printf '%s\n' "$undeclared" | sed 's/^/  /' >&2
  die "add PACKAGE=MINIMUM or PACKAGE=exempt for each"
fi
