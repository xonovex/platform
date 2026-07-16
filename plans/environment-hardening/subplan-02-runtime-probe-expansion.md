---
type: plan
has_subplans: false
parent_plan: environment-hardening
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
  - packages/skill/skill-codex/codex-guide/references/**
  - packages/skill/skill-kiro/kiro-guide/references/**
  - packages/skill/skill-copilot/copilot-guide/references/**
  - packages/skill/skill-pi/pi-guide/references/**
  - packages/skill/skill-opencode/opencode-guide/references/**
  - packages/skill/skill-{azure-devops,bitbucket,bitrise,aws,datadog}/*-guide/references/**
  - plans/composable-workflow-phases/VALIDATION.txt (shared with siblings — own section only)
skills_to_consult:
- skill-guide
- testing-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 02: Runtime Probe Expansion

## Objective

Convert documentation-derived candidates into honestly-scoped runtime evidence where the host already permits it, and executable runbooks where it does not — with every matrix row's claim matching exactly what ran, per Decision 4's boundary: "guard contract exercisable with credentials already present on this host."

## Tasks

1. **Inventory the host.** For each of Codex, Kiro, Copilot CLI, Pi, OpenCode: is a binary installed (`command -v`), does `--version` (or equivalent) run, and does a working credentialed configuration already exist (config files/keychain — do not create any)? Record the three-way expected outcome per harness (full probe / version-only / runbook) before touching any matrix.
2. **Full probes** where a credentialed install exists: follow the Claude Code precedent exactly — observed version + date, guard exit-code contract exercised locally, explicit non-claims (native hook registration not exercised), rerun-on-version-change trigger — recorded in that harness guide's capability matrix.
3. **Version-only observations** where the CLI runs but the contract is credential-blocked: record verbatim "CLI version observed; contract not exercised — hook semantics remain documentation-derived" with date. These rows stay candidates in the qualified count.
4. **Probe runbooks** for every remaining candidate (harnesses without even a version observation, plus the five enterprise platforms): an executable section in the owning guide's onboarding/conformance reference — exact commands, credentials required and their scope, expected evidence, which matrix fields to fill, and the rerun trigger. No credential acquisition; runbooks are for whoever holds them.
5. **Reconcile the qualified result line** in VALIDATION.txt (own section): candidate count equals the matrices exactly (only full probes leave the count); declare the permanent caveats — AWS trust-policy and Datadog data-collection org reviews, Services/Server and Cloud/DC detection, crosswalk non-certification — as standing items with owner "organization adopting the module", distinct from closable items.
6. **Spot-audit and gate.** Re-read every touched matrix row against what actually executed; run the release validator and full `:ci-check`; one conventional commit.

## Validation Steps

- Inventory table recorded in the subplan on completion (harness → outcome class → evidence).
- Each touched matrix row's evidence status is reproducible from the recorded commands.
- Release validator green (honesty guards active); `:ci-check` green; nothing pushed.

## Success Criteria

- [ ] Per-harness outcome recorded before matrix edits; no credential was created or acquired
- [ ] Full probes match the Claude Code evidence bar; version-only rows carry the verbatim non-claim; runbooks are executable as written
- [ ] All five enterprise skills carry probe runbooks with credential scope and expected evidence
- [ ] Qualified result line's candidate count equals the matrices exactly
- [ ] Permanent vs closable caveats split with named ownership in VALIDATION.txt
- [ ] Full gate green; one conventional commit; remote untouched

## Files Modified/Created

- Modified: capability-matrix/onboarding references in the five harness guides and five enterprise guides (as outcomes dictate), `plans/composable-workflow-phases/VALIDATION.txt`

## Dependencies

None (group 1). Shares only VALIDATION.txt with siblings — own section only.

## Estimated Duration

0.5–1 day (shrinks if the inventory resolves everything to version-only/runbook, which is the likely outcome).
