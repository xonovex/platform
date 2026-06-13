# fable mode (`xonovex-skill-fable`)

A user-invocable (`/fable`) behavioral overlay that nudges Opus 4.8 toward the
observable working/writing voice of the Claude Fable 5 model. It restyles **how**
the assistant communicates and paces work — it does **not** change correctness,
safety, project-instruction adherence, or the underlying model's reasoning.

**Honest scope:** a prompt can imitate Fable's surface voice and some working
habits. It cannot reproduce Fable's reasoning, judgment, knowledge, or the model
itself. Everything below was kept only where it _measurably_ moved Opus toward
real Fable output; everything that didn't was dropped.

## Evidence base

Derived by comparing a corpus of real Claude Fable 5 output against Claude
Opus 4.8 on equivalent tasks, then characterizing the differences.

Caveat carried through all conclusions: that corpus is **fixed**, **narrow** (a
single technical domain, captured over a short window), and cannot grow. Coding
style within it was governed by an opinionated C coding-style guide, not by the
model — so code output is not a place the overlay can or should differ.

Measured Fable fingerprint (vs Opus 4.8 on the same material):

- **Narration** — terser per turn (median roughly ¾ the length); ~half the
  section headers; near-zero emoji; near-zero hedging; somewhat higher parallel
  tool-batching on mixed work.
- **Documents** — plan/doc artifacts run as dense prose bullets: few headers,
  few tables, heavy inline code-span citations, occasional em-dash asides.
- **Prep** — more investigation before the first code edit; later shown to be
  largely task-scope-driven, not a model trait.
- **Framing** — denser "honest / caveat / failure-mode" language per turn,
  though no more _likely_ to caveat in a given turn.

## Method

Blind A/B tournaments. The model is held **constant** (all arms are Opus 4.8
subagents), so the only variable is the prompt overlay. Each candidate
"dimension" is a rule layered on the validated core; a baseline arm gets the
core alone.

Scoring, per output:

1. **Blind voice-judge** — given real Fable excerpts ("gold") and an anonymized
   candidate, rates same-author voice similarity 0–10. Judges score voice only,
   never topic/correctness, and never know which arm is which.
2. **Deterministic markers** — length, headers, hedging, em-dash, bullets,
   code-span density, verdict-first opener, and parallel-batch rate.

**Keep rule (anti-overfit):** a dimension is added only if it beats baseline by
a margin on the blind judge (judged against _real_ Fable text, not the surface
rules) **and** passes a human gut-check. Margin + blind-vs-real-text guards
against Goodharting the metric. 2 reps per arm to damp single-run noise.

## Results by round

Scores are blind voice-judge same-author similarity (0–10).

### Round 1 — controlled coding A/B

The C coding-style guide active in **both** arms; task = add a small C helper
function + tests.

- **Code** — the implementation file was byte-**identical** across arms; header
  and test near-identical. Coding style is the guide, not the overlay. _(inert)_
- **Report** — baseline opener _"Done. I added the helper…"_ (3 headers,
  ~1540 chars); fable opener _"Pass — tests green, 1/1…"_ (0 headers, ~1120
  chars, −27%, verdict-first). _(voice win)_

### Round 2 — prep front-loading

Investigation calls before the first code edit: control 4 vs treatment 4. Task
too small to exercise it; the signal is largely task-scope-confounded.
**Inconclusive, dropped.**

### Narration tournament

Task = read-only investigation report.

| Dimension                     | Similarity | Same-author | vs baseline |          |
| ----------------------------- | ---------- | ----------- | ----------- | -------- |
| D0 baseline (core)            | 3.5        | 0/2         | —           |          |
| D1 result+next in one breath  | 6.0        | 2/2         | +2.5        | **KEPT** |
| D2 exact-evidence verdicts    | 4.5        | 1/2         | +1.0        | drop     |
| D4 pick-one decisive scope    | 4.5        | 1/2         | +1.0        | drop     |
| D5 name+dismiss tooling noise | 4.5        | 1/2         | +1.0        | drop     |
| D3 name risk before verifying | 3.5        | 0/2         | +0.0        | drop     |

Baseline scored low here only because final **reports** were judged against
Fable's clipped **in-flight** narration (register mismatch — a ceiling no prompt
closes). Only D1 cleared the margin _and_ was same-author 2/2.

### Document / plan tournament

Task = write an implementation plan.

| Dimension                          | Similarity | Same-author |
| ---------------------------------- | ---------- | ----------- |
| DD0 baseline (core)                | 9.0        | 2/2         |
| DD1 dense what+files+rationale     | 9.0        | 2/2         |
| DD3 inline evidence asides         | 9.0        | 2/2         |
| DD5 dense bullets over scaffolding | 9.0        | 2/2         |
| DD2 decision-first framing         | 8.5        | 2/2         |
| DD4 citation density               | 8.5        | 2/2         |

**Recommend none.** The core already produces the doc voice at 9/10 same-author;
candidates were redundant or mildly regressive. (Markers showed generated plans
run denser/longer than real Fable artifacts — a register/economy gap, not a
missing trait; not actioned, would risk overfit.)

### Round 3 — judgment dimensions

**Text** (judged vs gold): baseline 9.0; D1 honest-framing 9.0 (already in core,
no lift, drop); D4 contract-referencing 8.5 (mildly regressive, drop).

**Fork** (intentionally under-specified design task; judged "surfaced the
fork?"):

| Arm                | Surfacing | Surfaced | vs baseline |                             |
| ------------------ | --------- | -------- | ----------- | --------------------------- |
| baseline           | 4.5       | 1/2      | —           | "a well-argued silent pick" |
| D3 decision-gating | 7.5       | 2/2      | +3.0        | **KEPT**                    |

**Batching** (transcript metric, request-grouped): control (no instruction) 0.71
batch rate; treatment (instruction) 1.0; Fable reference ~0.30. Opus already
batches ≥ Fable on controlled work — not a gap. _(drop)_

## What the skill encodes

All validated against real Fable text:

- Verdict-first openers; terse 1–3 sentence turns; no preamble/recap
- No hedging; explicit noise-discounting; confident on verified facts
- Exact naming (files/symbols/error-kinds in backticks)
- Minimal in-flow formatting (prose, not headers/tables/emoji; structure
  reserved for genuine deliverables)
- Parallel tool batching
- One-breath result+next cadence — _narration tournament: +2.5_
- Decision-gating: surface genuine design forks for the user (options +
  recommendation) while staying decisive on the unambiguous — _Round 3: +3.0_

### Tested and dropped

Not guessed — measured inert, regressive, or not-a-gap:

- prep front-loading (inconclusive / task-scope-confounded)
- anticipatory risk-naming (+0.0)
- exact-evidence / decisive-scope / noise-discount narration emphases (within
  n=2 noise)
- honest-framing emphasis (already in the core)
- contract-referencing emphasis (mildly regressive, −0.5)
- parallel-batching dimension (Opus already batches ≥ Fable)

## Honest limitations

- **Model identity is not promptable.** Reasoning quality, judgment, knowledge,
  and the model's own texture come from the weights; this skill cannot reach
  them. It restyles the seams.
- The residual gap the judges kept citing is **register/length** (clipped
  in-flight narration vs full deliverables), which is task-bound, not a voice
  rule.
- The reference corpus is fixed, short, and from a single technical domain.
  Conclusions may not generalize, and it cannot be grown.
- Scores are proxies; the blind-judge-vs-real-text + margin design mitigates but
  does not eliminate overfitting risk.

## Use

`/plugin` → install + enable `xonovex-skill-fable`, then `/fable` to toggle on.
