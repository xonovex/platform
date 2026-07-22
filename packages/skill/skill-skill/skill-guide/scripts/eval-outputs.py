#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Run output-quality evals against a skill: with-skill vs without-skill.

Usage:
    eval-outputs.py <evals.json> <skill_name> [iteration] [options]
        skill_name = bare ("git-commit") or plugin-namespaced ("myplugin:git-commit")
        iteration  = name for this run's workspace dir (default: auto "iteration-N")

evals.json shape:
    {
      "skill_name": "<skill-name>",
      "tier": "aggressive | moderate | conservative",
      "evals": [
      {
        "id": 1,
        "prompt": "<realistic user message>",
        "expected_output": "<human-readable success criterion (judge reference)>",
        "assertions": ["<verifiable check>", "..."],   # optional; falls back to expected_output
        "files": ["evals/files/input.csv"]             # optional; paths relative to evals.json
      }
      ]
    }

Options (flag overrides env; env keeps the loop/CI ergonomics):
    --runs N / RUNS=N                  runs per arm per eval (default: 1; maximum: 3)
    --concurrency N / CONCURRENCY=N    parallel claude invocations (default/maximum: 2)
    --model M / CLAUDE_MODEL=M         generation model
                                       (default: claude-haiku-4-5-20251001)
    --judge-model M / JUDGE_MODEL=M    grading model
                                       (default: claude-sonnet-4-6)
    --disallowed-tools L / DISALLOWED_TOOLS=L
                                       tools blocked in BOTH arms during generation
                                       (default: Bash,Edit,Write,NotebookEdit,WebFetch);
                                       the without-skill arm additionally blocks Skill
    --gen-timeout S / GEN_TIMEOUT=S    per-generation timeout in seconds (default: 600)
    --workspace DIR / WORKSPACE=DIR    workspace base dir (default: "<skill>-workspace")
    --eval-cwd DIR / EVAL_CWD=DIR      working dir for generation runs (default: current dir;
                                       must be where the skill resolves — installed plugin / project)
    --plugin-dir PATH / PLUGIN_DIR=PATH
                                       target-only local plugin directory
    --max-budget-usd N / MAX_BUDGET_USD=N  hard per-generation spend cap (default/max: 0.10)
    --judge-max-budget-usd N / JUDGE_MAX_BUDGET_USD=N
                                       hard per-judge spend cap (default/max: 0.10)

Method (mirrors SkillsBench / skill-creator 2.0):
    - Each eval runs in two arms, vanilla (Skill disallowed) and skill-augmented,
      in a fresh isolated `claude -p` context — no state bleeds between runs.
    - Generation uses stream-json so the runner records both the final result
      (text + token usage + duration) AND whether the target skill actually fired.
    - Grading is reference-guided, binary PASS/FAIL per assertion, via an
      LLM-as-judge that must cite evidence (verbosity-bias-resistant).
    - Aggregates pass rate, tokens, and duration per arm into benchmark.json,
      with deltas showing the skill's cost vs benefit.

Output:
    - One JSON object per eval on stdout (with/without pass_rate, tokens, delta).
    - A summary on stderr.
    - Workspace files: <workspace>/<iteration>/eval-<id>/<arm>/{outputs/response.md,
      timing.json, grading.json}, plus <iteration>/benchmark.json.

Exit code: 0 when the tier-aware absolute pass-rate, delta, and activation gates pass;
1 for valid evidence below a quality gate; 2 for invalid evidence or configuration.

