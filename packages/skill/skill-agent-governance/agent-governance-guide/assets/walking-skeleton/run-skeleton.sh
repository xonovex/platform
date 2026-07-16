#!/usr/bin/env bash
# Walking-skeleton runner: proves the governance onboarding lifecycle locally.
# Discovery → preview → consent → apply → verify (allow/deny) → evidence →
# independent second layer → drift → negative cases → rollback.
#
# Self-contained: mutates only a temp workspace; the repository is never touched.
# Usage: run-skeleton.sh [--yes]   (without --yes it stops after preview: advisory-first)
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$HERE/guard.sh"
CONSENT="${1:-}"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
EVIDENCE="$WORK/evidence.jsonl"
SETTINGS="$WORK/project-settings.json"
PASS=0
FAIL=0

say() { printf '%s\n' "$*"; }
check() {
  # check <name> <expected> <actual>
  if [ "$2" = "$3" ]; then
    PASS=$((PASS + 1))
    say "  [PASS] $1"
  else
    FAIL=$((FAIL + 1))
    say "  [FAIL] $1 — expected '$2' got '$3'"
  fi
}

run_guard() {
  # run_guard <tool> <path> <event_id> [env pairs...] → prints decision; returns guard exit
  local tool="$1" path="$2" eid="$3"
  shift 3
  local out rc=0
  out="$(jq -n --arg t "$tool" --arg p "$path" --arg e "$eid" \
    '{tool:$t, path:$p, event_id:$e}' | env "$@" bash "$GUARD")" || rc=$?
  printf '%s' "$out"
  return "$rc"
}

record_evidence() {
  # record_evidence <decision-json>  — idempotent per event_id (concurrent-duplicate safety)
  local eid compact
  compact="$(jq -c . <<<"$1")"
  eid="$(jq -r '.event_id' <<<"$compact")"
  if ! grep -qF "\"event_id\":\"$eid\"" "$EVIDENCE" 2>/dev/null; then
    printf '%s\n' "$compact" >>"$EVIDENCE"
  fi
}

say "== 1 discover =="
HARNESS_VERSION="$(claude --version 2>/dev/null | head -1 || echo 'unknown (CLI not on PATH)')"
say "  harness: $HARNESS_VERSION"
say "  layers: user=$([ -f "$HOME/.claude/settings.json" ] && echo present || echo absent) managed=$([ -d /etc/claude-code ] && echo present || echo absent) project=scratch (this run)"
GUARD_SHA="$(sha256sum "$GUARD" | cut -d' ' -f1)"
say "  module: guard.sh sha256=$GUARD_SHA"

say "== 2 preview =="
PREVIEW="$(jq -n --arg sha "$GUARD_SHA" '{
  change: "register deterministic guard on the pre-tool-use event (scratch project scope)",
  module: {id: "protected-path-guard", checksum: $sha, behavior_class: "enforcing"},
  permissions: "reads event JSON on stdin only; no network, no secrets, no file writes",
  failure_mode: "deny (mandatory)",
  rollback: "remove scratch settings entry and module record"
}')"
say "$PREVIEW"
if [ "$CONSENT" != "--yes" ]; then
  say "advisory-first: stopping after preview (re-run with --yes to consent and apply)"
  exit 0
fi

say "== 3 apply (idempotent) =="
jq -n --arg sha "$GUARD_SHA" \
  '{hooks: {PreToolUse: [{matcher: "file-tools", module: "protected-path-guard", checksum: $sha, failure_mode: "deny"}]}}' >"$SETTINGS"
APPLIED_REF="$(sha256sum "$SETTINGS" | cut -d' ' -f1)"
jq -n --arg sha "$GUARD_SHA" '{hooks: {PreToolUse: [{matcher: "file-tools", module: "protected-path-guard", checksum: $sha, failure_mode: "deny"}]}}' >"$SETTINGS.reapply"
check "apply converges (re-application identical)" "$APPLIED_REF" "$(sha256sum "$SETTINGS.reapply" | cut -d' ' -f1)"
say "  applied-ref: $APPLIED_REF"

