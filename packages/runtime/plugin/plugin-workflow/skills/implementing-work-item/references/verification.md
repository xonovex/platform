# Verification and self-checks

The checks decide whether the work is done, and the supervisor runs them. A worker's report is input to this step and never a substitute for it.

## Contents

- What to run
- Reading the results
- Labelling
- What the run does not claim

## What to run

1. Every command the `verification` of a criterion names, from the worktree.
2. The repository's own format, lint, type and test checks for what changed, found in its instructions or its build files. Compare a failure with the state before your edits so a new failure is not mistaken for an old one.
3. For a criterion with a visible result, open the artifact behind the assertion: the screenshot, the rendered page, the output file. An assertion can pass while the capture shows an empty panel.

Record each command exactly as run and its outcome in the `checks` list of the notes. A command you could not run (a missing tool, an unreachable service) is recorded with that as its result, and the criterion that depended on it is `blocked`.

## Reading the results

- A passing test does not prove a criterion. Read what the test asserts against the criterion's text.
- A criterion is `met` only when you ran its verification, or checked it by the means the criterion names, after the last edit that could affect it.
- Treat editor diagnostics and stale indexes as hints. A real build settles a dispute.
- A measured result that is close but short is not met. Record the number.

## Labelling

Everything this step concludes is a self-check. The notes say so in the heading `# Self-check against the criteria`, and nothing in them is worded as validated, verified or passed. Only **validating-work-item**, bound to the worktree it checked, supplies a mark that a pull request description may show.

## What the run does not claim

- It does not mark the work item done or change its status: only **validating-work-item** writes a verdict.
- It does not edit the work file. Criteria, readiness and dependencies are definitions, and a change to them goes through the plan.
- It does not claim the criteria are met when a check could not run.
