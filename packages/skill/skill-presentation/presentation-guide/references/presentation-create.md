# presentation-create: Create Presentation Document from Codebase

Generate a markdown presentation by exploring a codebase and structuring findings into slides separated by `---`.

## Example

```markdown
# Microservices Architecture

---

## Overview

- 12 services, 15 shared packages
- Kubernetes orchestration, GitOps deployment
- Found in: services/, clusters/

---

## Core Services

- API Gateway (packages/services/api-gateway)
- Auth Service (packages/services/auth)
- User Service (packages/services/users)

---

## Next Steps

- Scale to 20 services
- Implement service mesh
```

## Techniques

- Parse topic and identify relevant codebase areas and packages
- Clarify scope with questions: packages, depth, audience, style
- Scan codebase using parallel exploration: architecture, schemas, integration points
- Synthesize findings into logical narrative with file references
- Generate diagrams: ASCII art for markdown or GraphViz for motion
- Structure content into slides: title, summary, challenges, solutions, implementation
- Extract and document style preferences: colors, fonts, logo, themes
- Create markdown with `---` separators for slide boundaries
