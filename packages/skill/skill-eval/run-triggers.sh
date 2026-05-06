#!/usr/bin/env bash
# Run skill-triggering evals against an eval-queries.json file.
#
# Usage:
#   ./run-triggers.sh <queries.json> <skill_name> [split]
#     skill_name = bare ("git-commit") or plugin-namespaced ("xonovex-workflow:git-commit")
#     split      = train | validation | all   (default: all)
#
# Env:
#   RUNS=<n>               runs per query (default: 3)
#   THRESHOLD=<f>          trigger-rate cutoff for a passing query (default: 0.5)
#   CLAUDE_MODEL=<m>       model alias or full id passed to `claude --model`
#                          (e.g. haiku, sonnet, opus). Unset = claude default.
#   DISALLOWED_TOOLS=<l>   comma-separated tools blocked during the eval
#                          (default: Bash,Edit,Write,NotebookEdit,WebFetch).
#                          The variadic form of --disallowedTools eats positional
#                          args, so we always pass it as `--disallowedTools=...`.
#
# Safety model:
#   1. Each query launches `claude -p --output-format stream-json --verbose`
#      with side-effect tools disallowed. State-mutating tools like Bash, Edit,
#      and Write are blocked even before the kill fires.
#   2. The runner reads the stream line-by-line. The instant a `Skill` tool_use
#      OR a `Skill` permission_denial matching the target skill is observed,
#      the claude process is killed with SIGKILL — no further tools dispatch.
#   3. Skill names match three ways: exact, last-segment-after-colon, or
#      ":<short>" suffix. So both "git-commit" and "xonovex-workflow:git-commit"
#      work as the SKILL_NAME argument.
#
# Output: one JSON object per query on stdout, plus a summary on stderr.

set -euo pipefail

QUERIES_FILE="${1:?Usage: $0 <queries.json> <skill_name> [split]}"
SKILL_NAME="${2:?Usage: $0 <queries.json> <skill_name> [split]}"
SPLIT="${3:-all}"
RUNS="${RUNS:-3}"
THRESHOLD="${THRESHOLD:-0.5}"
CLAUDE_MODEL="${CLAUDE_MODEL:-}"
DISALLOWED_TOOLS="${DISALLOWED_TOOLS:-Bash,Edit,Write,NotebookEdit,WebFetch}"

command -v claude >/dev/null || { echo "claude CLI not found in PATH" >&2; exit 1; }
command -v jq     >/dev/null || { echo "jq not found in PATH" >&2; exit 1; }

SKILL_SHORT="${SKILL_NAME##*:}"

claude_args=(-p --output-format stream-json --verbose)
if [[ -n "$CLAUDE_MODEL" ]]; then
  claude_args+=(--model "$CLAUDE_MODEL")
fi
if [[ -n "$DISALLOWED_TOOLS" ]]; then
  # Use --opt=val form to avoid the variadic parser swallowing the prompt.
  claude_args+=("--disallowedTools=$DISALLOWED_TOOLS")
fi

filter='.[]'
if [[ "$SPLIT" != "all" ]]; then
  filter=".[] | select(.split == \"$SPLIT\")"
fi

# Per-line jq script that emits "match" when a target Skill call (allowed or
# denied) is observed. Otherwise emits nothing.
read -r -d '' MATCH_FILTER <<'JQ' || true
  (
    [.message?.content[]?
      | select(.type == "tool_use" and .name == "Skill")
      | .input.skill],
    [.permission_denials[]?
      | select(.tool_name == "Skill")
      | .tool_input.skill]
  )
  | flatten
  | map(select(. == $skill or . == $short or endswith(":" + $short)))
  | if length > 0 then "match" else empty end
JQ

# Returns 0 if a matching Skill call appears in the stream, 1 otherwise.
# Kills the claude process via SIGKILL on first match — no further tools fire.
check_triggered() {
  local query="$1"
  local fifo
  fifo=$(mktemp -u "/tmp/skill-eval-fifo.XXXXXX")
  mkfifo "$fifo"
  trap "rm -f '$fifo'" RETURN

  claude "${claude_args[@]}" "$query" </dev/null 2>/dev/null >"$fifo" &
  local cpid=$!

  local matched=0
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    local hit
    hit=$(echo "$line" | jq -nrc --arg skill "$SKILL_NAME" --arg short "$SKILL_SHORT" "input | $MATCH_FILTER" 2>/dev/null || true)
    if [[ "$hit" == "match" ]]; then
      matched=1
      kill -KILL "$cpid" 2>/dev/null || true
      break
    fi
  done < "$fifo"

  wait "$cpid" 2>/dev/null || true
  rm -f "$fifo"
  [[ "$matched" -eq 1 ]]
}

passed=0
failed=0
total=0

while IFS= read -r row; do
  query=$(jq -r '.query' <<<"$row")
  should=$(jq -r '.should_trigger' <<<"$row")
  rationale=$(jq -r '.rationale // ""' <<<"$row")
  triggers=0

  for _ in $(seq 1 "$RUNS"); do
    if check_triggered "$query"; then
      triggers=$((triggers + 1))
    fi
  done

  rate=$(awk -v t="$triggers" -v r="$RUNS" 'BEGIN{printf "%.3f", t/r}')
  triggered_majority=$(awk -v rate="$rate" -v thr="$THRESHOLD" 'BEGIN{print (rate>=thr) ? "true" : "false"}')
  pass=$(awk -v want="$should" -v got="$triggered_majority" 'BEGIN{print (want==got) ? "true" : "false"}')

  total=$((total + 1))
  if [[ "$pass" == "true" ]]; then passed=$((passed + 1)); else failed=$((failed + 1)); fi

  jq -nc \
    --arg query "$query" \
    --argjson should "$should" \
    --argjson triggers "$triggers" \
    --argjson runs "$RUNS" \
    --argjson rate "$rate" \
    --argjson pass "$pass" \
    --arg rationale "$rationale" \
    '{query: $query, should_trigger: $should, triggers: $triggers, runs: $runs, trigger_rate: $rate, pass: $pass, rationale: $rationale}'

done < <(jq -c "$filter" "$QUERIES_FILE")

echo "---" >&2
echo "skill: $SKILL_NAME  split: $SPLIT  runs: $RUNS  threshold: $THRESHOLD  model: ${CLAUDE_MODEL:-<default>}  disallowed: $DISALLOWED_TOOLS" >&2
echo "passed: $passed / $total   failed: $failed" >&2
