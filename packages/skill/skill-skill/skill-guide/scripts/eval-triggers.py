#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Run skill-triggering evals against an eval-queries.json file.

Usage:
    eval-triggers.py <queries.json> <skill_name> [split] [options]
        skill_name = bare ("git-commit") or plugin-namespaced ("myplugin:git-commit")
        split      = train | validation | all   (default: all)

Options (flag overrides env; env keeps the loop/CI ergonomics):
    --runs N             / RUNS=N            runs per query (default: 3)
    --threshold F        / THRESHOLD=F       trigger-rate cutoff for a pass (default: 0.5)
    --model M            / CLAUDE_MODEL=M    model for `claude --model` — use `haiku` to keep cost low
    --disallowed-tools L / DISALLOWED_TOOLS=L  tools blocked during the eval (default blocks
                                             everything but Skill, so non-triggering runs stay cheap)
    --max-budget-usd N   / MAX_BUDGET_USD=N  hard per-run spend cap (default: 0.10; 0 disables)

Cost: a run where the skill does NOT fire would otherwise execute the whole task.
The default tool-blocking keeps those short, `--model haiku` makes them cheap, and
`--max-budget-usd` caps each run. While iterating, prefer `--runs 1` on the train split.

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

import argparse
import json
import os
import shutil
import signal
import subprocess
import sys
from pathlib import Path


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        description="Run skill-triggering evals against an eval-queries.json file.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    p.add_argument("queries", help="path to eval-queries.json")
    p.add_argument(
        "skill_name",
        help="bare ('git-commit') or plugin-namespaced ('plugin:git-commit')",
    )
    p.add_argument(
        "split",
        nargs="?",
        choices=("train", "validation", "all"),
        default="all",
        help="which split to run (default: all)",
    )
    p.add_argument(
        "--runs",
        type=int,
        default=int(os.environ.get("RUNS", "3")),
        help="runs per query — model is nondeterministic (env RUNS, default 3)",
    )
    p.add_argument(
        "--threshold",
        type=float,
        default=float(os.environ.get("THRESHOLD", "0.5")),
        help="trigger-rate cutoff for a passing query (env THRESHOLD, default 0.5)",
    )
    p.add_argument(
        "--model",
        default=os.environ.get("CLAUDE_MODEL", "haiku"),
        help="model alias/id passed to `claude --model` (env CLAUDE_MODEL, default haiku)",
    )
    p.add_argument(
        "--disallowed-tools",
        default=os.environ.get(
            "DISALLOWED_TOOLS",
            "Bash,Edit,Write,NotebookEdit,WebFetch,WebSearch,Read,Glob,Grep,Task,TodoWrite",
        ),
        help="comma-separated tools blocked during the eval (env DISALLOWED_TOOLS); the "
        "default blocks everything but Skill so non-triggering runs stay short and cheap",
    )
    p.add_argument(
        "--max-budget-usd",
        type=float,
        default=float(os.environ.get("MAX_BUDGET_USD", "0.10")),
        help="hard per-run spend cap passed to `claude --max-budget-usd` "
        "(env MAX_BUDGET_USD, default 0.10; 0 disables)",
    )
    return p


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
    args = build_parser().parse_args(argv)
    queries_file = Path(args.queries)
    skill_name = args.skill_name
    split = args.split

    if not queries_file.is_file():
        sys.stderr.write(f"Error: queries file not found: {queries_file}\n")
        return 2
    if not shutil.which("claude"):
        sys.stderr.write("Error: 'claude' CLI not found in PATH\n")
        return 2

    runs = args.runs
    threshold = args.threshold
    claude_model = args.model
    disallowed = args.disallowed_tools
    budget = args.max_budget_usd

    short = skill_name.rsplit(":", 1)[-1]

    claude_args = ["-p", "--output-format", "stream-json", "--verbose"]
    if claude_model:
        claude_args.extend(["--model", claude_model])
    if disallowed:
        # Use --opt=val to avoid the variadic parser swallowing the prompt
        claude_args.append(f"--disallowedTools={disallowed}")
    if budget and budget > 0:
        # Hard per-run ceiling so a non-triggering run can't execute the whole task
        claude_args.extend(["--max-budget-usd", str(budget)])

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
        f"budget/run: {('$' + str(budget)) if budget and budget > 0 else 'none'}  "
        f"disallowed: {disallowed}",
        file=sys.stderr,
    )
    print(f"passed: {passed} / {total}   failed: {failed}", file=sys.stderr)
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
