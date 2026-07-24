# Capability Selection

The workflow guide has no hard dependency on a concrete domain skill. Commands may
hard-depend on this guide; the guide selects interchangeable implementation
capabilities softly at runtime.

## Selection Procedure

1. Derive capability needs from the operation, subject, explicit method, perspectives,
   criteria, repository context, and requested effects.
2. Inventory installed skills and their catalog provisions before choosing.
3. Prefer an exact compatible provision when one is requested; otherwise select the
   narrowest installed capability whose routing description fits the subject.
4. Resolve each selected skill's declared hard dependencies before the skill.
5. Resolve preferred requirements when compatible providers are installed.
6. Filter overlays to the active context and apply precedence from broad to narrow:
   global, organization, repository, language, framework, path, explicit.
7. Load each resulting guide once in dependency-first order.

Do not select by catalog order, duplicate a selected guide, or invent an unavailable
capability.

## Required and Preferred Needs

- **Required unavailable** — block before effects and name the missing capability,
  reason, and recovery action.
- **Preferred unavailable** — continue only when a safe baseline remains useful;
  report degraded coverage and its consequence.
- **No specialist needed** — proceed with baseline reasoning without manufacturing a
  dependency.

If two equally valid providers remain after explicit bindings and applicable overlays,
report the ambiguity instead of picking by prompt or filesystem order.

## Composition Boundary

Capability selection changes how the operation is performed, not:

- the requested operation;
- binding criteria or their authority;
- the requested effect mode;
- target identity or scope;
- whether another operation should follow.

An overlay may specialize guidance but cannot replace an unrelated skill, bypass a hard
dependency, or grant authority.

## Result Notes

Report selected capabilities only when they materially affect the result. Include:

- selected guide and why it matched;
- relevant hard-dependency or overlay ordering;
- unavailable preferred or required needs;
- any ambiguity that blocked deterministic selection.
