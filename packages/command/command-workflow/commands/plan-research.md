---
description: Research codebase and web for requirements without creating a plan
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - Task
  - TaskCreate
  - TaskUpdate
  - WebSearch
  - WebFetch
  - AskUserQuestion
argument-hint: "<requirements> [--interactive] [--save-to <file>]"
---

# /xonovex-workflow:plan-research – Research Codebase and Web

Research codebase and web for requirements, presenting findings for review. Does NOT create plans - run `/xonovex-workflow:plan-create` afterward to create an implementation plan.

## Goal

- Explore the codebase to understand context (stay in research mode — do not switch into plan-authoring)
- Use WebSearch/WebFetch for latest library versions and documentation
- Present research findings for review
- NO plan file created (unless `--save-to` specified)

## Arguments

- `requirements` (required): Description of what to research
- `--interactive` (optional): Ask clarifying questions
- `--save-to <file>` (optional): Save research to file

## Core Workflow

**Delegate codebase exploration to read-only search agents where available; otherwise use grep/find/file-read directly. Stay in research mode.**

1. **Gather requirements**: Parse input; ask clarifications if `--interactive`

2. **Codebase exploration**: Run focused, read-only searches (in parallel where possible):
   - Architecture, patterns, integration points
   - Existing library versions, similar implementations
   - Testing/build patterns, applicable coding guidelines

3. **Web research**: WebSearch for latest versions, WebFetch for docs

4. **Synthesize findings**: Combine codebase + web research

5. **Present or save**: Display findings OR save to file

## Implementation Details

**Version detection**: package.json -> WebSearch latest -> WebFetch docs

**Interactive mode**: Ask about preferences, libraries, error handling, testing depth

**Research output**: Current stack, recommended libraries (with versions/rationale), code locations, integration points, skills to consult, considerations

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

## Examples

```bash
/xonovex-workflow:plan-research "Add rate limiting to API"
/xonovex-workflow:plan-research "Implement WebSocket support" --interactive
/xonovex-workflow:plan-research "Add analytics tracking" --save-to research/analytics.md
```

## Error Handling

- **Error**: Requirements unclear, web research fails
- **Warning**: No existing patterns found, library version conflicts
- **Info**: Research saved to file, using cached results

## Gotchas

- A vague requirement turns research into a fishing trip — clarify before exploring
- Library version conflicts surface late if you don't read manifests first — start there
- Producing a plan during research conflates phases — keep research read-only and let the user decide whether to plan
