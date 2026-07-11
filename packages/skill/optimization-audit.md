# Skill Catalog Optimization Audit

Record of the pass that trimmed all 74 skills to their delta over baseline model
knowledge, and the knowledge-grounded ablation that verified it.

Commits: `6af54791` (trim), `48d551d9` (restore).

## Result

- 74 skills, **32,332 → 22,310** lines across `SKILL.md` + `references/` (**−31%**).
- Tiers: 25 aggressive, 25 moderate, 24 conservative.
- `validate.py` + `moon :ci-check` pass on every skill; `name`/`description` (routing) unchanged.
- Ablation on the weakest model (Haiku): **72/74 skills filler-only**; 3 removed facts were load-bearing and were restored.

## Method

Per skill: classify tier → trim at that depth → intra-skill dedup → fix genuine defects → `validate.py` gate.

Tier depths:

- **aggressive** — generic languages / SE concepts the model already owns: strip rationale + generic procedure, keep the rule + exact identifiers + exceptions + one example.
- **moderate** — frameworks/tools + opinionated overlays: keep the API/opinion delta, cut filler + duplicate examples.
- **conservative** — version-pinned / exact-fact / spec-heavy: light touch, dedupe only — the specifics are the value.

Verification: blind A/B with a binary per-rubric-point compliance metric, k=3 samples, across Haiku/Sonnet/Opus producers with Opus judges (pilot); then a knowledge-gap ablation over all removed content on Haiku (full catalog).

## Pilot A/B evidence (per-model compliance, higher = better)

| Producer | no-skill  | original | trimmed |
| -------- | --------- | -------- | ------- |
| Haiku    | 0.72–0.89 | 1.00     | 1.00    |
| Sonnet   | 0.67–0.79 | 1.00     | 1.00    |
| Opus     | 0.69–0.89 | 1.00     | 1.00    |

`no-skill` trailing the skill confirms the content is not redundant on these rules; `trimmed = original` confirms trimming kept the value.

## Ablation refinement (Haiku, over removed content)

Stage A audited every skill's removed lines — **72/74 filler-only**. Stage B ablated the 3 flagged items (3 arms × k=3 samples, correctness-graded):

| Removed item                       | no-skill | trimmed | original | Action                      |
| ---------------------------------- | -------- | ------- | -------- | --------------------------- |
| moon: distroless uid `65532` chown | 0.33     | 0.33    | 1.00     | restored (load-bearing)     |
| moon: `.moon/extensions.yml`       | 0.67     | 0.67    | 1.00     | restored (reliability)      |
| expressjs: JWT expiration          | 1.00     | 0.67    | 1.00     | restored (security default) |

## Defects fixed during the pass (10)

- react / remotion / threejs / strudel — broken `# filename:` template headers → proper titles
- c99-opinionated `safety-validations` — buffer-overflow check ordering (overflow guard before the capacity add that could itself overflow)
- python — `any` → `typing.Any` in `dict[...]` annotations
- kubernetes — deprecated Kustomize-2 fields → Kustomize-5 (`resources`/`patches`/`labels`)
- lua — global `unpack` → `table.unpack` (Lua 5.4 target)
- hono-opinionated — duplicate `OpenAPIHono` import
- testing — duplicated heading

## References merged (11, intra-skill dedup)

typescript (`avoid-barrel-exports` + `avoid-reexports` → `avoid-indirect-exports`), react, vitest (×3), typescript-to-lua (×3), zod, git, user-stories.

## Per-skill detail

Before/after are total lines across `SKILL.md` + `references/`, measured from git (`HEAD~2` original vs `HEAD` final).

