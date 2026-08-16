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
- **Moon**: a `moon.yml` tag `<name>` inherits `.moon/tasks/tag-<name>.yml`, and `extends` refines one (`tag-typescript-script.yml`, `tag-typescript-config.yml` -> `tag-typescript.yml`). A redefined task replaces the parent's `deps`, so restate every inherited dep or set `mergeDeps: replace`
- **Dependencies**: internal `@xonovex/*` exact, and a config package pins its plugins exactly; everything else caret. An exact pin elsewhere records its reason: `typescript` and `vite` in `.ncurc.cjs`, `@moonrepo/cli` because it resolves every task
- **Git**: no feature branches or pushes unless explicitly asked
- **Commits**: conventional. Rebuild a restructured series on a temporary branch and confirm `git diff <old tip> <new tip>` is empty before moving the branch; verify each commit with `npx moon run :ci-check --force` so the series stays bisectable
- **Delegation**: subagents only for large independent work such as a wide multi-file investigation, never to check your own work, and few
- **Scope**: deliver what the task asks. Judge routine calls yourself; ask only when two readings mean materially different work. Flag a mistaken request in one sentence, then continue and finish the whole task
- **Release**: reviewed `version packages` PR only; merging to `main` runs `.github/workflows/release.yml` (`:ci-publish` -> release/tag). Never bypass branch protection

### Code Style

- **Paradigm**: pure functions, immutability, composition, module-level functions, explicit state; no global mutable state
- **Imports**: direct from source; no re-exports, deprecated APIs, compatibility wrappers, or shims
- **Design**: small focused modules, explicit context, strict types, clear names, explicit error handling
- **Validation**: typecheck, lint, build, test must pass; fix warnings at source. Moon does not fold a dependency task's hash into its dependent, so a project's `dependencySources` file group lists the transitive `dependsOn` closure plus the config packages its `eslint.config.*`, `prettier.config.*`, `vitest.config.*`, `vite.config.*`, and `tsconfig*.json` name; a config package cannot be a `dependsOn` edge, which moon rejects as `would_cycle`. `script-moon-release-validate` fails on drift in either direction. Verify a signature change with `--force`; a cache hit hides a downstream break
- **Testing**: `test/specs/unit/` spawns no process and reads no repository content, reaching effects only through an injected port; `ts-test` and `ts-coverage` select it alone and every coverage floor measures it. `test/specs/acceptance/` asserts against live content such as the skill catalog, runs in `ci-check` through `ts-test-acceptance`, and declares that content as task inputs. `test/specs/integration/` drives real binaries through `ts-test-integration`, outside CI. Opt in by tag (`typescript-acceptance`, `typescript-integration`); Go uses the `integration` and `e2e` build tags. A floor moves with its code: a pure module behind a port takes a per-file floor in `vitest.config.ts`. A spec removes the temporary directory it makes. Reach the filesystem through `script-moon-common`'s `FileSystem` port as a defaulted last parameter, so a unit spec drives `memoryFileSystem`; a direct `node:fs` import under `src` fails validation unless `filesystem-allowlist.json` names it with a reason
- **Deprecation**: remove unused or deprecated code immediately; no `@deprecated` markers
- **Comments**: present behavior, naming the declaration, function, or module; never a plan, agent, doc path, porting history, `INTERIM`, or `TODO`

### Writing

- **Prose**: plain sentences; comma, colon, or full stop over an em dash; colon after a label; three periods for an ellipsis; straight quotes
- **Line breaks**: do not hard-wrap Markdown. Write one paragraph or list item on one line and leave the wrapping to whatever renders it
- **Simplified Technical English**: ASD-STE100 in replies to the user and in reports: approved words in their approved sense, one idea per sentence, active voice, simple present or past, one term per thing
- **Response length**: brief and focused; short caveats; a high-level summary unless asked for depth
- **Readability**: lead with the outcome, detail after. Complete sentences and spelled-out terms; no arrow chain, stacked compound, or invented label
- **Names in replies**: in replies to the user and in reports, name a plan, package, or document in full. Never a letter, an initial, or a document-number shorthand. A file path stays a file path
- **Deliverable length**: match the document to the task; no filler, redundant summary, or boilerplate
- **Progress updates**: one sentence before the first tool call, then only for an important find or a change of direction
- **Corrections**: correct only what changes a conclusion or a decision; state it plainly and continue

### Reporting

- **Completion**: report done or blocked: gates, what was done, what remains, suggestions
- **Gates**: name the command and its outcome; skipped, cached, or not covering the change is not a pass
- **Claims**: mark each verified, naming the evidence, or unverified
- **Interruption**: state the safe retry boundary: what is durable, what is partial, where a resumer restarts
- **Remainders**: label unfinished in-scope work `R1`, `R2`, ... and optional follow-ups `S1`, `S2`, ...

## Integration Points

- config -> shared -> agent

## Tone

<tone_preference> Keep outputs reasonably concise. </tone_preference>
