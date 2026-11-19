# Sources

## Agent Skills: Optimizing Skill Descriptions

- **URL:** https://agentskills.io/skill-creation/optimizing-descriptions.md
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `references/evaluating-triggers.md` → entire file
- **Aspects extracted:**
  - Trigger eval set design (~20 queries, 8-10 positive + 8-10 negative) → `evaluating-triggers.md` Trigger Eval Set
  - Near-miss negative queries → Strongest should-not-trigger queries
  - Realistic query content (file paths, personal context, casual language) → Making Eval Queries Realistic
  - Multi-run trigger rate measurement → Measuring Trigger Rate
  - Train / validation split methodology → Train / Validation Split
  - Optimization loop (5 iterations, pick best validation pass rate) → Optimization Loop

## Agent Skills: Evaluating Skill Output Quality

- **URL:** https://agentskills.io/skill-creation/evaluating-skills.md
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `references/evaluating-outputs.md` → entire file (incl. Automated Runner section)
  - `scripts/eval-outputs.py` → with-skill/without-skill arms, LLM-as-judge grading, `benchmark.json`
- **Aspects extracted:**
  - Test case shape (prompt, expected output, optional input files) → Test Case Shape
  - Reference-guided binary LLM-as-judge with evidence; with-skill vs without-skill arms in isolated runs → `scripts/eval-outputs.py` (method also informed by public SkillsBench / skill-creator output-eval practice)
  - `evals/evals.json` storage convention
  - Workspace layout (`iteration-N/`, `with_skill/`, `without_skill/`) → Workspace Layout
  - Run pair (with skill + baseline) → Running Eval Pairs
  - Fresh context per run → Workflow
  - Timing capture (`total_tokens`, `duration_ms`) → Running Eval Pairs
  - Good vs weak assertions, evidence requirement → Writing Assertions
  - Grading via script for mechanical checks, LLM for the rest → Grading
  - `benchmark.json` aggregation → Aggregating
  - Pattern analysis: always-pass, always-fail, with-vs-without delta, variance → Pattern Analysis
  - Human review + actionable feedback → Human Review
  - Iteration signals (failed assertions, feedback, transcripts) → Iteration Signals
  - "Stay lean, explain why, bundle repeated work" iteration principles

## OpenAI Codex skills and non-interactive mode

- **URLs:**
  - https://learn.chatgpt.com/docs/build-skills.md
  - https://learn.chatgpt.com/docs/non-interactive-mode.md
- **Last reviewed:** 2026-07-22
- **References:** references/evaluating-triggers.md, references/evaluating-outputs.md
- **Aspects extracted:** Codex skill discovery and progressive disclosure; implicit routing from skill descriptions; repository/user skill locations; explicit `$skill-name` invocation; ephemeral JSONL `codex exec`; ignored user configuration and execution rules; read-only sandbox selection.

## Xonovex skill-evaluation policy

- **Provenance:** Repository-original catalog routing, evidence-health, repetition, and tier-gate policy maintained with the evaluator implementations
- **Last reviewed:** 2026-07-22
- **References:** references/evaluating-triggers.md, references/evaluating-outputs.md
- **Aspects extracted:** Competitive multi-skill routing scenarios, two-run CI evidence, full activation, absolute pass-rate floors, and minimum quality deltas are repository conventions.

## Guide-level synthesis

- **Provenance:** Repository-original integration of the source blocks above; this reference combines multiple inputs or maintained conventions rather than one exclusive upstream
- **References:** references/optimize.md
- **Last reviewed:** 2026-05-27
