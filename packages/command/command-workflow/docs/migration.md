# Composition Contract Migration

The version 1 workflow contract removes ambiguous global selections and overlapping
effects. Commands do not provide compatibility aliases because aliases would preserve
the ambiguity this migration fixes.

## Flag mapping

| Previous shape              | Version 1 shape                                                                                                      |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `--provider`                | `--subject-provider`, `--destination-provider`, or `--workspace-provider`; use a request file for multiple providers |
| `--reference`               | Named `inputs` bindings in a request file                                                                            |
| `--revision`                | Revision on the exact subject, input, destination, or workspace binding                                              |
| `--kind`                    | `--subject-kind` or the kind on each request binding and result contract                                             |
| singular `--perspective`    | Repeatable `--perspective`                                                                                           |
| `--criteria`                | Repeatable `--criterion` or provenance-aware criteria in a request                                                   |
| `--capability`              | Runtime-derived skill, tool, or adapter selection reported in the result                                             |
| `--result` on non-publish   | Inline operation result, followed by a separate `publish` call when persistence is required                          |
| `--confirm` and `--dry-run` | Explicit `--effect inspect`, `preview`, or `apply` where supported                                                   |
| merge or abandon `--remove` | Separate `workspace-cleanup` call                                                                                    |

## Request migration

Use the simple subject shorthands only when one provider owns the subject and all
supporting information is inline. Move every advanced call to a request document:

1. Name each resource by its semantic slot.
2. Bind each resource to its own provider, opaque reference, revision, and kind.
3. Separate subject, evidence, policy, work record, and destination.
4. Make perspectives repeatable.
5. Mark criterion provenance and whether an authoritative source makes it binding.
6. Remove generic capability choices; declare an exact expert skill requirement only
   in the advanced request contract when deterministic selection is necessary.
7. Select an effect mode only for an operation that supports effects.

## Result migration

Consumers should stop scraping free-form prose for workflow state. Read the versioned
operation-result envelope and preserve:

- exact input bindings;
- resolved methods, perspectives, criteria, skills, and reasons;
- completed, partial, blocked, or failed status;
- evidence and observed effects;
- unresolved questions and uncertainty;
- safe retry boundary;
- provider-native result or checkpoint reference when one exists.

Cross-session continuation uses a persisted provider-native work record and exact
revision, never a conversational pointer.

## Workspace migration

Workspace effects are intentionally atomic:

- create creates;
- merge validates and integrates;
- abandon records why work stopped;
- cleanup removes.

Convenience automation may compose these operations, but its audit trail must retain
the individual previews, authorizations, results, and failures.