Cross-platform: works wherever the `claude` CLI is installed (macOS / Linux / Windows).
"""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import statistics
import subprocess
import sys
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

TOKEN_KEYS = (
    "input_tokens",
    "output_tokens",
    "cache_creation_input_tokens",
    "cache_read_input_tokens",
)

GENERATION_SYSTEM_PROMPT = (
    "Answer the user request directly. Use the explicitly invoked skill as "
    "authoritative guidance. Read only files that the skill itself identifies "
    "as necessary. Keep the final response under 1,000 words."
)
JUDGE_SYSTEM_PROMPT = (
    "Grade only the supplied response against the supplied assertions and "
    "return exactly the requested JSON."
)
MAX_OUTPUT_MODEL_CALLS = 24
CLAUDE_GENERATION_MODEL = "claude-haiku-4-5-20251001"
CLAUDE_JUDGE_MODEL = "claude-sonnet-4-6"
OUTPUT_GATE_POLICIES = {
    "aggressive": {"minimum_with_skill_pass_rate": 0.75,
                   "minimum_delta_pass_rate": 0.05},
    "moderate": {"minimum_with_skill_pass_rate": 0.80,
                 "minimum_delta_pass_rate": 0.05},
    "conservative": {"minimum_with_skill_pass_rate": 0.90,
                     "minimum_delta_pass_rate": 0.10},
}
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
        description="Run output-quality evals against a skill: with-skill vs without-skill.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    p.add_argument("evals", help="path to evals.json with skill_name, tier, and evals")
    p.add_argument(
        "skill_name",
        help="bare ('git-commit') or plugin-namespaced ('plugin:git-commit')",
    )
    p.add_argument(
        "iteration",
        nargs="?",
        default="",
        help="workspace dir name for this run (default: auto 'iteration-N')",
    )
    p.add_argument("--runs", type=int, default=int(os.environ.get("RUNS", "1")),
                   help="runs per arm per eval (env RUNS, default 1, maximum 3)")
    p.add_argument("--concurrency", type=int, default=int(os.environ.get("CONCURRENCY", "2")),
                   help="parallel claude invocations (env CONCURRENCY, default/maximum 2)")
    p.add_argument(
        "--model",
        default=os.environ.get("CLAUDE_MODEL") or CLAUDE_GENERATION_MODEL,
        help="generation model "
        f"(env CLAUDE_MODEL, default {CLAUDE_GENERATION_MODEL})",
    )
    p.add_argument(
        "--judge-model",
        default=os.environ.get("JUDGE_MODEL") or CLAUDE_JUDGE_MODEL,
        help="grading model "
        f"(env JUDGE_MODEL, default {CLAUDE_JUDGE_MODEL})",
    )
    p.add_argument("--disallowed-tools",
                   default=os.environ.get("DISALLOWED_TOOLS", "Bash,Edit,Write,NotebookEdit,WebFetch"),
                   help="tools blocked in both arms (env DISALLOWED_TOOLS); without-skill also blocks Skill")
    p.add_argument("--gen-timeout", type=int, default=int(os.environ.get("GEN_TIMEOUT", "600")),
                   help="per-generation timeout in seconds (env GEN_TIMEOUT, default 600)")
    p.add_argument("--workspace", default=os.environ.get("WORKSPACE"),
                   help="workspace base dir (env WORKSPACE, default '<skill>-workspace')")
    p.add_argument("--eval-cwd", default=os.environ.get("EVAL_CWD"),
                   help="working dir for generation runs (env EVAL_CWD, default current dir)")
    p.add_argument("--plugin-dir", default=os.environ.get("PLUGIN_DIR"),
                   help="target-only local plugin directory (env PLUGIN_DIR)")
    p.add_argument("--max-budget-usd", type=float,
                   default=float(os.environ.get("MAX_BUDGET_USD", "0.10")),
                   help="hard per-generation spend cap, maximum 0.10 "
                        "(env MAX_BUDGET_USD, default 0.10)")
    p.add_argument("--judge-max-budget-usd", type=float,
                   default=float(os.environ.get("JUDGE_MAX_BUDGET_USD", "0.10")),
                   help="hard per-judge spend cap, maximum 0.10 "
                        "(env JUDGE_MAX_BUDGET_USD, default 0.10)")
    return p


def match_skill(skill_field: object, target: str, short: str) -> bool:
    if not isinstance(skill_field, str):
        return False
    return (
        skill_field == target
        or skill_field == short
        or skill_field.endswith(":" + short)
    )


def skill_in_obj(obj: dict, target: str, short: str) -> bool:
    """True if a parsed stream-json line shows a matching Skill tool_use."""
    message = obj.get("message")
    if isinstance(message, dict):
        for item in message.get("content", []) or []:
            if (
                isinstance(item, dict)
                and item.get("type") == "tool_use"
                and item.get("name") == "Skill"
            ):
                inp = item.get("input")
                if isinstance(inp, dict) and match_skill(inp.get("skill"), target, short):
                    return True
    return False


def skill_available_in_obj(obj: dict, target: str, short: str) -> bool:
    """True when Claude's init event exposes the target skill."""
    if obj.get("type") != "system" or obj.get("subtype") != "init":
        return False
    skills = obj.get("skills")
    return isinstance(skills, list) and any(
        match_skill(skill, target, short) for skill in skills
    )


