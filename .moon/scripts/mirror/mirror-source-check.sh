#!/usr/bin/env bash
# Checks each file recorded in mirror.json against the source repository at the
# recorded commit. mirror-check only sees the mirror copies, so it cannot tell
# that the source has moved on.
set -euo pipefail

script_name=mirror-source-check
usage_line='usage: mirror-source-check.sh [--package DIR] [--source DIR]'

# shellcheck source-path=SCRIPTDIR source=mirror-lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/mirror-lib.sh"

package_dir=.
source_dir=
while (($# > 0)); do
  case "$1" in
  -h | --help)
    printf '%s\n' "$usage_line"
    exit 0
    ;;
  --package | --source)
    # A following option means the value is missing, not that it is a directory named -x.
    if (($# < 2)) || [[ "$2" == -* ]]; then
      printf '%s\n' "$usage_line" >&2
      exit 2
    fi
    if [[ "$1" == --package ]]; then
      package_dir=$2
    else
      source_dir=$2
    fi
    shift 2
    ;;
  *)
    printf '%s\n' "$usage_line" >&2
    exit 2
    ;;
  esac
done

# The helper output is captured in a variable so that its exit status decides.
# Its stderr goes straight to this script's stderr.
if ! rows=$(mirror_helper check "$package_dir/mirror.json"); then
  exit 1
fi
mapfile -t lines <<<"$rows"
commit=${lines[0]}

if [[ -z "$source_dir" ]]; then
  source_dir=$(mirror_default_source "$package_dir") ||
    mirror_die "$script_name" "cannot resolve the default source directory of $package_dir"
fi
mirror_require_source "$script_name" "$source_dir"

if ! git -C "$source_dir" cat-file -e "$commit^{commit}" 2>/dev/null; then
  mirror_die "$script_name" "commit not found in source repository $source_dir: $commit"
fi

failed=0
for ((i = 1; i < ${#lines[@]}; i++)); do
  IFS=$'\t' read -r _ source_path recorded <<<"${lines[i]}"
  if ! actual=$(mirror_blob_hash "$source_dir" "$commit" "$source_path"); then
    printf '%s: path not found at commit %s: %s\n' "$script_name" "$commit" "$source_path" >&2
    failed=1
    continue
  fi
  # A pointer has its own hash, so reporting the hash mismatch too would repeat the failure.
  if mirror_is_lfs_pointer "$source_dir" "$commit" "$source_path"; then
    printf '%s: source file is a Git LFS pointer at commit %s: %s\n' "$script_name" "$commit" "$source_path" >&2
    failed=1
    continue
  fi
  if [[ "$actual" != "$recorded" ]]; then
    printf '%s: source file differs from mirror.json at commit %s: %s has sha256 %s, manifest records %s\n' \
      "$script_name" "$commit" "$source_path" "$actual" "$recorded" >&2
    failed=1
  fi
done

if ((failed != 0)); then
  exit 1
fi
printf '%s: %d source files match mirror.json at %s\n' "$script_name" "$((${#lines[@]} - 1))" "$commit"
