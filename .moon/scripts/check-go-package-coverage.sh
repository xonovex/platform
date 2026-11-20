#!/usr/bin/env bash
# Enforces the coverage floors of a Go module: one floor per package, and the
# GO_COVERAGE_MIN floor over the whole module. It runs the unit tests of the
# module once and measures every figure from that one profile. Generated files
# are excluded, so a floor states what the hand-written code is held to, and the
# module figure leaves out the packages declared exempt, so a package that
# another test tier covers does not count as untested code.
#
# GO_PACKAGE_FLOORS holds one PACKAGE=MINIMUM per line, or PACKAGE=exempt for a
# package that is deliberately ungated; PACKAGE is relative to the module root.
# A package the module builds but no line names is an error: a new package
# cannot enter the module without declaring its posture. The stripped module
# profile is left in coverage.out.
set -euo pipefail

die() {
  printf 'ERROR: %s\n' "$1" >&2
  exit "${2:-1}"
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

[ "$#" -eq 0 ] || die "usage: GO_COVERAGE_MIN=MINIMUM GO_PACKAGE_FLOORS=PACKAGE=MINIMUM|PACKAGE=exempt... check-go-package-coverage.sh" 2

mapfile -t requirements < <(printf '%s\n' "${GO_PACKAGE_FLOORS:-}" | awk '{ for (i = 1; i <= NF; i++) print $i }')
[ "${#requirements[@]}" -gt 0 ] || die "GO_PACKAGE_FLOORS names no package: declare the floor of every package of the module" 2

module_minimum="${GO_COVERAGE_MIN:-}"
case "$module_minimum" in
*[!0-9.]* | "") die "GO_COVERAGE_MIN must be the module coverage floor as a number, got '$module_minimum'" 2 ;;
esac

module="$(go list -m)"
[ -n "$module" ] || die "cannot resolve the Go module path"

profile="coverage.out"
subset="$(mktemp)"
declared="$(mktemp)"
exempt="$(mktemp)"
trap 'rm -f "$subset" "$declared" "$exempt"' EXIT

# Each argument names a package relative to the module root; a profile names
# it by its import path.
import_path() {
  case "$1" in
  . | ./) printf '%s\n' "$module" ;;
  ./*) printf '%s/%s\n' "$module" "${1#./}" ;;
  *) die "expected a package path that starts with ./, got '$1'" 2 ;;
  esac
}

# select_packages KEEP LIST writes to $subset the profile header and the blocks
# whose package is in the file LIST (KEEP=1) or is not in it (KEEP=0). Every
# profile line is "<import path>/<file>.go:<start>,<end> <statements> <count>",
# so the package is the import path without the file name.
select_packages() {
  awk -v keep="$1" '
    NR == FNR { listed[$0] = 1; next }
    FNR == 1 { print; next }
    {
      package = $0
      sub(/\/[^\/]*:[0-9]+\.[0-9]+,[0-9]+\.[0-9]+ .*$/, "", package)
      if ((package in listed) == (keep == 1)) print
    }
  ' "$2" "$profile" >"$subset"
}

# coverage_of prints the statement coverage of $subset as go tool cover rounds
# it, or nothing when $subset holds no statement. A profile holding only its
# "mode:" header has nothing to measure, and go tool cover reports that as
# 0.0% rather than as no result at all.
coverage_of() {
  [ "$(wc -l <"$subset")" -gt 1 ] || return 0
  go tool cover -func="$subset" | awk '/^total:/ { gsub(/%/, "", $3); print $3 }'
}

check_floor() {
  local name="$1"
  local coverage="$2"
  local minimum="$3"

  printf '%s coverage %s%%\n' "$name" "$coverage"
  awk -v name="$name" -v coverage="$coverage" -v minimum="$minimum" 'BEGIN {
        if (coverage + 0 < minimum + 0) {
            printf "%s coverage %.1f%% is below %.1f%%\n", name, coverage, minimum > "/dev/stderr"
            exit 1
        }
    }'
}

for requirement in "${requirements[@]}"; do
  case "$requirement" in
  *=*)
    package_path="${requirement%%=*}"
    minimum="${requirement#*=}"
    [ -n "$package_path" ] || die "package path is empty in $requirement" 2
    [ -n "$minimum" ] || die "coverage minimum is empty in $requirement" 2
    case "$minimum" in
    exempt) import_path "$package_path" >>"$exempt" ;;
    *[!0-9.]*) die "expected a number or 'exempt', got '$minimum' in $requirement" 2 ;;
    esac
    printf '%s\n' "$package_path" >>"$declared"
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

# go test reports its own coverage figures, which still count generated code.
# Only the post-strip figures below are reported, so the two cannot be confused.
if ! test_output="$(go test -coverprofile="$profile" ./... 2>&1)"; then
  printf '%s\n' "$test_output" >&2
  die "tests failed"
fi
bash "$script_dir/strip-generated-coverage.sh" "$profile"

failed=0
for requirement in "${requirements[@]}"; do
  package_path="${requirement%%=*}"
  minimum="${requirement#*=}"
  if [ "$minimum" = exempt ]; then
    printf '%s is exempt from a coverage floor\n' "$package_path"
    continue
  fi
  import_path "$package_path" >"$declared"
  select_packages 1 "$declared"
  coverage="$(coverage_of)"
  if [ -z "$coverage" ]; then
    printf '%s has no statements outside generated code\n' "$package_path"
    continue
  fi
  check_floor "$package_path" "$coverage" "$minimum" || failed=1
done

select_packages 0 "$exempt"
coverage="$(coverage_of)"
[ -n "$coverage" ] || die "the module has no statements outside generated code and exempt packages"
check_floor "module (without the exempt packages)" "$coverage" "$module_minimum" || failed=1

[ "$failed" -eq 0 ] || die "coverage is below a floor"
