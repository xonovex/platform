#!/usr/bin/env bash
set -euo pipefail

decision_service_url="${DECISION_SERVICE_URL:-http://127.0.0.1:8787}"
policy_version="${GOVERNANCE_POLICY_VERSION:-governance-policy/1}"
mandatory="${GOVERNANCE_MANDATORY:-1}"

for dependency in curl jq sha256sum; do
  if ! command -v "$dependency" >/dev/null 2>&1; then
    printf 'Governance hook dependency unavailable: %s.\n' "$dependency" >&2
    exit 2
  fi
done

event="$(cat)"

if ! jq -e '
  type == "object"
  and (.session_id | type == "string" and length > 0)
  and (.tool_use_id | type == "string" and length > 0)
  and (.tool_name == "Edit" or .tool_name == "Write")
  and ((.tool_input.file_path // .tool_input.path) | type == "string" and length > 0)
' <<<"$event" >/dev/null; then
  printf 'Governance hook received an invalid or unsupported PreToolUse event.\n' >&2
  exit 2
fi

session_id="$(jq -r '.session_id' <<<"$event")"
tool_use_id="$(jq -r '.tool_use_id' <<<"$event")"
tool_name="$(jq -r '.tool_name' <<<"$event")"
path="$(jq -r '.tool_input.file_path // .tool_input.path // ""' <<<"$event")"
correlation_id="${session_id}:${tool_use_id}"
subject_reference="tool:${tool_name}:${path}"
subject_revision="tool-use:${tool_use_id}"

if [ "$mandatory" = "1" ]; then
  enforcement="mandatory"
else
  enforcement="advisory"
fi

request="$(jq -cn \
  --arg correlation "$correlation_id" \
  --arg subject "$subject_reference" \
  --arg revision "$subject_revision" \
  --arg policy "$policy_version" \
  --arg enforcement "$enforcement" \
  --arg path "$path" \
  '{
    apiVersion: "governance.xonovex.com/v1alpha1",
    correlationId: $correlation,
    subject: {reference: $subject, revision: $revision},
    policy: {version: $policy, enforcement: $enforcement},
    operation: {kind: "protected-path", input: {path: $path}}
  }')"
operation_digest="sha256:$(jq -cS '.operation' <<<"$request" | sha256sum | cut -d' ' -f1)"
protected_targets_digest="sha256:$(printf '[]' | sha256sum | cut -d' ' -f1)"

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

if ! jq -e \
  --arg api 'governance.xonovex.com/v1alpha1' \
  --arg correlation "$correlation_id" \
  --arg subject "$subject_reference" \
  --arg revision "$subject_revision" \
  --arg policy "$policy_version" \
  --arg evaluator 'governance-evaluator/1' \
  --arg digest "$operation_digest" \
  --arg protectedTargetsDigest "$protected_targets_digest" \
  '.apiVersion == $api
   and .correlationId == $correlation
   and .subjectReference == $subject
   and .subjectRevision == $revision
   and .policyVersion == $policy
   and .evaluatorVersion == $evaluator
   and .operationDigest == $digest
   and .protectedTargetsDigest == $protectedTargetsDigest
   and (.evidenceReference | type == "string" and test("#sha256:[0-9a-f]{64}$"))
   and (.decision == "allow" or .decision == "deny" or .decision == "observe")
   and (if .decision == "allow"
        then (has("failureCode") | not)
        else (.failureCode | type == "string" and length > 0)
        end)' \
  <<<"$verdict" >/dev/null; then
  printf 'Governance decision response contract mismatch; mandatory PreToolUse control fails closed.\n' >&2
  exit 2
fi

decision="$(jq -r '.decision' <<<"$verdict")"
failure_code="$(jq -r '.failureCode // empty' <<<"$verdict")"
evidence_reference="$(jq -r '.evidenceReference' <<<"$verdict")"

enforcement_signal="$(jq -cn \
  --arg correlation "$correlation_id" \
  --arg api 'governance.xonovex.com/v1alpha1' \
  --arg subject "$subject_reference" \
  --arg revision "$subject_revision" \
  --arg outcome "$decision" \
  --arg failure "$failure_code" \
  --arg policy "$policy_version" \
  --arg evaluator 'governance-evaluator/1' \
  --arg digest "$operation_digest" \
  --arg protectedTargetsDigest "$protected_targets_digest" \
  --arg decisionEvidence "$evidence_reference" \
  '{
     apiVersion: $api,
     correlationId: $correlation,
     subjectReference: $subject,
     subjectRevision: $revision,
     outcome: $outcome,
     policyVersion: $policy,
     evaluatorVersion: $evaluator,
     operationDigest: $digest,
     protectedTargetsDigest: $protectedTargetsDigest,
     decisionEvidenceReference: $decisionEvidence,
     enforcementPoint: "agent-harness:claude-code:PreToolUse"
   }
   + if $failure == "" then {} else {failureCode: $failure} end')"
if ! enforcement_receipt="$(curl --fail --silent --show-error \
  --connect-timeout 2 \
  --max-time 5 \
  --header 'content-type: application/json' \
  --data "$enforcement_signal" \
  "${decision_service_url%/}/v1/enforcements")"; then
  if [ "$mandatory" = "1" ]; then
    printf 'Governance enforcement evidence unavailable; mandatory PreToolUse control fails closed.\n' >&2
    exit 2
  fi
  printf 'Governance advisory observed unavailable enforcement evidence.\n' >&2
elif ! jq -e \
  --arg api 'governance.xonovex.com/v1alpha1' \
  --arg correlation "$correlation_id" \
  '.apiVersion == $api
   and .correlationId == $correlation
   and .status == "recorded"
   and (.evidenceReference | type == "string" and test("#sha256:[0-9a-f]{64}$"))' \
  <<<"$enforcement_receipt" >/dev/null; then
  if [ "$mandatory" = "1" ]; then
    printf 'Governance enforcement receipt was invalid; mandatory PreToolUse control fails closed.\n' >&2
    exit 2
  fi
  printf 'Governance advisory observed an invalid enforcement receipt.\n' >&2
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
