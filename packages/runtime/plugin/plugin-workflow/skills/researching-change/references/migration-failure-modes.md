# Known failure modes

Each mode has a mechanism and a fix. Check the draft research file against all five before saving it.

## Contents

- Stale stamps hide drift
- A design document claims behavior the code lacks
- Labels weaken in synthesis
- Wrapped paths cannot be spot-checked
- Collapsed commands cannot be reconstructed

## Stale stamps hide drift

Mechanism: a page records the commit it was written at, the repository has moved past that commit, and the page keeps reading as current. A reader trusts a description of code that has changed.

Fix: compare each stamped commit with the repository head. Report a page that trails as a correction to that page, never as a fact about the system, and read the code for the fact.

## A design document claims behavior the code lacks

Mechanism: a design record describes persistence, deduplication or a retry that was planned, removed or never merged. An explorer that reads only the record reports the behavior as present.

Fix: have the target explorer look for the code that does what the document claims. Report what the code does, name the document under `# Corrections`, and list the behavior as a gap when a consumer depends on it.

## Labels weaken in synthesis

Mechanism: an explorer writes `Inference`, the first summary calls it "probably", and the next summary drops the hedge. An unverified claim ends up as a fact in the bottom line.

Fix: keep the label wherever the claim is repeated, including the bottom line and the gap table. Re-verify the few claims that would change a decision, and say which side you checked.

## Wrapped paths cannot be spot-checked

Mechanism: a path wrapped over two lines in an explorer report cannot be searched or pasted, so the claim it supports is never checked.

Fix: require one anchor per line as repository root plus path and line in every brief, and reject a report that wraps them.

## Collapsed commands cannot be reconstructed

Mechanism: a summary says that the repositories were checked but not how, so a later reader cannot repeat the check or tell which commit it ran against.

Fix: list the verification commands and the commit of each repository in the research file, in `# Not checked` or in the footnotes, as plain lines.
