# sections: Writing each ADR section

## Two modes: new vs retroactive

- **Not-yet-implemented (true proposal)**: Status starts `Pending`. Context describes the need and the current gap; Proposal describes the intended design. Flip to `Accepted` after sign-off.
- **Existing-but-undocumented**: Write it retroactively, but still in proposal voice and present tense as if to-be-built. Status is usually `Accepted`. Context still contrasts the standing design with the ad-hoc approach the now-documented design replaces; ground Proposal and detail sections in the real code that exists.
- Both modes use the identical structure. Only the Status value and how much existing code you cite differ.

## Context

- Open with the need / forces in one or two sentences: what is missing and what it must carry.
- State the standing architecture or principles as a short bulleted list.
- Mark the single deviation with "Today, though, ...": how the current implementation falls short.
- Close with the need: "We need ..." and the property the proposal must guarantee (no drift, type safety, offline support, etc.).
- Keep target-state claims out of this section.

## Proposal

- One paragraph stating the decision in present tense: "This ADR proposes ...".
- Follow with `**With the following principles:**` and one bullet per principle the solution upholds (single source of truth, typed contract, who owns what, generation/lock-step, etc., as fits the decision).
- Add `###` detail subsections only when they pin down the contract: schemas, type/code excerpts, or a catalogue table of variants. Use real, minimal excerpts; do not paste whole files.

## Example uses

- One `###` heading per concrete scenario.
- Show a real example payload in a fenced block (e.g. `json5`), then a numbered walkthrough of what the producer emits and how the consumer reacts, ending in the result.
- Two short scenarios usually beat one exhaustive one.

## Decision

- A single line mirroring Status in brackets: `[Pending]` or `[Accepted]`.
- For a superseded decision, state it plainly here and reference the superseding ADR.

## Optional extra sections

- Some teams add **Consequences** (trade-offs accepted) or **Alternatives considered**. Include them only if the house examples do; otherwise the four core sections above suffice. When in doubt, match the sibling ADRs in the same folder.