def sum_tokens(usage: object) -> int:
    if not isinstance(usage, dict):
        return 0
    return sum(int(usage.get(k, 0) or 0) for k in TOKEN_KEYS)


def extract_json(text: str) -> dict | None:
    """Pull the first JSON object out of a model response (tolerates fences/prose)."""
    if not text:
        return None
    fenced = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", text, re.DOTALL)
    candidate = fenced.group(1) if fenced else None
    if candidate is None:
        start = text.find("{")
        end = text.rfind("}")
        candidate = text[start : end + 1] if start != -1 and end > start else None
    if candidate is None:
        return None
    try:
        obj = json.loads(candidate)
        return obj if isinstance(obj, dict) else None
    except json.JSONDecodeError:
        return None


def stream_text_delta_length(value: object) -> int:
    if not isinstance(value, dict) or value.get("type") != "stream_event":
        return 0
    event = value.get("event")
    if not isinstance(event, dict) or event.get("type") != "content_block_delta":
        return 0
    delta = event.get("delta")
    if not isinstance(delta, dict) or delta.get("type") != "text_delta":
        return 0
    text = delta.get("text")
    return len(text) if isinstance(text, str) else 0


def claude_failure_detail(stdout: str) -> str:
    for line in reversed(stdout.strip().splitlines()):
        try:
            value = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not isinstance(value, dict):
            continue
        result = value.get("result")
        subtype = value.get("subtype")
        detail = result.strip() if isinstance(result, str) else ""
        kind = subtype.strip() if isinstance(subtype, str) else ""
        if detail and (value.get("is_error") is True or kind.startswith("error")):
            return detail[:500]
        if kind.startswith("error"):
            return kind[:500]
    return ""


def run_generation_process(
    command: list[str], timeout: int, cwd: str | None, max_output_chars: int,
) -> dict:
    proc = subprocess.Popen(
        command,
        stdin=subprocess.DEVNULL,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        cwd=cwd,
    )
    stdout_lines: list[str] = []
    stderr_chunks: list[str] = []
    output_limit_exceeded = threading.Event()

    def read_stdout() -> None:
        output_chars = 0
        assert proc.stdout is not None
        for line in proc.stdout:
            stdout_lines.append(line)
            try:
                event = json.loads(line)
            except json.JSONDecodeError:
                continue
            output_chars += stream_text_delta_length(event)
            if output_chars > max_output_chars:
                output_limit_exceeded.set()
                proc.kill()
                break

    def read_stderr() -> None:
        assert proc.stderr is not None
        stderr_chunks.append(proc.stderr.read())

    stdout_thread = threading.Thread(target=read_stdout)
    stderr_thread = threading.Thread(target=read_stderr)
    stdout_thread.start()
    stderr_thread.start()
    timed_out = False
    try:
        return_code = proc.wait(timeout=timeout)
    except subprocess.TimeoutExpired:
        timed_out = True
        proc.kill()
        return_code = proc.wait()
    stdout_thread.join()
    stderr_thread.join()
    return {
        "returncode": return_code,
        "stdout": "".join(stdout_lines),
        "stderr": "".join(stderr_chunks),
        "timed_out": timed_out,
        "output_limit_exceeded": output_limit_exceeded.is_set(),
    }


