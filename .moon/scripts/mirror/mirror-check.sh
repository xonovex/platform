#!/usr/bin/env bash
# Checks a mirror package against its mirror.json without a source checkout or
# network access: every listed file must be a regular file with the recorded
# hash, and no other file may sit in the package.
set -euo pipefail

# shellcheck source-path=SCRIPTDIR
# shellcheck source=mirror-lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/mirror-lib.sh"

script_name=mirror-check
usage_line='usage: mirror-check.sh [--package DIR]'
package_dir=.

while (($# > 0)); do
  case $1 in
  --package)
    if (($# < 2)); then
      printf '%s\n' "$usage_line" >&2
      exit 2
    fi
    package_dir=$2
    shift 2
    ;;
  -h | --help)
    printf '%s\n' "$usage_line"
    exit 0
    ;;
  *)
    printf '%s\n' "$usage_line" >&2
    exit 2
    ;;
  esac
done

# The helper's problems go straight to stderr; its exit status is the verdict.
if ! rows=$(mirror_helper check "$package_dir/mirror.json"); then
  exit 1
fi
commit=${rows%%$'\n'*}

cd -- "$package_dir"

failed=0
checked=0
declare -A listed=()

while IFS=$'\t' read -r mirror_path _ recorded; do
  listed[$mirror_path]=1
  checked=$((checked + 1))
  if [[ -L $mirror_path ]]; then
    printf '%s: file listed in mirror.json is not a regular file: %s\n' "$script_name" "$mirror_path" >&2
    failed=1
  elif [[ ! -e $mirror_path ]]; then
    printf '%s: file listed in mirror.json is missing: %s\n' "$script_name" "$mirror_path" >&2
    failed=1
  elif [[ ! -f $mirror_path ]]; then
    printf '%s: file listed in mirror.json is not a regular file: %s\n' "$script_name" "$mirror_path" >&2
    failed=1
  else
    actual=$(mirror_file_hash "$mirror_path")
    if [[ $actual != "$recorded" ]]; then
      printf '%s: file differs from mirror.json: %s has sha256 %s, manifest records %s\n' \
        "$script_name" "$mirror_path" "$actual" "$recorded" >&2
      failed=1
    fi
  fi
done < <(tail -n +2 <<<"$rows")

# find writes to a file rather than a pipe so that a failing find stops the script.
found=$(mktemp)
trap 'rm -f -- "$found"' EXIT
find . \( -type f -o -type l \) -print0 >"$found"

while IFS= read -r -d '' path; do
  path=${path#./}
  case $path in
  moon.yml | package.json | mirror.json) continue ;;
  esac
  if [[ -z ${listed[$path]+x} ]]; then
    printf '%s: file is not in mirror.json: %s\n' "$script_name" "$path" >&2
    failed=1
  fi
done <"$found"

if ((failed)); then
  exit 1
fi
printf '%s: %d files match mirror.json at %s\n' "$script_name" "$checked" "$commit"
