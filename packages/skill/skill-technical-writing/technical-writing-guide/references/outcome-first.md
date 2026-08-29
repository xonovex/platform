# outcome-first: Put the Main Message First

## Choose the Lead

Open with the information the reader needs to act or understand:

- **Answer**: Give the direct answer to a question.
- **Result**: State what changed or what the evidence shows.
- **Decision**: State the selected option and its main consequence.
- **Action**: State what must happen next when action is the purpose.
- **Status**: State the current condition, impact, and blocker for ongoing work.

Use the first complete sentence or short paragraph. A heading can identify the topic, but it does not replace the message. Do not add a visible `BLUF` label unless the format requires it.

## Order Supporting Detail

Use this default order:

1. Main message.
2. Required action or practical implication.
3. A caveat that can change the message.
4. Strongest evidence.
5. Secondary detail and chronology.

Do not repeat the lead as a closing summary unless the document is long enough to require one.

## Write Under Uncertainty

Do not force a conclusion from incomplete evidence. Lead with the decision status and the missing evidence:

> The cache choice is not final because peak memory for Cache A is unknown. Measure it under production load before selection.

Separate verified facts, inferences, and recommendations. Use confidence language only when the source provides a basis for it.

## Place Caveats by Consequence

Put a caveat beside the lead when it can reverse the decision, block the action, or materially narrow the claim. Put minor limitations after the primary evidence. A long list of edge cases at the start hides the result as effectively as chronology does.

## Adapt to the Artifact

- In a report, lead with the result and its consequence.
- In a status update or handoff, lead with current status, impact, blocker, and next action.
- In technical documentation, open each page and section with the answer, constraint, or reader action it exists to provide.
- In a tutorial, lead with the task and expected outcome, then give prerequisites and steps.
- In an incident account, lead with impact and current status, then give cause, remediation, and chronology.

Keep any required section structure from the applicable artifact skill. Apply outcome-first order inside each section.

## Before and After

Before:

> We tested two caches over three runs. Cache A was faster. Cache B used less memory. We still need peak-memory data for Cache A before choosing.

After:

> The cache choice remains open until Cache A's peak memory is measured. Cache A reduced latency more, while Cache B used less memory.

## Check

- Can the first sentence stand alone as the answer, result, decision, action, or status?
- Does decisive contrary evidence appear beside it?
- Does each later paragraph add evidence or necessary detail instead of repeating the lead?
- Can a reader stop after the opening and still know what matters?
