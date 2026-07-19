#!/usr/bin/env bash
set -euo pipefail

decision_service_url="${DECISION_SERVICE_URL:-http://127.0.0.1:8787}"
policy_version="${GOVERNANCE_POLICY_VERSION:-governance-policy/1}"
mandatory="${GOVERNANCE_MANDATORY:-1}"
event="$(cat)"

session_id="$(jq -r '.session_id // "unknown-session"' <<<"$event")"
tool_use_id="$(jq -r '.tool_use_id // "unknown-tool-use"' <<<"$event")"
tool_name="$(jq -r '.tool_name // "unknown-tool"' <<<"$event")"
path="$(jq -r '.tool_input.file_path // .tool_input.path // ""' <<<"$event")"
correlation_id="${session_id}:${tool_use_id}"

if [ "$mandatory" = "1" ]; then
  enforcement="mandatory"
else
  enforcement="advisory"
fi

request="$(jq -cn \
  --arg correlation "$correlation_id" \
  --arg subject "tool:${tool_name}:${path}" \
  --arg policy "$policy_version" \
  --arg enforcement "$enforcement" \
  --arg path "$path" \
  '{
    apiVersion: "governance.xonovex.com/v1alpha1",
    correlationId: $correlation,
    subject: {reference: $subject},
    policy: {version: $policy, enforcement: $enforcement},
    operation: {kind: "protected-path", input: {path: $path}}
  }')"

if ! verdict="$(curl --fail --silent --show-error \
  --connect-timeout 2 \
  --max-time 5 \
  --header 'content-type: application/json' \
  --data "$request" \
  "${decision_service_url%/}/v1/decisions")"; then
  if [ "$mandatory" = "1" ]; then
    printf 'Governance decision unavailable; mandatory PreToolUse control fails closed.\n' >&2
    exit 2
  fi
  printf 'Governance decision unavailable; advisory PreToolUse control observed the outage.\n' >&2
  exit 0
fi

decision="$(jq -r '.decision // "deny"' <<<"$verdict")"
failure_code="$(jq -r '.failureCode // empty' <<<"$verdict")"
evidence_reference="$(jq -r '.evidenceReference // "unavailable"' <<<"$verdict")"

if [ "$decision" = "deny" ] && [ -z "$failure_code" ]; then
  failure_code="governance-decision-invalid"
fi

enforcement_signal="$(jq -cn \
  --arg correlation "$correlation_id" \
  --arg outcome "$decision" \
  --arg failure "$failure_code" \
  '{correlationId: $correlation, outcome: $outcome}
   + if $failure == "" then {} else {failureCode: $failure} end')"
if ! curl --fail --silent --show-error \
  --connect-timeout 2 \
  --max-time 5 \
  --header 'content-type: application/json' \
  --data "$enforcement_signal" \
  "${decision_service_url%/}/v1/enforcements" >/dev/null; then
  if [ "$mandatory" = "1" ]; then
    printf 'Governance enforcement evidence unavailable; mandatory PreToolUse control fails closed.\n' >&2
    exit 2
  fi
  printf 'Governance advisory observed unavailable enforcement evidence.\n' >&2
fi

case "$decision" in
allow)
  exit 0
  ;;
observe)
  printf 'Governance advisory observed %s; evidence %s.\n' "$failure_code" "$evidence_reference" >&2
  exit 0
  ;;
deny)
  printf 'Governance denied the tool call: %s; evidence %s.\n' "$failure_code" "$evidence_reference" >&2
  exit 2
  ;;
*)
  printf 'Governance decision response was invalid; mandatory PreToolUse control fails closed.\n' >&2
  exit 2
  ;;
esac
