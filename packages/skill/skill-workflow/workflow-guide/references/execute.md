# Execute

## Goal

Inspect, preview, or apply one bounded action while reporting all observed effects
and returning the operation result inline.

## Procedure

- [ ] Bound the subject, named inputs, completion criteria, targets, and allowed
      effect mode.
- [ ] Resolve exact bindings, implementation, policy constraints, and current
      revisions.
- [ ] For inspect, gather evidence only; for preview, return the exact proposed effect
      set.
- [ ] For apply, require runtime authorization and unchanged approved preview, then
      perform only the bounded effects.
- [ ] Verify observed outcomes in proportion to risk and return inline changes,
      evidence, failures, and remaining work.

Execute may change its bounded subject only in apply mode. It never persists its
`OperationResult`, publishes another domain result, or manages workspace lifecycle.

## Error handling

- Return blocked when the subject, target, policy, or authority is unresolved.
- Preserve verified successful effects after partial failure.
- Mark an uncertain effect unknown and reconcile it before retry.
- Report the exact retry boundary and never replay confirmed effects.