def generate(
    prompt: str, claude_args: list[str], cwd: str | None, timeout: int,
    target: str, short: str, expect_skill: bool,
) -> dict:
    """Run one generation; return final text, tokens, duration, and trigger flag."""
    proc = run_generation_process(
        ["claude", *claude_args, prompt], timeout, cwd, 10_000,
    )
    if proc["output_limit_exceeded"]:
        return {"text": "", "total_tokens": 0, "duration_ms": 0,
                "skill_triggered": False, "error": "output-limit"}
    if proc["timed_out"]:
        return {"text": "", "total_tokens": 0, "duration_ms": timeout * 1000,
                "skill_triggered": False, "error": "timeout"}
    if proc["returncode"] != 0:
        detail = proc["stderr"].strip() or claude_failure_detail(proc["stdout"])
        error = f"claude exited {proc['returncode']}{': ' + detail if detail else ''}"
        return {"text": "", "total_tokens": 0, "duration_ms": 0,
                "skill_triggered": False, "error": error}

    text, usage, duration, triggered, available = "", {}, 0, False, False
    for line in proc["stdout"].splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not isinstance(obj, dict):
            continue
        available = available or skill_available_in_obj(obj, target, short)
        triggered = triggered or skill_in_obj(obj, target, short)
        if obj.get("type") == "result":
            text = obj.get("result", "") or ""
            usage = obj.get("usage") or {}
            duration = obj.get("duration_ms", 0) or 0
    return {
        "text": text,
        "total_tokens": sum_tokens(usage),
        "duration_ms": duration,
        "skill_triggered": expect_skill and (triggered or available),
        "error": None if text else "no-result",
    }


JUDGE_RUBRIC = """\
You are a strict output evaluator. Grade the ASSISTANT RESPONSE against each \
assertion independently.

Rules:
- Binary verdict per assertion: passed = true or false. No partial credit.
- Cite concrete evidence: quote the response or name the specific gap.
- No benefit of the doubt — vagueness, omission, or a hedge is FAIL.
- Judge ONLY against the assertion. Ignore response length, tone, and style.
- If the response lacks the information to decide, mark FAIL, evidence "insufficient".
- Use the EXPECTED OUTPUT only as a reference for what success looks like; the \
response need not match it word for word.

TASK PROMPT:
{prompt}

EXPECTED OUTPUT (reference):
{expected}

ASSERTIONS (grade each, in order):
{assertions}

ASSISTANT RESPONSE:
{response}

Return ONLY minified JSON, no markdown fences, one object per assertion in order:
{{"assertion_results":[{{"text":"<assertion>","passed":true,"evidence":"<quote or reason>"}}]}}
"""


def grade(
    prompt: str, expected: str, assertions: list[str], response: str, model: str,
    budget: float,
) -> dict:
    """Reference-guided, binary LLM-as-judge grading of one response."""
    def all_fail(reason: str) -> dict:
        results = [{"text": a, "passed": False, "evidence": reason} for a in assertions]
        return summarize(results, reason)

    if not response.strip():
        return all_fail("empty response")

    numbered = "\n".join(f"{i + 1}. {a}" for i, a in enumerate(assertions))
    rubric = JUDGE_RUBRIC.format(
        prompt=prompt, expected=expected or "(none provided)",
        assertions=numbered, response=response,
    )
    schema = json.dumps({
        "type": "object",
        "properties": {
            "assertion_results": {
                "type": "array",
                "minItems": len(assertions),
                "maxItems": len(assertions),
                "items": {
                    "type": "object",
                    "properties": {
                        "text": {"type": "string"},
                        "passed": {"type": "boolean"},
                        "evidence": {"type": "string"},
                    },
                    "required": ["text", "passed", "evidence"],
                    "additionalProperties": False,
                },
            },
        },
        "required": ["assertion_results"],
        "additionalProperties": False,
    }, separators=(",", ":"))
    args = [
        "-p", "--output-format", "json",
        "--setting-sources", "", "--strict-mcp-config", "--mcp-config",
        '{"mcpServers":{}}', "--no-session-persistence", "--no-chrome",
        "--tools", "", "--max-budget-usd", str(budget), "--max-turns", "1",
        "--system-prompt", JUDGE_SYSTEM_PROMPT,
        "--json-schema", schema,
    ]
    if model:
        args.extend(["--model", model])
    verdict = None
    try:
        proc = subprocess.run(
            ["claude", *args, rubric],
            stdin=subprocess.DEVNULL, capture_output=True, text=True, timeout=300,
        )
        if proc.returncode != 0:
            detail = proc.stderr.strip() or claude_failure_detail(proc.stdout)
            suffix = f": {detail}" if detail else ""
            return all_fail(
                f"judge process error: claude exited {proc.returncode}{suffix}"
            )
        outer = json.loads(proc.stdout)
        structured = outer.get("structured_output") if isinstance(outer, dict) else None
        verdict = structured if isinstance(structured, dict) else extract_json(
            outer.get("result", "") if isinstance(outer, dict) else ""
        )
    except subprocess.TimeoutExpired:
        return all_fail("judge timeout")
    except json.JSONDecodeError:
        verdict = None
    if not verdict or not isinstance(verdict.get("assertion_results"), list):
        return all_fail("unparseable judge output")

    results = []
    for i, a in enumerate(assertions):
        item = verdict["assertion_results"][i] if i < len(verdict["assertion_results"]) else {}
        results.append({
            "text": a,
            "passed": item.get("passed") is True if isinstance(item, dict) else False,
            "evidence": (item.get("evidence", "") if isinstance(item, dict) else "") or "no evidence",
        })
    return summarize(results)


