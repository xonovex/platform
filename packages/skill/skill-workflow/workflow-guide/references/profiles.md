# Workflow Profile Composition

A workflow profile is an optional topology preset. It may declare:

- identity, version, owner, scope, and applicability;
- included results and preserved result kinds;
- sequence, concurrency, iteration, and presentation;
- required inputs, publication boundaries, exit rules, and completion evidence;
- workflow methods and provider requirements;
- failure behavior for unavailable workflow dependencies.

It does not select an executable, trigger, host, control, evidence sink, or maturity model.
Those dimensions belong to an independent runtime composition when one is used.

## Composition rules

1. Preserve every included result contract and publication boundary.
2. Resolve workflow requirements by semantic identity, not provider configuration.
3. Reject invalid edges, missing prerequisites, and unsatisfied exact revisions.
4. Fail visibly when an explicitly selected provider is unavailable.
5. Do not turn a runtime trace into persistent workflow identity.

The worked `assets/profiles/workflow-only.json` example can be copied and modified. It is
not a required default and has no cross-reference to a control profile.
