# Xonovex Platform Monorepo

## Structure

### Packages

- **`agent`**: Agent wrapper CLI (`agent-cli-go`) with its five per-platform binary packages and GitHub release orchestration (`agent-cli-go-github`), plus `agent-operator-go` and its image
- **`asset`**: Shared diagrams and images; private, pinned at version `0.0.0`, and outside the lockstep release line
- **`command`**: Slash commands (`command-utility`, `command-workflow`); one lockstep version with the skill plugins and `.claude-plugin/marketplace.json`
- **`config`**: Shared ESLint, Prettier, TypeScript, Vite, and Vitest configuration
- **`moon`**: moonrepo nix plugin family (`moon-nix-toolchain`, `moon-nix-extension`, `moon-nix-runtime`), versioned from `Cargo.toml`
- **`script`**: moon task binaries named `moon-<domain>-<action>-<subject>`; a package that ships no binary is shared code (`script-moon-common`)
- **`shared`**: Shared libraries for TypeScript (`shared-core`) and Go (`shared-core-go`, `shared-agent-go`)
- **`skill`**: Agent skills catalog (`skill-*`)

### Workflow

- **Setup**: `npm install`
- **Tasks**: `npx moon run <project>:<task>` or `npx moon run #<tag>:<task>`
- **Query**: `moon query projects --tags "<pattern>"`
- **Moon**: task templates are tag-based: a project inherits `.moon/tasks/tag-<name>.yml` by listing `<name>` in its `moon.yml` `tags`, and a template refines another with `extends`, as `tag-typescript-script.yml` and `tag-typescript-config.yml` extend `tag-typescript.yml`. A task redefined across `extends` replaces the parent's `deps` rather than appending to them, so restate every inherited dep or declare `mergeDeps: replace`; `script-moon-release-validate` fails the build otherwise.
- **Dependencies**: internal `@xonovex/*` dependencies are exact, and a config package pins the plugins it configures exactly so consumers get identical behaviour. Every other external dependency takes a caret range; an exact pin elsewhere records its reason: `typescript` and `vite` in `.ncurc.cjs`, `@moonrepo/cli` because it resolves every task.
- **Git**: Do not create feature branches or push unless explicitly asked.
- **Commits**: use conventional commits. When a series is restructured, rebuild it on a temporary branch and confirm `git diff <old tip> <new tip>` is empty before moving the branch, which separates a history change from a content change; verify each commit with `npx moon run :ci-check --force`, since a series whose intermediate commits do not build cannot be bisected.
- **Delegation**: delegate to a subagent only for large, genuinely independent work such as a wide multi-file investigation; do not delegate what a handful of tool calls finishes, do not spawn one to check your own work, and keep the count low
- **Scope**: deliver what the task asks, at the scope it intends. Make a routine judgment call yourself. Ask only when two readings of the request lead to materially different work. If the request looks mistaken, say so in one sentence and continue as asked. Finish the whole task
- **Release**: only through a reviewed `version packages` PR; merging to `main` runs `.github/workflows/release.yml` (`:ci-publish` -> release/tag). Never bypass branch protection.

### Code Style

