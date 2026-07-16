# GitHub Issues as an Optional Result Provider

Use GitHub Issues only when the workflow profile explicitly selects it or the GitHub adapter is the resolved work-item/result provider. The semantic result-provider port remains owned by **workflow-guide**; this adapter owns GitHub authentication, API resources, permissions, native identifiers, mutations, and limitations.

## Capability mapping

| Result-provider capability | GitHub-native realization                                                                                                               |
| -------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `resolve` / `read`         | Read `repos/{owner}/{repo}/issues/{number}` plus required issue comments or timeline resources                                          |
| `publish`                  | Create an issue or issue comment containing the selected result's semantic content and return the API/HTML/native node references       |
| `revise`                   | Update the owning issue or append/update the provider-native comment selected by the adapter                                            |
| `relate`                   | Record a GitHub-native cross-reference or comment containing the related opaque native reference; declare that this is not a typed edge |
| `version`                  | Report native IDs and `updated_at` freshness tokens; do not represent them as immutable issue revisions                                 |
| `capabilities`             | Report tested host/version, enabled Issues features, permissions, supported operations, rate limits, and exact-revision limitations     |

GitHub Issues do not provide an immutable revision for the whole mutable issue result. Re-read before revision-sensitive work and bind code review, QA, acceptance, or release evidence to the immutable commit/check/artifact revision that GitHub natively provides. Mark the issue `updated_at` token as freshness-only.

## Adapter rules

- Keep `{host, owner, repository, issue number or node id, selected comment id}` as opaque provider context; never turn a runtime trace into the result identity.
- Reconstruct the handle after process restart by reading GitHub. Do not depend on conversation state, a local mirror, or a sidecar result file.
- Use an idempotency marker or reconcile an existing issue/comment before retrying publish. GitHub's create endpoints do not make an arbitrary repeated result publication idempotent.
- Detect an `updated_at` or selected-comment change between read and mutation and report a stale-result conflict. Do not claim atomic compare-and-swap unless a tested native mechanism guarantees it.
- Return authentication, permission, rate-limit, unavailable-feature, and partial-publication failures explicitly. Never fall back to a local file or another provider for a requested side effect.
- Treat repository rules, checks, approvals, and protected environments as separate enforcement/evidence resources; an issue or comment is not enforcement proof.

Use the shared provider assertions from **workflow-guide**. GitHub is an optional hosted fixture; the required portability fixture remains the self-controlled non-file task-system provider.
