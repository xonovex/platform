#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
guide_dir="$(cd "$script_dir/../.." && pwd)"
repository_root="$(cd "$guide_dir/../../../.." && pwd)"
decision_service="$guide_dir/scripts/decision-service.ts"
hook_source="$repository_root/packages/skill/skill-claude-code/code-harness-guide/scripts/governance-pre-tool-use.sh"
settings_source="$repository_root/packages/skill/skill-claude-code/code-harness-guide/assets/pre-tool-use-settings.json"
capability_probe="$repository_root/packages/skill/skill-claude-code/code-harness-guide/assets/pre-tool-use-probe-2026-07-19.json"
command_source="$repository_root/packages/command/command-workflow/commands"
port="${DECISION_SERVICE_PORT:-18788}"
decider="${LIVE_PROBE_DECIDER:-}"
workspace="$(mktemp -d)"
evidence="$workspace/verdicts.jsonl"
telemetry="$workspace/telemetry.jsonl"
service_log="$workspace/decision-service.log"
service_pid=""

die() {
  printf 'ERROR: %s\n' "$1" >&2
  exit "${2:-1}"
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1" 127
}

cleanup() {
  local code=$?
  if [ -n "$service_pid" ]; then
    kill "$service_pid" 2>/dev/null || true
    wait "$service_pid" 2>/dev/null || true
  fi
  if [ -n "${LIVE_PROBE_OUTPUT_DIR:-}" ]; then
    mkdir -p "$LIVE_PROBE_OUTPUT_DIR"
    cp "$workspace"/*.json "$LIVE_PROBE_OUTPUT_DIR/" 2>/dev/null || true
    cp "$evidence" "$LIVE_PROBE_OUTPUT_DIR/verdicts.jsonl" 2>/dev/null || true
    cp "$telemetry" "$LIVE_PROBE_OUTPUT_DIR/telemetry.jsonl" 2>/dev/null || true
    cp "$service_log" "$LIVE_PROBE_OUTPUT_DIR/decision-service.log" 2>/dev/null || true
  fi
  rm -rf "$workspace"
  exit "$code"
}

trap cleanup EXIT INT TERM

[ -n "$decider" ] || die 'set LIVE_PROBE_DECIDER to the accountable maintainer identity running this live probe'
require_command curl
require_command git
require_command jq
require_command node
require_command npx

jq -e '.nativeEvent == "PreToolUse" and .denied.fileAbsent == true and .denied.failureCode == "protected-path-denied"' "$capability_probe" >/dev/null || die 'the Phase 2 native capability probe is missing or did not prove blocking'

mkdir -p "$workspace/.claude/commands/xonovex-workflow" "$workspace/.claude/hooks" "$workspace/lifecycle"
cp "$hook_source" "$workspace/.claude/hooks/governance-pre-tool-use.sh"
chmod 0755 "$workspace/.claude/hooks/governance-pre-tool-use.sh"
cp "$settings_source" "$workspace/.claude/settings.json"
cp "$command_source/discovery-run.md" "$workspace/.claude/commands/xonovex-workflow/discovery-run.md"
cp "$command_source/acceptance-validate.md" "$workspace/.claude/commands/xonovex-workflow/acceptance-validate.md"
cp "$command_source/integration-validate.md" "$workspace/.claude/commands/xonovex-workflow/integration-validate.md"

revision="$(git -C "$repository_root" rev-parse HEAD)"
policy_version="governance-policy/1"
author="agent:claude-live-probe"
service_url="http://127.0.0.1:${port}"

DECISION_SERVICE_PORT="$port" DECISION_EVIDENCE_PATH="$evidence" DECISION_TELEMETRY_PATH="$telemetry" node "$decision_service" >"$service_log" 2>&1 &
service_pid=$!
for _ in $(seq 1 30); do
  if curl --fail --silent "$service_url/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 0.2
done
curl --fail --silent "$service_url/healthz" >/dev/null || die 'decision service did not become ready'

claude_command=(npx --yes @anthropic-ai/claude-code@2.1.211)
runtime="$("${claude_command[@]}" --version)"

run_lifecycle_command() {
  local command_name="$1" output_file="$2" artifact_path="$3"
  local prompt="/$command_name repository@$revision. This is a harness wiring probe, not a governance attestation or external mutation. Use the Write tool exactly once to create $artifact_path containing a compact JSON observation with command, revision, and probeStatus observed. Do not use Bash, do not retry, and stop after the tool result."
  (
    cd "$workspace"
    DECISION_SERVICE_URL="$service_url" "${claude_command[@]}" -p "$prompt" \
      --tools Write \
      --allowedTools Write \
      --permission-mode bypassPermissions \
      --dangerously-skip-permissions \
      --max-turns 2 \
      --output-format json >"$output_file"
  )
  [ -f "$workspace/$artifact_path" ] || die "$command_name did not produce $artifact_path"
}

gate() {
  local gate_name="$1" gate_decider="$2" correlation_id="$3"
  local request response
  request="$(jq -cn \
    --arg correlation "$correlation_id" \
    --arg subject "deliverable:${revision}:${gate_name}" \
    --arg revision "$revision" \
    --arg policy "$policy_version" \
    --arg decider "$gate_decider" \
    --arg author "$author" \
    --arg failure "${gate_name}-independence-failed" \
    '{
      apiVersion: "governance.xonovex.com/v1alpha1",
      correlationId: $correlation,
      subject: {reference: $subject, revision: $revision},
      policy: {version: $policy, enforcement: "mandatory"},
      operation: {kind: "independence", input: {
        required: "distinct-identity", decider: $decider, author: $author, failureCode: $failure
      }}
    }')"
  response="$(curl --fail --silent --show-error \
    --header 'content-type: application/json' \
    --data "$request" \
    "$service_url/v1/decisions")"
  printf '%s\n' "$response" >"$workspace/${correlation_id}.json"
  jq -r '.decision' <<<"$response"
}

run_lifecycle_command 'xonovex-workflow:discovery-run' "$workspace/discovery.json" 'lifecycle/discovery.json'

[ "$(gate acceptance "$decider" live-acceptance)" = allow ] || die 'independent acceptance gate denied the live lifecycle'
run_lifecycle_command 'xonovex-workflow:acceptance-validate' "$workspace/acceptance.json" 'lifecycle/acceptance.json'

[ "$(gate integration "$decider" live-integration)" = allow ] || die 'independent integration gate denied the live lifecycle'
run_lifecycle_command 'xonovex-workflow:integration-validate' "$workspace/integration.json" 'lifecycle/integration.json'

[ "$(gate acceptance "$author" live-negative-self-approval)" = deny ] || die 'self-approved acceptance was not denied'
[ ! -e "$workspace/lifecycle/negative-integration.json" ] || die 'the denied negative path advanced past its gate'
jq -e '.failureCode == "acceptance-independence-failed"' "$workspace/live-negative-self-approval.json" >/dev/null || die 'negative path did not retain the exact failure code'

jq -n \
  --arg runtime "$runtime" \
  --arg revision "$revision" \
  --arg capabilityProbe "$(sha256sum "$capability_probe" | cut -d' ' -f1)" \
  --argjson verdictRecords "$(wc -l <"$evidence")" \
  '{
    mode: "live",
    runtime: $runtime,
    revision: $revision,
    capabilityProbeSha256: $capabilityProbe,
    lifecycle: ["discovery", "acceptance", "integration"],
    gates: {acceptance: "allow", integration: "allow"},
    negative: {gate: "acceptance", decision: "deny", failureCode: "acceptance-independence-failed", advanced: false},
    verdictRecords: $verdictRecords
  }'
