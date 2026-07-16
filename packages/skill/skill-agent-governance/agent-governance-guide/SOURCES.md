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

## S-HARNESS-CLAUDE — Claude Code hooks

- **URLs:** https://code.claude.com/docs/en/hooks · https://code.claude.com/docs/en/configuration · https://code.claude.com/docs/en/plugins-reference
- **Last reviewed:** 2026-07-16
- **Used for:** `references/harness-adapters.md`, `assets/harness-conformance-fixtures.json`, `assets/harness-module-templates.json`
- **Aspects extracted:** Handler types, event-specific blocking and context behavior, parallel matching handlers, settings scopes, managed hook restrictions, plugin packaging, and experimental agent handlers.

## S-HARNESS-CODEX — Codex hooks

- **URLs:** https://learn.chatgpt.com/docs/hooks · https://learn.chatgpt.com/docs/plugins · https://learn.chatgpt.com/docs/managed-configuration
- **Last reviewed:** 2026-07-16
- **Used for:** `references/harness-adapters.md`, `assets/harness-conformance-fixtures.json`, `assets/harness-module-templates.json`
- **Aspects extracted:** Hook discovery, configuration scopes, trust hashes, concurrent command handlers, managed requirements, plugin hooks, covered tool paths, and parsed-but-skipped prompt, agent, and asynchronous handlers.

## S-HARNESS-KIRO — Kiro hooks

- **URLs:** https://kiro.dev/docs/hooks/ · https://kiro.dev/docs/hooks/actions/ · https://kiro.dev/docs/hooks/types/ · https://kiro.dev/docs/hooks/management/ · https://kiro.dev/docs/cli/v3/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/harness-adapters.md`, `assets/harness-conformance-fixtures.json`, `assets/harness-module-templates.json`
- **Aspects extracted:** Versioned workspace hook files, command and agent actions, triggers, event-specific exit-code blocking, disable behavior, generated-hook review, and CLI v3 alignment.

## S-HARNESS-COPILOT — GitHub Copilot hooks

- **URLs:** https://docs.github.com/en/copilot/reference/hooks-reference · https://docs.github.com/en/copilot/concepts/agents/hooks · https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-plugin-reference · https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills
- **Last reviewed:** 2026-07-16
- **Used for:** `references/harness-adapters.md`, `assets/harness-conformance-fixtures.json`, `assets/harness-module-templates.json`
- **Aspects extracted:** CLI/cloud surface differences, policy/user/project/plugin scopes, command and HTTP handlers, limited prompt handlers, event decisions, cloud sandbox constraints, plugin packaging, and skill trust warnings.

## S-HARNESS-PI — Pi extensions and packages

- **URLs:** https://pi.dev/docs/latest/extensions · https://pi.dev/docs/latest/packages · https://pi.dev/docs/latest/security · https://pi.dev/docs/latest/settings · https://pi.dev/docs/latest/skills
- **Last reviewed:** 2026-07-16
- **Used for:** `references/harness-adapters.md`, `assets/harness-conformance-fixtures.json`, `assets/harness-module-templates.json`
- **Aspects extracted:** Extension events and ordering, project trust, package scopes and pinning, context injection, tool interception, full-process permissions, and absence of a built-in sandbox.

## S-HARNESS-OPENCODE — OpenCode plugins

- **URL:** https://opencode.ai/docs/plugins/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/harness-adapters.md`, `assets/harness-conformance-fixtures.json`, `assets/harness-module-templates.json`
- **Aspects extracted:** Global/project plugin loading, sequential load order, JavaScript/TypeScript events, tool interception, permissions and session events, shell environment mutation, custom tools, and experimental compaction context.

Harness capabilities and limitations require adapter-specific official sources and version-pinned conformance probes before release. The 2026-07-16 fixture snapshot records documentation conformance; its `not-installed` runtime probes are explicit non-results rather than passing runtime evidence.
