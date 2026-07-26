# Execute

Execute carries out work that was already specified. It expects an antecedent — a
plan, an accepted request, a recorded decision, or review feedback — and its job is to
realize that antecedent, not to decide what the work should be.

When no antecedent exists, Execute is the wrong operation. Producing the specification
is Create; deciding between options is Decide; open-ended work with no operation shape
belongs in a freeform session.

## Procedure

1. Resolve the antecedent and the bounded subject it specifies, plus the optional
   subject revision, completion criteria, effect mode, exact targets, applicable
   capabilities, and retry identity.
2. Default to `inspect`; use `preview` or `apply` only when explicitly requested.
3. Adapt the selected method to the effect mode; block before effects when its useful
   procedure cannot run without a broader mode.
4. Perform only work the antecedent specifies and the effect mode permits.
5. Record material implementation choices as decisions with what, why, and a code
   anchor. Stop for a separate Decide operation when a choice changes approved scope,
   architecture, authority, or accepted risk.
6. Verify completion criteria and report every planned, applied, failed, or unknown
   effect.

For an externally submitted apply, require the supplied idempotency key when the
selected provider supports one. If the provider does not support idempotency,
reconcile the exact target before retrying any unknown effect.

Do not publish results or manage workspace resources implicitly. Effect modes are
defined in [effects.md](effects.md); authority rules are in
[governance.md](governance.md).
