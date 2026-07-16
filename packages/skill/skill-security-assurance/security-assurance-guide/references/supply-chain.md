# Software and Executable Supply-Chain Assurance

## Inventory the effective chain

Use deterministic sources to record source repositories/revisions, dependencies, lockfiles, build definitions, runners/builders, environments, identities, inputs, generated code, artifacts, containers, deployment bundles, models/data where selected, plugins, hooks, MCP servers, policies, CI modules, and update channels.

Missing identities, versions, digests, relationships, or provenance are gaps. A generated SBOM or inventory is evidence with a declared scope and generator; it is not proof of completeness or safety.

## Verify provenance and signatures against policy

Before trusting an artifact or executable module:

1. bind the expected subject identity and digest;
2. verify signature or attestation integrity using an approved trust root and identity policy;
3. validate builder, source, revision, build parameters, dependencies, environment, and reproducibility/isolation claims selected by the profile;
4. inspect transparency or provider-native immutable records where required;
5. preserve verifier, policy, trust-root, source and artifact versions, result, time, limitations, and native references;
6. reject moving versions, mismatched subjects/digests, untrusted identities, unverifiable claims, or stale evidence before activation.

## Review executable authority

Compare declared and observed tools, filesystem, network, secret, model, provider, data, side-effect, concurrency, retry, timeout, failure, update, disable, and rollback behavior. Require repository trust or user consent at the appropriate authority zone, plus organization change control for managed modules.

Test unexpected permissions, injection boundaries, duplicate/concurrent invocation, reentrancy, partial application, outage, evidence leakage, rollback, emergency disable, and drift. A trusted repository does not automatically authorize a later permission expansion.

## Update safely

Pin the candidate and rollback versions, preview provenance/capability/permission/data-flow changes, validate compatibility, canary representative contexts, monitor explicit success/abort criteria, authorize expansion, and verify effective state. Separate disablement from rollback and retain an independently reachable emergency-disable path for privileged modules.
