#!/usr/bin/env bash
# Renders every Graphviz source of a directory with the mirrored Xonovex style,
# in its light mode on a white background.
#
# Usage: render-diagrams.sh DIRECTORY FORMAT...
# A format is svg or png.
#
# An export that is already current stays untouched. git-lfs can leave its
# pointer in the working tree in place of the export, and git reports a pointer
# that is replaced by its own object as a modified file. The pointer names the
# digest of the export, so the comparison needs no download.
set -euo pipefail

die() {
  printf 'ERROR: %s\n' "$1" >&2
  exit 1
}

# Prints the SHA-256 digest of an export, or of the object that its git-lfs
# pointer names. Prints nothing when the export does not exist.
export_digest() {
  local file="$1"
  [ -f "$file" ] || return 0
  if head -n 1 -- "$file" | grep -qax 'version https://git-lfs\.github\.com/spec/v1'; then
    sed -n 's/^oid sha256:\([0-9a-f]\{64\}\)$/\1/p' "$file"
  else
    sha256sum -- "$file" | cut -d ' ' -f 1
  fi
}

render() {
  local format="$1"
  local target="$2"
  case "$format" in
  svg) dot -Gbgcolor=white -Tsvg -o "$target" ;;
  png) dot -Gbgcolor=white -Gdpi=150 -Tpng -o "$target" ;;
  *) die "Unknown format: $format" ;;
  esac
}

[ "$#" -ge 2 ] || die 'Usage: render-diagrams.sh DIRECTORY FORMAT...'
directory="$1"
shift
formats=("$@")

export FONTCONFIG_FILE="${XONOVEX_DIAGRAM_FONTCONFIG_FILE:?The Nix environment must supply the diagram fonts}"
for family in Inter 'Space Grotesk'; do
  if [ "$(fc-match --format '%{family[0]}' "$family")" != "$family" ]; then
    die "Missing diagram font: $family"
  fi
done

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
style="$script_dir/../../packages/asset/asset-brand/diagrams/style.gvpr"

shopt -s nullglob
sources=("$directory"/*.dot)
[ "${#sources[@]}" -gt 0 ] || die "No .dot sources found in $directory"

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT

for source in "${sources[@]}"; do
  styled="$(gvpr -c -a light -f "$style" "$source")"
  for format in "${formats[@]}"; do
    target="${source%.dot}.$format"
    fresh="$scratch/render.$format"
    printf '%s\n' "$styled" | render "$format" "$fresh"
    if [ "$(sha256sum -- "$fresh" | cut -d ' ' -f 1)" = "$(export_digest "$target")" ]; then
      printf 'current:  %s\n' "$target"
    else
      mv -- "$fresh" "$target"
      printf 'rendered: %s\n' "$target"
    fi
  done
done