def summarize(results: list[dict], error: str | None = None) -> dict:
    passed = sum(1 for r in results if r["passed"])
    total = len(results)
    return {
        "assertion_results": results,
        "summary": {
            "passed": passed, "failed": total - passed, "total": total,
            "pass_rate": round(passed / total, 3) if total else 0.0,
        },
        "error": error,
    }


def run_job(eval_obj: dict, arm: str, run_idx: int, ctx: dict) -> dict:
    """Generate + grade one (eval, arm, run); write artifacts; return a record."""
    prompt = ctx["build_prompt"](eval_obj)
    if arm == "with_skill":
        prompt = f'/{ctx["target"]} {prompt}'
    args = ctx["with_args"] if arm == "with_skill" else ctx["without_args"]
    gen = generate(
        prompt, args, ctx["cwd"], ctx["timeout"], ctx["target"], ctx["short"],
        arm == "with_skill",
    )
    generation_healthy = gen["error"] is None and (
        arm != "with_skill" or gen["skill_triggered"]
    )
    if generation_healthy:
        graded = grade(
            eval_obj["prompt"], eval_obj.get("expected_output", ""),
            eval_obj["assertions"], gen["text"], ctx["judge_model"],
            ctx["judge_budget"],
        )
    else:
        graded = summarize([
            {
                "text": assertion,
                "passed": False,
                "evidence": "not graded because generation evidence is invalid",
            }
            for assertion in eval_obj["assertions"]
        ])

    arm_dir = ctx["iter_dir"] / f"eval-{eval_obj['id']}" / arm
    if ctx["runs"] > 1:
        arm_dir = arm_dir / f"run-{run_idx + 1}"
    (arm_dir / "outputs").mkdir(parents=True, exist_ok=True)
    (arm_dir / "outputs" / "response.md").write_text(gen["text"], encoding="utf-8")
    (arm_dir / "timing.json").write_text(json.dumps({
        "total_tokens": gen["total_tokens"], "duration_ms": gen["duration_ms"],
        "skill_triggered": gen["skill_triggered"], "error": gen["error"],
    }, indent=2), encoding="utf-8")
    (arm_dir / "grading.json").write_text(json.dumps(graded, indent=2), encoding="utf-8")

    sys.stderr.write(
        f"  [{eval_obj['id']}/{arm}{'' if ctx['runs'] == 1 else f'/run-{run_idx+1}'}] "
        f"pass_rate={graded['summary']['pass_rate']} tokens={gen['total_tokens']}"
        f"{' (skill fired)' if arm == 'with_skill' and gen['skill_triggered'] else ''}"
        f"{' [' + gen['error'] + ']' if gen['error'] else ''}\n"
    )
    return {
        "id": eval_obj["id"], "arm": arm,
        "pass_rate": graded["summary"]["pass_rate"],
        "tokens": gen["total_tokens"], "duration_ms": gen["duration_ms"],
        "skill_triggered": gen["skill_triggered"],
        "error": gen["error"] or graded.get("error"),
    }


def mean_block(values: list[float], runs: int) -> dict:
    block = {"mean": round(statistics.fmean(values), 3) if values else 0.0}
    if runs > 1 and len(values) > 1:
        block["stddev"] = round(statistics.pstdev(values), 3)
    return block


