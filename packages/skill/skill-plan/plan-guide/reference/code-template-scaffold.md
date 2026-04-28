# code-template-scaffold: Analyze Package Scaffolding Options

**Guideline:** Analyze available templates and generate research report on scaffolding options.

**Rationale:** Understanding available templates before scaffolding ensures correct foundation. Reduces setup time and ensures consistency with established patterns.

**Example:**

```
Available templates in .templates/:
1. api-typescript:
   - Features: Express, Zod validation, Jest tests, Docker
   - Use for: RESTful API services
   - Dependencies: 45, Size: 2.1MB

2. react-vite:
   - Features: Vite, React 19, TailwindCSS, Vitest
   - Use for: Frontend applications
   - Dependencies: 38, Size: 1.8MB

3. shared-library:
   - Features: Monorepo-friendly, TypeScript, ESM, minimal deps
   - Use for: Reusable utilities and components
   - Dependencies: 12, Size: 0.8MB

Recommendation: For new API → use api-typescript
```

**Techniques:**

- Discover templates in monorepo (usually in `templates/`, `.templates/`, or `_templates/`)
- List template types: website, API, library, utility, service
- Document features and capabilities for each template
- Identify configuration options and substitution placeholders
- Catalog dependencies, build systems, and deployment setup
- Map best use cases and target scenarios for each
- Compare templates across categories to guide selection
- Provide recommendations matching requirements to template fit
