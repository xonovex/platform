# plan-research: Research Codebase and Web for Requirements

Research codebase and web for requirements; present findings for review. Does **not** create plans — run `plan-clarify` afterward to settle open decisions one by one, then `plan-create` to author one.

## Goal

- Explore the codebase to understand context (stay in research mode — do not switch into plan-authoring)
- Search the web for latest library versions and documentation
- Present research findings for review
- No plan file created (unless the user asks to save findings to a file)

## Core Workflow

**Delegate codebase exploration to read-only search agents where available; otherwise use grep/find/file-read directly. Stay in research mode.**

1. **Gather requirements** — parse input; ask clarifications if interactive mode was requested
2. **Codebase exploration** — run focused, read-only searches (in parallel where possible):
   - Architecture, patterns, integration points
   - Existing library versions, similar implementations
   - Testing / build patterns, applicable coding guidelines
3. **Web research** — search for latest versions; fetch official docs
4. **Synthesize** — combine codebase + web research
5. **Present or save** — display findings or save to a file; when findings contain open decisions (ambiguities, contradictions, unsettled trade-offs), say so and recommend `plan-clarify`

## Code-quality audits

When the request is an inward code-quality audit rather than forward / web research — hardening (type safety, validation, error handling, logging), simplification (duplicates, dead code, unused dependencies, over-abstraction), or alignment / consistency between two implementations — load the **code-quality-guide** skill and apply its matching dimension, producing the same read-only report grouped by category and graded by severity (no edits, no plan). Then continue with `plan-clarify` / `plan-create` as usual.

## Implementation Details

### Version detection

read project manifests (`package.json`, `pyproject.toml`, `go.mod`, `Cargo.toml`) → web search for latest → fetch official docs

### Interactive mode

ask about preferences, libraries, error handling, testing depth

### Output

current stack, recommended libraries (with versions/rationale), code locations, integration points, skills to consult, considerations

## Output

```
Research: Add rate limiting to API

Current Stack:
- Hono 4.0.2, Redis 4.6.0

Code Locations:
- Middleware pattern: packages/api/src/middleware/
- Similar: packages/auth/src/middleware/throttle.ts

Recommended:
- @upstash/ratelimit@1.0.0 (latest stable)
- https://upstash.com/docs/ratelimit
- Fits existing Redis infrastructure

Skills: typescript-guide, hono-guide
```

## Error Handling

- **Error:** requirements unclear, web research fails
- **Warning:** no existing patterns found, library version conflicts
- **Info:** research saved to file, using cached results

## Gotchas

- A vague requirement turns research into a fishing trip — clarify before exploring
- Library version conflicts surface late if you don't read manifests first — start there
- Producing a plan during research conflates phases — keep research read-only and let the user decide whether to plan
