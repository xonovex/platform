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
    --model M            / CLAUDE_MODEL=M    model for `claude --model`
                                               (default: claude-haiku-4-5-20251001)
    --plugin-dir PATH    / PLUGIN_DIR=PATH    target-only local plugin directory
    --max-budget-usd N   / MAX_BUDGET_USD=N  hard per-run spend cap (default/max: 0.05)

Cost: a run where the skill does NOT fire would otherwise execute the whole task.
Only Skill is exposed, negative responses are bounded, and each run has hard
budget/time/output ceilings. While iterating, prefer `--runs 1` on the train split.

Safety model:
    1. Each query launches `claude -p --output-format stream-json --verbose`
       with only the target plugin and Skill tool available.
    2. The runner reads the stream line-by-line. The instant a `Skill` tool_use
       OR a `Skill` permission_denial matching the target skill is observed,
       the claude process is terminated, no further tools dispatch.
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
import subprocess
import sys
import threading
from pathlib import Path

TRIGGER_TIMEOUT = 60
TRIGGER_OUTPUT_LIMIT = 2_000
MAX_MODEL_RUNS = 24
CLAUDE_GENERATION_MODEL = "claude-haiku-4-5-20251001"
TRIGGER_SYSTEM_PROMPT = (
    "Decide only whether the available skill applies to the user request. "
    "If it applies, invoke Skill immediately. Otherwise reply with one short sentence. "
    "Do not perform the requested task."
)
CLAUDE_SKILL_PLUGIN_FIELDS = {
    "author", "dependencies", "description", "name", "skills", "version",
}
CLAUDE_EXECUTABLE_COMPONENTS = (
    "commands", "agents", "hooks", ".mcp.json", ".lsp.json", "settings.json",
    ".claude/settings.json",
)


