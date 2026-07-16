# plan-research: Research Codebase and Web for Requirements

Research codebase and web for a planning-oriented report. This specialized read-only analysis does not create Planning or canonical Research results: use `research-run` when the result must be reusable/provider-native, `decision-create` for authority-bound choices, then `plan-create`.

## Core Workflow

**Delegate codebase exploration to read-only search agents where available; otherwise grep/find/read directly. Stay in research mode — do not switch into plan-authoring.**

1. **Gather requirements** — parse input; ask clarifications only if interactive mode was requested
2. **Codebase exploration** (parallel where possible) — architecture, patterns, integration points; existing library versions and similar implementations; testing/build patterns and applicable coding guidelines
3. **Web research** — search for latest versions; fetch official docs
4. **Synthesize** codebase + web
5. **Present or save** — when findings contain open decisions, recommend `decision-create`; when the report needs a durable cross-context handoff, publish it through `research-run`

## Code-quality audits

When the request is an inward code-quality audit rather than forward/web research — hardening (type safety, validation, error handling, logging), simplification (duplicates, dead code, unused deps, over-abstraction), or alignment/consistency between two implementations — load the **code-quality-guide** skill and apply its matching dimension, producing the same read-only report grouped by category and graded by severity (no edits, no Planning mutation). Then use `decision-create` / `plan-create` as needed.

## Details

- **Version detection** — read manifests (`package.json`, `pyproject.toml`, `go.mod`, `Cargo.toml`) → web search for latest → fetch docs
- **Output** — current stack, recommended libraries (versions + rationale), code locations, integration points, skills to consult, considerations

```
Research: Add rate limiting to API
Current Stack: Hono 4.0.2, Redis 4.6.0
Code Locations:
- Middleware pattern: packages/api/src/middleware/
- Similar: packages/auth/src/middleware/throttle.ts
Recommended: @upstash/ratelimit@1.0.0 (latest stable) — fits existing Redis
  https://upstash.com/docs/ratelimit
Skills: typescript-guide, hono-guide
```

## Gotchas

- A vague requirement turns research into a fishing trip — clarify before exploring
- Library version conflicts surface late if you don't read manifests first — start there
- Producing a plan during research conflates phases — keep research read-only and let the user decide whether to plan
