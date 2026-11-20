# Results and evidence

## Contents

- One result for each criterion
- Pass, fail and blocked
- Evidence and its limits
- What the validator never does

## One result for each criterion

Every `criteria[].id` of the work file gets exactly one result, in the order of the work file. A criterion is judged on its own: a pass on one says nothing about another, and a failing criterion does not stop the rest from being checked. The work file lists the ids; compare them with the verdict's `criterion` values before writing, because the schema cannot.

## Pass, fail and blocked

| Result | Meaning |
| --- | --- |
| `pass` | You ran the verification, or checked the criterion by the means it names, against this tree, and observed what the criterion states. |
| `fail` | You ran the verification and the criterion does not hold, or the diff shows the behavior the criterion needs is absent. |
| `blocked` | You could not run or observe it: a tool is missing, a service is unreachable, an input is absent, or the verification is too vague to run. |

- `blocked` is not `pass` and is not `fail`. State what would unblock it and who can provide that.
- A measured result that is close but short is `fail`. Record the number.
- A criterion whose verification names a means you cannot use (a device, an account, a pre-production environment) is `blocked` even when a lighter check looks fine. A lighter check may be recorded as evidence, never as the pass.
- Context may explain intent. It cannot make a criterion pass.
- How to run a verification and read what it shows is in [running-checks.md](running-checks.md).

## Evidence and its limits

Evidence is what another person can use to get the same result without you.

- In the frontmatter, `evidence` is one sentence: the command or anchor and what it showed, for example "npm test -- retry: 12 passed, includes the limit case in src/retry.test.ts:41".
- In the body, each criterion has `Command`, `Observed` and `Limits`. `Observed` is a short excerpt of output, a measured number or a file and line, not a transcript. `Limits` says what the check does not cover: a platform not tried, an input not exercised, a baseline not known.
- An anchor is the repository root, the path and a line or a line range (`src/retry.ts:10-14`), one anchor per line. Keep the path on one line, because a wrapped path cannot be spot-checked.
- A secret that appears in output stays out of the file, by the rule for secrets of **recording-handoff**.
- A claim you did not observe is labelled an inference wherever it is repeated, and an inference cannot carry a `pass`.

## What the validator never does

- Fix a failure, adjust a test, update a snapshot or change a fixture to get a pass.
- Edit the work file, the plan or the implementation notes.
- Commit, stage or push.
- Run a check that writes outside the worktree or the system's temporary directory, or that calls a service the verification does not name.
- Report a pass it did not observe, however likely it seems.
