# Xonovex Platform Monorepo

## Structure

### Packages

- **`agent`**: Agent CLI (`agent-cli-go`, five per-platform binary packages, `agent-cli-go-github`), `agent-operator-go` + image
- **`asset`**: Diagrams and images; private, pinned `0.0.0`, outside the lockstep release line
- **`command`**: Slash commands (`command-utility`, `command-workflow`); lockstep with the skill plugins and `.claude-plugin/marketplace.json`
- **`config`**: Shared ESLint, Prettier, TypeScript, Vite, Vitest config
- **`moon`**: moonrepo nix plugins (`moon-nix-toolchain`, `moon-nix-extension`, `moon-nix-runtime`), versioned from `Cargo.toml`
- **`script`**: moon task binaries `moon-<domain>-<action>-<subject>`; no binary means shared code (`script-moon-common`)
- **`shared`**: TypeScript (`shared-core`) and Go (`shared-core-go`, `shared-agent-go`) libraries
- **`skill`**: Agent skills catalog (`skill-*`)

### Workflow

- **Setup**: `npm install`
- **Tasks**: `npx moon run <project>:<task>` or `npx moon run #<tag>:<task>`
- **Query**: `moon query projects --tags "<pattern>"`
- **Moon**: a `moon.yml` tag `<name>` inherits `.moon/tasks/tag-<name>.yml`; `extends` refines one (`tag-typescript-script.yml`, `tag-typescript-config.yml` -> `tag-typescript.yml`). A redefined task replaces the parent's `deps`, so restate every inherited dep or set `mergeDeps: replace`
- **Dependencies**: internal `@xonovex/*` exact, a config package pins its plugins exactly, everything else caret. An exact pin elsewhere records its reason: `typescript` and `vite` in `.ncurc.cjs`, `@moonrepo/cli` because it resolves every task
- **Git**: no feature branches or pushes unless asked
- **Commits**: conventional. Rebuild a restructured series on a temporary branch, confirm `git diff <old tip> <new tip>` is empty before moving the branch, and verify each commit with `npx moon run :ci-check --force` so the series stays bisectable
- **Release**: reviewed `version packages` PR only; merging to `main` runs `.github/workflows/release.yml` (`:ci-publish` -> release/tag). Never bypass branch protection
- **Delegation**: subagents only for large independent work such as a wide multi-file investigation; never to check your own work, and few
- **Scope**: deliver what the task asks; ask only when two readings mean materially different work. Flag a mistaken request in one sentence, then finish the whole task

### Code Style

- **Design**: pure functions, immutability, composition, module-level functions, explicit state and context, strict types, clear names, small focused modules, explicit error handling; no global mutable state
- **Imports**: direct from source; no re-exports, deprecated APIs, compatibility wrappers, or shims
- **Deprecation**: remove unused or deprecated code immediately; no `@deprecated` markers
- **Comments**: present behavior, naming the declaration, function, or module; never a plan, agent, doc path, porting history, `INTERIM`, or `TODO`
- **Validation**: typecheck, lint, build, test must pass; fix warnings at source. Keep a project's `dependencySources` file group in sync with its dependencies; `script-moon-release-validate` enforces it. Verify a signature change with `--force`
- **Testing**: specs live in `test/specs/<tier>/`: `unit` runs in `ci-check`, `acceptance` and `integration` opt in by tag

### Writing

- **Prose**: plain sentences; comma, colon, or full stop over an em dash; colon after a label; three periods for an ellipsis; straight quotes. ASD-STE100 in replies and reports: approved words in their approved sense, one idea per sentence, active voice, simple present or past, one term per thing
- **Line breaks**: no hard-wrapped Markdown; one paragraph or list item per line
- **Length**: match the output to the task: brief and focused, outcome first then detail, short caveats, a high-level summary unless asked for depth; no filler, redundant summary, or boilerplate. One progress sentence before the first tool call, then only an important find or a change of direction
- **Names**: complete sentences and spelled-out terms; name a plan, package, or document in full, never a letter, an initial, or a document-number shorthand. A file path stays a file path. No arrow chain, stacked compound, or invented label
- **Corrections**: correct only what changes a conclusion or a decision; state it plainly and continue

### Reporting

- **Completion**: report done or blocked: gates, what was done, what remains, suggestions
- **Gates**: name the command and its outcome; skipped, cached, or not covering the change is not a pass
- **Claims**: mark each verified, naming the evidence, or unverified
- **Interruption**: state the safe retry boundary: what is durable, what is partial, where a resumer restarts

## Integration Points

- config -> shared -> agent