| Skill                        | Tier         | Before | After | Reduction | Ablation         |
| ---------------------------- | ------------ | -----: | ----: | --------: | ---------------- |
| `skill-shell-scripting`      | aggressive   |    665 |   178 |       73% | filler-only      |
| `skill-vitest`               | moderate     |    480 |   132 |       72% | filler-only      |
| `skill-python`               | aggressive   |    589 |   176 |       70% | filler-only      |
| `skill-sql-postgresql`       | aggressive   |    471 |   155 |       67% | filler-only      |
| `skill-typescript-to-lua`    | moderate     |    803 |   270 |       66% | filler-only      |
| `skill-zod`                  | moderate     |    330 |   114 |       65% | filler-only      |
| `skill-lua`                  | aggressive   |    419 |   154 |       63% | filler-only      |
| `skill-remotion`             | moderate     |    887 |   327 |       63% | filler-only      |
| `skill-git`                  | moderate     |    655 |   246 |       62% | filler-only      |
| `skill-terraform`            | moderate     |    536 |   208 |       61% | filler-only      |
| `skill-cmake`                | moderate     |    388 |   168 |       57% | filler-only      |
| `skill-typescript`           | aggressive   |    355 |   152 |       57% | filler-only      |
| `skill-hono-opinionated`     | moderate     |    461 |   201 |       56% | filler-only      |
| `skill-microkernel-pattern`  | aggressive   |    416 |   191 |       54% | filler-only      |
| `skill-hexagonal-pattern`    | aggressive   |    361 |   169 |       53% | filler-only      |
| `skill-reflect`              | aggressive   |    215 |   104 |       52% | filler-only      |
| `skill-moon`                 | moderate     |    670 |   335 |       50% | 2 facts restored |
| `skill-hono`                 | moderate     |    661 |   339 |       49% | filler-only      |
| `skill-lua-opinionated`      | moderate     |    140 |    71 |       49% | filler-only      |
| `skill-pull-request`         | aggressive   |    296 |   155 |       48% | filler-only      |
| `skill-connascence`          | aggressive   |    318 |   173 |       46% | filler-only      |
| `skill-plan`                 | moderate     |    984 |   530 |       46% | filler-only      |
| `skill-strudel`              | moderate     |    540 |   297 |       45% | filler-only      |
| `skill-c99-game-opinionated` | moderate     |    374 |   215 |       43% | filler-only      |
| `skill-user-stories`         | aggressive   |    432 |   247 |       43% | filler-only      |
| `skill-astro`                | moderate     |    228 |   135 |       41% | filler-only      |
| `skill-fp`                   | aggressive   |     34 |    20 |       41% | filler-only      |
| `skill-expressjs`            | moderate     |    475 |   286 |       40% | 1 fact restored  |
| `skill-tdd`                  | aggressive   |    240 |   143 |       40% | filler-only      |
| `skill-docker`               | moderate     |    216 |   132 |       39% | filler-only      |
| `skill-fdd`                  | aggressive   |    183 |   113 |       38% | filler-only      |
| `skill-bdd`                  | aggressive   |    257 |   165 |       36% | filler-only      |
| `skill-c99`                  | aggressive   |    508 |   327 |       36% | filler-only      |
| `skill-npm`                  | moderate     |    141 |    90 |       36% | filler-only      |
| `skill-orthogonal-pattern`   | aggressive   |    919 |   591 |       36% | filler-only      |
| `skill-testing`              | aggressive   |    361 |   235 |       35% | filler-only      |
| `skill-kubernetes`           | moderate     |    316 |   208 |       34% | filler-only      |
| `skill-ddd`                  | aggressive   |    244 |   172 |       30% | filler-only      |
| `skill-oop`                  | aggressive   |     46 |    32 |       30% | filler-only      |
| `skill-motion-react`         | moderate     |    570 |   411 |       28% | filler-only      |
| `skill-instruction`          | conservative |    383 |   280 |       27% | filler-only      |
| `skill-react`                | moderate     |   1031 |   768 |       26% | filler-only      |
| `skill-code-review`          | aggressive   |    293 |   219 |       25% | filler-only      |
| `skill-code-quality`         | aggressive   |    189 |   146 |       23% | filler-only      |
| `skill-debugging`            | aggressive   |    308 |   238 |       23% | filler-only      |
| `skill-threejs`              | moderate     |    620 |   479 |       23% | filler-only      |
| `skill-c99-opinionated`      | moderate     |    700 |   589 |       16% | filler-only      |
| `skill-command`              | conservative |    454 |   398 |       12% | filler-only      |
| `skill-presentation`         | conservative |    124 |   112 |       10% | filler-only      |
| `skill-content`              | conservative |    268 |   245 |        9% | filler-only      |
| `skill-game-networking`      | conservative |    348 |   315 |        9% | filler-only      |
| `skill-versioning`           | conservative |    215 |   198 |        8% | filler-only      |
| `skill-llmstxt`              | conservative |    378 |   351 |        7% | filler-only      |
| `skill-data-oriented-design` | conservative |    700 |   657 |        6% | filler-only      |
| `skill-android-analytics`    | conservative |    433 |   411 |        5% | filler-only      |
| `skill-android-wcag`         | conservative |   1235 |  1174 |        5% | filler-only      |
| `skill-adr`                  | aggressive   |    137 |   132 |        4% | filler-only      |
| `skill-cross-platform`       | conservative |    337 |   322 |        4% | filler-only      |
| `skill-node-graph`           | conservative |    375 |   359 |        4% | filler-only      |
| `skill-caveman`              | aggressive   |     37 |    36 |        3% | filler-only      |
| `skill-github`               | moderate     |    481 |   467 |        3% | filler-only      |
| `skill-memory-management`    | conservative |    234 |   227 |        3% | filler-only      |
| `skill-editor-viewport`      | conservative |    371 |   362 |        2% | filler-only      |
| `skill-fable`                | conservative |     61 |    60 |        2% | filler-only      |
| `skill-lock-free`            | conservative |    894 |   880 |        2% | filler-only      |
| `skill-skill`                | conservative |   1284 |  1263 |        2% | filler-only      |
| `skill-gitlab`               | moderate     |    435 |   432 |        1% | filler-only      |
| `skill-asset-pipeline`       | conservative |    318 |   318 |        0% | filler-only      |
| `skill-audio`                | conservative |    326 |   326 |        0% | filler-only      |
| `skill-data-model`           | conservative |    515 |   515 |        0% | filler-only      |
| `skill-ecs`                  | conservative |    253 |   253 |        0% | filler-only      |
| `skill-gpu-rendering`        | conservative |    566 |   566 |        0% | filler-only      |
| `skill-gpu-rendering-vulkan` | conservative |    456 |   456 |        0% | filler-only      |
| `skill-imgui`                | conservative |    389 |   389 |        0% | filler-only      |

