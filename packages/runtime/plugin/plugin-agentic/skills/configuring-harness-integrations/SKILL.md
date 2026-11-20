---
name: configuring-harness-integrations
description: "Use when comparing agent harnesses or applying the contract every harness shares: a versioned capability matrix of semantic intents, a guarantee boundary, the preview-then-apply lifecycle with explicit authorization, rollback and drift, or the executor class of a hook (deterministic command, model evaluator, specialist agent). Triggers on which harnesses can block an action, matrix identity fields, mapping an intent to native events across harnesses, onboarding or rolling back one integration on several harnesses, even when the user doesn't say 'harness'. Not for setting up or debugging one named product such as Claude Code, Codex, Copilot, Kiro, OpenCode, or Pi: its own configuring skill owns its events, settings, and syntax."
---

# Harness Integrations

Treat each harness as an adapter behind one portable contract, so every integration is mapped, previewed, applied, verified and removed the same way.

## Essentials

- **Pin the surface** - Record the documentation snapshot and the observed runtime version before selecting a guarantee; absence is `not-installed`, never a guess.
- **Map intents, not event names** - Translate each requested behavior through the harness's versioned capability matrix, see [references/capability-matrix.md](references/capability-matrix.md).
- **Claim only the covered guarantee** - A hook is not a sandbox; enforce independently what its native event cannot cover.
- **Onboard transactionally** - Discover, preview the exact change, apply after explicit authorization, verify by probe, see [references/lifecycle.md](references/lifecycle.md).
- **Keep executor classes apart** - Deterministic command, model evaluator and specialist agent stay separate and declared, see [references/patterns.md](references/patterns.md).
- **Own only what you added** - Disable, roll back and update touch the owned entry and nothing else.

## Gotchas

- Event-name similarity across harnesses is not equivalence: blocking, ordering and coverage differ per event.
- Documentation-verified is not runtime-verified; a version probe alone proves neither the event nor the blocking behavior.
- A handler that starts a model or child agent is not deterministic because its outer handler is a command.
- A managed configuration restricts sources but does not prove the script it names was securely distributed.
- Generated or proposed configuration is untrusted until its command, matcher and timeout are reviewed.

## Progressive Disclosure

- Read [references/lifecycle.md](references/lifecycle.md) - Load when discovering, previewing, applying, verifying, disabling, rolling back, updating, or detecting drift of an integration
- Read [references/capability-matrix.md](references/capability-matrix.md) - Load when building or reading a harness capability matrix, its identity fields, its semantic intents, or its guarantee boundary
- Read [references/patterns.md](references/patterns.md) - Load when choosing a portable pattern, an executor class, or an organization-managed adoption recipe
