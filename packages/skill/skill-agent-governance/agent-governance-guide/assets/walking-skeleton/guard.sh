#!/usr/bin/env bash
# Deterministic policy guard for the walking skeleton.
# Reads a JSON tool event on stdin and decides allow/deny with an explanation.
#
# Environment contract (all optional):
#   PROTECTED_PATTERN   ERE of protected paths      (default: '(^|/)secrets/|\.key$')
#   MANDATORY           1 = fail closed on outage   (default: 1)
#   POLICY_UNAVAILABLE  1 = simulate decision-point outage
#   EXCEPTION_ID        exception record id weakening this control
#   EXCEPTION_EXPIRY    ISO date the exception expires (required with EXCEPTION_ID)
#   POLICY_VERSION      policy identity              (default: skeleton-policy/1)
#
# Output: one JSON decision on stdout. Exit 0 = allow, 2 = deny.
set -euo pipefail

PROTECTED_PATTERN="${PROTECTED_PATTERN:-(^|/)secrets/|\.key$}"
MANDATORY="${MANDATORY:-1}"
POLICY_UNAVAILABLE="${POLICY_UNAVAILABLE:-0}"
POLICY_VERSION="${POLICY_VERSION:-skeleton-policy/1}"

event="$(cat)"
tool="$(jq -r '.tool // "unknown"' <<<"$event")"
path="$(jq -r '.path // ""' <<<"$event")"
event_id="$(jq -r '.event_id // "none"' <<<"$event")"

emit() {
  # emit <decision> <reason> — compact JSON so decisions are JSONL-safe
  jq -cn --arg d "$1" --arg r "$2" --arg p "$POLICY_VERSION" \
    --arg t "$tool" --arg pa "$path" --arg e "$event_id" \
    '{decision:$d, reason:$r, policy_version:$p, tool:$t, path:$pa, event_id:$e}'
}

# Decision-point outage: mandatory controls fail closed, advisory ones observe.
if [ "$POLICY_UNAVAILABLE" = "1" ]; then
  if [ "$MANDATORY" = "1" ]; then
    emit deny "policy decision point unavailable — mandatory control fails closed; retry or invoke break-glass explicitly"
    exit 2
  fi
  emit observe "policy decision point unavailable — advisory control proceeds as observe; staleness reported"
  exit 0
fi

if [[ "$path" =~ $PROTECTED_PATTERN ]]; then
  # An exception may weaken the control only while unexpired.
  if [ -n "${EXCEPTION_ID:-}" ]; then
    expiry="${EXCEPTION_EXPIRY:?EXCEPTION_ID requires EXCEPTION_EXPIRY}"
    today="$(date -u +%F)"
    if [[ "$expiry" > "$today" || "$expiry" == "$today" ]]; then
      emit allow "protected path permitted under exception ${EXCEPTION_ID} (expires ${expiry}); compensating controls apply"
      exit 0
    fi
    emit deny "exception ${EXCEPTION_ID} expired ${expiry} — control 'protected-path' applies unweakened (no grace period); renew via explicit exception review"
    exit 2
  fi
  emit deny "tool '${tool}' on protected path '${path}' violates control 'protected-path' (${POLICY_VERSION}); remediation: use a non-protected location or request an exception"
  exit 2
fi

emit allow "no protected pattern matched; control 'protected-path' satisfied"
exit 0
