# Sources

The governance contract is an Xonovex architectural synthesis. Sources support concerns and constraints, not automatic equivalence, certification, legal compliance, or one mandatory implementation.

## S-NIST-80053 — Security and privacy controls

- **URL:** https://csrc.nist.gov/pubs/sp/800/53/r5/upd1/final
- **Last reviewed:** 2026-07-14
- **Used for:** `references/architecture.md`, `references/execution.md`, `references/policy-and-authority.md`, `references/modules.md`, `references/profiles.md`, `references/onboarding.md`, `references/inspect.md`, `references/conformance.md`, `references/drift.md`, `references/module-management.md`
- **Aspects extracted:** Access, audit, configuration, incident, contingency, and supply-chain control families. Mappings remain contextual.

## S-NIST-800207 — Zero Trust Architecture

- **URL:** https://csrc.nist.gov/pubs/sp/800/207/final
- **Last reviewed:** 2026-07-14
- **Used for:** `references/architecture.md`, `references/execution.md`, `references/policy-and-authority.md`
- **Aspects extracted:** Least authority, resource-centric authorization, and continuous evaluation.

## S-NIST-80061 — Incident response

- **URL:** https://csrc.nist.gov/pubs/sp/800/61/r3/final
- **Last reviewed:** 2026-07-14
- **Used for:** `references/policy-and-authority.md`, `references/conformance.md`
- **Aspects extracted:** Incident preparation, response, recovery, exception evidence, and post-event review.

## S-NIST-PRIVACY — Privacy Framework

- **URL:** https://www.nist.gov/privacy-framework
- **Last reviewed:** 2026-07-14
- **Used for:** `references/policy-and-authority.md`, `references/profiles.md`, `references/onboarding.md`, `references/inspect.md`
- **Aspects extracted:** Data minimization and privacy-risk management; version and applicability remain explicit.

## S-AGENTSKILLS — Agent Skills specification

- **URL:** https://agentskills.io/specification
- **Last reviewed:** 2026-07-14
- **Used for:** `SKILL.md`, `references/modules.md`, `references/conformance.md`
- **Aspects extracted:** Skills contain instructions, scripts, references, and assets; installation is not enforcement proof.

## S-MCP — Model Context Protocol

- **URL:** https://modelcontextprotocol.io/specification/2025-06-18
- **Last reviewed:** 2026-07-14
- **Used for:** `references/policy-and-authority.md`, `references/modules.md`, `references/conformance.md`
- **Aspects extracted:** Capability negotiation, consent, and tool-safety principles; MCP is not a complete policy or sandbox boundary.

## S-OPA — Open Policy Agent

- **URL:** https://www.openpolicyagent.org/docs
- **Last reviewed:** 2026-07-14
- **Used for:** `references/policy-and-authority.md`, `references/profiles.md`
- **Aspects extracted:** Policy decision and enforcement separation while keeping the policy engine optional.

## S-SUPPLY-CHAIN — Executable-module provenance options

- **URLs:** https://slsa.dev/spec/v1.2/ · https://in-toto.io/ · https://theupdateframework.io/ · https://docs.sigstore.dev/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/modules.md`, `references/module-trust.md`, `references/module-management.md`, `references/conformance.md`
- **Aspects extracted:** Pinned subjects, verifiable provenance/signatures/digests, secure update metadata, and rollback-aware distribution. Each mechanism remains optional and requires a declared trust policy.

## S-INVENTORY — Optional AI and governance inventory

- **URLs:** https://spdx.dev/learn/areas-of-interest/ai/ · https://cyclonedx.org/capabilities/mlbom/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/inspect.md`, `references/conformance.md`
- **Aspects extracted:** Inventories may include models, data, tools, agents, and software relationships alongside governance modules without requiring one BOM representation.

## S-OTEL-GENAI — OpenTelemetry GenAI semantic conventions

- **URL:** https://opentelemetry.io/docs/specs/semconv/gen-ai/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/events-and-capabilities.md`, `references/profiles.md`, `references/inspect.md`, `references/drift.md`
- **Aspects extracted:** Interoperable agent/model/tool telemetry with stability, cost, and sensitive-data caveats.

## S-HEXAGONAL — Hexagonal architecture

- **URL:** https://alistair.cockburn.us/hexagonal-architecture
- **Last reviewed:** 2026-07-14
- **Used for:** `references/architecture.md`, `references/modules.md`, `references/onboarding.md`, `references/module-management.md`
- **Aspects extracted:** Core-owned ports, native adapters, inward dependencies, and composition-root wiring.

Harness capabilities and limitations require adapter-specific official sources and version-pinned conformance probes before release.
