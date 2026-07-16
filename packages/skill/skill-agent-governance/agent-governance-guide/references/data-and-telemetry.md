# Data and Telemetry Governance

## Decide before collection or routing

For prompt, context, memory, retrieval, model, agent, tool, MCP, CI/CD, provider, policy, configuration, privileged-operation, and evidence data, record:

- classification, owner, subject, purpose, authority or consent where required, and prohibited uses;
- minimum fields/content, source/destination, model/provider route, region/residency, external transfer, and onward use;
- redaction/tokenization/hashing, access roles, sampling, retention, deletion, legal hold, and cost;
- schema/semantic-convention version, stability, correlation, evidence origin, limitations, and failure behavior.

Apply rules before content crosses a model, tool, telemetry, storage, or provider boundary. Reject a route that cannot satisfy classification, residency, transfer, retention, access, or provider-use requirements. Never send the content first and rely on downstream deletion as the primary control.

## Keep sensitive content off by default

Default to opaque references, component/version identifiers, operation kinds, status, durations, token/resource counts, policy/control/module versions, error classes, redaction results, and non-sensitive hashes. Do not collect raw prompts, model outputs, tool arguments/results, source, secrets, personal data, customer payloads, or credentials unless an explicit purpose and authorization require it.

Hashing is not anonymity when the input domain is small or linkable. Record algorithm, salt/key handling, rotation, collision/linkability risk, and whether the hash is used for integrity, deduplication, or correlation.

Test secret/personal-data canaries at instrumentation, collector, exporter, provider, retention, query, and support boundaries. A redacted application log does not prove traces, events, baggage, exception messages, provider payloads, or retries are safe.

## Use interoperable semantics cautiously

Pin the selected OpenTelemetry and GenAI/agent semantic-convention version and stability level. Map model, agent, MCP, tool, policy, CI/CD, provider, configuration, and privileged operations where a stable or explicitly accepted experimental convention fits; use a versioned organization namespace for gaps.

Preserve trace/span correlation without turning a trace, session, or conversation identifier into workflow identity. Keep policy decision, enforcement, evidence, source revision, deployment, and provider-native references distinct.

The GenAI conventions moved from the core semantic-conventions tree to their own repository after the recorded 1.43.0 baseline. Treat location, schema, names, stability, and version as drift-sensitive inputs rather than copying current attributes into a permanent contract.

## Handle outages and access

Telemetry and advisory sinks shed load, buffer within bounded encrypted limits, or fail visibly without blocking mandatory control execution by default. Never fall back to raw sensitive logging when redaction, exporter, or provider paths fail.

Enforce role-based access at the native provider, audit queries and configuration changes, bound support access, and test unauthorized reads. Preserve sampling/retention limitations so absent telemetry is not misrepresented as evidence that no event occurred.