- **Paradigm**: prefer pure functions, immutability, composition, module-level functions, and explicit state; avoid global mutable state
- **Imports**: import directly from source; no re-exports, deprecated APIs, compatibility wrappers, or shims
- **Design**: keep modules small and focused, with explicit context
- **Quality**: strict types, clear names, explicit error handling
- **Validation**: typecheck, lint, build, and test must pass; fix warnings at their source. Moon does not fold a dependency task's hash into its dependent, so a TypeScript project declares the source of every internal package it reads in its `dependencySources` file group: the transitive `dependsOn` closure, plus the config packages its own `eslint.config.*`, `prettier.config.*`, `vitest.config.*`, `vite.config.*`, and `tsconfig*.json` name. A config package cannot be a `dependsOn` edge, because it is tagged `npm` and its publish tasks depend on the script packages, so the reverse edge makes moon reject the graph with `would_cycle`. `script-moon-release-validate` derives both sets and fails the build when the group drifts from them in either direction. Verify a change that alters a signature with `--force`, since a cache hit can otherwise hide a downstream break.
- **Testing**: TypeScript specs are tiered by what they touch. `test/specs/unit/` spawns no process and reads no repository content, so it may only reach effects through an injected port; `ts-test` and `ts-coverage` select it alone, and it is the tier every coverage floor measures. `test/specs/acceptance/` asserts against live repository content such as the skill catalog, runs inside `ci-check` through `ts-test-acceptance`, and declares the content it reads as task inputs. `test/specs/integration/` drives real binaries through `ts-test-integration` and stays out of CI. A project opts into a tier with the `typescript-acceptance` or `typescript-integration` tag. When plumbing moves behind a port, its floor moves with it: give the pure module a per-file floor in `vitest.config.ts` and let the project total fall to what the unit tier still reaches. Go tiers the same way with the `integration` and `e2e` build tags. A spec that creates a temporary directory removes it after the case that made it. Reach the filesystem through the `FileSystem` port in `script-moon-common`, taken as a defaulted last parameter, so a unit spec drives `memoryFileSystem` and only the integration tier touches a disk; `script-moon-release-validate` fails the build on a direct `node:fs` import under `src` that is not named, with its reason, in `filesystem-allowlist.json`.
- **Deprecation**: remove unused or deprecated code immediately; no `@deprecated` markers
- **Comments**: describe present behavior and name the declaration, function, or module; never reference a plan, agent, doc path, porting history, `INTERIM`, or `TODO`

### Writing

- **Prose**: write in plain sentences; prefer a comma, colon, or full stop to an em dash, use a colon after a label, write an ellipsis as three periods rather than the single ellipsis character, and use straight quotes and apostrophes rather than the typographic ones
- **Simplified Technical English**: write documentation, comments, plans, and reports in ASD-STE100 Simplified Technical English: use an approved word in its approved meaning and part of speech, keep one instruction or one idea per sentence, use the active voice and a simple present or past tense, name the same thing by the same term every time, and hold a sentence to 20 words in a procedure and 25 words in a description
- **Response length**: keep a response focused and brief. Keep a disclaimer or a caveat short, and give most of the response to the main answer. When you explain something, give a high level summary unless the request asks for depth
- **Readability**: lead with the outcome. The first sentence after you finish answers what happened or what you found, and the supporting detail comes after it. Write complete sentences and spell out a term instead of abbreviating it. Do not use an arrow chain, a hyphen stacked compound, or a label you invented earlier
- **Deliverable length**: match the length of a written document to what the task needs. Cover the substance, and leave out a filler section, a redundant summary, and boilerplate
- **Progress updates**: before the first tool call, say in one sentence what you do next. While you work, give a brief update only when you find something important or when you change direction
- **Corrections**: correct an earlier statement only when the error changes a conclusion or a decision. State the correction plainly and briefly, then continue the task. For a slip that changes nothing for the user, make the fix and continue without a note

### Reporting

- **Completion**: report when the requested work is done or blocked: quality gates, what was done, what remains, suggestions
- **Gates**: report each gate by its command and outcome; a gate that was skipped, cached, or does not cover the change is not a pass
- **Claims**: mark each claim as verified, naming the evidence, or as unverified
- **Interruption**: when work stops before completion, state the safe retry boundary: what is already durable, what is partial, and where a resumer restarts without redoing or corrupting it
- **Remainders**: keep what remains, meaning unfinished work inside the agreed scope, separate from suggestions, which are optional follow-ups; label what remains `R1`, `R2`, ... and suggestions `S1`, `S2`, ... so either can be referenced by its identifier

## Integration Points

- config -> shared -> agent

## Tone

<tone_preference>
Keep outputs reasonably concise.
</tone_preference>
