# shellcheck shell=bash
# Functions for the mirror scripts. Sourcing defines functions only.

mirror_die() {
  local script_name=$1 message=$2 code=${3:-1}
  printf '%s: %s\n' "$script_name" "$message" >&2
  exit "$code"
}

mirror_helper() {
  local lib_dir
  lib_dir=$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)
  node "$lib_dir/mirror-manifest.mjs" "$@"
}

mirror_default_source() {
  local package_dir=$1 top
  top=$(git -C "$package_dir" rev-parse --show-toplevel)
  realpath -m "$top/../../drodan/drodan-platform"
}

mirror_require_source() {
  local script_name=$1 source_dir=$2
  if [[ ! -d "$source_dir" ]]; then
    mirror_die "$script_name" "source directory not found: $source_dir"
  fi
  if ! git -C "$source_dir" rev-parse --git-dir >/dev/null 2>&1; then
    mirror_die "$script_name" "source directory is not a git repository: $source_dir"
  fi
}

# cat-file reads the exact blob. git show would also run any textconv driver.
mirror_blob_hash() {
  local source_dir=$1 commit=$2 source_path=$3 type
  type=$(git -C "$source_dir" cat-file -t "$commit:$source_path" 2>/dev/null) || return 1
  [[ "$type" == blob ]] || return 1
  git -C "$source_dir" cat-file blob "$commit:$source_path" | sha256sum | cut -d ' ' -f 1
}

# A pipe into head can end early and fail git with SIGPIPE, so the first
# bytes are read with the builtin instead. NUL stops the read, which keeps
# binary blobs from matching the ASCII header.
mirror_is_lfs_pointer() {
  local source_dir=$1 commit=$2 source_path=$3 header=''
  IFS= read -r -d '' -n 42 header < <(git -C "$source_dir" cat-file blob "$commit:$source_path" 2>/dev/null) || true
  [[ "$header" == "version https://git-lfs.github.com/spec/v1" ]]
}

mirror_file_hash() {
  sha256sum -- "$1" | cut -d ' ' -f 1
}
