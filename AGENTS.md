# Xonovex Platform

Monorepo for Xonovex tools and configuration packages, all code under `packages/`.

## Structure

### Packages

- **`config`**: shared configs (ESLint, TypeScript, Vitest, Prettier, Vite)
- **`agent`**: CLI tools (agent-cli-go) and K8s operator (agent-operator-go)
- **`shared`**: shared libraries (shared-core, shared-core-go, shared-agent-go)
- **`skill`**: coding guidelines and skills
- **`diagram`**: diagrams (action graph, workflow)
- **`asset`**: static assets

### Workflow

- **Setup**: `npm install`
- **Tasks**: `npx moon run <project>:<task>` or `npx moon run #<tag>:<task>`
- **Moon**: `.moon/tasks/*.yml` templates auto-inherit by type/language/tags
- **Query**: `moon query projects --tags "<pattern>"`
- **Git**: Do not create feature branches unless explicitly asked. Never push unless explicitly asked.

### Code Style

- **Paradigm**: functional where practical (pure functions, immutability, no side effects); module-level functions over classes; compose simple functions; pass state explicitly, no global/shared mutable state
- **Imports**: direct from source; no re-exports or backwards-compat wrappers
- **Design**: modular functions, explicit context, small focused files
- **Quality**: strict types, clear naming, explicit error handling
- **Validation**: typecheck, lint, build, test must pass; fix warnings at root cause
- **Deprecation**: remove unused/deprecated code immediately; no @deprecated markers or backwards-compat shims
- **Comments**: state present behavior only; name the declaration/function/module, never a plan, agent, doc, or other file by path; no porting-provenance or `INTERIM`/`TODO` markers
- **Commits**: conventional commits

## Integration Points

- config -> shared -> agent
- command plugins depend on guideline skills via `plugin.json` `dependencies` (auto-installed on Claude Code): `command-utility` -> `skill-{content,insights,instruction,skill,command}`; `command-workflow` -> `skill-{plan,git,pull-request,code-review}`; the `pr-*` commands additionally load an optional host-delivery skill detected from the remote (`skill-github`, `skill-gitlab`, or another `skill-<host>`), installed separately rather than as a hard dependency
