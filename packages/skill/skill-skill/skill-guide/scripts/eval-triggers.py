#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Run skill-triggering evals against an eval-queries.json file.

Usage:
    eval-triggers.py <queries.json> <skill_name> [split]
        skill_name = bare ("git-commit") or plugin-namespaced ("xonovex-git:git-commit")
        split      = train | validation | all   (default: all)

Env:
    RUNS=<n>               runs per query (default: 3)
    THRESHOLD=<f>          trigger-rate cutoff for a passing query (default: 0.5)
    CLAUDE_MODEL=<m>       model alias or full id passed to `claude --model`
                           (e.g. haiku, sonnet, opus). Unset = claude default.
    DISALLOWED_TOOLS=<l>   comma-separated tools blocked during the eval
                           (default: Bash,Edit,Write,NotebookEdit,WebFetch).

Safety model:
    1. Each query launches `claude -p --output-format stream-json --verbose`
       with side-effect tools disallowed. State-mutating tools are blocked
       even before the kill fires.
    2. The runner reads the stream line-by-line. The instant a `Skill` tool_use
       OR a `Skill` permission_denial matching the target skill is observed,
       the claude process is terminated — no further tools dispatch.
    3. Skill names match three ways: exact, last-segment-after-colon, or
       ":<short>" suffix.

Output: one JSON object per query on stdout, plus a summary on stderr.

Cross-platform: works on macOS, Linux, and Windows (where `claude` CLI is installed).
"""

from __future__ import annotations

import json
import os
import shutil
import signal
import subprocess
import sys
from pathlib import Path


def usage() -> None:
    sys.stderr.write(
        "Usage: eval-triggers.py <queries.json> <skill_name> [split]\n"
    )
    sys.exit(2)


def match_skill(skill_field: object, target: str, short: str) -> bool:
    if not isinstance(skill_field, str):
        return False
    return (
        skill_field == target
        or skill_field == short
        or skill_field.endswith(":" + short)
    )


def check_line(line: str, target: str, short: str) -> bool:
    """Return True if the JSON line indicates a matching Skill call."""
    try:
        obj = json.loads(line)
    except json.JSONDecodeError:
        return False

    # Check message.content[].type == "tool_use" and .name == "Skill"
    message = obj.get("message")
    if isinstance(message, dict):
        for item in message.get("content", []) or []:
            if (
                isinstance(item, dict)
                and item.get("type") == "tool_use"
                and item.get("name") == "Skill"
            ):
                input_field = item.get("input")
                if isinstance(input_field, dict):
                    if match_skill(input_field.get("skill"), target, short):
                        return True

    # Check permission_denials[].tool_name == "Skill"
    for denial in obj.get("permission_denials", []) or []:
        if isinstance(denial, dict) and denial.get("tool_name") == "Skill":
            tool_input = denial.get("tool_input")
            if isinstance(tool_input, dict):
                if match_skill(tool_input.get("skill"), target, short):
                    return True

    return False


def check_triggered(
    query: str, claude_args: list[str], target: str, short: str
) -> bool:
    """Return True if a matching Skill call appears in the stream.

    Terminates the claude process on first match — no further tools fire.
    """
    proc = subprocess.Popen(
        ["claude", *claude_args, query],
        stdin=subprocess.DEVNULL,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
        text=True,
        bufsize=1,  # line-buffered
    )
    matched = False
    try:
        assert proc.stdout is not None
        for line in proc.stdout:
            line = line.strip()
            if not line:
                continue
            if check_line(line, target, short):
                matched = True
                break
    finally:
        # Terminate forcefully — SIGKILL on POSIX, TerminateProcess on Windows
        if proc.poll() is None:
            if sys.platform == "win32":
                proc.terminate()
            else:
                try:
                    proc.send_signal(signal.SIGKILL)
                except ProcessLookupError:
                    pass
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.wait()

    return matched


def main(argv: list[str]) -> int:
    if len(argv) < 3 or len(argv) > 4:
        usage()
        return 2

    queries_file = Path(argv[1])
    skill_name = argv[2]
    split = argv[3] if len(argv) == 4 else "all"

    if not queries_file.is_file():
        sys.stderr.write(f"Error: queries file not found: {queries_file}\n")
        return 2
    if not shutil.which("claude"):
        sys.stderr.write("Error: 'claude' CLI not found in PATH\n")
        return 2

    runs = int(os.environ.get("RUNS", "3"))
    threshold = float(os.environ.get("THRESHOLD", "0.5"))
    claude_model = os.environ.get("CLAUDE_MODEL", "")
    disallowed = os.environ.get(
        "DISALLOWED_TOOLS", "Bash,Edit,Write,NotebookEdit,WebFetch"
    )

    short = skill_name.rsplit(":", 1)[-1]

    claude_args = ["-p", "--output-format", "stream-json", "--verbose"]
    if claude_model:
        claude_args.extend(["--model", claude_model])
    if disallowed:
        # Use --opt=val to avoid the variadic parser swallowing the prompt
        claude_args.append(f"--disallowedTools={disallowed}")

    try:
        queries = json.loads(queries_file.read_text(encoding="utf-8"))
    except json.JSONDecodeError as e:
        sys.stderr.write(f"Error: invalid JSON in {queries_file}: {e}\n")
        return 2
    if not isinstance(queries, list):
        sys.stderr.write(f"Error: {queries_file} must contain a top-level array\n")
        return 2

    if split != "all":
        queries = [q for q in queries if q.get("split") == split]

    passed = 0
    failed = 0
    total = 0

    for q in queries:
        if not isinstance(q, dict):
            continue
        query = q.get("query", "")
        should_trigger = bool(q.get("should_trigger", False))
        rationale = q.get("rationale", "")

        triggers = 0
        for _ in range(runs):
            if check_triggered(query, claude_args, skill_name, short):
                triggers += 1

        rate = triggers / runs if runs else 0.0
        triggered_majority = rate >= threshold
        passes = triggered_majority == should_trigger

        total += 1
        if passes:
            passed += 1
        else:
            failed += 1

        result = {
            "query": query,
            "should_trigger": should_trigger,
            "triggers": triggers,
            "runs": runs,
            "trigger_rate": round(rate, 3),
            "pass": passes,
            "rationale": rationale,
        }
        print(json.dumps(result, ensure_ascii=False))

    print("---", file=sys.stderr)
    print(
        f"skill: {skill_name}  split: {split}  runs: {runs}  "
        f"threshold: {threshold}  model: {claude_model or '<default>'}  "
        f"disallowed: {disallowed}",
        file=sys.stderr,
    )
    print(f"passed: {passed} / {total}   failed: {failed}", file=sys.stderr)
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
