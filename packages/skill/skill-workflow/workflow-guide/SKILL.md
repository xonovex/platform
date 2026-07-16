---
name: workflow-guide
description: "Use when defining, inspecting, composing, or validating lifecycle workflow capabilities, provider-native result contracts, ephemeral result handles, workflow profiles, completion evidence, or cross-provider handoffs. Triggers on workflow result, phase result, PhaseResultHandle, lifecycle profile, provider-native evidence, workflow conformance, or a request to keep workflow semantics independent from storage, methods, agents, and governance runtimes, even when the user doesn't say 'workflow architecture'."
---

# Composable Workflow Contracts

Keep lifecycle meaning stable while methods, executors, providers, topology, and governance vary independently.

## Core Principles

- **Semantic results, native persistence** — each capability owns required meaning; provider skills own representation, authentication, native identifiers, revisions, and side effects.
- **Opaque handoffs** — exchange provider context plus opaque native references; never invent a central workflow identity, database, or universal persisted envelope.
- **Ephemeral reconstruction** — use `PhaseResultHandle` only in runtime handoffs and reconstruct it from provider-native state after context loss.
- **Profiles compose** — profiles include, omit, sequence, parallelize, loop, and present capabilities without erasing their independent results or publication boundaries.
- **Axes resolve independently** — method, artifact, work-item, code-host, workspace, policy, and learning selections never collapse into one provider-method choice.
- **Independent adoption** — workflow contracts do not require governance hooks; governance may consume workflow facts without becoming a lifecycle prerequisite.

## Operations

- **Inspect** — resolve native references and report effective capabilities, profile topology, result freshness, and gaps — see [references/inspect.md](references/inspect.md)
- **Conformance** — validate result semantics, ephemeral handles, provider handoffs, profile topology, publication boundaries, evidence, and cross-plane requirements — see [references/conformance.md](references/conformance.md)

## Gotchas

- A runtime trace or session ID correlates execution; it is not a workflow identity.
- A local path is one provider-native reference, not the fallback for an unavailable explicitly selected hosted provider.
- Composite labels such as Discover or Review are presentation; constituent capability results remain independently publishable.
- Installing this skill explains contracts but does not enforce policy or prove a control is active.

## Progressive Disclosure

- Read [references/architecture.md](references/architecture.md) - Load when assigning ownership, checking adoption modes, or deciding allowed dependency directions between planes
- Read [references/results.md](references/results.md) - Load when authoring or interpreting a lifecycle result or `PhaseResultHandle`
- Read [references/providers.md](references/providers.md) - Load when defining or validating result-provider resolve, read, publish, revise, relate, version, capability, or restart behavior
- Read [references/profiles.md](references/profiles.md) - Load when composing topology, variation axes, exit rules, cumulative completion, or cross-plane requirements
- Read [references/inspect.md](references/inspect.md) - Load when inspecting an effective workflow from native references and environment facts
- Read [references/conformance.md](references/conformance.md) - Load when validating a result, handle, provider handoff, or workflow profile
