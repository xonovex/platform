# GitLab Issues as an Optional Result Provider

Use GitLab Issues only when the workflow profile explicitly selects it or the GitLab adapter is the resolved work-item/result provider. The semantic result-provider port remains owned by **workflow-guide**; this adapter owns GitLab authentication, edition/version capabilities, API resources, permissions, native identifiers, mutations, and limitations.

## Capability mapping

| Result-provider capability | GitLab-native realization                                                                                                                             |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `resolve` / `read`         | Read `projects/:id/issues/:issue_iid` plus required issue notes, discussions, or resource events                                                      |
| `publish`                  | Create an issue or issue note containing the selected result's semantic content and return project/id/iid/note references                             |
| `revise`                   | Update the owning issue or update/append the provider-native note selected by the adapter                                                             |
| `relate`                   | Record a GitLab-native cross-reference, link, discussion, or note containing the related opaque reference and declare the relationship strength       |
| `version`                  | Report native IDs and `updated_at` freshness tokens; do not represent them as immutable issue revisions                                               |
| `capabilities`             | Report offering, edition/tier, tested version/date, permissions, supported operations, pagination/rate limits, and exact-revision/relationship limits |

GitLab Issues do not provide an immutable revision for the whole mutable issue result. Re-read before revision-sensitive work and bind code review, QA, acceptance, or release evidence to immutable commit, pipeline, job, artifact, or deployment references. Mark issue and note `updated_at` values as freshness-only.

## Adapter rules

- Keep `{host, project id or encoded path, issue iid/id, selected note/discussion id}` as opaque provider context; never turn a runtime trace into result identity.
- Reconstruct the handle after process restart by reading GitLab. Do not require conversation memory, a checkout, or a universal result file.
- Reconcile an existing issue/note before retrying publish and disclose duplicate risk when no native idempotency facility covers the operation.
- Re-read the issue and selected note before mutation; report a stale-result conflict when freshness changes. Do not claim atomic compare-and-swap without a tested native guarantee.
- Preserve confidential/internal visibility, permissions, and data classification. Minimize copied issue/note content in telemetry and evidence.
- Return authentication, role, tier/edition, rate-limit, provider-outage, and partial-publication failures explicitly. Never silently fall back to another provider.
- Treat approval rules, pipeline policies, protected environments, and compliance controls as separate enforcement/evidence resources; an issue or note is not enforcement proof.

Use the shared provider assertions from **workflow-guide**. GitLab is an optional hosted fixture; the required portability fixture remains the self-controlled non-file task-system provider.
