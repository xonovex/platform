---
type: decision-brief
parent_plan: plans/skill-catalog-eval-sweeps.md
status: open
---

# Should the output-gate tier floors be adjusted?

The one decision holding stabilization subplan 7. The sweep left 27 of 72 output
gates failing, and the go/no-go called this a calibration question. Measured
against the results, it mostly is not.

## Recommendation: leave the floors alone

Lowering the floor does not clear the failures. Applying a single floor to every
tier, against the 22 skills that miss their tier's absolute bar:

| floor applied to all tiers | failures cleared | still failing |
| -------------------------- | ---------------- | ------------- |
| 0.75                       | 9 of 22          | 13            |
| 0.70                       | 10 of 22         | 12            |
| 0.60                       | 11 of 22         | 11            |
| 0.50                       | 14 of 22         | 8             |

At 0.50 — a bar meaning "wrong half the time is acceptable" — 8 of the 22
still fail. A remedy that does not remedy the thing is the wrong remedy.

The floors are also demonstrably doing their job elsewhere: the five gates that
fail on delta alone score 0.75 to 1.00 absolute and fail correctly, because the
bare model already performs the task and the skill adds nothing measurable.

| skill                   | tier       | with | without | delta |
| ----------------------- | ---------- | ---- | ------- | ----- |
| `c99-guide`             | aggressive | 0.75 | 0.79    | -0.04 |
| `code-quality-guide`    | aggressive | 1.00 | 1.00    | +0.00 |
| `docker-guide`          | moderate   | 0.94 | 1.00    | -0.06 |
| `oop-guide`             | aggressive | 0.75 | 0.75    | +0.00 |
| `shell-scripting-guide` | aggressive | 0.92 | 1.00    | -0.08 |

## The failures are two problems, not one

### 9 near-misses — a tier-assignment question

Within 0.15 of their own floor, each with a healthy delta. Four sit at exactly
0.83 against conservative's 0.90; three at 0.75 against moderate's 0.80.

| skill                        | tier         | with | floor | gap  | delta |
| ---------------------------- | ------------ | ---- | ----- | ---- | ----- |
| `asset-pipeline-guide`       | conservative | 0.83 | 0.90  | 0.07 | +0.48 |
| `content-guide`              | conservative | 0.83 | 0.90  | 0.07 | +0.42 |
| `gpu-rendering-vulkan-guide` | conservative | 0.83 | 0.90  | 0.07 | +0.38 |
| `memory-management-guide`    | conservative | 0.83 | 0.90  | 0.07 | +0.33 |
| `skill-guide`                | conservative | 0.77 | 0.90  | 0.13 | +0.57 |
| `hono-opinionated-guide`     | moderate     | 0.75 | 0.80  | 0.05 | +0.38 |
| `moon-guide`                 | moderate     | 0.75 | 0.80  | 0.05 | +0.56 |
| `vitest-guide`               | moderate     | 0.75 | 0.80  | 0.05 | -0.12 |
| `gitlab-guide`               | moderate     | 0.71 | 0.80  | 0.09 | +0.55 |

Re-tiering only these would take the catalog from 45 to 54 of 72. Each move must
be argued from why that skill is lower-risk, never from the fact that it fails —
which is the one form of adjustment the standing rule permits.

### 13 far below — an eval-quality question

No defensible floor reaches these, and the without-skill arm is the tell: it is
0.00 for six of them, so the evals are unanswerable without the skill and only
partly answerable with it. That points at assertions demanding more than the
skill contains, or content that is genuinely absent — not at a floor set too high.

| skill               | tier         | with | floor | gap  | without |
| ------------------- | ------------ | ---- | ----- | ---- | ------- |
| `copilot-guide`     | moderate     | 0.25 | 0.80  | 0.55 | 0.00    |
| `opencode-guide`    | moderate     | 0.25 | 0.80  | 0.55 | 0.00    |
| `claude-code-guide` | moderate     | 0.33 | 0.80  | 0.47 | 0.00    |
| `command-guide`     | conservative | 0.33 | 0.90  | 0.57 | 0.00    |
| `kiro-guide`        | moderate     | 0.33 | 0.80  | 0.47 | 0.00    |
| `codex-guide`       | moderate     | 0.42 | 0.80  | 0.38 | 0.00    |
| `instruction-guide` | conservative | 0.44 | 0.90  | 0.46 | 0.25    |
| `plan-guide`        | moderate     | 0.48 | 0.80  | 0.32 | 0.50    |
| `pi-guide`          | moderate     | 0.50 | 0.80  | 0.30 | 0.00    |
| `workflow-guide`    | moderate     | 0.51 | 0.80  | 0.29 | 0.11    |
| `github-guide`      | moderate     | 0.54 | 0.80  | 0.26 | 0.04    |
| `git-guide`         | moderate     | 0.63 | 0.80  | 0.17 | 0.39    |
| `ecs-guide`         | conservative | 0.75 | 0.90  | 0.15 | 0.25    |

Most of these are one family — the harness and provider guides (`copilot`,
`opencode`, `kiro`, `codex`, `claude-code`, `pi`, `command`), the same cluster
that lost the trigger gate's boundary questions to Claude Code's bundled skills.
Something systematic affects that family's evals and deserves diagnosis before
anything is loosened. `plan-guide` is the worked precedent: the identical shape
turned out to be prompts naming artifacts the eval never supplied, not a weak
skill.

## Suggested order

1. Leave the floors as they are.
2. Re-tier the near-misses individually, each with a recorded risk argument.
3. Diagnose one provider guide by reading its judge evidence before touching the
   other twelve.
4. Ship subplan 7 with the far-below gates red and recorded as pre-existing
   eval-quality debt rather than regressions from this work.

## Caveat on this analysis

It reads tier assignments as deliberate risk labels, which was never verified. If
tiers were assigned by rule of thumb, step 2 is not re-tiering but assigning tiers
properly for the first time, and that should be settled before acting on it.
