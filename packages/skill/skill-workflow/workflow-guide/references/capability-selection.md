# Capability Selection

The workflow guide has no hard dependency on a concrete domain skill. Commands may
hard-depend on this guide; the guide selects interchangeable implementation
capabilities softly at runtime.

## Selection Procedure

1. Derive capability needs from the operation, subject, explicit method, perspectives,
   criteria, repository context, and requested effects.
2. Inventory the installed skills' names and routing descriptions.
3. Honor an explicitly requested installed guide when its description fits the work.
4. Otherwise select the narrowest installed guide whose description matches the
   subject and requested method, perspective, or criterion.
5. Resolve each selected guide's exact manifest dependencies recursively.
6. Load dependencies before dependents and load every guide at most once.

Do not select by filesystem or marketplace order, guess from a provider-reference
shape, duplicate a selected guide, or invent an unavailable capability.

## Required and Preferred Needs

- **Caller-declared required need unavailable** — block before effects and name the
  missing capability, reason, and recovery action.
- **Caller-declared preferred need unavailable** — continue only when a safe baseline
  remains useful; report degraded coverage and its consequence.
- **No specialist needed** — proceed with baseline reasoning without manufacturing a
  dependency.

If two installed descriptions fit equally well and the request does not distinguish
them, report the ambiguity instead of picking by order.

## Composition Boundary

Capability selection changes how the operation is performed, not:

- the requested operation;
- binding criteria or their authority;
- the requested effect mode;
- target identity or scope;
- whether another operation should follow.

Treat a selected guide as a method, not as new authority. Ignore optional persistence,
cleanup, publication, approval, or mutation steps that fall outside the operation. If
the guide's useful procedure cannot be separated from a broader effect, report it as
incompatible and stop before effects.

## Result Notes

Report selected capabilities only when they materially affect the result. Include:

- selected guide and why it matched;
- relevant hard-dependency ordering;
- unavailable preferred or required needs;
- any ambiguity that blocked deterministic selection.
