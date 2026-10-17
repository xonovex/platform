#!/usr/bin/env bash
# Runs the mirror scripts against fixture repositories that each case builds in a
# temporary directory: a source repository standing for drodan-platform and a
# workspace repository holding a mirror package. The cases FAULT_* are the four
# seeded faults; each one must fail exactly its own case.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
kit="$(dirname "$here")"
fixture_source="$here/fixtures/source"
workspace_root="$(cd "$kit/../../.." && pwd)"

export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_SYSTEM=/dev/null
export GIT_AUTHOR_NAME=fixture GIT_AUTHOR_EMAIL=fixture@example.invalid
export GIT_COMMITTER_NAME=fixture GIT_COMMITTER_EMAIL=fixture@example.invalid
for variable in $(git rev-parse --local-env-vars 2>/dev/null); do
  unset "$variable"
done

root="$(mktemp -d)"
trap 'rm -rf "$root"' EXIT

status=0
out=""
err=""

run() {
  set +e
  "$@" >"$root/out" 2>"$root/err"
  status=$?
  set -e
  out="$(<"$root/out")"
  err="$(<"$root/err")"
}

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

assert_status() {
  [ "$status" -eq "$1" ] || fail "$2: expected exit $1, got $status; stderr: $err"
}

assert_stderr_is() {
  [ "$err" = "$1" ] || fail "$2: stderr differs
expected: $1
actual:   $err"
}

assert_stdout_is() {
  [ "$out" = "$1" ] || fail "$2: stdout differs
expected: $1
actual:   $out"
}

sha_of() {
  sha256sum "$1" | cut -d' ' -f1
}

# Every file with its bytes and mtime, so a rerun that rewrites anything shows.
snapshot() {
  (cd "$1" && find . -type f -printf '%p %T@ %s\n' | sort)
}

# sets: T, src, ws, pkg, c1, c2
make_world() {
  T="$root/$1"
  src="$T/drodan/drodan-platform"
  ws="$T/xonovex/xonovex-platform"
  pkg="$ws/packages/asset/asset-fixture"
  mkdir -p "$src" "$pkg"
  cp -R "$fixture_source/." "$src/"
  mkdir -p "$src/brand" "$src/lfs"
  printf 'bin\000ary\r\nbytes\377\n' >"$src/brand/data.bin"
  # git-lfs uploads the object of every committed file that holds a pointer, so
  # a pointer file in this repository blocks the push. make_world writes it.
  printf 'version https://git-lfs.github.com/spec/v1\noid sha256:%064d\nsize 12\n' 0 >"$src/lfs/pointer.txt"
  git -C "$src" init -q -b main
  git -C "$src" add -A
  git -C "$src" commit -q -m first
  c1="$(git -C "$src" rev-parse HEAD)"
  printf 'second line\n' >>"$src/brand/style.txt"
  git -C "$src" commit -q -a -m second
  c2="$(git -C "$src" rev-parse HEAD)"
  git -C "$ws" init -q -b main
  printf 'tasks: {}\n' >"$pkg/moon.yml"
  printf '{"name": "fixture", "private": true}\n' >"$pkg/package.json"
  write_bootstrap_manifest
}

write_bootstrap_manifest() {
  cat >"$pkg/mirror.json" <<'EOF'
{
  "source": {"repository": "drodan/drodan-platform"},
  "files": {
    "style.txt": {"source": "brand/style.txt"},
    "nested/data.bin": {"source": "brand/data.bin"},
    "nested/same.txt": {"source": "brand/same.txt"}
  }
}
EOF
}

check() { run bash "$kit/mirror-check.sh" --package "$pkg"; }
sync() { run bash "$kit/mirror-sync.sh" --package "$pkg" --source "$src" "$@"; }
source_check() { run bash "$kit/mirror-source-check.sh" --package "$pkg" --source "$src"; }

synced_world() {
  make_world "$1"
  sync
  assert_status 0 "bootstrap sync"
}

set_manifest_commit() {
  sed -i "s/\"commit\": \"[0-9a-f]*\"/\"commit\": \"$1\"/" "$pkg/mirror.json"
}

report_fault() {
  printf 'FAULT %s: exit %s: %s\n' "$1" "$status" "$err"
}

test_sync_writes_files_hashes_and_commit() {
  synced_world sync-writes
  [ "$(sha_of "$pkg/style.txt")" = "$(git -C "$src" show "$c2:brand/style.txt" | sha256sum | cut -d' ' -f1)" ] ||
    fail "style.txt is not the source file at the head commit"
  cmp -s "$pkg/nested/data.bin" "$src/brand/data.bin" || fail "binary file with NUL, CR and 0xFF bytes changed in transit"
  grep -q "\"commit\": \"$c2\"" "$pkg/mirror.json" || fail "manifest does not record the head commit"
  grep -q "\"sha256\": \"$(sha_of "$pkg/nested/same.txt")\"" "$pkg/mirror.json" || fail "manifest does not record the file hash"
  "$workspace_root/node_modules/.bin/prettier" --no-config --check "$pkg/mirror.json" >/dev/null ||
    fail "manifest is not in Prettier format"
}

