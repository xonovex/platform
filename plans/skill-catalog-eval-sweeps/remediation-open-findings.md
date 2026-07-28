---
type: report
parent_plan: plans/skill-catalog-eval-sweeps.md
status: active
---

# Remediation plan: open eval findings after the trigger sweep

Input to subplan 2. Produced by a multi-agent investigation of the 21 findings left open
by subplan 1 (12 trigger, 9 routing), each proposal adversarially verified against the
catalog's constraints, then spot-checked by hand.

## Status of the recommendations

Landed already, in subplan 1's branch:

- **S1 / H1 / H3** — `0333a195`. Three-way classification of a Skill invocation, turn cap
  1 -> 2, and the corrected comment about which plugins a per-skill run loads.
- **S2 (core)** — `0333a195`. `selected_skills` records each run's outcome. The wider
  `loaded_skills` / `available_skills` / CLI-version fields are not implemented.
- **F1 / F2 / F3** — `45594ee1`. The three routing-owner pairing adds. Verified after the
  fact: all three owners win their new scenario (3/3, 3/3, 2/3).

Still open: **S3** (adaptive run escalation), **F4** (workflow train positives),
**D1-D7** (the ownership decisions), and everything in ACCEPT AND RECORD.

## Corrections to the analysis below

Two claims in the report did not survive checking, and the report is left unedited so the
disagreement stays visible:

1. **E1's frequency claim is not established.** The defect is real — the captured stream
   and the code path both confirm it. But the report states memory-management re-probed
   3/3 with the bare plugin name; seven probes run afterwards (3 on memory-management, 4
   on data-oriented-design) produced the correct full name every time. The slip is
   intermittent, and the attribution of four routing failures "entirely" to it was not
   proven at the time of writing. Evidence since: after `0333a195`, data-oriented-design's
   broad-phase scenario moved 0.333 -> 1.0, which supports the attribution for that one.
   The likely mechanism for the discrepancy is that a probe loads a single plugin while a
   routing scenario loads several similarly-named ones, so abbreviation pressure differs.

2. **S1a versus S1b (U1) was settled by implementation, not by the report.** The landed
   fix holds a non-target invocation until the harness says whether it launched, so an
   unresolvable name settles nothing and the second turn lets the model name a skill that
   exists. Nothing is scored as a trigger for a run in which no skill loaded.

---

# Skill-catalog eval remediation plan — 21 open findings

**Read section 0 first.** Seven probes run during this review changed the disposition of five findings, including one where the investigator and the verifier were both wrong about the mechanism.

---

## 0. New evidence that changes the plan

