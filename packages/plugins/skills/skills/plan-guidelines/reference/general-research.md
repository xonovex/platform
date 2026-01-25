# general-research: Research Codebase and Web

**Guideline:** Research codebase and web for requirements, presenting findings without creating plans.

**Rationale:** Separates research from planning, allowing exploration and decision-making before implementation. Enables saving research independently from multiple sources (codebase + web).

**Example:**
```
Research Topic: "Add OAuth2 authentication to API"

Codebase Findings:
- Current: JWT-based auth in packages/auth-service
- Version: @auth/core v5.1
- Similar: GitHub OAuth in 2 places (admin, web)

Web Research:
- Latest: @auth/core v6.0 (released Jan 2026)
- Breaking: Session handling changed
- Docs: oauth.net/2 reference implementation

Recommendations:
- Upgrade to v6.0 (more secure)
- Plan migration of session layer
- Consolidate OAuth logic to shared-auth package
```

**Techniques:**
- Parse research topic and clarify scope with specific requirements
- Explore codebase patterns using parallel search: grep, glob, file reading
- Document current library versions and identify similar implementations
- Research latest versions and releases via web search
- Fetch and analyze documentation from official sources
- Map integration points and existing patterns in the codebase
- Identify relevant guidelines and best practices
- Synthesize codebase findings and web research into recommendations
