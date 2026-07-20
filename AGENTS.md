# Xonovex Platform

- Monorepo for Xonovex tools and configuration packages; code lives under `packages/`.
- Setup: `npm install`.
- Run tasks with `npx moon run <project>:<task>` or `npx moon run #<tag>:<task>`; query projects with `moon query projects --tags "<pattern>"`.
- `.moon/tasks/*.yml` templates inherit by project type, language, and tags.
- Do not create feature branches or push unless explicitly asked.
- Release only through a reviewed `version packages` PR; merging to `main` runs `.github/workflows/release.yml` (`:ci-publish` -> release/tag). Never bypass branch protection.
- Prefer pure functions, immutability, composition, module-level functions, and explicit state; avoid global mutable state.
- Import directly from source; do not add re-exports, deprecated APIs, compatibility wrappers, or shims.
- Keep modules small and focused, with strict types, clear names, explicit context, and explicit error handling.
- Typecheck, lint, build, and test must pass; fix warnings at their source.
- Remove unused or deprecated code immediately; do not add `@deprecated` markers.
- Comments describe present behavior and name the declaration, function, or module; never reference a plan, agent, doc path, porting history, `INTERIM`, or `TODO`.
- Use conventional commits.
- Dependency direction: `config -> shared -> agent`.