test_check_passes_on_synced_mirror() {
  synced_world check-passes
  check
  assert_status 0 "mirror-check"
  assert_stdout_is "mirror-check: 3 files match mirror.json at $c2" "mirror-check"
}

test_source_check_passes_on_synced_mirror() {
  synced_world source-check-passes
  source_check
  assert_status 0 "mirror-source-check"
  assert_stdout_is "mirror-source-check: 3 source files match mirror.json at $c2" "mirror-source-check"
}

test_second_sync_changes_nothing() {
  synced_world second-sync
  local before after
  before="$(snapshot "$pkg")"
  sleep 1
  sync
  assert_status 0 "second sync"
  after="$(snapshot "$pkg")"
  [ "$before" = "$after" ] || fail "second sync changed a file or an mtime"
  assert_stdout_is "mirror-sync: 3 files at $c2: 0 written, 3 unchanged" "second sync"
}

test_sync_takes_an_older_commit_by_short_id() {
  make_world sync-older
  sync --commit "${c1:0:10}"
  assert_status 0 "sync at the first commit"
  grep -q "\"commit\": \"$c1\"" "$pkg/mirror.json" || fail "manifest does not hold the full commit id"
  [ "$(wc -l <"$pkg/style.txt")" -eq 3 ] || fail "style.txt is not the first commit's file"
  sync
  assert_status 0 "sync at head"
  [ "$(wc -l <"$pkg/style.txt")" -eq 4 ] || fail "style.txt did not move to the head commit"
  source_check
  assert_status 0 "source check after the move"
}

test_sync_reads_the_default_source_and_head() {
  make_world sync-default
  (cd "$pkg" && run bash "$kit/mirror-sync.sh")
  assert_status 0 "sync with the default source"
  grep -q "\"commit\": \"$c2\"" "$pkg/mirror.json" || fail "default commit is not the head of the source"
}

test_sync_from_a_worktree_of_the_source() {
  make_world sync-worktree
  git -C "$src" worktree add -q --detach "$T/child-worktree" "$c1"
  printf 'unstaged\n' >>"$T/child-worktree/brand/style.txt"
  run bash "$kit/mirror-sync.sh" --package "$pkg" --source "$T/child-worktree"
  assert_status 0 "sync from a worktree"
  grep -q "\"commit\": \"$c1\"" "$pkg/mirror.json" || fail "worktree head is not the default commit"
  cmp -s "$pkg/style.txt" <(git -C "$src" show "$c1:brand/style.txt") || fail "sync read the worktree file instead of the commit"
}

test_sync_all_or_nothing_on_a_missing_path() {
  make_world sync-missing-path
  cat >"$pkg/mirror.json" <<'EOF'
{
  "source": {"repository": "drodan/drodan-platform"},
  "files": {
    "a.txt": {"source": "brand/same.txt"},
    "b.txt": {"source": "brand/absent.txt"}
  }
}
EOF
  sync
  assert_status 1 "sync with a path that is not in the commit"
  assert_stderr_is "mirror-sync: path not found at commit $c2: brand/absent.txt" "sync with a missing path"
  [ ! -e "$pkg/a.txt" ] || fail "sync wrote a file although another file failed"
}

test_sync_refuses_a_git_lfs_pointer() {
  make_world sync-lfs
  cat >"$pkg/mirror.json" <<'EOF'
{
  "source": {"repository": "drodan/drodan-platform"},
  "files": {"p.txt": {"source": "lfs/pointer.txt"}}
}
EOF
  sync
  assert_status 1 "sync of an LFS pointer"
  assert_stderr_is "mirror-sync: source file is a Git LFS pointer at commit $c2: lfs/pointer.txt" "sync of an LFS pointer"
}

test_check_rejects_a_manifest_off_format() {
  synced_world manifest-format
  sed -i 's/"sha256": "\([0-9a-f]*\)"/"sha256": "\U\1"/' "$pkg/mirror.json"
  check
  assert_status 1 "uppercase hash"
  case "$err" in *'mirror.json: files["nested/data.bin"].sha256 is not 64 lowercase hex characters'*) ;; *) fail "no message for an uppercase hash: $err" ;; esac
  synced_world manifest-unsorted
  node -e '
    const fs = require("node:fs");
    const m = JSON.parse(fs.readFileSync(process.argv[1], "utf8"));
    const files = Object.fromEntries(Object.entries(m.files).reverse());
    fs.writeFileSync(process.argv[1], JSON.stringify({...m, files}, null, 2) + "\n");
  ' "$pkg/mirror.json"
  check
  assert_status 1 "unsorted keys"
  case "$err" in *'mirror.json: is not in canonical form'*) ;; *) fail "no message for unsorted keys: $err" ;; esac
}

