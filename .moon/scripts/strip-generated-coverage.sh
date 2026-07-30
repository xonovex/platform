#!/usr/bin/env bash
# Rewrites a Go coverage profile in place, dropping every block that belongs to a
# generated file, so the coverage gates measure hand-written code only. A file
# counts as generated when it carries the marker Go tooling defines for the
# purpose: a "// Code generated ... DO NOT EDIT." line.
set -euo pipefail

die() {
  printf 'ERROR: %s\n' "$1" >&2
  exit "${2:-1}"
}

[ "$#" -eq 1 ] || die "usage: strip-generated-coverage.sh PROFILE" 2

profile="$1"
[ -f "$profile" ] || die "coverage profile not found: $profile"

module="$(go list -m)"
[ -n "$module" ] || die "cannot resolve the Go module path"

module_dir="$(go list -m -f '{{.Dir}}')"
[ -n "$module_dir" ] || die "cannot resolve the Go module directory"

generated_list="$(mktemp)"
filtered="$(mktemp)"
trap 'rm -f "$generated_list" "$filtered"' EXIT

# grep reports no match with status 1, which for a module that generates nothing
# is the ordinary case rather than a failure.
grep_status=0
(cd "$module_dir" && grep -rlE '^// Code generated .* DO NOT EDIT\.$' --include='*.go' .) \
  >"$generated_list" || grep_status=$?
[ "$grep_status" -le 1 ] || die "cannot scan $module_dir for generated files"

[ -s "$generated_list" ] || exit 0

# Profile paths are import paths, so the on-disk names need the module prefix to
# match them.
sed "s|^\./|${module}/|" "$generated_list" >"$filtered"
cat "$filtered" >"$generated_list"

# Every profile line is "<import path>:<start>,<end> <statements> <count>"; the
# "mode:" header carries no range, so it never matches and is always kept.
awk '
  NR == FNR { generated[$0] = 1; next }
  {
    path = $0
    sub(/:[0-9]+\.[0-9]+,[0-9]+\.[0-9]+ .*$/, "", path)
    if (!(path in generated)) print
  }
' "$generated_list" "$profile" >"$filtered"

cat "$filtered" >"$profile"
