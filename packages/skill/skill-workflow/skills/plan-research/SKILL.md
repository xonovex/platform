---
description: "Research the codebase and web for requirements without producing a plan document. Use when the user asks to research, investigate, scope, or gather context on a topic. Keywords: research, investigate, scope, codebase research, web research, requirements gathering."
---

# /xonovex-workflow:plan-research – Research Codebase and Web

Research codebase and web for requirements, presenting findings for review. Does NOT create plans - run `/xonovex-workflow:plan-create` afterward to create an implementation plan.

## Goal

- Use Task/Explore agents to understand codebase (do NOT use EnterPlanMode)
- Use WebSearch/WebFetch for latest library versions and documentation
- Present research findings for review
- NO plan file created (unless the user asks to save to a file)

## Core Workflow

**Use Task agents with subagent_type=Explore and model=haiku for codebase analysis. Do NOT use EnterPlanMode.**

1. **Gather requirements**: Parse input; ask clarifications if interactive mode was requested

2. **Codebase exploration**: Use Task with subagent_type=Explore and model=haiku (up to 3 parallel):
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

Integration Points:
- Add to Hono middleware chain
- Use existing Redis connection

Skills: typescript-guidelines, hono-guidelines

Considerations: Rate limit keys per user/IP, env config, metrics
```

## Error Handling

- **Error**: Requirements unclear, web research fails
- **Warning**: No existing patterns found, library version conflicts
- **Info**: Research saved to file, using cached results
