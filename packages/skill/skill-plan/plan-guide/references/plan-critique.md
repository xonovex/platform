# plan-critique: Adversarially Stress-Test a Plan

Attack an existing plan to expose its weaknesses before any code is written — flawed assumptions, failure modes, missing cases, unmanaged risks. Read-only: produces a findings report for `plan-revise` to act on; never edits the plan or the codebase. Run it as an independent agent — a fresh session that did not author the plan — so it attacks the approach rather than defending it.

## Critique vs Validate

- **plan-critique** attacks a plan that has _not_ been built — it predicts how the approach will fail.
- **plan-validate** checks a plan that _has_ been implemented — it verifies success criteria and structural match.
- Critique is forward-looking and adversarial; validate is backward-looking and confirmatory.

## Modes

Run one or more adversarial lenses (default: red-team + pre-mortem):

- **red-team** — attack the essential assumptions: what must be true for this to work, and which of those is unproven?
- **pre-mortem** — assume it shipped and failed; enumerate the most likely causes, working backward to the decision behind each
- **falsify** — hunt the single disconfirming case: the input, scale, or sequence the plan does not handle
- **steelman** — strengthen the weakest section: state the strongest version of the concern, then the minimal change that removes it

## Core Workflow

**Stay in critique mode — no plan authoring, no implementation, read-only.**

1. **Load plan** — read the document; extract approach, assumptions, success criteria, subplan structure
2. **Ground in the codebase** — verify each essential claim against the actual code; a critique built on a misread of the plan is noise
3. **Run each selected mode** — produce concrete findings, not vague worries
4. **Rate each finding** — severity (blocking / major / minor) and confidence; a finding without a severity can't be triaged
5. **Tie findings to the plan** — name the section / decision / subplan each one attacks
6. **Report** — grouped by mode, severity-ordered, each with the weakness, why it bites, and a suggested direction; STOP (feeds `plan-revise`)

## Finding Format

```
[<mode>] <severity> — <the weakness>
  Where:     <plan section / decision>
  Why:       <the concrete failure it causes>
  Direction: <the change that would remove it>
```

## Example Output

```
Critique: 6 findings (2 blocking, 3 major, 1 minor)

[red-team]   blocking — SSE assumes <1k concurrent; plan targets 10k
  Where: Transport decision · Why: connection pool exhausts · Direction: per-user queue + backpressure
[pre-mortem] major    — no migration path for existing rows
  Where: Schema change · Why: deploy breaks reads mid-rollout · Direction: expand/contract migration
...
Next: plan-revise (resolve blocking + major), then re-run plan-critique
```

## Gotchas

- Vague worries ("this might not scale") are noise — every finding names a concrete input, scale, or sequence that breaks it
- Critiquing a misread of the plan wastes the round — verify each claim against the code before attacking it
- Findings without severity can't be triaged — always rate and order
- A critique that edits the plan oversteps — it reports; `plan-revise` applies the fixes
- Running every mode on a tiny plan is busywork — default to red-team + pre-mortem, escalate to falsify / steelman when the stakes warrant
- The plan's author critiquing their own plan defends it — run plan-critique in a fresh session (or a separate agent) that did not write the plan
