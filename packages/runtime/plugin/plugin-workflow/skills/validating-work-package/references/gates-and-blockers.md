# Blockers and gates

## Contents

- Kinds of blocker
- Describing a blocker
- The gate chain
- What is no longer a blocker

## Kinds of blocker

Open with a count: how many things stop the plan from being done, and how many of them are code. The ones that decide the schedule are rarely the code.

| Kind | What it is | Example |
| --- | --- | --- |
| human or external gate | A step no code can satisfy: an approval, an enrollment, a release by another party | A signing certificate that another department issues |
| design gate | A problem with no known workable design yet | A contract the two sides cannot both meet |
| residue | A remainder that needs a person or a device | A manual check on hardware the team does not hold |

## Describing a blocker

For each blocker:

- Name, kind, owner, and what waits on it.
- Human or external gate: the exact steps, who performs each, before which milestone, and which step fails silently if it is skipped.
- Design gate: the sharpest assessment so far, the sub-problems that no tooling catches, and any workable design already found, so the next attempt starts from there and not from zero.
- State the mechanism, not the symptom. "The build fails" misdirects; "the packaging step reads the old manifest field" points at the fix.

## The gate chain

When more than one gate exists, draw which gates feed which and state which tracks are independent and can run in parallel. A short list of arrows in the body is enough; a diagram file is the job of **illustrating-plan**, and the close-out may point to one that exists.

## What is no longer a blocker

List the risks that closed during the work, each with the evidence that closed it, so nobody argues them again. A risk is closed only by evidence gathered in this run or cited from a file that still shows it.