def aggregate_arm(records: list[dict], arm: str, runs: int) -> dict:
    rs = [r for r in records if r["arm"] == arm]
    by_eval: dict[object, list[dict]] = {}
    for r in rs:
        by_eval.setdefault(r["id"], []).append(r)
    pass_rates = [statistics.fmean([r["pass_rate"] for r in g]) for g in by_eval.values()]
    tokens = [statistics.fmean([r["tokens"] for r in g]) for g in by_eval.values()]
    durations = [statistics.fmean([r["duration_ms"] for r in g]) for g in by_eval.values()]
    block = {
        "pass_rate": mean_block(pass_rates, runs),
        "tokens": mean_block(tokens, runs),
        "duration_ms": mean_block(durations, runs),
    }
    if arm == "with_skill":
        fired = [1.0 if r["skill_triggered"] else 0.0 for r in rs]
        block["skill_trigger_rate"] = {"mean": round(statistics.fmean(fired), 3) if fired else 0.0}
    return block


def evaluate_output_gate(
    tier: str, with_skill_pass_rate: float, without_skill_pass_rate: float,
    skill_trigger_rate: float,
) -> dict:
    policy = OUTPUT_GATE_POLICIES[tier]
    checks = {
        "with_skill_pass_rate": (
            with_skill_pass_rate >= policy["minimum_with_skill_pass_rate"]
        ),
        "delta_pass_rate": (
            with_skill_pass_rate - without_skill_pass_rate
            >= policy["minimum_delta_pass_rate"]
        ),
        "skill_trigger_rate": skill_trigger_rate == 1.0,
    }
    return {"passed": all(checks.values()), "policy": policy, "checks": checks}


