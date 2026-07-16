# Migrating from 5.x to 6.0.0

## 6.0.1

Patch release. Nine skill packages moved their guide directories from a
nested `skills/<topic>-guide/` layout to the flat `<topic>-guide/` layout
their plugin manifests now reference: skill-accessibility,
skill-ai-governance, skill-aws, skill-azure-devops, skill-bitbucket,
skill-bitrise, skill-datadog, skill-reliability, and
skill-security-assurance. Harnesses resolve skills through each plugin's
manifest, so no consumer action is needed beyond upgrading; the previous
nested layout was unreliable for the Codex loader. The release also adds
the executable governance walking skeleton, the adversarial conformance
scenario corpus with machine runners, and repository-decoupled package
validation.

Version 6 is a coordinated breaking release of every command and skill plugin. Keep command plugins, skill plugins, plugin manifests, and the marketplace on one major version; mixed 5.x/6.x installations are unsupported.

## Why this is breaking

The workflow changed from a mostly local plan-and-worktree sequence into provider-native lifecycle results plus an independently adoptable governance plane. Removed commands have no compatibility wrappers, acceptance now separates evidence assembly from accountable decision, and commands preserve opaque native result references rather than assuming local files are the only source of truth.

## Command migration

| 5.x command or assumption                         | 6.0.0 replacement                                                                                                                                                              |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `story-refine` owns discovery and story structure | Use `discovery-run`, then `formulation-run`; select `user-stories-guide` only when stories are the chosen method                                                               |
| `acceptance-formalize` requires Given-When-Then   | Use `formulation-run`; select BDD/example mapping only when the profile chooses it                                                                                             |
| `plan-decide` stores an inline planning decision  | Use `decision-create`, `decision-critique`, `decision-revise`, and authority-bound `decision-accept` against exact revisions                                                   |
| `acceptance-validate` implies acceptance          | Use `acceptance-validate` to assemble fresh evidence and `acceptance-decide` for the accountable human decision                                                                |
| Planning always begins from local plan files      | Planning commands accept opaque lifecycle/provider references; local Markdown remains one valid provider                                                                       |
| Development is implicit inside `plan-continue`    | Use `develop-run` for exact assignments and `develop-consolidate` or `develop-abandon` for explicit outcomes; `plan-continue` remains the plan-document continuation operation |
| PR publication completes delivery                 | Use `deliver-publish`, review/QA/assessment results, acceptance, integration, transition, and release capabilities selected by the profile                                     |
| Hooks and lifecycle form one stack                | Install governance, onboarding, or external enforcement independently; lifecycle commands are not prerequisites                                                                |

All existing `plan-*`, worktree, PR, and review command names that remain in the package keep thin command contracts, but their owning skills now reconstruct provider-native state and preserve exact revisions, limitations, and evidence.

## Upgrade procedure

1. Inventory installed command/skill plugin versions, native harness configuration, provider adapters, executable modules, and retained result/evidence references.
2. Read the [architecture guide](docs/architecture-and-composition.md) and choose workflow-only, governance-only, enablement-only, external-enforcement-only, or integrated adoption.
3. Preview package changes and the effective module/profile composition. Review new permissions, data flows, native files/settings, external controls, verification probes, and rollback targets.
4. Upgrade the marketplace plus every selected command and skill plugin to `6.0.0` in one transaction.
5. Replace removed command invocations and automation with the mappings above. Do not add wrappers that preserve the removed semantics.
6. Run native harness diagnostics, module/provider conformance, allow/deny and unsupported-path probes, workflow fixture tests, and the repository release validation.
7. Record observed versions, effective composition, limitations, native references, and rollback evidence.

## Rollback

Pin the marketplace and every installed command/skill plugin back to the last verified `5.0.0` set. Restore captured native configuration and executable digests through each onboarding adapter, remove only v6-owned entries, and re-run v5 diagnostics. Preserve provider-native workflow, authorization, policy, enforcement, and audit evidence even when the v5 client cannot interpret a v6 semantic result.

Do not translate a v6 acceptance, enforcement, exception, break-glass, or privileged-operation result into a weaker v5 state. Pause the affected operation or retain a read-only evidence link until an authorized migration path is available.
