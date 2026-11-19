# Xonovex Platform Monorepo

Work within the package boundaries below, run the required checks, and follow the release and Git rules before delivering a change.

## Structure

### Packages

- **`runtime/plugin`**: Grouped plugins that hold only skills, under `skills/`; lockstep with both marketplaces
- **`tooling/cli`**: Agent CLI, platform binary packages, and GitHub Action wrapper
- **`tooling/script`**: Moon task binaries and their shared code
- **`sandbox/operator`**: Kubernetes agent operator
- **`sandbox/image`**: Operator image build and publishing
- **`library`**: TypeScript and Go libraries
- **`config`**: Shared ESLint, Prettier, TypeScript, Vite, and Vitest configuration
- **`asset`**: Private diagrams and images, pinned at `0.0.0`

### Workflow

- **Setup**: `npm install` -> `git lfs pull`
- **Nix shell**: the root `flake.nix` is the only provider of build tools, and it composes one development shell from the modules in `nix/`. direnv enters it through the root `.envrc`, and CI enters it with `nix develop --command`. No package has a flake, and no task names a toolchain for Nix. `.moon/tasks/all.yml` makes the flake, its lock and the modules inputs of every task
- **Tasks**: `npx moon run <project>:<task>` or `npx moon run #<tag>:<task>`. `moon run` stops at the first failed task; `npx moon exec <target> --on-failure continue` runs every task and reports each failure
- **Query**: `moon query projects --tags "<pattern>"`
- **Moon**: a `moon.yml` tag `<name>` inherits `.moon/tasks/tag-<name>.yml`; `extends` refines one (`tag-typescript-script.yml`, `tag-typescript-config.yml` -> `tag-typescript.yml`). A redefined task replaces the parent's `deps`, so restate every inherited dep or set `mergeDeps: replace`
- **Dependencies**: internal `@xonovex/*` exact, a config package pins its plugins exactly, everything else caret. An exact pin elsewhere records its reason: `typescript` in `.ncurc.cjs`, `@moonrepo/cli` because it resolves every task. `npm run deps:update` moves every range and obeys the holds in `.ncurc.cjs`
- **Git**: no feature branches or pushes unless asked. Commit your own finished work by path without asking
- **Commits**: conventional, enforced by `commitlint.config.ts`. Rebuild a restructured series on a temporary branch, confirm `git diff <old tip> <new tip>` is empty before moving the branch, and verify each commit with `npx moon run :ci-check --force` so the series stays bisectable
- **Release**: reviewed `version packages` PR only; merging to `main` runs `.github/workflows/release.yml` (`:ci-publish` -> release/tag). Never bypass branch protection
- **Subagent model**: set the model and the effort of each subagent in the call that starts it: the cheapest that does its part with no loss of quality, and the most capable for a design decision or a hard fault
- **Scope**: deliver what the task asks; ask only when two readings mean materially different work. Flag a mistaken request in one sentence, then finish the whole task
- **Decisions**: do the next step that is in scope and can be undone, without asking. Ask only before a remote, destructive or paid action, or for a judgement of taste. Give a recommendation with each open decision, and report no open item that you can do yourself
- **Effort**: the cost and the time of the work are not a factor in a design or a recommendation. Name the best option and do it; do not trade a better design for a smaller change
- **Finality**: no temporary solution, and no deferred work that can be done now. Deliver each change in its final form in the same task; a follow-up item is for work that is blocked or out of scope, never for work that is merely larger
- **Plans**: this repository holds no plan documents. They live in `plans/` of the `drodan-platform` repository, and after each step of a plan its roadmap records all that a new session needs to continue
- **Gates**: after each step, run only the fast gates (`ci-check` of the touched projects) and the scenarios that prove that step. Every other check is a slow gate (`ci-integration`, `ci-e2e*`): run the slow gates that cover the touched code one time, when the whole change is implemented
- **Searching files**: never search a whole filesystem or home directory, with any tool. Search the repository or a directory you know, name it in the command, and bound the search with a timeout. For a file outside the repository, ask the tool or package manager that owns it, then whoever gave you the task

### Code Style

- **Design**: pure functions, immutability, composition, module-level functions, explicit state and context, strict types, clear names, small focused modules, explicit error handling; no global mutable state; hexagonal as far as the language allows (a pure core with I/O at the edge, ports and adapters where the language has interfaces), orthogonal (independent axes, one concern per module)
- **Imports**: direct from source; no re-exports, deprecated APIs, compatibility wrappers, or shims
- **Deprecation**: remove unused or deprecated code immediately; no `@deprecated` markers
- **Comments**: only the non-obvious _why_, never what the code or config already shows. Present behavior, naming the declaration, function, or module; never a plan, agent, doc path, porting history, `INTERIM`, or `TODO`
- **Validation**: typecheck, lint, build, test must pass; fix warnings at source. Keep a project's `dependencySources` file group in sync with its dependencies; `script-moon-release-validate` enforces it. Verify a signature change with `--force`
- **Testing**: specs live in `test/specs/<tier>/`: `unit` runs in `ci-check`, `acceptance` and `integration` opt in by tag. Unit tests use in-memory inputs and test doubles, with no filesystem, network, subprocess, or other external side effects

### Writing

- **Visual design**: diagrams follow the Xonovex design system through the mirrored style (`packages/asset/asset-brand/diagrams/style.gvpr`), and diagram sources name roles only. The Xonovex design system guide lives in the `drodan-platform` repository (`packages/xonovex/xonovex-design-system`). Keep diagram sources and rendered exports together.
- **Lead**: put the answer, result, decision, action, or status in the first sentence. Keep a decisive caveat beside it, then give evidence and background in order of importance. Use task tables and links to detailed procedures when they help the reader choose an action.
- **Prose**: plain sentences; comma, colon, or full stop over an em dash; colon after a label; three periods for an ellipsis; straight quotes. ASD-STE100 in replies and reports: approved words in their approved sense, one idea per sentence, active voice, simple present or past, one term per thing
- **Line breaks**: no hard-wrapped Markdown; one paragraph or list item per line
- **Length**: match the output to the task: brief and focused, outcome first then detail, short caveats, a high-level summary unless asked for depth; no filler, redundant summary, or boilerplate. One progress sentence before the first tool call, then only an important find or a change of direction
- **Names**: complete sentences and spelled-out terms; name a plan, package, or document in full, never a letter, an initial, or a document-number shorthand. A file path stays a file path. No arrow chain, stacked compound, or invented label
- **Corrections**: correct only what changes a conclusion or a decision; state it plainly and continue
- **Rules**: write a rule as one general sentence that states the behavior. Name no tool and give no example or reason unless the rule fails without it, and leave the judgement to the model

### Reporting

- **Completion**: report done or blocked: gates, what was done, what remains, suggestions
- **Gates**: name the command and its outcome; skipped, cached, or not covering the change is not a pass
- **Claims**: mark each verified, naming the evidence, or unverified
- **Interruption**: state the safe retry boundary: what is durable, what is partial, where a resumer restarts

## Integration Points

- The library packages use the config packages. The CLI and operator packages use the library packages.
