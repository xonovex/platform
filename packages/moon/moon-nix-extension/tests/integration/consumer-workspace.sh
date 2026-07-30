#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '[moon-nix-extension] %s\n' "$*"
}

fail() {
  printf 'ERROR: %s\n' "$1" >&2
  return 1
}

require_command() {
  command -v "$1" > /dev/null 2>&1 || fail "required command not found: $1"
}

assert_contains() {
  local file="$1"
  local expected="$2"

  rg -F --quiet -- "$expected" "$file" || {
    sed -n '1,200p' "$file" >&2
    fail "expected $file to contain: $expected"
  }
}

assert_not_contains() {
  local file="$1"
  local rejected="$2"

  if rg -F --quiet -- "$rejected" "$file"; then
    sed -n '1,200p' "$file" >&2
    fail "expected $file not to contain: $rejected"
  fi
}

assert_json() {
  local file="$1"
  local expression="$2"

  jq -e "$expression" "$file" > /dev/null || {
    jq '.' "$file" >&2
    fail "JSON assertion failed for $file: $expression"
  }
}

counter_lines() {
  local counter="$1"

  if [[ -f "$counter" ]]; then
    wc -l < "$counter"
  else
    printf '0\n'
  fi
}

repo_root="$(git rev-parse --show-toplevel)"
fixture_source="$repo_root/packages/moon/moon-nix-extension/tests/fixtures/consumer-workspace"
wasm="$repo_root/packages/moon/moon-nix-extension/target/wasm32-wasip1/release/moon_nix_extension.wasm"
moon_entry="$repo_root/node_modules/@moonrepo/cli/moon.js"
node_bin="$(command -v node)"
temp_parent="${TMPDIR:-/tmp}"
temp_root="$(mktemp -d "$temp_parent/moon-nix-extension.XXXXXX")"
workspace="$temp_root/workspace"
evidence="$temp_root/evidence"
counters="$temp_root/counters"

