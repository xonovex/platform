# code-template-extract: Analyze Package for Template Extraction

Extract a reusable scaffold template from a proven package. Read-only: produces a template structure with parameterized substitution markers.

## Technique

- Capture structure (directories, naming), config templates (package.json, tsconfig, build), code patterns (handlers, middleware), deployment (Dockerfile, CI, k8s), and testing patterns
- Compare across packages to confirm reusability before extracting
- Parameterize package-specific values as `{{PACKAGE_NAME}}`, `{{PORT}}`, `{{DATABASE_URL}}`, etc.
