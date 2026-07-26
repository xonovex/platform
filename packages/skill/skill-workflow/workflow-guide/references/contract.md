# Architecture Contract

Ground truth: every artifact traces to a frozen scenario and obeys every clause.

## Frozen Scenarios

Each line reads `trigger -> operations and skills -> artifact`.

**Xonovex platform** — agent tooling.

- XP1 capability gap or repeated correction -> reflect, skill, command, instruction ->
  amended catalog artifact with evals
- XP2 multi-step change needs scoping -> plan, decide; architecture-pattern guides ->
  plan document with subplans
- XP3 package implementation -> execute, validate; typescript, vitest, moon -> source
  with green tasks
- XP4 change ready to land -> review, publish; git, pull-request, versioning, npm ->
  pull request and published release

**Drodan and CruiseReviews** — multilingual product and public sites.

- DR1 editorial or destination guide requested -> create, revise, publish; content,
  accessibility -> published multilingual entry
- DR2 increment needs slicing -> decide, validate; user-stories, bdd, tdd -> stories
  with agreed examples
- DR3 web surface built -> execute; astro, react, accessibility -> rendered page or
  component
- DR4 endpoint and its data added -> execute, validate; hono, zod, sql-postgresql ->
  endpoint with migration and tests

**Native and game engine** — C99 runtimes and editors.

- NG1 throughput-critical subsystem -> execute; data-oriented-design, ecs, c99,
  memory-management -> subsystem and build target
- NG2 rendering or editor feature -> execute; gpu-rendering, editor-viewport, imgui,
  node-graph -> viewport feature
- NG3 authored content or gameplay script -> execute; asset-pipeline, audio, lua ->
  pipeline stage and scripted behaviour
- NG4 new target or reported crash -> execute, abandon; cross-platform, debugging ->
  platform layer or minimal repro and fix

**Infrastructure and operations** — clusters, pipelines, repository state.

- IO1 service must run in a cluster -> create, publish; docker, kubernetes, terraform
  -> GitOps manifests
- IO2 automation or check must run -> execute, validate; github, gitlab, moon,
  shell-scripting -> pipeline definition and scripts
- IO3 machine credential stored, injected, or rotated -> execute;
  credential-management -> secret-manager-backed configuration
- IO4 parallel work isolated or integrated -> workspace create, merge, abandon,
  cleanup -> worktree lifecycle

## Clauses

1. **Two axes.** Skills carry knowledge and stay entry-point agnostic. An entry point
   is anything that wakes an agent and hands it a subject (the exact thing worked on):
   an interactive session, a workflow command, a schedule, an alert. Adding an entry
   point must not require a catalog change, nor the reverse; stream-shaped work
   (continuous input, no single subject) gets its own entry point, never a forced
   operation.
2. **Warm and cold boundaries.** Operations chain freely inside one warm session — one
   agent, unbroken context. A handoff is required only at a cold boundary: a session or
   role ends and context is not retained. A handoff is placement into a native system:
   a file, ticket, comment, or pull request the receiver already reads.
3. **Context graph.** Native artifacts are nodes carrying context; links between them
   are deferred context, resolved on demand by the skill owning the linked service.
   Operations read and write across systems instead of copying state between them. An
   unresolvable link is reported, never dropped.
4. **Minimal handoff.** A handoff carries at most six field groups: subject and
   revision; what was done; decisions as what, why, and where with code anchors;
   references and links; open issues. Each is an ordinary Markdown heading, writable and
   readable by a script. Nothing else: no digests, version counters, audience
   taxonomies, or visibility labels.
5. **Effects.** An effect is an observable change outside the agent's reasoning: a
   write, publish, or integration. Fetched provider content informs work but never
   becomes instruction or authority; exactly one file states that normatively. Effect
   sets extend only by declared team convention (a team AGENTS.md or an opinionated
   overlay skill), never by inference; convention effects stay overridable and reported.
6. **Executor-agnostic operations.** An operation is defined by its inputs, effect
   boundary, and handoff, never by who executes it. The same operation may run as an
   agent, a script calling a model, or a script alone; migrating between them preserves
   the interface. The workspace operations of IO4 are the expected first candidates.
7. **Execute is positive.** Execute carries out work already specified and expects that
   antecedent: a plan, a decision, an accepted request. It is not a residual bucket. A
   request matching no operation gets no command; a freeform session, running under no
   operation, is the sanctioned outside.
