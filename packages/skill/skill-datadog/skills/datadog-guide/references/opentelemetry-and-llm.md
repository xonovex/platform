# OpenTelemetry and LLM Observability

## OpenTelemetry

Inventory SDK/auto-instrumentation, Collector, exporters, endpoints/site, transport/auth, resource attributes, semantic-convention versions, processors, sampling, tail-sampling dependencies, routing, retries/queues, storage, and downstream exports.

Keep service/environment/version and external provider references consistent. Avoid high-cardinality or sensitive attributes by default. Preview which logs, metrics, traces, baggage, events, exemplars, and payload attributes cross each network/data boundary.

## LLM and agent data

Classify prompts, completions, system instructions, tool inputs/outputs, retrieved documents, source snippets, user identifiers, model/provider IDs, token/cost fields, evaluations, and feedback independently. Metadata-only collection is the default.

Content capture requires a documented purpose, data owner, legal/policy basis where applicable, allowed fields, redaction before egress, sampling, retention, residency, access roles, deletion path, incident response, and cost budget. Do not rely only on UI masking after ingestion.

Model/agent traces and evaluations are inference/observation evidence. They do not authorize a tool action or replace authoritative inspection.

## Verification

Inject safe canaries representing secrets, personal identifiers, source, prompt content, and prohibited tags; prove they are absent or redacted at the provider boundary. Verify allowed metadata arrives with correct service/version/source references. Test sampling math, dropped/queued data, exporter outage, rate limit, backpressure, duplicate spans, clock skew, unauthorized access, deletion, and rollback.

Record semantic-convention/configuration versions and state when sampling or retention makes absence inconclusive.
