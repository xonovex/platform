# Evaluate Governance Drift

## Compare

- intended versus observed source, profile, policy, module, adapter, and platform versions;
- declared versus observed event support, handler execution, blocking, ordering, and managed configuration;
- expected versus effective permissions, data flows, secrets, network, side effects, and failure behavior;
- required versus fresh decision, enforcement, configuration, and evidence references;
- active exceptions and break-glass records versus expiry and review state.

## Workflow

1. Capture native observed state read-only and timestamp the probe.
2. Resolve intended state and its authority source.
3. Classify changes as strengthening, neutral, weakening, unsupported, stale, or unknown.
4. Re-run applicable conformance probes.
5. Fail visibly when drift invalidates a mandatory guarantee; otherwise advise with impact and remediation.
6. Publish a provider-native drift result and link the exact compared versions.

Never auto-promote a discovered difference into policy. Remediation follows preview, authorization, apply, verify, and rollback semantics.