def main(argv: list[str]) -> int:
    args = build_parser().parse_args(argv)
    evals_file = Path(args.evals)
    skill_name = args.skill_name
    iteration = args.iteration

    if not evals_file.is_file():
        sys.stderr.write(f"Error: evals file not found: {evals_file}\n")
        return 2
    if not shutil.which("claude"):
        sys.stderr.write("Error: 'claude' CLI not found in PATH\n")
        return 2

    try:
        data = json.loads(evals_file.read_text(encoding="utf-8"))
    except json.JSONDecodeError as e:
        sys.stderr.write(f"Error: invalid JSON in {evals_file}: {e}\n")
        return 2
    if not isinstance(data, dict):
        sys.stderr.write(
            f"Error: {evals_file} must contain skill_name, tier, and evals\n"
        )
        return 2
    declared_skill_name = data.get("skill_name")
    tier = data.get("tier")
    if declared_skill_name != skill_name.rsplit(":", 1)[-1]:
        sys.stderr.write(
            f"Error: skill name mismatch: requested '{skill_name}' but evals declare "
            f"'{declared_skill_name}'\n"
        )
        return 2
    if tier not in OUTPUT_GATE_POLICIES:
        sys.stderr.write(
            "Error: tier must be aggressive, moderate, or conservative\n"
        )
        return 2
    evals = data.get("evals", [])
    if not isinstance(evals, list) or not evals:
        sys.stderr.write(f"Error: {evals_file} has no evals\n")
        return 2

    # Normalize each eval: require id + prompt + (assertions | expected_output).
    norm: list[dict] = []
    for i, e in enumerate(evals):
        if not isinstance(e, dict) or "prompt" not in e:
            sys.stderr.write(f"Skipping eval #{i}: missing prompt\n")
            continue
        assertions = e.get("assertions")
        if not assertions and e.get("expected_output"):
            assertions = [e["expected_output"]]
        if not assertions:
            sys.stderr.write(f"Skipping eval {e.get('id', i)}: no assertions or expected_output\n")
            continue
        norm.append({**e, "id": e.get("id", i + 1), "assertions": list(assertions)})
    if not norm:
        sys.stderr.write("Error: no gradable evals\n")
        return 2

    runs = args.runs
    concurrency = args.concurrency
    claude_model = args.model or CLAUDE_GENERATION_MODEL
    judge_model = args.judge_model or CLAUDE_JUDGE_MODEL
    disallowed = args.disallowed_tools
    timeout = args.gen_timeout
    cwd = args.eval_cwd or None
    budget = args.max_budget_usd
    judge_budget = args.judge_max_budget_usd
    if not 1 <= runs <= 3:
        sys.stderr.write("Error: --runs must be between 1 and 3\n")
        return 2
    if not 1 <= concurrency <= 2:
        sys.stderr.write("Error: --concurrency must be between 1 and 2\n")
        return 2
    if not 0 < budget <= 0.10:
        sys.stderr.write("Error: --max-budget-usd must be > 0 and <= 0.10\n")
        return 2
    if not 0 < judge_budget <= 0.10:
        sys.stderr.write("Error: --judge-max-budget-usd must be > 0 and <= 0.10\n")
        return 2
    model_calls = len(norm) * runs * 4
    if model_calls > MAX_OUTPUT_MODEL_CALLS:
        sys.stderr.write(
            f"Error: output eval would launch {model_calls} model calls; "
            f"maximum is {MAX_OUTPUT_MODEL_CALLS}. "
            "Split the eval set into bounded batches.\n"
        )
        return 2
    short = skill_name.rsplit(":", 1)[-1]

    plugin_dir = Path(args.plugin_dir) if args.plugin_dir else evals_file.parent.parent
    if not plugin_dir.is_dir():
        sys.stderr.write(f"Error: target plugin directory is invalid: {plugin_dir}\n")
        return 2
    try:
        plugin_directories = resolve_plugin_directories(plugin_dir)
    except ValueError as error:
        sys.stderr.write(f"Error: {error}\n")
        return 2

    base = Path(args.workspace) if args.workspace else Path(f"{short}-workspace")
    if not iteration:
        existing = [int(m.group(1)) for p in base.glob("iteration-*")
                    if (m := re.match(r"iteration-(\d+)$", p.name))]
        iteration = f"iteration-{max(existing, default=0) + 1}"
    iter_dir = base / iteration
    iter_dir.mkdir(parents=True, exist_ok=True)
    benchmark_path = iter_dir / "benchmark.json"
    invalid_run_path = iter_dir / "invalid-run.json"
    benchmark_path.unlink(missing_ok=True)
    invalid_run_path.unlink(missing_ok=True)

    gen_base = [
        "-p", "--output-format", "stream-json", "--verbose",
        "--include-partial-messages",
        "--setting-sources", "", "--strict-mcp-config", "--mcp-config",
        '{"mcpServers":{}}', "--no-session-persistence", "--no-chrome",
        "--max-turns", "6", "--system-prompt", GENERATION_SYSTEM_PROMPT,
    ]
    if claude_model:
        gen_base.extend(["--model", claude_model])
    gen_base.extend(["--max-budget-usd", str(budget)])
    with_args = [*gen_base, "--tools", "Skill,Read"]
    if disallowed:
        with_args.append(f"--disallowedTools={disallowed}")
    for plugin_directory in plugin_directories:
        with_args.extend(["--plugin-dir", str(plugin_directory)])
    without_disallowed = ",".join(filter(None, [disallowed, "Skill"]))
    without_args = [
        *gen_base, "--tools", "Read", f"--disallowedTools={without_disallowed}"
    ]

    evals_dir = evals_file.parent

    def build_prompt(e: dict) -> str:
        prompt = e["prompt"]
        files = e.get("files") or []
        if files:
            paths = [str((evals_dir / f).resolve()) for f in files]
            prompt += "\n\nRelevant input files (read them as needed):\n" + "\n".join(
                f"- {p}" for p in paths
            )
        return prompt

    ctx = {
        "with_args": with_args, "without_args": without_args, "cwd": cwd,
        "timeout": timeout, "target": skill_name, "short": short,
        "judge_model": judge_model, "judge_budget": judge_budget,
        "iter_dir": iter_dir, "runs": runs,
        "build_prompt": build_prompt,
    }

    sys.stderr.write(
        f"skill: {skill_name}  evals: {len(norm)}  runs/arm: {runs}  "
        f"concurrency: {concurrency}  workspace: {iter_dir}\n"
        f"gen model: {claude_model or '<default>'}  judge model: {judge_model or '<default>'}\n"
        f"caps: generation=${budget}/6 turns  judge=${judge_budget}/1 turn  "
        f"model_calls={model_calls}/{MAX_OUTPUT_MODEL_CALLS}  retries=0\n---\n"
    )

    jobs = [(e, arm, r) for e in norm for arm in ("with_skill", "without_skill")
            for r in range(runs)]
    records: list[dict] = []
    with ThreadPoolExecutor(max_workers=concurrency) as pool:
        futures = [pool.submit(run_job, e, arm, r, ctx) for e, arm, r in jobs]
        for fut in as_completed(futures):
            record = fut.result()
            records.append(record)
            failed = record["error"] or (
                record["arm"] == "with_skill" and not record["skill_triggered"]
            )
            if failed:
                for pending in futures:
                    pending.cancel()
                break

    failures = [
        {"id": record["id"], "arm": record["arm"],
         "reason": record["error"] or "target skill did not activate"}
        for record in records
        if record["error"] or (
            record["arm"] == "with_skill" and not record["skill_triggered"]
        )
    ]
    if failures:
        invalid_run_path.write_text(json.dumps({
            "skill": skill_name, "iteration": iteration,
            "status": "invalid", "failures": failures,
        }, indent=2), encoding="utf-8")
        sys.stderr.write(
            f"invalid benchmark evidence: {len(failures)} infrastructure failure(s)\n"
            f"diagnostic: {invalid_run_path}\n"
        )
        return 2

    # Per-eval stdout lines.
    for e in norm:
        w = [r for r in records if r["id"] == e["id"] and r["arm"] == "with_skill"]
        wo = [r for r in records if r["id"] == e["id"] and r["arm"] == "without_skill"]
        w_pr = statistics.fmean([r["pass_rate"] for r in w]) if w else 0.0
        wo_pr = statistics.fmean([r["pass_rate"] for r in wo]) if wo else 0.0
        w_tok = statistics.fmean([r["tokens"] for r in w]) if w else 0.0
        wo_tok = statistics.fmean([r["tokens"] for r in wo]) if wo else 0.0
        print(json.dumps({
            "id": e["id"], "prompt": e["prompt"],
            "with_skill": {"pass_rate": round(w_pr, 3), "tokens": round(w_tok),
                           "skill_triggered": any(r["skill_triggered"] for r in w)},
            "without_skill": {"pass_rate": round(wo_pr, 3), "tokens": round(wo_tok)},
            "delta_pass_rate": round(w_pr - wo_pr, 3),
            "delta_tokens": round(w_tok - wo_tok),
        }, ensure_ascii=False))

    with_block = aggregate_arm(records, "with_skill", runs)
    without_block = aggregate_arm(records, "without_skill", runs)
    benchmark = {
        "skill": skill_name, "tier": tier, "iteration": iteration,
        "model": claude_model, "judge_model": judge_model,
        "runs_per_arm": runs, "eval_count": len(norm),
        "run_summary": {
            "with_skill": with_block, "without_skill": without_block,
            "delta": {
                "pass_rate": round(with_block["pass_rate"]["mean"]
                                   - without_block["pass_rate"]["mean"], 3),
                "tokens": round(with_block["tokens"]["mean"]
                                - without_block["tokens"]["mean"]),
                "duration_ms": round(with_block["duration_ms"]["mean"]
                                     - without_block["duration_ms"]["mean"]),
            },
        },
    }
    benchmark["quality_gate"] = evaluate_output_gate(
        tier,
        with_block["pass_rate"]["mean"],
        without_block["pass_rate"]["mean"],
        with_block.get("skill_trigger_rate", {}).get("mean", 0.0),
    )
    benchmark_path.write_text(json.dumps(benchmark, indent=2), encoding="utf-8")

    delta = benchmark["run_summary"]["delta"]
    sys.stderr.write(
        f"---\nwith_skill pass_rate: {with_block['pass_rate']['mean']}  "
        f"(skill fired: {with_block.get('skill_trigger_rate', {}).get('mean')})  "
        f"tokens: {with_block['tokens']['mean']}\n"
        f"without_skill pass_rate: {without_block['pass_rate']['mean']}  "
        f"tokens: {without_block['tokens']['mean']}\n"
        f"delta pass_rate: {delta['pass_rate']}  tokens: {delta['tokens']}\n"
        f"quality gate ({tier}): "
        f"{'PASS' if benchmark['quality_gate']['passed'] else 'FAIL'}\n"
        f"benchmark: {benchmark_path}\n"
    )
    return 0 if benchmark["quality_gate"]["passed"] else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
