# Operation Composition

## Independent dimensions

| Dimension   | Question                                               | Rule                                                |
| ----------- | ------------------------------------------------------ | --------------------------------------------------- |
| Operation   | What should happen?                                    | Select one core verb per call.                      |
| Kind        | What kind of subject or result is involved?            | Keep open; infer only when unambiguous.             |
| Perspective | Which evidence or stakeholder lens matters?            | Change emphasis, not operation semantics.           |
| Method      | Which procedure should guide the operation?            | Load the selected method capability.                |
| Executor    | Who or what performs the call?                         | Never change the selected operation.                |
| Capability  | Which installed domain, method, or adapter is needed?  | Load only when selected or unambiguous.             |
| Trigger     | What initiated the call?                               | Supply context without adding a stage or authority. |
| Provider    | Which native system reads or persists?                 | Own authentication, revisions, and effects.         |
| Reference   | Where is the native subject, evidence, or destination? | Remain opaque outside the provider.                 |

Criteria, supporting references, source revisions, result destinations, and dry-run or confirmation choices remain explicit inputs rather than hidden dimensions.

## Selection procedure

1. Resolve the exact subject and requested operation.
2. Resolve every explicit dimension independently; infer kind or provider only from unambiguous evidence and report the basis.
3. Load only the selected or unambiguous domain, method, execution, workspace, and provider capabilities.
4. Stop and name any unavailable explicit capability instead of substituting another.
5. Perform only the selected operation and preserve its side-effect boundary.
6. Return the result inline unless the caller explicitly selects a destination provider and reference.

## Composition

Examples such as `create → review → revise`, `review → publish`, `execute → validate`, and `create → decide` are optional compositions. Calls may be omitted, repeated, reordered, or used alone when the subject and selected method require it. A role lens may suggest a composition but never grants permission, selects a provider, or changes a verb's semantics.
