#!/usr/bin/env bash
# Copies the files that mirror.json lists from one source commit into the package,
# then records that commit and the hashes of the copies. Every file is read before
# the package is written, so a failing file leaves the package untouched.
set -euo pipefail

# shellcheck source-path=SCRIPTDIR source=mirror-lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/mirror-lib.sh"

script_name=mirror-sync
usage='usage: mirror-sync.sh [--package DIR] [--commit REV] [--source DIR]'

usage_error() {
  printf '%s\n' "$usage" >&2
  exit 2
}

package=.
commit_rev=HEAD
while (($# > 0)); do
  case "$1" in
  -h | --help)
    printf '%s\n' "$usage"
    exit 0
    ;;
  --package)
    (($# >= 2)) || usage_error
    package=$2
    shift 2
    ;;
  --commit)
    (($# >= 2)) || usage_error
    commit_rev=$2
    shift 2
    ;;
  --source)
    (($# >= 2)) || usage_error
    source_dir=$2
    shift 2
    ;;
  *) usage_error ;;
  esac
done

manifest="$package/mirror.json"
tmp=$(mktemp -d)
trap 'rm -rf -- "$tmp"' EXIT

mirror_helper list "$manifest" >"$tmp/list"

# mirror_default_source fails inside a condition, where errexit is off in its body,
# so the repository check runs here in the main shell.
if [[ ! -v source_dir ]]; then
  git -C "$package" rev-parse --show-toplevel >/dev/null 2>&1 ||
    mirror_die "$script_name" "package is not in a git repository: $package"
  source_dir=$(mirror_default_source "$package")
fi
mirror_require_source "$script_name" "$source_dir"

if ! full_commit=$(git -C "$source_dir" rev-parse --verify --quiet "$commit_rev^{commit}"); then
  mirror_die "$script_name" "commit not found in source repository $source_dir: $commit_rev"
fi

mkdir -p "$tmp/files"
while IFS=$'\t' read -r mirror_path source_path; do
  staged="$tmp/files/$mirror_path"
  if [[ "$(git -C "$source_dir" cat-file -t "$full_commit:$source_path" 2>/dev/null || true)" != blob ]]; then
    mirror_die "$script_name" "path not found at commit $full_commit: $source_path"
  fi
  if mirror_is_lfs_pointer "$source_dir" "$full_commit" "$source_path"; then
    mirror_die "$script_name" "source file is a Git LFS pointer at commit $full_commit: $source_path"
  fi
  mkdir -p "$(dirname "$staged")"
  git -C "$source_dir" cat-file blob "$full_commit:$source_path" >"$staged"
done <"$tmp/list"

rows=''
written=0
unchanged=0
while IFS=$'\t' read -r mirror_path source_path; do
  staged="$tmp/files/$mirror_path"
  target="$package/$mirror_path"
  if [[ -f "$target" && ! -L "$target" ]] && cmp -s -- "$staged" "$target"; then
    unchanged=$((unchanged + 1))
  else
    mkdir -p -- "$(dirname -- "$target")"
    rm -f -- "$target"
    cp -- "$staged" "$target"
    chmod 0644 -- "$target"
    written=$((written + 1))
  fi
  rows+="$mirror_path"$'\t'"$source_path"$'\t'"$(mirror_file_hash "$staged")"$'\n'
done <"$tmp/list"

printf '%s' "$rows" | mirror_helper write "$manifest" "$full_commit"
printf '%s: %s files at %s: %s written, %s unchanged\n' \
  "$script_name" "$((written + unchanged))" "$full_commit" "$written" "$unchanged"
