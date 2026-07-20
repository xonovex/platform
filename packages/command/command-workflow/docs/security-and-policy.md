# Controls and Trust

Executable plugins run with the authority of their host. Review and pin executable paths,
images, dependencies, permissions, filesystem and network access, secrets, model/provider
use, data flow, timeouts, output bounds, update path, and removal path.

## Control effect is selected, not inferred

A control plugin returns `allow`, `deny`, or `abstain`. Each invocation selects its mode:

- `observe` records the result and never blocks;
- `enforce` blocks on `deny`.

Before-phase controls can prevent execution. After-phase controls can fail the result but
cannot undo side effects. Put irreversible gates at a native before-action point when the
owner chooses that guarantee.

Controls should each own one concern. Approval, budget, protected-target checks,
provenance, independent critique, and escalation remain independent plugins. None is
implied by a trigger, executor, host, evidence sink, or maturity label.

## Evidence failures

Every selected sink states `ignore` or `fail`. There is no mandatory audit sink. Avoid
secret or raw-content capture unless the selected sink explicitly requires and protects
it.

Use the explanation operation to show the exact selected enforcement points and missing
required capabilities before running a composition.
