# transition-run: Plan, Execute, Verify, or Roll Back a Transition

## Core workflow

1. Resolve the requested `plan`, `execute`, `verify`, or `rollback` mode plus exact source
   and target contexts. Keep Transition separate from Integration and Release.
2. For planning, cover data and lineage, users and access, providers and contracts, feature
   flags, training, support, accessibility, monitoring, resilience, continuity, readiness
   thresholds, dependencies, communications, rollback triggers, and irreversible steps.
3. Before execution or rollback, apply [operational-contracts.md](operational-contracts.md)
   to current authorization, target state, external enforcement, short-lived credentials,
   evidence, and policy/profile bindings.
4. Execute through the selected native transition capability in bounded stages. Record
   checkpoints, partial outcomes, target drift, user/data impact, notifications, and each
   provider-native action reference.
5. Verify reconciliations, readiness/outcome thresholds, user and support state, monitoring,
   resilience, access, data integrity, and residual risk. Trigger the authorized rollback or
   fail visibly when a threshold is missed.
6. Publish an independent Transition result for every plan, execution, verification, or
   rollback revision with exact contexts, actor/executor, outcome, evidence, and follow-up.