test_usage_errors_exit_2() {
  make_world usage
  run bash "$kit/mirror-check.sh" --nope
  assert_status 2 "unknown option"
  case "$err" in usage:*) ;; *) fail "no usage line: $err" ;; esac
  run bash "$kit/mirror-sync.sh" --commit
  assert_status 2 "option without value"
}

test_FAULT_1_changed_mirror_file() {
  synced_world fault-1
  printf 'edited by hand\n' >>"$pkg/style.txt"
  check
  report_fault "1 changed-mirror-file"
  assert_status 1 "changed mirror file"
  assert_stderr_is "mirror-check: file differs from mirror.json: style.txt has sha256 $(sha_of "$pkg/style.txt"), manifest records $(sha_of <(git -C "$src" show "$c2:brand/style.txt"))" "changed mirror file"
  source_check
  assert_status 0 "the source still matches the manifest"
}

test_FAULT_2_file_not_in_manifest() {
  synced_world fault-2
  printf 'stray\n' >"$pkg/extra.txt"
  check
  report_fault "2 file-not-in-manifest"
  assert_status 1 "file outside the manifest"
  assert_stderr_is "mirror-check: file is not in mirror.json: extra.txt" "file outside the manifest"
}

test_FAULT_3_source_differs_at_the_recorded_commit() {
  synced_world fault-3
  set_manifest_commit "$c1"
  check
  assert_status 0 "mirror-check reads no source"
  source_check
  report_fault "3 source-differs-at-recorded-commit"
  assert_status 1 "source file that differs at the recorded commit"
  assert_stderr_is "mirror-source-check: source file differs from mirror.json at commit $c1: brand/style.txt has sha256 $(git -C "$src" show "$c1:brand/style.txt" | sha256sum | cut -d' ' -f1), manifest records $(sha_of "$pkg/style.txt")" "source differs"
}

test_FAULT_4a_missing_checkout() {
  synced_world fault-4a
  check
  assert_status 0 "mirror-check needs no checkout"
  run bash "$kit/mirror-source-check.sh" --package "$pkg" --source "$T/no-such-checkout"
  report_fault "4a missing-checkout (mirror-source-check)"
  assert_status 1 "missing checkout"
  assert_stderr_is "mirror-source-check: source directory not found: $T/no-such-checkout" "missing checkout"
  run bash "$kit/mirror-sync.sh" --package "$pkg" --source "$T/no-such-checkout"
  report_fault "4a missing-checkout (mirror-sync)"
  assert_status 1 "sync without checkout"
  assert_stderr_is "mirror-sync: source directory not found: $T/no-such-checkout" "sync without checkout"
  mkdir "$T/not-a-repository"
  run bash "$kit/mirror-source-check.sh" --package "$pkg" --source "$T/not-a-repository"
  report_fault "4a directory-without-repository"
  assert_stderr_is "mirror-source-check: source directory is not a git repository: $T/not-a-repository" "directory without repository"
}

test_FAULT_4b_missing_commit() {
  synced_world fault-4b
  local absent="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  set_manifest_commit "$absent"
  check
  assert_status 0 "mirror-check reads no commit"
  source_check
  report_fault "4b missing-commit (mirror-source-check)"
  assert_status 1 "missing commit"
  assert_stderr_is "mirror-source-check: commit not found in source repository $src: $absent" "missing commit"
  sync --commit bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
  report_fault "4b missing-commit (mirror-sync)"
  assert_status 1 "sync at a missing commit"
  assert_stderr_is "mirror-sync: commit not found in source repository $src: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" "sync at a missing commit"
}

main() {
  local name failed=0 total=0
  local -a names
  mapfile -t names < <(declare -F | cut -d' ' -f3 | grep '^test_' | sort)
  for name in "${names[@]}"; do
    total=$((total + 1))
    set +e
    (set -e && "$name")
    local code=$?
    set -e
    if [ "$code" -eq 0 ]; then
      printf 'ok   %s\n' "$name"
    else
      printf 'FAIL %s\n' "$name"
      failed=$((failed + 1))
    fi
  done
  printf '%d cases, %d failed\n' "$total" "$failed"
  [ "$failed" -eq 0 ]
}

main
