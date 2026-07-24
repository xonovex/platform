# Execute

1. Resolve the bounded subject, optional subject revision, completion criteria,
   requested method, effect mode, exact targets, applicable capabilities, and retry
   identity.
2. Default to `inspect`; use `preview` or `apply` only when explicitly requested.
3. Adapt the selected method to the effect mode; block before effects when its useful
   procedure cannot run without a broader mode.
4. Perform only work permitted by the selected effect mode.
5. Verify completion criteria and report every planned, applied, failed, or unknown
   effect.

For an externally submitted apply, require the supplied idempotency key when the
selected provider supports one. If the provider does not support idempotency, reconcile
the exact target before retrying any unknown effect.

Do not publish results or manage workspace resources implicitly.