## Method notes

- The trim decision was made by a strong model's judgment of "filler vs delta"; the ablation re-checked the **removed** content against the **weakest** model (Haiku), so the final catalog is trimmed to what Haiku provably does not already produce — the three exceptions restored.
- Conservative-tier skills show near-zero reduction by design: their value is exact facts, so only obvious duplication was removed.

## Re-running the evals (and adding a new model)

Every skill carries an `evals.json` next to its `SKILL.md` — the output-eval seed consumed by `eval-outputs.py`. It holds the knowledge probes (`prompt` + binary `assertions`) that seed any A/B or ablation run. **291 evals across the catalog.** This is the durable "something to start from": no probe needs re-authoring to test a new model.

### Test one skill on any model

```bash
uv run packages/skill/skill-skill/skill-guide/scripts/eval-outputs.py \
  packages/skill/<skill>/<guide>/evals.json <plugin>:<guide> --model <model>
# e.g. …/eval-outputs.py packages/skill/skill-typescript/typescript-guide/evals.json \
#        xonovex-skill-typescript:typescript-guide --model haiku
```

Each eval runs twice — **with** the skill and **without** it (no-skill baseline) — graded on its assertions by an LLM judge. The script exits `0` iff with-skill beats no-skill. Needs the `claude` CLI, and `--eval-cwd` set to where the skill resolves (installed plugin or this repo).

### When a new model is added

The **weakest model you deploy is the gate** — "what the model already knows" is model-specific, so re-verify against the new model. Re-run the seed across the catalog:

```bash
for d in packages/skill/skill-*/*/evals.json; do
  guide=$(basename "$(dirname "$d")")
  uv run packages/skill/skill-skill/skill-guide/scripts/eval-outputs.py \
    "$d" "xonovex-skill-${guide%-guide}:$guide" --model <new-model>
done
```

Read each eval's `with_skill` vs `without_skill` pass rate:

- `without_skill` **high** → the new model already knows it → that content is now redundant → **trim-further** candidate (only act if this model is the weakest you run).
- `with_skill` high, `without_skill` low → still **load-bearing** → keep (restore if a trim removed it).
- `with_skill` low → the skill or the eval needs attention.

Load-bearing is measured, not declared: after any future trimming pass, re-run the seed on the weakest model (or the two-stage ablation: diff removed content → ablate on the weakest model); an eval a model fails _with_ the skill is one a trim broke — restore it, exactly as this session did.
