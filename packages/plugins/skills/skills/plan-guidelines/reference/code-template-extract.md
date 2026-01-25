# code-template-extract: Analyze Package for Template Extraction

**Guideline:** Analyze packages to identify core patterns for template extraction.

**Rationale:** Templates enable rapid scaffolding while maintaining consistency. Extracting from proven implementations ensures new packages follow established patterns and best practices.

**Example:**
```
Analyzing packages/api-auth (successful, reusable):
├── src/
│   ├── routes/index.ts
│   ├── middleware/auth.ts
│   ├── services/TokenService.ts
│   └── types/index.ts
├── tests/
├── Dockerfile
├── docker-compose.yml
└── moon.yml

Template parameters:
- {{PACKAGE_NAME}}: api-auth
- {{PORT}}: 3000
- {{DATABASE_URL}}: postgres://...
```

**Techniques:**
- Analyze successful package structure: directories, naming, organization patterns
- Document configuration templates: package.json, tsconfig.json, build settings
- Extract code patterns: common handlers, middleware, components, hooks
- Identify deployment setup: Dockerfile patterns, CI/CD, Kubernetes manifests
- Catalog testing patterns: utilities, fixtures, test organization
- Compare across packages to validate reusability and frequency
- Define parameterization: variables for names, ports, URLs, package-specific values
- Create template scaffold with substitution markers
