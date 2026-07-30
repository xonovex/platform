# Architecture Contract

Ground truth: every operation, skill, and entry point obeys every clause below.

## Clauses

1. **Two axes.** Skills carry knowledge and stay entry-point agnostic. An entry point
   is anything that wakes an agent and hands it a subject (the exact thing worked on):
   an interactive session, a workflow command, a schedule, an alert. Adding an entry
   point must not require a catalog change, nor the reverse; stream-shaped work
   (continuous input, no single subject) gets its own entry point, never a forced
   operation.
2. **Warm and cold boundaries.** Operations chain freely inside one warm session: one
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