cleanup() {
  local status=$?

  trap - EXIT
  if [[ "$status" -eq 0 ]]; then
    if [[ -d "$temp_root" && "$temp_root" == "$temp_parent"/moon-nix-extension.* ]]; then
      rm -rf -- "$temp_root"
    else
      printf 'ERROR: refusing to remove unvalidated temp directory: %s\n' "$temp_root" >&2
      status=1
    fi
  else
    printf 'Integration workspace retained at: %s\n' "$temp_root" >&2
  fi
  exit "$status"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

require_command git
require_command jq
require_command nix
require_command rg
[[ -d "$fixture_source" ]] || fail "fixture source not found: $fixture_source"
[[ -f "$wasm" ]] || fail "release WASM not found: $wasm"
[[ -f "$moon_entry" ]] || fail "repository Moon entry point not found: $moon_entry"

mkdir -p "$workspace/.moon/plugins" "$evidence" "$counters"
cp -R "$fixture_source/." "$workspace/"
cp "$wasm" "$workspace/.moon/plugins/moon_nix_extension.wasm"
git -C "$workspace" init -q
git -C "$workspace" add --all
git -C "$workspace" -c user.name=Fixture -c user.email=fixture@example.test commit -qm fixture

export MOON_HOME="$temp_root/moon-home"
export PROTO_HOME="$temp_root/proto-home"
export MOON_FIXTURE_COUNTER_DIR="$counters"
export MOON_SKIP_INSTALL_DEPS=true
export MOON_TELEMETRY=false
export PROTO_BYPASS_VERSION_CHECK=true
export PROTO_OFFLINE=true
export PROTO_TELEMETRY=false
export PROTO_VERSION_CHECKS=false
export NO_COLOR=1
unset \
  IN_NIX_SHELL \
  MOON_CACHE_DIR \
  MOON_NIX_WRAPPED \
  MOON_PROJECT_ID \
  MOON_PROJECT_ROOT \
  MOON_PROJECT_SNAPSHOT \
  MOON_PROJECT_SOURCE \
  MOON_TARGET \
  MOON_TASK_HASH \
  MOON_TASK_ID \
  MOON_TASK_RETRY_ATTEMPT \
  MOON_TASK_RETRY_TOTAL \
  MOON_WORKING_DIR \
  MOON_WORKSPACE_ROOT

moon() {
  (
    cd "$workspace"
    "$node_bin" "$moon_entry" "$@"
  )
}

run_moon() {
  local output="$1"
  shift

  if ! moon "$@" > "$output" 2>&1; then
    sed -n '1,240p' "$output" >&2
    fail "Moon command failed: $*"
  fi
}

capture_hash() {
  local target="$1"
  local project="${target%%:*}"
  local task="${target#*:}"

  jq -er '.hash' "$workspace/.moon/cache/states/$project/$task/lastRun.json"
}

capture_hash_manifest() {
  local hash="$1"
  local output="$2"

  if ! moon hash "$hash" --json > "$output" 2> "$output.stderr"; then
    sed -n '1,200p' "$output.stderr" >&2
    fail "unable to inspect Moon hash: $hash"
  fi
}

restore_extensions() {
  cp "$fixture_source/.moon/extensions.yml" "$workspace/.moon/extensions.yml"
}

expect_failure() {
  local case_name="$1"
  local target="$2"
  local expected="$3"
  local output="$evidence/failure-$case_name.log"

  cp "$fixture_source/.moon/cases/$case_name.yml" "$workspace/.moon/extensions.yml"
  if moon run "$target" --force > "$output" 2>&1; then
    fail "expected $case_name to fail for $target"
  fi
  assert_contains "$output" "$target"
  assert_contains "$output" "$expected"
  assert_not_contains "$output" 'let flake = builtins.getFlake "path:'
  restore_extensions
}

expect_current_failure() {
  local case_name="$1"
  local target="$2"
  local expected="$3"
  local output="$evidence/failure-$case_name.log"

  if moon run "$target" --force > "$output" 2>&1; then
    fail "expected $case_name to fail for $target"
  fi
  assert_contains "$output" "$target"
  assert_contains "$output" "$expected"
  assert_not_contains "$output" 'let flake = builtins.getFlake "path:'
}

log 'checking pinned Moon and standalone fixture configuration'
run_moon "$evidence/moon-version.log" --version
assert_contains "$evidence/moon-version.log" 'moon 2.4.5'
[[ "$(find "$workspace/.moon/plugins" -maxdepth 1 -type f -printf '%f\n')" == 'moon_nix_extension.wasm' ]] \
  || fail 'the fixture must stage only moon_nix_extension.wasm'
if rg --quiet '^[[:space:]]*nix:|moon_nix_toolchain|toolchains:[[:space:]]*\[[^]]*nix' \
  "$workspace/.moon/toolchains.yml" \
  "$workspace/.moon/extensions.yml" \
  "$workspace/.moon/extensions-base.yml"; then
  fail 'the standalone fixture must not configure the Nix toolchain plugin'
fi

log 'checking Moon native toolchain detection'
run_moon "$evidence/node-task.json" task node-basic:detected-node --json
run_moon "$evidence/polyglot-task.json" task polyglot:detected-polyglot --json
assert_json "$evidence/node-task.json" '.toolchains == ["javascript", "npm", "node"]'
assert_json "$evidence/polyglot-task.json" '.toolchains == ["go", "javascript", "npm", "node"]'

log 'checking composition, overrides, installables, and Proto opt-out'
run_moon "$evidence/positive-matrix.log" run \
  node-basic:detected-node \
  polyglot:detected-polyglot \
  overrides:project-append \
  overrides:task-append \
  overrides:task-replace \
  legacy:project-replace \
  special-compiler:project-shell \
  special-compiler:bootstrap \
  special-compiler:supersede \
  system:scoped-only \
  --force \
  --summary normal \
  --log debug \
  --log-file "$evidence/positive-trace.log"
for marker in \
  component:general \
  component:go \
  component:node \
  component:node20 \
  component:node24 \
  component:postgresql \
  component:protobuf \
  component:special-moon \
  component:special-bootstrap; do
  assert_contains "$evidence/positive-matrix.log" "$marker"
done
cp "$workspace/.moon/cache/runReport.json" "$evidence/positive-run-report.json"
assert_json "$evidence/positive-run-report.json" \
  '[.actions[] | select(.node.action == "setup-toolchain") | select(.node.params.toolchain.id == "node" or .node.params.toolchain.id == "npm" or .node.params.toolchain.id == "go")] | length >= 3 and all(.[]; .status == "skipped" and .node.params.toolchain.req == null)'
assert_not_contains "$evidence/positive-trace.log" 'moon_nix_toolchain'

run_moon "$evidence/unmapped.log" run system:unmapped --force --log debug --log-file "$evidence/unmapped-trace.log"
assert_contains "$evidence/unmapped.log" 'system-unmapped'
assert_not_contains "$evidence/unmapped.log" 'component:'
assert_not_contains "$evidence/unmapped-trace.log" 'nix develop'

MOON_NIX_WRAPPED=1 run_moon "$evidence/reentry.log" run system:reentry --force --log debug --log-file "$evidence/reentry-trace.log"
assert_contains "$evidence/reentry.log" 'sentinel-reentry'
assert_not_contains "$evidence/reentry-trace.log" 'nix develop'

log 'checking invalid configuration and Nix failure diagnostics'
outside_flake="$temp_root/outside"
mkdir -p "$outside_flake"
ln -s "$outside_flake" "$workspace/escape"
expect_failure remote-installable system:negative 'github:NixOS/nixpkgs#default'
expect_failure inline-installable system:negative 'let shell = {}; in shell'
expect_failure mixed-union system:negative 'did not match any variant'
expect_failure parent-escape system:negative 'path:../outside#moon'
expect_failure symlink-escape system:negative 'outside workspace'
expect_failure missing-lock system:negative 'has no flake.lock'
expect_failure missing-attribute system:negative 'absent'
expect_failure append-over-installable special-compiler:negative 'cannot append components'
expect_failure invalid-component system:negative 'bad component'

expect_current_failure missing-component system:missing-component 'unknown fixture component(s): missing'
expect_current_failure broken-evaluation system:broken-evaluation 'deliberate fixture Nix evaluation error'
sed -i 's/lib\.mkMoonShell = mkMoonShell;/lib = {};/' "$workspace/flake.nix"
rg -F --quiet 'lib = {};' "$workspace/flake.nix" || fail 'failed to remove mkMoonShell for the negative case'
expect_current_failure missing-mk-moon-shell system:negative 'mkMoonShell'
cp "$fixture_source/flake.nix" "$workspace/flake.nix"

restricted_path="$temp_root/restricted-path"
mkdir -p "$restricted_path"
for executable in bash git go node npm sh uname which; do
  executable_path="$(command -v "$executable")"
  ln -s "$executable_path" "$restricted_path/$executable"
done
moon_without_nix() {
  (
    export PATH="$restricted_path"
    moon "$@"
  )
}
if moon_without_nix run system:negative --force > "$evidence/failure-missing-nix-closed.log" 2>&1; then
  fail 'expected fail-closed execution without Nix to fail'
fi
assert_contains "$evidence/failure-missing-nix-closed.log" 'system:negative'
assert_contains "$evidence/failure-missing-nix-closed.log" 'nix is required'
assert_not_contains "$evidence/failure-missing-nix-closed.log" 'let flake = builtins.getFlake "path:'

cp "$fixture_source/.moon/cases/fail-open.yml" "$workspace/.moon/extensions.yml"
if ! moon_without_nix run system:negative --force > "$evidence/missing-nix-open.log" 2>&1; then
  sed -n '1,200p' "$evidence/missing-nix-open.log" >&2
  fail 'expected fail-open execution without Nix to remain unchanged'
fi
assert_contains "$evidence/missing-nix-open.log" 'negative-control'
assert_not_contains "$evidence/missing-nix-open.log" 'nix is required'
restore_extensions

log 'checking central cache ownership'
run_moon "$evidence/cache-central-first.log" run node-basic:cacheable --summary normal
central_hash="$(capture_hash node-basic:cacheable)"
capture_hash_manifest "$central_hash" "$evidence/hash-central-base.json"
assert_json "$evidence/hash-central-base.json" \
  '.[0].inputs | has("flake.nix") and has("flake.lock") and has("nix/components.nix") and has(".moon/extensions.yml") and (has("unrelated.txt") | not)'
run_moon "$evidence/cache-central-second.log" run node-basic:cacheable --summary normal
[[ "$(capture_hash node-basic:cacheable)" == "$central_hash" ]] || fail 'warm central task hash changed without an input change'
[[ "$(counter_lines "$counters/node-basic")" -eq 1 ]] || fail 'warm central task executed instead of using cache'
assert_contains "$evidence/cache-central-second.log" '(cached,'

central_runs=1
for relative_input in flake.nix flake.lock nix/components.nix .moon/extensions.yml; do
  printf '\n' >> "$workspace/$relative_input"
  output_name="${relative_input//\//-}"
  run_moon "$evidence/cache-central-$output_name.log" run node-basic:cacheable --summary normal
  changed_hash="$(capture_hash node-basic:cacheable)"
  [[ "$changed_hash" != "$central_hash" ]] || fail "$relative_input did not invalidate the central cache consumer"
  capture_hash_manifest "$changed_hash" "$evidence/hash-central-$output_name.json"
  base_input_hash="$(jq -er --arg input "$relative_input" '.[0].inputs[$input]' "$evidence/hash-central-base.json")"
  changed_input_hash="$(jq -er --arg input "$relative_input" '.[0].inputs[$input]' "$evidence/hash-central-$output_name.json")"
  [[ "$base_input_hash" != "$changed_input_hash" ]] || fail "$relative_input did not change its hash manifest entry"
  central_runs=$((central_runs + 1))
  [[ "$(counter_lines "$counters/node-basic")" -eq "$central_runs" ]] || fail "$relative_input cache miss did not execute"
  cp "$fixture_source/$relative_input" "$workspace/$relative_input"
done

printf '\n' >> "$workspace/unrelated.txt"
run_moon "$evidence/cache-unrelated.log" run node-basic:cacheable --summary normal
[[ "$(capture_hash node-basic:cacheable)" == "$central_hash" ]] || fail 'unrelated input changed the central task hash'
[[ "$(counter_lines "$counters/node-basic")" -eq "$central_runs" ]] || fail 'unrelated input caused central task execution'
assert_contains "$evidence/cache-unrelated.log" '(cached,'
cp "$fixture_source/unrelated.txt" "$workspace/unrelated.txt"

run_moon "$evidence/uncached-first.log" run node-basic:uncached --summary normal
run_moon "$evidence/uncached-second.log" run node-basic:uncached --summary normal
[[ "$(counter_lines "$counters/uncached")" -eq 2 ]] || fail 'uncached task did not execute twice'
assert_not_contains "$evidence/uncached-second.log" '(cached,'

log 'checking selected project cache ownership'
run_moon "$evidence/cache-project-first.log" run special-compiler:cacheable --summary normal
project_hash="$(capture_hash special-compiler:cacheable)"
capture_hash_manifest "$project_hash" "$evidence/hash-project-base.json"
assert_json "$evidence/hash-project-base.json" \
  '.[0].inputs | has("projects/special-compiler/flake.nix") and has("projects/special-compiler/flake.lock") and has("projects/special-compiler/nix/shells.nix")'
run_moon "$evidence/cache-project-second.log" run special-compiler:cacheable --summary normal
[[ "$(counter_lines "$counters/special-compiler")" -eq 1 ]] || fail 'warm project task executed instead of using cache'
assert_contains "$evidence/cache-project-second.log" '(cached,'

project_runs=1
for relative_input in \
  projects/special-compiler/flake.nix \
  projects/special-compiler/flake.lock \
  projects/special-compiler/nix/shells.nix; do
  printf '\n' >> "$workspace/$relative_input"
  output_name="${relative_input//\//-}"
  run_moon "$evidence/cache-project-$output_name.log" run special-compiler:cacheable --summary normal
  changed_hash="$(capture_hash special-compiler:cacheable)"
  [[ "$changed_hash" != "$project_hash" ]] || fail "$relative_input did not invalidate the project cache consumer"
  project_runs=$((project_runs + 1))
  [[ "$(counter_lines "$counters/special-compiler")" -eq "$project_runs" ]] || fail "$relative_input cache miss did not execute"

  run_moon "$evidence/cache-central-during-$output_name.log" run node-basic:cacheable --summary normal
  [[ "$(capture_hash node-basic:cacheable)" == "$central_hash" ]] || fail "$relative_input invalidated an unrelated central consumer"
  [[ "$(counter_lines "$counters/node-basic")" -eq "$central_runs" ]] || fail "$relative_input executed an unrelated central consumer"
  assert_contains "$evidence/cache-central-during-$output_name.log" '(cached,'
  cp "$fixture_source/$relative_input" "$workspace/$relative_input"
done

log 'consumer workspace integration passed'