def read_plugin_manifest(directory: Path) -> tuple[str, list[str]]:
    manifest_path = directory / ".claude-plugin" / "plugin.json"
    try:
        value = json.loads(manifest_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        raise ValueError(f"invalid Claude plugin manifest: {manifest_path}") from error
    if not isinstance(value, dict) or not isinstance(value.get("name"), str):
        raise ValueError(f"Claude plugin manifest has no name: {manifest_path}")
    unsupported = sorted(set(value) - CLAUDE_SKILL_PLUGIN_FIELDS)
    if unsupported:
        raise ValueError(
            f"Claude eval plugin is not skill-only: {manifest_path} has "
            f"{unsupported[0]}"
        )
    skills = value.get("skills")
    if not isinstance(skills, list) or not skills or not all(
        isinstance(skill, str) for skill in skills
    ):
        raise ValueError(f"Claude plugin skills are invalid: {manifest_path}")
    dependencies = value.get("dependencies", [])
    if not isinstance(dependencies, list) or not all(
        isinstance(dependency, str) for dependency in dependencies
    ):
        raise ValueError(f"Claude plugin dependencies are invalid: {manifest_path}")
    executable = next(
        (entry for entry in CLAUDE_EXECUTABLE_COMPONENTS if (directory / entry).exists()),
        None,
    )
    if executable is not None:
        raise ValueError(
            f"Claude eval plugin is not skill-only: {directory} contains {executable}"
        )
    return value["name"], dependencies


def resolve_plugin_directories(target: Path) -> list[Path]:
    target = target.resolve()
    directories: dict[str, Path] = {}
    for directory in target.parent.iterdir():
        if not (directory / ".claude-plugin" / "plugin.json").is_file():
            continue
        name, _ = read_plugin_manifest(directory)
        directories[name] = directory

    ordered: list[Path] = []
    visited: set[str] = set()

    def visit(directory: Path) -> None:
        name, dependencies = read_plugin_manifest(directory)
        if name in visited:
            return
        visited.add(name)
        for dependency in dependencies:
            dependency_directory = directories.get(dependency)
            if dependency_directory is None:
                raise ValueError(
                    f"local Claude plugin dependency not found: {name} -> {dependency}"
                )
            visit(dependency_directory)
        ordered.append(directory)

    visit(target)
    return ordered


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
        help="runs per query: model is nondeterministic (env RUNS, default 3)",
    )
    p.add_argument(
        "--threshold",
        type=float,
        default=float(os.environ.get("THRESHOLD", "0.5")),
        help="trigger-rate cutoff for a passing query (env THRESHOLD, default 0.5)",
    )
    p.add_argument(
        "--model",
        default=os.environ.get("CLAUDE_MODEL") or CLAUDE_GENERATION_MODEL,
        help="model id passed to `claude --model` "
        f"(env CLAUDE_MODEL, default {CLAUDE_GENERATION_MODEL})",
    )
    p.add_argument("--plugin-dir", default=os.environ.get("PLUGIN_DIR"),
                   help="target-only local plugin directory (env PLUGIN_DIR)")
    p.add_argument(
        "--max-budget-usd",
        type=float,
        default=float(os.environ.get("MAX_BUDGET_USD", "0.05")),
        help="hard per-run spend cap, maximum 0.05 "
        "(env MAX_BUDGET_USD, default 0.05)",
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


def skill_available_line(line: str, target: str, short: str) -> bool:
    try:
        obj = json.loads(line)
    except json.JSONDecodeError:
        return False
    if not isinstance(obj, dict):
        return False
    if obj.get("type") != "system" or obj.get("subtype") != "init":
        return False
    skills = obj.get("skills")
    return isinstance(skills, list) and any(
        match_skill(skill, target, short) for skill in skills
    )


def stream_text_delta_length(line: str) -> int:
    try:
        obj = json.loads(line)
    except json.JSONDecodeError:
        return 0
    if not isinstance(obj, dict) or obj.get("type") != "stream_event":
        return 0
    event = obj.get("event")
    if not isinstance(event, dict) or event.get("type") != "content_block_delta":
        return 0
    delta = event.get("delta")
    if not isinstance(delta, dict) or delta.get("type") != "text_delta":
        return 0
    text = delta.get("text")
    return len(text) if isinstance(text, str) else 0


def check_triggered(
    query: str, claude_args: list[str], target: str, short: str
) -> tuple[bool, str | None]:
    """Return the target trigger result or an infrastructure error.

    Terminates the claude process on first match, no further tools fire.
    """
    proc = subprocess.Popen(
        ["claude", *claude_args, query],
        stdin=subprocess.DEVNULL,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    state = {
        "matched": False,
        "target_available": False,
        "output_limit_exceeded": False,
    }
    stderr_chunks: list[str] = []

    def kill_process() -> None:
        if proc.poll() is None:
            try:
                proc.kill()
            except ProcessLookupError:
                pass

    def read_stdout() -> None:
        output_chars = 0
        assert proc.stdout is not None
        for raw_line in proc.stdout:
            line = raw_line.strip()
            if not line:
                continue
            state["target_available"] = state["target_available"] or skill_available_line(
                line, target, short,
            )
            if check_line(line, target, short):
                state["matched"] = True
                kill_process()
                return
            output_chars += stream_text_delta_length(line)
            if output_chars > TRIGGER_OUTPUT_LIMIT:
                state["output_limit_exceeded"] = True
                kill_process()
                return

    def read_stderr() -> None:
        assert proc.stderr is not None
        stderr_chunks.append(proc.stderr.read())

    stdout_thread = threading.Thread(target=read_stdout)
    stderr_thread = threading.Thread(target=read_stderr)
    stdout_thread.start()
    stderr_thread.start()
    timed_out = False
    try:
        return_code = proc.wait(timeout=TRIGGER_TIMEOUT)
    except subprocess.TimeoutExpired:
        timed_out = True
        kill_process()
        return_code = proc.wait()
    stdout_thread.join()
    stderr_thread.join()

    if state["matched"]:
        return True, None
    if timed_out:
        return False, "timeout"
    if state["output_limit_exceeded"]:
        return False, "output-limit"
    if return_code != 0:
        detail = "".join(stderr_chunks).strip()
        suffix = f": {detail}" if detail else ""
        return False, f"claude exited {return_code}{suffix}"
    if not state["target_available"]:
        return False, "target skill unavailable"
    return False, None


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
    claude_model = args.model or CLAUDE_GENERATION_MODEL
    budget = args.max_budget_usd
    if not 1 <= runs <= 3:
        sys.stderr.write("Error: --runs must be between 1 and 3\n")
        return 2
    if not 0 <= threshold <= 1:
        sys.stderr.write("Error: --threshold must be between 0 and 1\n")
        return 2
    if not 0 < budget <= 0.05:
        sys.stderr.write("Error: --max-budget-usd must be > 0 and <= 0.05\n")
        return 2

    plugin_dir = Path(args.plugin_dir) if args.plugin_dir else queries_file.parent.parent
    if not plugin_dir.is_dir():
        sys.stderr.write(f"Error: target plugin directory is invalid: {plugin_dir}\n")
        return 2
    try:
        plugin_directories = resolve_plugin_directories(plugin_dir)
    except ValueError as error:
        sys.stderr.write(f"Error: {error}\n")
        return 2

    short = skill_name.rsplit(":", 1)[-1]

    claude_args = [
        "-p", "--output-format", "stream-json", "--verbose",
        "--include-partial-messages", "--setting-sources", "",
        "--strict-mcp-config", "--mcp-config", '{"mcpServers":{}}',
        "--no-session-persistence", "--no-chrome", "--model", claude_model,
        "--max-budget-usd", str(budget), "--max-turns", "1",
        "--system-prompt", TRIGGER_SYSTEM_PROMPT, "--tools", "Skill",
    ]
    for plugin_directory in plugin_directories:
        claude_args.extend(["--plugin-dir", str(plugin_directory)])

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
    model_runs = len(queries) * runs
    if model_runs > MAX_MODEL_RUNS:
        sys.stderr.write(
            f"Error: trigger eval would launch {model_runs} model runs; "
            f"maximum is {MAX_MODEL_RUNS}\n"
        )
        return 2

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
            triggered, error = check_triggered(
                query, claude_args, skill_name, short,
            )
            if error is not None:
                sys.stderr.write(
                    f"Error: trigger infrastructure failure for query "
                    f"{json.dumps(query)}: {error}\n"
                )
                return 2
            if triggered:
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
            "model": claude_model,
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
        f"budget/run: ${budget}  tools: Skill  timeout: {TRIGGER_TIMEOUT}s  "
        f"output-limit: {TRIGGER_OUTPUT_LIMIT} chars",
        file=sys.stderr,
    )
    print(f"passed: {passed} / {total}   failed: {failed}", file=sys.stderr)
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