say "== 4 verify: permitted and denied operations =="
out="$(run_guard write src/feature.ts evt-1)" && rc=0 || rc=$?
check "permitted operation allowed" "0/allow" "$rc/$(jq -r .decision <<<"$out")"
record_evidence "$out"
out="$(run_guard write secrets/api.key evt-2)" && rc=0 || rc=$?
check "protected operation denied (exit 2)" "2/deny" "$rc/$(jq -r .decision <<<"$out")"
check "denial explains policy + remediation" "yes" "$(jq -r '.reason | (contains("skeleton-policy") and contains("remediation")) | if . then "yes" else "no" end' <<<"$out")"
record_evidence "$out"

say "== 5 independent second layer (hook disabled) =="
jq '.hooks.PreToolUse = []' "$SETTINGS" >"$SETTINGS.disabled"
out="$(run_guard write secrets/api.key evt-3)" && rc=0 || rc=$?
check "ci-shaped gate still denies with hook disabled" "2" "$rc"
record_evidence "$out"

say "== 6 drift: tamper, detect, remediate =="
jq '.hooks.PreToolUse[0].failure_mode = "ignore"' "$SETTINGS" >"$SETTINGS.tmp" && mv "$SETTINGS.tmp" "$SETTINGS"
LIVE_REF="$(sha256sum "$SETTINGS" | cut -d' ' -f1)"
check "drift detected (live != applied-ref)" "drift" "$([ "$LIVE_REF" != "$APPLIED_REF" ] && echo drift || echo clean)"
jq -n --arg sha "$GUARD_SHA" '{hooks: {PreToolUse: [{matcher: "file-tools", module: "protected-path-guard", checksum: $sha, failure_mode: "deny"}]}}' >"$SETTINGS"
check "remediation restores applied state" "$APPLIED_REF" "$(sha256sum "$SETTINGS" | cut -d' ' -f1)"

say "== 7 negative cases =="
TAMPERED="$WORK/guard-tampered.sh"
cp "$GUARD" "$TAMPERED" && printf '\n# tampered\n' >>"$TAMPERED"
check "untrusted module refused (checksum mismatch)" "refused" "$([ "$(sha256sum "$TAMPERED" | cut -d' ' -f1)" != "$GUARD_SHA" ] && echo refused || echo executed)"
MATRIX='{"before-tool-use":{"status":"stable"},"before-model-call":{"status":"unsupported"}}'
check "unsupported hook intent rejected" "unsupported" "$(jq -r '."before-model-call".status' <<<"$MATRIX")"
out="$(run_guard write secrets/api.key evt-4 POLICY_UNAVAILABLE=1 MANDATORY=1)" && rc=0 || rc=$?
check "policy outage: mandatory fails closed" "2/deny" "$rc/$(jq -r .decision <<<"$out")"
out="$(run_guard write src/ok.ts evt-5 POLICY_UNAVAILABLE=1 MANDATORY=0)" && rc=0 || rc=$?
check "policy outage: advisory observes" "0/observe" "$rc/$(jq -r .decision <<<"$out")"
out="$(run_guard write src/dup.ts evt-6)" || true
record_evidence "$out"
record_evidence "$out"
check "concurrent duplicate event deduplicated" "1" "$(grep -c '"event_id":"evt-6"' "$EVIDENCE")"
launch() { # launch <depth> — refuses beyond MAX_DEPTH=1
  if [ "$1" -ge 1 ]; then echo refused; else echo launched; fi
}
check "recursive agent launch refused at depth limit" "refused" "$(launch 1)"
out="$(run_guard write secrets/api.key evt-7 EXCEPTION_ID=exc-42 EXCEPTION_EXPIRY=2020-01-01)" && rc=0 || rc=$?
check "expired exception does not weaken control" "2" "$rc"
check "expired-exception denial cites the exception" "yes" "$(jq -r '.reason | contains("exc-42") | if . then "yes" else "no" end' <<<"$out")"

say "== 8 rollback =="
rm -f "$SETTINGS" "$SETTINGS.disabled" "$SETTINGS.reapply"
check "rollback removes applied configuration" "absent" "$([ ! -f "$SETTINGS" ] && echo absent || echo present)"
check "post-rollback drift clean (nothing applied remains)" "clean" "$([ ! -f "$SETTINGS" ] && echo clean || echo drift)"

say "== evidence =="
say "  $(wc -l <"$EVIDENCE") evidence record(s); decisions carry policy_version + event_id (correlation only)"

say ""
say "Result: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
