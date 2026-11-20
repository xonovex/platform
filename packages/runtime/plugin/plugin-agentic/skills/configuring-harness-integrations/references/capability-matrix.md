# Capability Matrix

A capability matrix is the adapter table of one harness: it maps each portable semantic intent to the native event that implements it, with the support level and the limits that remain.

## Matrix identity

Every matrix opens with these fields:

| Field | Meaning |
| --- | --- |
| Matrix version | Version of the table itself |
| Documentation snapshot | Date of the documentation the rows come from |
| Runtime probe | The command that reports the installed version |
| Observed runtime | Version and date of the last probe, or `not installed` |
| Evidence status | `Documentation mapping only`, `Documentation-verified; runtime-unverified`, or runtime-verified |
| Refresh trigger | Product update, schema change, handler change, or 90 days |

A version probe alone leaves hook rows documentation-verified. Running the deterministic guard contract (a JSON event on standard input, exit `0` allow, exit `2` deny) locally proves the contract, not the native registration; only an exercised native registration moves a row to runtime-verified.

## Semantic intents

Rows use one vocabulary, so two harnesses compare row by row:

- Session start or resume
- Prompt before submission
- Tool before use
- Permission request
- Tool after use or failure
- Context compaction
- Subagent lifecycle
- Model evaluator
- Agent verifier

A harness adds a row for an intent only it has. It adds columns for native mapping, support, blocking or context behavior, and ordering or limits.

## Guarantee boundary

A hook is not an independent sandbox. Confirm the exact event, the handler type that executes, the covered tools, the output schema, the exit behavior, the permission mode, the timeout, the sibling behavior, and the active configuration sources. Use independent enforcement for every operation the native event cannot cover, and claim a mandatory control only for the operations the matrix shows as intercepted.

Ordering and concurrency that the snapshot does not state stay unknown until a probe establishes them.
