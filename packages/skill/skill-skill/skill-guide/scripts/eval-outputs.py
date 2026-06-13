#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Run output-quality evals against a skill: with-skill vs without-skill.

Usage:
    eval-outputs.py <evals.json> <skill_name> [iteration] [options]
        skill_name = bare ("git-commit") or plugin-namespaced ("int-poc-git:git-commit")
        iteration  = name for this run's workspace dir (default: auto "iteration-N")

evals.json shape (either a bare array of evals, or {"skill_name", "evals": [...]}):
    [
      {
        "id": 1,
        "prompt": "<realistic user message>",
        "expected_output": "<human-readable success criterion (judge reference)>",
        "assertions": ["<verifiable check>", "..."],   # optional; falls back to expected_output
        "files": ["evals/files/input.csv"]             # optional; paths relative to evals.json
      }
    ]

Options (flag overrides env; env keeps the loop/CI ergonomics):
    --runs N / RUNS=N                  runs per arm per eval (default: 1; >1 measures variance)
    --concurrency N / CONCURRENCY=N    parallel claude invocations (default: 4)
    --model M / CLAUDE_MODEL=M         model for the generation runs (haiku/sonnet/opus)
    --judge-model M / JUDGE_MODEL=M    model for grading (default: claude default)
    --disallowed-tools L / DISALLOWED_TOOLS=L
                                       tools blocked in BOTH arms during generation
                                       (default: Bash,Edit,Write,NotebookEdit,WebFetch);
                                       the without-skill arm additionally blocks Skill
    --gen-timeout S / GEN_TIMEOUT=S    per-generation timeout in seconds (default: 600)
    --workspace DIR / WORKSPACE=DIR    workspace base dir (default: "<skill>-workspace")
    --eval-cwd DIR / EVAL_CWD=DIR      working dir for generation runs (default: current dir;
                                       must be where the skill resolves — installed plugin / project)
    --max-budget-usd N / MAX_BUDGET_USD=N  optional hard per-generation spend cap (unset = no cap)

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

Exit code: 0 if with-skill mean pass rate exceeds without-skill, else 1.

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
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

TOKEN_KEYS = (
    "input_tokens",
    "output_tokens",
    "cache_creation_input_tokens",
    "cache_read_input_tokens",
)


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        description="Run output-quality evals against a skill: with-skill vs without-skill.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    p.add_argument("evals", help="path to evals.json (array, or {\"evals\": [...]})")
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
                   help="runs per arm per eval (env RUNS, default 1; >1 measures variance)")
    p.add_argument("--concurrency", type=int, default=int(os.environ.get("CONCURRENCY", "4")),
                   help="parallel claude invocations (env CONCURRENCY, default 4)")
    p.add_argument("--model", default=os.environ.get("CLAUDE_MODEL", "haiku"),
                   help="model for the generation runs (env CLAUDE_MODEL, default haiku)")
    p.add_argument("--judge-model", default=os.environ.get("JUDGE_MODEL", ""),
                   help="model for grading (env JUDGE_MODEL)")
    p.add_argument("--disallowed-tools",
                   default=os.environ.get("DISALLOWED_TOOLS", "Bash,Edit,Write,NotebookEdit,WebFetch"),
                   help="tools blocked in both arms (env DISALLOWED_TOOLS); without-skill also blocks Skill")
    p.add_argument("--gen-timeout", type=int, default=int(os.environ.get("GEN_TIMEOUT", "600")),
                   help="per-generation timeout in seconds (env GEN_TIMEOUT, default 600)")
    p.add_argument("--workspace", default=os.environ.get("WORKSPACE"),
                   help="workspace base dir (env WORKSPACE, default '<skill>-workspace')")
    p.add_argument("--eval-cwd", default=os.environ.get("EVAL_CWD"),
                   help="working dir for generation runs (env EVAL_CWD, default current dir)")
    p.add_argument("--max-budget-usd", type=float,
                   default=(float(os.environ["MAX_BUDGET_USD"]) if os.environ.get("MAX_BUDGET_USD") else None),
                   help="optional hard per-generation spend cap passed to `claude --max-budget-usd` "
                        "(env MAX_BUDGET_USD; unset = no cap — output eval needs the task to finish)")
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