| #   | Claim under review                                                                                                                             | What I measured                                                                                                                                                                                                                                                                                                                                                                                                                                           | Consequence                                                                                                                                                                                                                                                                                                                                    |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| E1  | "matchSkill rejects the bare plugin name and misfiles it as a competitor win"                                                                  | Captured raw stream (`scratchpad/mm_raw.jsonl`): `tool_use {skill:"xonovex-skill-memory-management"}` → `tool_result is_error:true "<tool_use_error>Unknown skill: xonovex-skill-memory-management</tool_use_error>"`, `error_max_turns`, turns 2. `matchSkill` (trigger-process.ts:43-56) rejects it, `invokedSkillNames` returns it non-empty, so trigger-process.ts:196 sets `competingSkill`.                                                         | **CONFIRMED.** No competitor won; the run is scored as if one did. Real defect, affects both gates.                                                                                                                                                                                                                                            |
| E2  | "Dropping dependency plugin dirs from the per-skill run dissolves both c99 findings" (cluster-2 verifier's headline `missedOption`)            | 3 probes of c99-opinionated's physical-design query with the **overlay alone** → `PICKED: NONE`, `NONE`, `NONE`. Pooled with the earlier investigator probe: 0/4.                                                                                                                                                                                                                                                                                         | **FALSIFIED.** Dropping the dependency takes that query from 0.333 to **0.0**. The dependency-loaded base establishes the language context that makes the overlay reachable at all. Do not do it.                                                                                                                                              |
| E3  | "hono-opinionated loses the bodyLimit query, so it needs a description edit + train query first" (cluster-3 investigator)                      | 2 probes with the true routing field `{hono-opinionated, hono (dependency), typescript}` → `xonovex-skill-hono-opinionated:hono-opinionated-guide` **2/2**. Pooled with the verifier's 2: **4/4**.                                                                                                                                                                                                                                                        | **FALSIFIED.** The fix is a one-line pairing add. No description edit, no budgets.json bump, no train-query sequencing.                                                                                                                                                                                                                        |
| E4  | "claude-code-guide's description promise of 'settings, or managed configuration' is unbacked; narrow it" (cluster-1 verifier's `missedOption`) | Body grep: SKILL.md:16 "Respect settings authority — Managed, user, project, local, plugin, and component hooks have native precedence and trust rules"; SKILL.md:26 managed-settings restrictions; capabilities.md:18 full settings-scope list; onboarding.md:25 "source precedence".                                                                                                                                                                    | **VERIFIER WRONG.** Settings-source authority _is_ backed. What is absent is permission-**rule authoring**. The label flip is sufficient; **do not narrow the description** (8 passing queries depend on it).                                                                                                                                  |
| E5  | "instruction-guide loses to bundled `init`"                                                                                                    | The harness's advertised `skills` array does **not** contain `init`: `['deep-research','xonovex-skill-instruction:instruction-guide','design-sync','dataviz','update-config','verify','debug','code-review','simplify','batch','fewer-permission-prompts','doctor','loop','claude-api','run','run-skill-generator']`. But at `--max-turns 2`: `TOOL_USE: init` → `tool_result "Launching skill: init"` (is_error absent) → model re-invoked `init` again. | **Premise holds, but the advertised list is incomplete.** `init` launches despite being unlisted. Two consequences: (a) D2 is a genuine bundled loss and the concession stands; (b) **the init `skills` array cannot be used to classify competitors** — only to resolve the target. This kills one variant of the S1 fix and rescues another. |
| E6  | "The degenerate lua-opinionated query fails independently of routing"                                                                          | `skill-lua-opinionated/.skill-eval-results/trigger/claude/results.jsonl` (written 01:40, 3 min before the routing sweep at 01:43): `triggers:2, runs:3, trigger_rate:0.667, pass:true`. Counts: **8 positives / 8 negatives**, validation positives 3.                                                                                                                                                                                                    | **FALSIFIED, and deletion is blocked.** `checkTriggerMinimums` (catalog-files.ts:134) rejects <8 positives. Deleting drops it to 7.                                                                                                                                                                                                            |

Corrected catalog counts (both reviews had these wrong): `c99-opinionated-guide` has **17** positives (not 12 or 9); `workflow-guide` has **16** (8 train / 8 validation).

---

## 1. Systemic options

### S1 — Distinguish "unresolvable skill name" from "a competitor won" · **DO FIRST**

**Bucket 4. Closes up to 4 of 9 routing failures. Zero extra model calls.**

`checkTriggered` has three outcomes but only two buckets. A `Skill` call naming something the harness cannot resolve is filed under "another skill won the request" (trigger-process.ts:192-201), which is false.

The stream distinguishes the cases unambiguously (E1, E5):

| tool_result                                                           | meaning                                |
| --------------------------------------------------------------------- | -------------------------------------- |
| `is_error: true`, `<tool_use_error>Unknown skill: X</tool_use_error>` | nothing launched — **unresolvable**    |
| `Launching skill: X` (no `is_error`)                                  | a real skill launched — **competitor** |

**Do not** classify using the init `skills` array — E5 proves it is incomplete (`init` launches while unlisted). Use it only to resolve the _target's_ plugin prefix: accept invoked name `X` when `X + ":" + short` equals the target's advertised entry. That lookup cannot collide across the 8 prefix-colliding plugin pairs (`xonovex-skill-typescript` + `:typescript-guide` can only equal typescript-guide's own entry), so the `startsWith` hazard the cluster-5 investigator warned about never arises.

Two variants; **the owner must pick one** (see §5, U1):

- **S1a — three-way classification + `--max-turns 2`** (recommended). An unresolvable name no longer settles the run; the model gets the harness's `Unknown skill` error and one turn to self-correct, and the score then reflects a skill that actually loaded. No metric redefinition. Cost: one extra turn only on runs that slip (~5-13% by probe evidence); `--max-budget-usd 0.05` and `TRIGGER_OUTPUT_LIMIT` still bound it. Caveat from E5: on a genuine competitor win the model re-invokes the same competitor, so turn 2 does not launder real losses.
- **S1b — count a target-resolving unresolvable name as a trigger.** Cheaper, no turn change, but records a trigger for a run in which no skill loaded, and will flip some currently-passing `should_trigger:false` queries to failing.

**Do NOT route this through `runOnce` retry** (proposed by the cluster-3 verifier). `runOnce` retries on `outcome.error !== null`; three consecutive slips make `evaluateQuery` return `undefined`, which makes `runTriggerEvaluation` return `success:false` (trigger-evaluation.ts:76, 106-107) and **aborts the entire skill's sweep**. A nondeterministic formatting slip would become a gate outage.

**Argument against:** measured rates move on both sides, so the 12/9 baselines are invalidated and both sweeps must be re-baselined in the same change (~$7.3, see §2).

### S2 — Record what won and what was loaded · **DO FIRST**

**Bucket 4. Closes 0 findings by itself. Zero model calls. Purely additive JSON.**

`competingSkill` is computed at trigger-process.ts:142/197 and discarded at :228; `ResultRecord` (trigger-evaluation.ts:6-14) and `RoutingRecord` (routing-evaluate.ts:61-70) have no field for it; `candidate_skills` is built from `scenario.candidates` while the harness is fed `resolveClaudePluginDirectories(...)`, which pulls in declared `dependencies`.

Add: `selected_skills` (one entry per run, with the S1 three-way tag: `target` / `competitor:<name>` / `unresolvable:<name>` / `none` / `output-limit`), `loaded_skills` (the dependency-expanded plugin set), `available_skills` (the init array, **annotated as incomplete** per E5), and the Claude CLI version. Mirror into `RoutingRecord` (routing-evaluate.ts:282).

This is the prerequisite that turns triage from probe archaeology into data — recovering "who won" cost this review and the two prior ones roughly 30 probe calls. **It must not be counted as remediating any of the 21.**

**Argument against:** none material. Do not extend `routing-check` to enforce declared-vs-loaded parity: loading a base alongside its overlay is by design for every `*-opinionated` pair, so the warning would fire on nearly every overlay scenario.

### S3 — Adaptive run escalation for non-unanimous queries · **AFTER S1**

**Bucket 4. Would decide the 5 marginal findings + skill-guide. +14% trigger cost.**

Measured from the 72 stored `results.jsonl`: 317 rows, `{1.0: 245, 0.0: 50, 0.667: 17, 0.333: 5}` — only **22 rows (6.9%) are non-unanimous**. Run 3; stop on a unanimous 0/3 or 3/3; escalate only non-unanimous queries to 9.

| scheme           | trigger runs | Δ cost              | routing runs   |
| ---------------- | ------------ | ------------------- | -------------- |
| today (runs=3)   | 951 (~$5.90) | —                   | 231 (~$1.43)   |
| flat runs=5      | 1585         | +67%                | 385            |
| flat runs=7      | 2219         | +133%               | 539            |
| **adaptive 3→9** | **1083**     | **+13.9% (~$0.82)** | **309 (+34%)** |

Three blockers must move together:

1. `validation.ts:28` — `.pipe(z.int().positive().max(3))`.
2. `maxBatchModelRuns` (trigger-config.ts:257-266 and routing-evaluate.ts:202-209) is computed **statically from the configured `runs`**, so escalation inside `evaluateQuery` makes the accounting cap silently lie (`--batch-size 8 --runs 3` reads 24/24 while real spend reaches 8×9=72). Fold the escalation ceiling into that computation instead of raising `.max(3)` separately.
3. Routing already accepts `SKILL_ROUTING_RUNS` (moon.yml:49) — blocked only by the same Zod cap and `--batch-size 8`. Apply escalation to `evaluateScenario` too, or stop calling this systemic.

**Reject the third `unstable` verdict.** An `unstable` bucket that CI reports but does not fail on is an unratcheted gate weakening in a repo built on ratchets, and it misreads the sentence it cites: `skill-guide/references/evaluating-triggers.md:67` says "a query that triggers 1/3 of the time still indicates instability; **widen the eval set or tighten the description**" — it prescribes content work, not a bucket CI ignores.

**Argument against:** sequence after S1. Four of the five "variance" failures are a deterministic matcher defect (E1); escalating first spends money confirming a bug.

**Wall-clock warning:** routing runs are fully sequential (`evaluateScenario` awaits each run, `main` awaits each batch). 231 runs × ~25 s ≈ **1.6 h** today; 309 ≈ **2.1 h**. If CI has a job timeout, that is the binding constraint, not cost.

### Systemic options to REJECT

| option                                                                                                    | why reject                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Drop dependency plugin dirs from the per-skill run** (`trigger-config.ts:288` → `[pluginDirectory]`)    | **E2: falsified.** 0/4 overlay-alone probes fire on c99-opinionated's own physical-design query; this makes 0.333 → 0.0. Instead fix the **factually false comment** at trigger-config.ts:87-92 ("This evaluator loads one skill") — untrue for the 8 plugins declaring `dependencies` — and record `loaded_skills` (S2).                                                                                                                                                       |
| **Defer every registered routing scenario out of the per-skill gate**                                     | Removes **77 of 317** kept validation rows (24%) and deletes 5 of the 12 trigger failures by construction. The accessibility case is the proof it hides signal: routing 3/3, trigger 0/3, pooled probes 2/7 — the routing 3/3 is the lucky sample, and deferral keeps only the flattering number. If ever adopted it needs a `kept.length === 0` guard: `boundedBatches([])` → `maxBatchModelRuns` 0 → `passed 0 / failed 0` → `evaluate.ts:105` returns 0 = **vacuous green**. |
| **Scale routing threshold by candidate count, or use plurality**                                          | No dilution effect exists: pass-by-candidate-count is non-monotonic (2: 44/46, 3: 8/11, 4: 10/11, 5: 3/4, 6: 3/4) and a 2-candidate scenario is among the failures while 6-candidate ones pass. A chance-corrected bar at 6 candidates ≈ 0.17 would pass a genuine 1-in-4 mis-route. At runs=3, plurality is satisfied by a 1-1-1 split.                                                                                                                                        |
| **`conclusive_runs` denominator** (rate = owner ÷ runs-that-invoked-something)                            | E1 shows the dominant loss mode is an unresolvable-name invocation, which is "conclusive" under that definition — memory-management stays 0/3. Simultaneously it lets a 1-owner/2-nothing scenario score 1.0, a weaker bar than the trigger gate applies to the same text.                                                                                                                                                                                                      |
| **Standing rule: "query wording overlapping a bundled description ⇒ negative unless a probe rescues it"** | Not mechanically checkable, inverts the burden so relabelling fires by default, and bundled descriptions change per release (E5: the advertised set already differs from the sweep brief). The defensible narrow rule is: **relabel only when the skill's body has no content for the concept, with a probe as required evidence.**                                                                                                                                             |
| **`defer_to` schema field + lint**                                                                        | Redundant by construction — once the named owner carries the text, `catalogQueryOwners` already defers it. And the lint would demand ~170 new positives catalog-wide for the 3 that actually failed.                                                                                                                                                                                                                                                                            |
| **Exclude bundled skills from the per-skill run**                                                         | Not possible. `--disable-slash-commands` ("Disable all skills") and `--safe-mode` both remove the catalog plugin under measurement; `--disallowed-tools` still scores a denial as settled, because `invokedSkillNames` (trigger-process.ts:86-96) harvests `permission_denials`. Record the decision (see FIX-NOW H4).                                                                                                                                                          |

---

## 2. Order of operations

**Wave 0 — harness only, no catalog edits.** S1 + S2 + the trigger-config.ts:87-92 comment correction, with tests (`trigger-process.test.ts` for the three-way classification incl. the multi-skill/unresolvable cases; `trigger-evaluation.test.ts`, `routing-evaluate.test.ts`, `index.test.ts` for the record shape). Then a **full re-baseline sweep**: 951 trigger runs + 231 routing runs ≈ **$7.3**, routing wall-clock ≈ 1.6 h. Everything downstream is measured against the new baseline.

**Wave 1 — three pairing adds (FIX NOW, catalog).** No query text changes, no description changes, no budget bumps. Then `routing-check` + `ci-routing --owners hexagonal-pattern-guide,data-oriented-design-guide,hono-opinionated-guide` + `ci-skill-eval-trigger` for those three skills (a new self-owned positive enters the owner's own per-skill gate — self-owned queries are kept, not deferred).

**Wave 2 — decisions (D1-D5), after the owner rules.** Each is its own commit citing its specific failing row.

**Wave 3 — S3 adaptive escalation, then re-measure the residual marginals** (§ACCEPT AND RECORD) and settle skill-guide and accessibility at n=9.

### What must land together

| atomic unit                                                                                                   | why                                                                                                                                                                                                                                                                                       |
| ------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D1 flip **+** the new train positive                                                                          | `claude-code-guide` has exactly **8 positives** (5 train / 3 validation). Flipping alone → 7 → `catalog-files.ts:134` fails. Post-change validation positives = **2**, exactly the per-split floor (`:141`).                                                                              |
| D2's three flips                                                                                              | Keep atomic. Counts verified legal: 12/8 → **9/11**; train pos 8→6, validation pos 4→3, train neg 5→7, validation neg 3→4 — all clear of the ≥8 and ≥2-per-split floors.                                                                                                                  |
| D3 description **+** `budgets.json` bump **+** re-run of _both_ debugging's and memory-management's scenarios | `budgets.json` `packages/skill/skill-debugging/debugging-guide/SKILL.md: 605` is exactly at the current count; any growth fails the drift lint. Reseed with `validate-drift --seed`, do not hand-count. The two skills contest the same word and their scenarios must be judged together. |
| D4's label flip **+** the description narrowing                                                               | Flipping alone leaves `c99-opinionated-guide`'s description still advertising "caller-owned string views/builders", so it keeps competing on every strings query. Do both or neither.                                                                                                     |
| Any query **text** change                                                                                     | Every file carrying the text, in one commit. **None is recommended in this plan.**                                                                                                                                                                                                        |
| S1 + S2 + full re-sweep                                                                                       | Committed `.skill-eval-results/*` become stale the moment S1 lands.                                                                                                                                                                                                                       |

`budgets.json` has **555 entries, all `.md`** — `eval-queries.json` is never budgeted, so every pairing add and label flip is budget-free.

---

## 3. FIX NOW

High confidence, low risk, demonstrably a real defect.

### Harness

| id  | change                                                                                                                                                               | bucket | effort  | conf | risk                                                                                                                                                      |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------- | ---- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| H1  | S1 three-way classification in `checkTriggered` (+ `--max-turns 2` if S1a)                                                                                           | 4      | M       | high | Rates move both ways; full re-baseline mandatory. Never route through `runOnce` retry.                                                                    |
| H2  | S2 record widening: `selected_skills`, `loaded_skills`, `available_skills` (flagged incomplete), CLI version; mirror to `RoutingRecord`                              | 4      | S       | high | Additive; update the three test files. Codex path supplies `null` — keep the asymmetry explicit.                                                          |
| H3  | Correct the false comment at `trigger-config.ts:87-92`; state that 8 plugins load declared `dependencies` and that a dependency-loaded sibling can win a kept query  | 4      | trivial | high | none                                                                                                                                                      |
| H4  | Record the bundled-skill decision in `skill-guide/references/evaluating-triggers.md` (budget 638, bump in the same commit), citing the **claude-code-guide 0/3** row | 4      | trivial | high | Must cite that specific failure, not be written as standalone doctrine. Must **not** encode the rejected "surface vocabulary" auto-relabel rule (see §1). |

### Catalog — three pairing adds (bucket 3, one JSON entry each, no text change, no budget)

Each is an unpaired negative whose rationale names an owner that does not carry the text, so nothing defers it and the target is scored as the only plausible candidate.

**F1 — `packages/skill/skill-hexagonal-pattern/hexagonal-pattern-guide/eval-queries.json`** — add:

```json
{
  "query": "the core package imports the concrete email and SMS adapters directly; invert it so the core depends only on interfaces",
  "should_trigger": true,
  "split": "validation",
  "rationale": "dependency inversion at the core boundary"
}
```

Cites `skill-orthogonal-pattern` trigger_rate 1 on a `should_trigger:false` row. Head-to-head probe → hexagonal. Leave orthogonal's negative byte-identical. Effort trivial, confidence high.

**F2 — `packages/skill/skill-data-oriented-design/data-oriented-design-guide/eval-queries.json`** — add:

```json
{
  "query": "Lay out the mesh and transform component arrays as SoA so the update loop stays cache-friendly.",
  "should_trigger": true,
  "split": "validation",
  "rationale": "SoA component layout for cache-friendly iteration"
}
```

Cites `skill-threejs` trigger_rate 0.667 on a `should_trigger:false` row. Do **not** also add the text as a negative to `ecs-guide` — untested, and it materially hardens the scenario. Keep this visibly separate in the commit from DOD's still-failing broad-phase scenario.

**F3 — `packages/skill/skill-hono-opinionated/hono-opinionated-guide/eval-queries.json`** — add:

```json
{
  "query": "add bodyLimit middleware to all routes in our Hono app",
  "should_trigger": true,
  "split": "validation",
  "rationale": "request payload cap; bodyLimit is an opinionated-overlay decision"
}
```

Cites `skill-typescript` trigger_rate 1 on a `should_trigger:false` row. **E3 supersedes the investigator's 3-step sequencing**: with the true field (hono-opinionated + its `hono` dependency + typescript) the owner wins **4/4** pooled. No description edit, no `budgets.json` 295 bump, no train query. `bodyLimit` content exists only under `hono-opinionated-guide` (SKILL.md + `references/body-limit.md`); `skill-hono` has zero occurrences. Optionally correct typescript-guide's now-stale rationale string ("should match hono-guide") — rationale text is not load-bearing for pairing.

All three: `routing-check` stays green (verified — 147 scenarios / 77 validation; 67 skills own exactly 1, 5 own 2; orthogonal, threejs and typescript each keep their own passing scenario). Cost: +9 routing runs.

### Catalog — one eval-set gap (bucket 1)

**F4 — `workflow-guide`: add 2 train positives for the Validate operation.** All 8 train positives are review / compose / execute / workspace-create / workspace-merge / abandon / SDLC / resume; **both** Validate queries sit in the validation split, so there is no legitimate lever to iterate on and any description tune today would be fitted to a validation failure. Author them across the axis the evidence identifies — **subject opacity**, not Validate vocabulary. (The passing validation item names an opaque revision the model cannot read and routes; the failing one names a plausible file path and the model tries to open it. Probes of the failing text → `NONE` twice, with the model asking for the file.)

```json
{
  "query": "Judge PR 4417 against its acceptance criteria independently and return pass, fail, or blocked for each one — the author's summary says everything is covered.",
  "should_trigger": true,
  "split": "train",
  "rationale": "independent per-criterion verdict against an opaque change reference"
},
{
  "query": "Acceptance criteria: (1) uploads over 10 MB are rejected with 413, (2) the limit is configurable per route, (3) existing routes keep their current behaviour. The implementer reports all three met. Give me your own verdict per criterion with the evidence you relied on.",
  "should_trigger": true,
  "split": "train",
  "rationale": "criteria supplied inline; verdict must not inherit the implementer's conclusion"
}
```

Then run `--split train`. If they fail, a description edit is **train-authorized**. If they pass, the validation query is recorded as a genuine generalization miss, not patched. Do **not** use the earlier proposals — both were the failing validation item with its subject stripped, and one was probed as behaving identically (`NONE`); tuning until those pass launders validation tuning through the train split. No budget impact (`eval-queries.json` is unbudgeted). Effort small, confidence high.

---

## 4. NEEDS A DECISION

Genuine ownership questions. Each has a recommendation, but only the repo owner can settle the catalog-vs-bundled boundary.

### D1 — `claude-code-guide` vs bundled `update-config` (bucket 3)

_Query:_ "Configure Claude Code permissions so repository reads and the test command are allowed, destructive shell commands require confirmation, and network tools stay blocked." — trigger 0/3.

**`update-config` is genuinely on offer** (confirmed present in the advertised skill list, E5) and its description names the literal action ("permissions ('allow X', 'add permission', 'move permission to')… any changes to settings.json/settings.local.json"). The verifier built a scratchpad copy whose description named `settings.json`/`settings.local.json` permission rules literally, lint-clean, and re-probed → still `update-config`. **Unwinnable by description.**

**Recommendation: concede the mechanical permission-rule write to the harness.** Flip to `should_trigger:false` with rationale `"the harness update-config skill writes settings.json permission rules; this guide covers hook and settings-source authority"` (deliberately not matching `^near miss owned by (\S+)$`, so `generatedNegativeOwners` at catalog-files.ts:248-252 is not perturbed).

**Mandatory in the same commit** (positives are exactly 8) — a train positive that exercises what the body actually delivers, with an outcome that is _not_ pre-verified:

```json
{
  "query": "Our org ships managed settings that only allow managed hooks — can a project-level PreToolUse handler still run, and how do I confirm which source is active?",
  "should_trigger": true,
  "split": "train",
  "rationale": "settings-source authority and managed hook restrictions"
}
```

(Grounded in SKILL.md:16, SKILL.md:26, `references/capabilities.md:18`, `references/onboarding.md:18`.)

**Per E4, do NOT narrow the description.** The verifier's `missedOption` misread the body: settings-source precedence and managed configuration _are_ substantively covered. Only permission-rule authoring is absent. Narrowing would put 8 currently-passing queries at risk for no gain.

Residual to state in the rationale: after the flip the catalog asserts claude-code-guide must not fire on this text while its description still says "settings, or managed configuration". That assertion is true only while a bundled skill outranks it, and the routing evaluator cannot adjudicate it (no catalog skill owns the text). Validation positives land on exactly 2 — the floor — so any later positive removal in that split breaks the gate.

_Effort small · confidence high · risk: bundled descriptions change per release._

### D2 — `instruction-guide` vs bundled `init` (bucket 3)

_Queries:_ "bootstrap a CLAUDE.md for the new services/notifications package i just scaffolded" (train), "add a CLAUDE.md to the new monorepo workspace at packages/billing" (train), "no claude.md in services/payments — analyze the code and write one" (validation, 0/3).

**E5 nearly overturned this and then confirmed it.** `init` is _not_ in the advertised `skills` array — but at `--max-turns 2` the harness returned `Launching skill: init` (no error) and the model re-invoked it. So `init` is real, it wins, and turn 2 does not rescue it.

Corroborating: the **train** queries fail identically, so the decision is not driven by the validation result; a description variant adding `CLAUDE.md` as a trigger token was probed by both reviewers → still `init`; and `grep -i claude` over `instruction-guide/SKILL.md` + all five references returns **zero** matches, so the three positives assert coverage the body never had. That is squarely inside constraint 2's carve-out.

**Recommendation: flip all three** to `should_trigger:false`, rationale `"bundled init owns CLAUDE.md bootstrap; this guide owns AGENTS.md"`.

**Drop the "optional" SKILL.md Core Principle line.** The flip deletes the only queries that could have cited it, so adding it would grow a budgeted file (504) with no failing eval behind it — speculative hardening under the standing rule. If the pointer-file behaviour is worth owning, add a train positive that exercises it and let _that_ failure authorize the body edit, separately.

**The decision is a product decision, not an eval one:** this repo's own `CLAUDE.md` is a one-line `@AGENTS.md` pointer, so conceding routes plain Claude Code sessions toward writing a standalone `CLAUDE.md` — the format instruction-guide's "Open standard" principle exists to displace. Record it as a decision taken, not as a residual.

_Effort small · confidence high · counts verified legal._

### D3 — `debugging-guide` vs `memory-management-guide` on leak detection (bucket 2/3)

_Query:_ "I want to tag every allocation with file and line and check the counter is zero on shutdown to catch leaks" — routing **0/3**, and the only clean, repeatable competitor win in the whole set (4 probes across two reviewers, all → `memory-management-guide`).

The bodies already implement a clean split — debugging owns detection instrumentation (`references/bug-taxonomy.md:9`: "tag allocations with `__FILE__`/`__LINE__`, total bytes per subsystem, assert the counter is zero on that system's shutdown", repeated in `instrumentation-and-checks.md:8`); every `leak` mention under memory-management is ownership-design prose. The **descriptions do not encode that split**: debugging lists sanitizers/assertions/fill patterns but none of the allocation-tagging vocabulary, while memory-management leads with "how memory is allocated" plus a bare "leaks" token.

**Recommendation, with two conditions.** Bucket 2 prescribes a _deletion-first_ edit — attempt a net-zero-word swap inside the existing 605 budget before growing it. If that fails, the growth version (+11 words → 616) is:

> `Use when chasing a bug in native or low-level software: a crash, access violation, use-after-free, leak, intermittent/heisenbug, or 'works on my machine' failure, and when deciding how to prevent a whole class of bugs by design. Triggers on segfaults, callstacks/.dmp files, freed-memory fill patterns (0xdddddddd), allocation tagging with file/line, per-subsystem byte counters asserted zero at shutdown, git bisect, minimal repro, assertions, sanitizers (ASan/UBSan/TSan), determinism, and record/replay, even when the user doesn't say 'debugging'.`

Lint-clean (550 chars; no sentence-initial Skip/Do-not-use/out-of-scope; names no other `*-guide`). Wording tracks debugging's own reference vocabulary rather than the query's word order.

**Condition 1:** the claim that editing memory-management would be "a drive-by" is wrong — memory-management's own single validation scenario is _also_ a cited failure in this sweep (0.333), so editing it is covered by the citation rule. These two contest the word "leak" and should be resolved as one boundary decision with one owner for leak-**detection** and one for ownership-**design**.
**Condition 2:** re-run both scenarios together, plus debugging's train split, to show the change generalizes beyond the motivating query.

_Effort small · confidence high · risk: broadening debugging toward allocation vocabulary can pull it into memory-management's ownership-design scenario, which is already at 0.333._

### D4 — `c99-opinionated-guide` vs `c99-guide` on strings (bucket 2, not 3)

_Query:_ "the parser in src/cfg/ calls strlen and strtok all over a big buffer and load time is terrible — how should I restructure string handling?" — trigger 0.333.

The bodies are unambiguous: `c99-guide/SKILL.md:20` owns "Borrow length-carrying views, write through bounded builders over caller memory — not strlen/strcat/strtok rescans" with `references/string-views.md`; c99-opinionated hands the concept off at its own SKILL.md:31 and :46 and has no strings reference. **But the router never reads bodies** — and c99-opinionated's _description_ literally advertises "caller-owned string views/builders" while c99-guide's description does not contain the word "string" at all. So this is a **content gap in the routing surface**, not a pure label dispute.

**Recommendation: do both edits in one commit, or neither.**

1. Delete `caller-owned string views/builders, ` from c99-opinionated's description (a pure shrink → no budget bump). Result:
   > `Use when editing systems or embedded C99 code in projects that follow the opinionated caller-owns-memory, data-oriented style. A focused overlay that covers only house-style decisions, not generic C99 idioms. Triggers on \`.c\`/\`.h\` files in systems/embedded/DOD projects and on prompts about caller-owns-memory, SoA/SIMD variants, alignment, index/handle references, physical design, plugin architecture, strict file naming, and shaping a C API to be bound from other languages (Lua/C#/Python/Rust FFI, generated wrappers), even when the user doesn't say 'opinionated'.`
2. Flip this entry to `should_trigger:false` (rationale: "strings are owned by c99-guide; this overlay adds only caller-owned storage, which the prompt does not ask for") and add the identical text to `c99-guide/eval-queries.json` as `should_trigger:true, split:validation` — making it a real routing scenario.

Counts are safe: c99-opinionated has **17** positives (validation 6 → 5). c99-opinionated keeps its own validation scenario ("review the file naming under src/comms/…"), so `routing-check` stays green.

**Caveat the flip alone does not fix:** pooled probes are 4/6 to c99-guide, i.e. inside the same indifference band. Flipping without the description edit relocates the flake into the routing gate. **Requires a full per-skill re-sweep of c99-opinionated** (the description change touches its other 16 positives), so schedule it after Wave 0.

_Effort medium · confidence medium._

### D5 — `lua-opinionated-guide` degenerate query (bucket 1, blocked)

_Query:_ "Review this Lua Opinionated change for essentials and fix the concrete issues you find." — routing 0/3.

Auto-generated from a SKILL.md heading (its own rationale says "implicit task cue derived from the Essentials section"); "essentials" names a document section and "Lua Opinionated" is the title-cased skill name.

**But deletion is blocked and the stated evidence is false (E6).** The skill has exactly **8 positives / 8 negatives**; deleting one fails `checkTriggerMinimums`. And the per-skill trigger gate scored this exact query **0.667, pass** three minutes before the routing sweep scored it 0/3 — it does _not_ fail independently of routing. Across two reviewers' routing probes no competitor ever won; the losses were `NONE` with the model asking for the diff. Promoting an existing positive into a routing scenario restores the routing-owner invariant but does **nothing** for the positive count.

**Recommendation: defer to Wave 3.** Re-run after S1 and at n=9. If it still loses with `NONE` and no competitor, the legitimate remedy is an **in-place, byte-identical text repair** across all three files that carry it — `skill-lua-opinionated/lua-opinionated-guide/eval-queries.json`, `skill-lua/lua-guide/eval-queries.json`, `skill-typescript-to-lua/typescript-to-lua-guide/eval-queries.json` — inlining the Lua snippet under review. That preserves the pairing, the routing-owner invariant **and** the 8/8 counts, which deletion does not. If deletion is preferred instead, the same commit must add a new curated validation positive.

**Drop the "audit the same generator family" rider.** 14 queries share the shape and most currently pass (fp-guide's three are its own validation positives); touching them is drive-by rewording of passing queries.

### D6 — `skill-guide` "audit this third-party skill" (bucket contested)

_Query:_ "before we install this third-party skill, audit its scripts and the URLs it fetches and lock down what tools it's allowed to use" — the **only finding failing in both gates** (trigger 0.333, routing 0/3).

Pooled across three reviewers: **4/9 owner selections** — right at the threshold, not a clean loss. Three incompatible diagnoses, each with real support:

- _Content gap:_ `references/security.md` exists and is indexed with exactly this trigger, but the description contains no security vocabulary at all — discovery sees only name+description. Proposed remedy: +20 words, `budgets.json` 1687 → 1707.
- _Eval quality:_ "this third-party skill" has no antecedent and the eval supplies nothing. **Weak** — of 268 positives in the sweep, 127 use a bare deictic with no path or code block and only 4 of those score below 0.5.
- _Harness confound:_ one probe routed correctly and then declined pending the artifact ("I can help with that using the skill-guide skill… however I need you to provide the third-party skill first"). The eval prompt (`validation.ts:131-133`) says "invoke only that Skill immediately. Otherwise reply with one short sentence" — it has no branch for "a skill applies but I want the artifact first."

**Recommendation: change nothing yet; this is S3's best test case.** Both edits are risky — the description grows a 1687-word file and makes "audit" appear twice (greedier on generic "audit this" prompts, pulling toward code-review-guide/command-guide), while the query rewrite touches skill-guide's **only** validation routing scenario across four files (`skill-guide` positive/validation, `command-guide` negative/validation, `llmstxt-guide` negative/validation, `reflect-guide` negative/train) and makes the query strictly easier, masking future regressions. Re-measure at n=9 after S1, then decide.

One thing to confirm on the re-run rather than assume: `skill-guide` declares `allowed-tools: "Read Bash(git:*) Bash(uv:*) Bash(claude:*) Bash(codex:*)"` while the evaluator runs `--tools Skill`, so its invocations come back as `permission_denials`. `invokedSkillNames` does count denials, so it still scores as triggered — but verify empirically. An asymmetry here would be a harness defect, not a content gap.

### D7 — `accessibility-guide` vs bundled `run` (bucket 3)

_Query:_ "retest focus, zoom, errors, and dynamic status after this UI revision" — trigger 0/3 while the **same text** passes routing 3/3.

`run` is confirmed in the advertised skill list and demonstrably wins ("I'll help you run and test the app to verify that focus, zoom, errors, and dynamic status work correctly after your UI revision"). Pooled 2/7 to accessibility across seven probes, with one bare-plugin-name emission (an S1 case) among them. The routing 3/3 is the lucky sample, not evidence the trigger gate is wrong.

**Recommendation: do not suppress via deferral** (see §1). Re-measure after S1 at n=9. If it still loses, this is a third catalog-vs-bundled boundary call of the same kind as D1/D2, and the decision axis is whether "re-verify accessibility behaviour after a change" is a category accessibility-guide should own in its description (it currently says "exact-revision reassessment") or a request the harness's `run` skill legitimately takes. Do **not** rewrite the query text: it would require the same edit in `skill-react/react-guide/eval-queries.json` or `routing-check` fails for accessibility-guide.

---

## 5. ACCEPT AND RECORD

Nothing should change. Each is recorded with its reason; several are conditional on S1/S3 landing, so **their gates stay red until then** — that is the price of not tuning against validation.

| skill / query                                                                           | bucket   | why nothing changes                                                                                                                                                                                                                                                                                                                                                                                                  |
| --------------------------------------------------------------------------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `c99-opinionated-guide` — "editing one header … near-full rebuild" (0.333)              | 1        | No content defect: "physical design" is in the description and SKILL.md:22, :42 + `references/physical-design.md`; c99-guide has no equivalent. Pooled 5/8. **E2 correction:** re-attribute as dependency competition, not "the dependency helps the overlay win" — a dep-loaded c99-guide stole a pick in re-probes. Re-measure after S1/S3.                                                                        |
| `code-quality-guide` — "riddled with `any` … what is the right fix?" (0.333)            | 1        | 5 probes, 5 owner picks, **no competitor ever observed**; pooled 6/8. **Delete the unmeasured "code-review and simplify are the competitors" causal claim** — writing an unmeasured competitor into the record seeds a false lead. Do **not** reword: the prompt is byte-identical (476 chars) in `eval-queries.json` and `evals.json`, so a rewrite silently re-targets the output eval.                            |
| `github-guide` — "private remote, API returns 404, what permission is missing" (0.333)  | 1        | Description already ends "or GitHub permissions"; the concept is owned in `references/auth.md`. The `credential-management` dependency **is** loaded as an unregistered competitor, but **nobody ever observed it winning** — that is a hypothesis, not a finding. Do not add the negative: it would also make this a registered routing scenario and pull it out of github-guide's per-skill run. Re-measure first. |
| `python-guide` — "Tighten src/ingest/config.py until pyright strict passes" (0.333)     | 1        | 4/4 pooled probes, correctly qualified, no competitor and no bare-name slip. Do **not** add "pyright"/"mypy"/"strict typing" to the description — keyword-copying from a failed validation query, which `evaluating-triggers.md` calls overfitting and constraint 2 forbids. Pure runs=3 variance.                                                                                                                   |
| `lua-guide` — "A Lua loop accidentally writes counters into _G…" (0.333)                | 1        | The "broken as a test" carve-out **does not hold**: 3/3 re-probes with all six candidates loaded routed to `lua-guide` immediately, none asked for the file. That leaves only "it failed", which is never sufficient. The proposed replacement also imported "our stats module" straight out of lua-guide's own description. Text change would need all six files atomically. Do nothing.                            |
| `data-oriented-design-guide` — "profile why the broad-phase is slow" (0.333)            | 4, not 1 | Re-attributed: a re-probe returned the **bare plugin name** `xonovex-skill-data-oriented-design`, which the current matcher scores as a competitor win. Not sampling noise. Folds into S1; skill and query untouched. Also: refuse to edit `c99-game-opinionated-guide`'s description on this cluster's evidence — no probe ever produced a mis-route to it.                                                         |
| `memory-management-guide` — "design the buffer-ownership convention…" (0.333)           | 4, not 1 | Re-attributed: **3/3** re-probes emitted the bare plugin name; raw stream shows `Unknown skill`, `error_max_turns`, turns 2 (E1). Under today's scoring this is 0/3, worse than recorded. Skill and query untouched; folds entirely into S1.                                                                                                                                                                         |
| `typescript-guide` — "rewrite this string concat with template literals…" (routing 0/3) | 4        | Same defect, reproduced against the real 8-plugin field. Note the quoted query in the original finding was **truncated** — the real scenario carries a second line, `if (retries > 3) { throw new Error('too many'); }`. Folds into S1. Do not claim it "should score 1.0" afterwards; the acceptable form appeared in only 1 of 2 runs.                                                                             |
| `editor-viewport-guide` — "my translate gizmo jumps to the cursor…" (routing 0/3)       | 4        | Description already names "transform handles/gizmos", "dragging an axis at a steep camera angle", "ray/plane projection"; re-probes selected it, one of them via the bare plugin name. Second-site evidence for S1, not an unexplained non-reproduction. **Do not edit the description** — speculative hardening against an unreproducible result.                                                                   |
| Harness: can bundled skills be excluded?                                                | 4        | No — `--disable-slash-commands` and `--safe-mode` both remove the catalog plugin; `--disallowed-tools` still scores a denial as settled. Record via H4.                                                                                                                                                                                                                                                              |

---

## 6. What the investigation could NOT determine

| #   | Open question                                                                                                                                                                                                                                                                                                                                                                                                            | Evidence that would settle it                                                                                                                                                                | Cost                       |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------- |
| U1  | **Should a target-resolving unresolvable Skill name count as a trigger?** This is a metric definition, not a bug — S1a (max-turns 2, measure real loads) and S1b (count routing intent) give different numbers for memory-management (0/3 vs 3/3). E5 shows max-turns 2 does **not** rescue genuine competitor losses, so S1a stays honest; but it is untested on the bare-plugin-name case specifically.                | Run the 4 affected routing scenarios at `--max-turns 2` with the S1 classification and see whether the model self-corrects after `Unknown skill`.                                            | ~12 runs, ~$0.10           |
| U2  | **True selection rates for the 6 near-threshold items** (c99-opinionated ×1, code-quality, github, python, lua-guide, skill-guide). All sit at 3-6/9 pooled, where a 3-run vote at threshold 0.5 is maximally brittle.                                                                                                                                                                                                   | S3 escalation to n=9 after S1. Currently blocked by `validation.ts:28` `.max(3)` **and** the `batchSize × runs ≤ 24` accounting cap.                                                         | +14% trigger, +34% routing |
| U3  | **Is the accessibility loss to bundled `run` stable?** 2/7 pooled spans two harness configurations and includes one S1 case.                                                                                                                                                                                                                                                                                             | Post-S1 re-measure at n=9.                                                                                                                                                                   | ~9 runs                    |
| U4  | **How much has the bundled competitor set already drifted?** The advertised list I captured (16 entries incl. `deep-research`, `design-sync`, `debug`, `batch`, `doctor`, `run-skill-generator`) does not match the set named in the sweep brief (no `security-review`), **and E5 proves the list is incomplete** — `init` launches while unlisted. So neither the recorded set nor the advertised set is authoritative. | S2's `available_skills` + CLI version, recorded across two sweeps, cross-referenced against observed `Launching skill:` results rather than the init array. May force pinning the CLI in CI. | free with S2               |
| U5  | **Does D3's debugging edit pull memory-management further down?** Its own scenario is already at 0.333 and the two contest the word "leak".                                                                                                                                                                                                                                                                              | Re-run both scenarios together plus debugging's train split, before merging.                                                                                                                 | ~18 runs                   |
| U6  | **Does D4's description narrowing hold c99-opinionated's other 16 positives?** The shrink removes a trigger phrase that other passing queries may depend on.                                                                                                                                                                                                                                                             | Full `ci-skill-eval-trigger` for `skill-c99-opinionated` before and after.                                                                                                                   | ~2×18 runs                 |
| U7  | **Should dependency plugins be loaded in the per-skill gate at all?** E2 shows dropping them is harmful for c99-opinionated, but the comment at trigger-config.ts:87-92 asserts the opposite invariant, and 8 plugins are affected. The question is now _policy_, not measurement.                                                                                                                                       | Land H3 (correct the comment) + S2 (`loaded_skills`), then decide from a sweep where dependency-loaded wins are visible in the record rather than inferred.                                  | free with S2               |