def generate(
    prompt: str, claude_args: list[str], cwd: str | None, timeout: int, target: str, short: str
) -> dict:
    """Run one generation; return final text, tokens, duration, and trigger flag."""
    try:
        proc = subprocess.run(
            ["claude", *claude_args, prompt],
            stdin=subprocess.DEVNULL,
            capture_output=True,
            text=True,
            timeout=timeout,
            cwd=cwd,
        )
    except subprocess.TimeoutExpired:
        return {"text": "", "total_tokens": 0, "duration_ms": timeout * 1000,
                "skill_triggered": False, "error": "timeout"}

    text, usage, duration, triggered = "", {}, 0, False
    for line in proc.stdout.splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not isinstance(obj, dict):
            continue
        if skill_in_obj(obj, target, short):
            triggered = True
        if obj.get("type") == "result":
            text = obj.get("result", "") or ""
            usage = obj.get("usage") or {}
            duration = obj.get("duration_ms", 0) or 0

    return {
        "text": text,
        "total_tokens": sum_tokens(usage),
        "duration_ms": duration,
        "skill_triggered": triggered,
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
    prompt: str, expected: str, assertions: list[str], response: str, model: str
) -> dict:
    """Reference-guided, binary LLM-as-judge grading of one response."""
    def all_fail(reason: str) -> dict:
        results = [{"text": a, "passed": False, "evidence": reason} for a in assertions]
        return summarize(results)

    if not response.strip():
        return all_fail("empty response")

    numbered = "\n".join(f"{i + 1}. {a}" for i, a in enumerate(assertions))
    rubric = JUDGE_RUBRIC.format(
        prompt=prompt, expected=expected or "(none provided)",
        assertions=numbered, response=response,
    )
    args = [
        "-p", "--output-format", "json",
        "--disallowedTools=Bash,Edit,Write,Read,NotebookEdit,WebFetch,"
        "WebSearch,Glob,Grep,Task,Skill,TodoWrite",
    ]
    if model:
        args.extend(["--model", model])
    try:
        proc = subprocess.run(
            ["claude", *args, rubric],
            stdin=subprocess.DEVNULL, capture_output=True, text=True, timeout=300,
        )
        outer = json.loads(proc.stdout)
        verdict = extract_json(outer.get("result", "") if isinstance(outer, dict) else "")
    except (subprocess.TimeoutExpired, json.JSONDecodeError):
        verdict = None

    if not verdict or not isinstance(verdict.get("assertion_results"), list):
        return all_fail("unparseable judge output")

    results = []
    for i, a in enumerate(assertions):
        item = verdict["assertion_results"][i] if i < len(verdict["assertion_results"]) else {}
        results.append({
            "text": a,
            "passed": bool(item.get("passed", False)) if isinstance(item, dict) else False,
            "evidence": (item.get("evidence", "") if isinstance(item, dict) else "") or "no evidence",
        })
    return summarize(results)


def summarize(results: list[dict]) -> dict:
    passed = sum(1 for r in results if r["passed"])
    total = len(results)
    return {
        "assertion_results": results,
        "summary": {
            "passed": passed, "failed": total - passed, "total": total,
            "pass_rate": round(passed / total, 3) if total else 0.0,
        },
    }


def run_job(eval_obj: dict, arm: str, run_idx: int, ctx: dict) -> dict:
    """Generate + grade one (eval, arm, run); write artifacts; return a record."""
    prompt = ctx["build_prompt"](eval_obj)
    args = ctx["with_args"] if arm == "with_skill" else ctx["without_args"]
    gen = generate(prompt, args, ctx["cwd"], ctx["timeout"], ctx["target"], ctx["short"])
    graded = grade(
        eval_obj["prompt"], eval_obj.get("expected_output", ""),
        eval_obj["assertions"], gen["text"], ctx["judge_model"],
    )

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
    evals = data.get("evals", []) if isinstance(data, dict) else data
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
    concurrency = max(1, args.concurrency)
    claude_model = args.model
    judge_model = args.judge_model
    disallowed = args.disallowed_tools
    timeout = args.gen_timeout
    cwd = args.eval_cwd or None
    budget = args.max_budget_usd
    short = skill_name.rsplit(":", 1)[-1]

    base = Path(args.workspace) if args.workspace else Path(f"{short}-workspace")
    if not iteration:
        existing = [int(m.group(1)) for p in base.glob("iteration-*")
                    if (m := re.match(r"iteration-(\d+)$", p.name))]
        iteration = f"iteration-{max(existing, default=0) + 1}"
    iter_dir = base / iteration
    iter_dir.mkdir(parents=True, exist_ok=True)

    gen_base = ["-p", "--output-format", "stream-json", "--verbose"]
    if claude_model:
        gen_base.extend(["--model", claude_model])
    if budget and budget > 0:
        gen_base.extend(["--max-budget-usd", str(budget)])
    with_args = [*gen_base, f"--disallowedTools={disallowed}"] if disallowed else list(gen_base)
    without_disallowed = ",".join(filter(None, [disallowed, "Skill"]))
    without_args = [*gen_base, f"--disallowedTools={without_disallowed}"]

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
        "judge_model": judge_model, "iter_dir": iter_dir, "runs": runs,
        "build_prompt": build_prompt,
    }

    sys.stderr.write(
        f"skill: {skill_name}  evals: {len(norm)}  runs/arm: {runs}  "
        f"concurrency: {concurrency}  workspace: {iter_dir}\n"
        f"gen model: {claude_model or '<default>'}  judge model: {judge_model or '<default>'}\n---\n"
    )

    jobs = [(e, arm, r) for e in norm for arm in ("with_skill", "without_skill")
            for r in range(runs)]
    records: list[dict] = []
    with ThreadPoolExecutor(max_workers=concurrency) as pool:
        futures = [pool.submit(run_job, e, arm, r, ctx) for e, arm, r in jobs]
        for fut in as_completed(futures):
            records.append(fut.result())

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
        "skill": skill_name, "iteration": iteration,
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
    (iter_dir / "benchmark.json").write_text(json.dumps(benchmark, indent=2), encoding="utf-8")

    delta = benchmark["run_summary"]["delta"]
    sys.stderr.write(
        f"---\nwith_skill pass_rate: {with_block['pass_rate']['mean']}  "
        f"(skill fired: {with_block.get('skill_trigger_rate', {}).get('mean')})  "
        f"tokens: {with_block['tokens']['mean']}\n"
        f"without_skill pass_rate: {without_block['pass_rate']['mean']}  "
        f"tokens: {without_block['tokens']['mean']}\n"
        f"delta pass_rate: {delta['pass_rate']}  tokens: {delta['tokens']}\n"
        f"benchmark: {iter_dir / 'benchmark.json'}\n"
    )
    return 0 if delta["pass_rate"] > 0 else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
