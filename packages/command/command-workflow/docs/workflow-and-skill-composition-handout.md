# Xonovex Workflow, Slash Command, and Skill Composition Handout

Snapshot date: 2026-07-23

Status: historical design input. The repository implementation that followed this
handout is documented in [the command package README](../README.md) and
[the migration guide](migration.md). Appendix argument shapes describe the
pre-migration snapshot and are intentionally retained as evidence.

This handout gives another model enough context to form an independent solution for
Xonovex workflow commands, multi-axis composition, role lenses, criteria,
perspectives, durable workflows, story decomposition, and composable skills.

It is deliberately not a target specification. It separates repository facts from
ideas discussed so that maintainers can retain, alter, or reject any proposal.
An entirely different architecture is acceptable if it better satisfies the
underlying goals.

## Requested task for the receiving model

Study the current implementation and the complete discussion summarized here, then
propose the architecture you think Xonovex should use.

Your answer should:

1. State the user problems and invariants before selecting abstractions.
2. Distinguish public user concepts from internal runtime concepts.
3. Decide whether the current eight workflow operations and four workspace
   operations are the right primitives.
4. Define composition without assuming that every concern is a peer axis.
5. Explain how criteria, perspectives, methods, skills, providers, capabilities,
   policy, authority, and durable state relate.
6. Show simple and advanced slash-command experiences.
7. Show how work survives context windows and continues across sessions.
8. Show how stories split into child stories or implementation subtasks.
9. Show how the primitives can support several SDLC shapes without encoding one
   mandatory lifecycle.
10. Define migration, validation, evaluation, and failure behavior.
11. Identify which ideas in this handout you reject and why.

Do not optimize for compatibility with the proposals in this document. Preserve
current behavior only where it remains valuable. Inspect the repository rather than
treating this summary as the sole source of truth.

## Reading guide

The document uses four labels:

- Current: behavior or structure observed in the repository.
- Observed issue: a concrete ambiguity, overlap, defect, or missing check.
- Discussed direction: an idea raised during the design conversation.
- Open decision: a choice intentionally left for the receiving model.

Code-shaped blocks are illustrations of machine contracts or invocations. They are
not intended to replace plain-language definitions of user concepts. Parameters
should first be explained in the language of the work; syntax is secondary.

## Executive summary

Xonovex currently has a promising set of stable workflow verbs:

- create
- review
- revise
- decide
- execute
- validate
- publish
- abandon

It also has four workspace commands:

- workspace-create
- workspace-merge
- workspace-abandon
- workspace-cleanup

These commands delegate most semantics to workflow-guide. That is a sound direction:
slash commands can remain thin user entry points while skills carry reusable
guidance.

The central unresolved problem is composition. Current documentation describes
several dimensions, but prose and arrows do not yet establish a durable composition
contract. Global provider and revision options cannot faithfully bind a subject,
evidence, and result that live on different systems. Perspectives are singular,
capability overlaps other concepts, role lenses prescribe more workflow than a lens
should, and result persistence overlaps publish.

The discussion moved toward:

- small stable operations rather than lifecycle-specific mega-commands;
- provider-native references rather than a Xonovex universal identifier;
- exact per-resource bindings rather than one global provider;
- repeatable perspectives that add questions and assurance, not authority;
- suggested or derived criteria with visible provenance;
- durable operation-result records for composition and cross-session continuation;
- an SDLC defined as policy-driven composition above the operations;
- a two-axis taxonomy for small composable skills;
- deterministic runtime enforcement for authority, policy, retries, and effects.

These remain proposals. A stronger architecture may use a typed request document,
a workflow manifest, an event model, a graph, a state machine, or another interface
instead of adding many command-line flags.

## Repository and product constraints

Current repository guidance:

- The monorepo contains tools and configuration packages under packages/.
- Tasks run through moon.
- Dependency direction is config to shared to agent.
- Code favors pure functions, immutability, composition, module-level functions,
  explicit state, strict types, focused modules, and explicit error handling.
- Source imports are direct; compatibility wrappers and re-export layers are not
  desired.
- Typecheck, lint, build, and test should pass without warnings.
- Command, skill, and marketplace versions are maintained in lockstep.
- Releases happen only through the reviewed version-packages process.

Current relevant locations:

- packages/command/command-workflow/commands/
- packages/command/command-workflow/docs/
- packages/command/command-utility/commands/
- packages/skill/skill-workflow/workflow-guide/
- packages/skill/skill-command/command-guide/
- packages/skill/
- packages/config/config-claude/
- packages/config/config-codex/

## Current command surfaces

### Harness distinction

Current:

- Claude exposes the Markdown files under command packages as slash commands.
- Codex installs and invokes skills directly; the command namespace is not a
  universal cross-harness API.
- The Claude marketplace currently contains 94 plugins: 92 skill plugins plus the
  xonovex-utility and xonovex-workflow command plugins.

This distinction matters. The semantic workflow contract may be portable while the
entry-point syntax remains harness-specific.

### Core workflow operations

Current:

| Operation | Intended semantic role                                   |
| --------- | -------------------------------------------------------- |
| create    | Produce a new work product from a subject or request.    |
| review    | Examine an existing subject and report findings.         |
| revise    | Change a subject in response to feedback or constraints. |
| decide    | Compare available information and record an outcome.     |
| execute   | Perform an action or implementation.                     |
| validate  | Test a subject against criteria and evidence.            |
| publish   | Persist or transmit a result to an external destination. |
| abandon   | Stop work and record why it should not continue.         |

The operations are intended to be reusable across planning, implementation,
investigation, review, delivery, and other processes.

### Workspace operations

Current:

| Operation         | Intended semantic role                                |
| ----------------- | ----------------------------------------------------- |
| workspace-create  | Create an isolated workspace from a source reference. |
| workspace-merge   | Integrate workspace work into a destination.          |
| workspace-abandon | Stop workspace work and record the reason.            |
| workspace-cleanup | Remove workspace artifacts and related metadata.      |

### Current composition vocabulary

Current documentation uses concepts including:

- subject
- kind
- method
- perspective
- provider
- reference
- revision
- capability
- trigger
- executor
- result

The terms are useful, but they do not all describe the same kind of variation.
Treating them as peer axes makes the model harder to explain and creates flag
overlap.

## The semantic layers discussed

A proposed reclassification is:

| Layer                          | Concepts                                                            | Purpose                                                     |
| ------------------------------ | ------------------------------------------------------------------- | ----------------------------------------------------------- |
| Operation                      | create, review, revise, decide, execute, validate, publish, abandon | What kind of transformation or judgment is requested.       |
| Semantic selection             | subject kind, result kind, perspectives, method                     | What the work means and how it should be approached.        |
| Resource binding               | provider, native reference, revision, input/output slot             | Where each concrete artifact is and which version is meant. |
| Invocation context             | trigger, executor identity, effect mode, authorization context      | How and under whose authority this invocation runs.         |
| Derived implementation binding | capabilities, adapters, selected skills                             | How the runtime can carry out the request.                  |

This avoids presenting trigger, provider, perspective, and capability as equivalent
user choices.

Open decision: keep this layering, replace it with another model, or collapse layers
that do not earn their complexity.

## Terminology

### Operation

A stable semantic verb describing the immediate kind of work. It should not imply a
whole lifecycle or a particular provider.

### Subject

The thing being acted on. It can be an inline request, a local file, a repository
artifact, a ticket, a pull request, a deployment, a prior result, or another
provider-owned object.

### Kind

The semantic type of an input or output. A single global kind is ambiguous when the
input is a story and the output is a story set, report, decision, or code change.

Open decision: distinguish subject kind and result kind explicitly, infer either
from schemas, or make kind a property of each resource binding.

### Method

A chosen procedure or technique, such as story splitting, example mapping,
functional-parity migration, or incident tracing. A method may propose steps,
criteria, evidence, and supporting skills.

### Perspective

A semantic lens that asks additional questions or requires additional evidence,
such as accessibility, security, reliability, operability, user impact, privacy,
or maintainability.

A perspective should not grant authority, impersonate a role, or silently prescribe
an entire operation chain.

### Role lens

A convenience mapping from a familiar organizational role to likely perspectives,
criteria, and evidence needs.

For example, a product role might suggest value, scope, and acceptance perspectives.
A security role might suggest threat, privacy, and abuse perspectives. The role is
not an executor identity, permission, owner, or workflow.

### Provider

The system that owns a resource, such as a repository, issue tracker, filesystem,
document store, CI system, or deployment platform.

### Reference

The provider-native locator of a resource. It should remain opaque outside the
provider adapter. A wrapper can carry the provider and revision but should not
invent a universal Xonovex object identity.

### Revision

The exact version, commit, generation, ETag, updated-at token, or other concurrency
marker for a referenced resource.

### Capability

An available implementation mechanism: an installed skill, runtime tool, provider
adapter, API operation, or other means the runtime can use.

Capability is not:

- the requested operation;
- a perspective;
- a method;
- a provider;
- a role;
- permission or authority.

The ordinary caller usually should not select a generic capability. The runtime
should derive it from the operation, bindings, policy, and installed mechanisms,
then report what it selected. Expert overrides may be useful when they refer to an
exact known capability.

### Skill

A reusable body of knowledge or procedure that helps the model reason or act
reliably. A skill does not grant access or authority.

### Policy

A deterministic rule about what is required or allowed. Skills may explain policy,
but they must not enforce or override authorization, approvals, budgets, or other
mandatory controls.

## The eight main design questions discussed

### 1. Result persistence versus publish

Observed issue:

Create, review, revise, decide, execute, validate, and abandon can describe a result
destination while publish is also defined as the operation that persists a result.
This creates two paths to the same effect and makes auditing and authorization less
clear.

Discussed direction A:

- Core operations produce inline or ephemeral results.
- Publish is the only operation that writes a result to an external destination.
- A workflow host may automatically run operation then publish when destination and
  authority are predefined, but the record still contains two operations.

Discussed direction B:

- Remove publish as a first-class operation.
- Treat output binding and persistence as a cross-cutting part of every operation.
- Make preview, apply, idempotency, and authority consistent across all operations.

Question for the receiving model:

Which model produces clearer semantics, fewer accidental effects, and better
composability? Do not retain both unless their boundaries are unambiguous.

### 2. Bind provider and revision to each resource

Observed issue:

One global provider and one global revision cannot represent a subject from Jira,
evidence from Datadog, code in GitHub, and a report published to Confluence.

Discussed direction:

Each named input and output slot carries its own:

- provider;
- native reference;
- revision;
- semantic kind;
- optional schema;
- read or write intent.

Simple invocations can still use a positional subject plus subject-provider and
subject-revision shorthand. Advanced requests can use labeled bindings or a request
document.

Illustrative logical request:

```yaml
operation: review
subject:
  provider: github
  reference: pull-request-native-reference
  revision: exact-head-revision
  kind: pull-request
evidence:
  - provider: datadog
    reference: trace-native-reference
    revision: immutable-event-set
result:
  kind: review-report
```

The field names are illustrative, not prescribed.

### 3. Define an operation-result contract

Observed issue:

Arrows between operations explain intent to a human but do not provide a stable
handoff between sessions, models, providers, or workflow hosts.

Discussed direction:

Every operation returns a human-readable summary and a durable machine-readable
envelope. The envelope should preserve:

- contract version;
- operation;
- exact subject and input bindings;
- resolved method, perspectives, criteria, and skills;
- provenance and reason for every derived selection;
- completion status: completed, partial, blocked, or failed;
- inline result or provider-native result reference;
- evidence references;
- observed effects;
- unresolved questions and uncertainty;
- safe retry boundary;
- concurrency or revision information.

This envelope is a composition protocol, not necessarily a central database.

Open decisions:

- Must every harness render the same envelope?
- Is it embedded in Markdown, emitted as structured output, or both?
- Is the envelope an event, an artifact, a state transition, or a combination?
- Which fields are stable public contract and which are runtime diagnostics?

### 4. Stop treating every concern as a peer axis

Observed issue:

Operation, provider, perspective, role, trigger, executor, capability, and revision
have different semantics and ownership. Calling all of them axes hides those
differences.

Discussed direction:

Use semantic layers, typed slots, or another explicit model. Keep the public surface
small and derive implementation choices. Report derivations rather than forcing
every internal decision into a user-facing flag.

### 5. Narrow role lenses

Observed issue:

Current role guidance says a role is not permission, but some role descriptions
still imply fixed operation chains, ownership, or responsibilities.

Discussed direction:

A role lens can suggest:

- perspectives;
- evidence sources;
- likely criteria;
- vocabulary and communication depth.

It cannot:

- grant access;
- authorize effects;
- determine executor identity;
- claim organizational ownership;
- force a fixed workflow;
- replace a named perspective or method.

The command surface should prefer repeatable perspectives. Role lenses, if retained,
are an optional convenience resolved into those perspectives.

### 6. Make constraint resolution explicit

Observed issue:

It is unclear when an inferred method, criterion, perspective, or capability becomes
binding and how conflicting selections are resolved.

Discussed precedence:

1. Explicit caller selection.
2. An accepted proposal made by a selected method or workflow.
3. A reported, unambiguous inference from authoritative context.
4. Unresolved, requiring a question, safe default, or blocked result.

Suggested requirements:

- Record the source and confidence of each derived value.
- Do not silently turn a suggestion into a binding acceptance gate.
- Surface provider and capability constraints.
- Distinguish a default from an inference.
- Preserve conflicting evidence rather than resolving it by prompt order.

### 7. Clarify workspace effects

Observed issue:

Workspace-merge and workspace-abandon overlap workspace-cleanup when they also remove
branches, worktrees, references, or metadata.

Discussed direction A:

- workspace-create only creates;
- workspace-merge inspects, preflights, updates, and integrates but does not remove;
- workspace-abandon records and snapshots abandonment but does not remove;
- workspace-cleanup exclusively removes workspace artifacts.

Discussed direction B:

Split workspace behavior into smaller atomic primitives and let workflows compose
them.

Discussed direction C:

Keep convenience macros, but define them as explicit transactions composed from
named primitives with preview, compensation, and per-effect reporting.

### 8. Align permissions and effect modes

Observed issue:

Read-oriented commands can preauthorize broad shell or write behavior. Prompt-level
tool declarations are not a substitute for runtime policy.

Discussed direction:

- Use effect modes such as inspect, preview, and apply.
- Minimize capabilities preauthorized for each operation.
- Require provider or runtime authorization at the actual effect.
- Record approvals and effects in the result.
- Keep credentials, grants, approval enforcement, and kill switches outside skills.

## Criteria: derivation, suggestion, and authority

Criteria can and often should be derived or suggested. The important distinction is
whether a criterion is binding.

### Potential criterion sources

| Source                | Typical status             | Example                                                         |
| --------------------- | -------------------------- | --------------------------------------------------------------- |
| Explicit caller input | Binding                    | The endpoint must remain backward compatible.                   |
| Exact parent artifact | Binding when authoritative | Acceptance criteria inherited from the selected story revision. |
| Mandatory policy      | Binding                    | Required security or release gate.                              |
| Selected method       | Suggested until accepted   | INVEST or functional-parity checks.                             |
| Perspective           | Advisory by default        | Accessibility or operability questions.                         |
| Model inference       | Advisory by default        | Likely performance concern derived from a hot path.             |

An advisory finding should not silently fail an acceptance gate. The caller or an
authoritative policy must promote it to binding status.

### Suggested resolution modes

Strict:

- Use only explicit or authoritative criteria.
- Report missing criteria.
- Suitable for regulated or tightly controlled gates.

Assisted:

- Suggest criteria and perspectives.
- Automatically use high-confidence advisory checks while reporting them.
- Require confirmation before making new criteria binding.
- This was the recommended default in the discussion.

Automatic:

- Apply high-confidence suggestions and report all derivations.
- Never invent binding validation, release, or merge criteria.
- Stop or escalate when uncertainty changes safety or authority.

Open decision: whether these should be public modes, policy profiles, or internal
resolution behavior.

## Perspective suggestion and selection

Perspectives may be explicit, inherited, suggested, or derived from the subject and
change shape.

Illustrative suggestions:

| Change shape                    | Likely perspectives                                         |
| ------------------------------- | ----------------------------------------------------------- |
| Public API                      | compatibility, security, reliability, consumer impact       |
| User interface                  | usability, accessibility, localization                      |
| Data migration                  | integrity, reversibility, operations, privacy               |
| Authentication or authorization | security, privacy, abuse prevention, auditability           |
| Build or dependency change      | reproducibility, supply-chain security, portability         |
| Production incident             | customer impact, reliability, operability, evidence quality |

Selection should include:

- selected perspective;
- whether it was explicit, inherited, suggested, or inferred;
- why it applies;
- supporting skill or baseline behavior;
- questions, criteria, and evidence it added;
- whether any added criterion is advisory or binding.

Repeatable perspectives are preferable to a singular perspective flag when more than
one concern applies.

## Perspectives and skills

A perspective is a semantic request, not necessarily a skill family.

The relationship is many-to-many:

- One perspective may resolve to no skill, one skill, or several skills.
- One skill may support several perspectives.
- A method skill owns the procedure.
- Perspective-oriented or assurance skills add questions, checks, and evidence.
- The operation owns the final result contract.
- Provider adapters resolve references and perform authorized effects.

Specialist perspectives such as accessibility, security, reliability, privacy, or
data integrity will often benefit from dedicated skills. Broad perspectives such as
clarity, correctness, or user impact may be baseline behavior or part of the chosen
method.

Workflow skills should normally use soft semantic requirements instead of hard
dependencies on every possible lens. An unavailable specialist skill should produce
a visible resolution result, not silently remove the concern.

Avoid persona bundles such as developer-reviewer that combine domain, procedure,
permissions, and communication into one monolithic prompt.

## Proposed two-axis skill taxonomy

The user proposed two independent classification axes.

### Axis 1: lifecycle type

Original terms:

- Preference skill: durable organizational knowledge.
- Procedural skill: temporary instructions that compensate for model limitations.

Naming concern:

Preference is also a functional family below, while durable knowledge includes facts
and topology that are not preferences. Durable and Procedural may therefore be
clearer lifecycle names.

Open decision: preserve the original terminology or rename the lifecycle axis.

### Axis 2: functional role

- Domain
- Context
- Preferences
- Procedure
- Capability usage
- Assurance
- Recovery
- Communication

### Family summary

| Skill family              | Purpose                                                      | Proposed typical lifecycle |
| ------------------------- | ------------------------------------------------------------ | -------------------------- |
| Domain knowledge          | Explain company-specific concepts and invariants.            | Durable                    |
| Environment context       | Explain repositories, services, ownership, and topology.     | Durable                    |
| Conventions               | Define how the organization or team prefers work to be done. | Durable                    |
| Capability usage          | Teach reliable use of one or more capabilities.              | Procedural                 |
| Task procedure            | Describe how to complete a task class.                       | Procedural                 |
| Assurance                 | Define how to verify correctness, safety, and completeness.  | Split or composed          |
| Recovery and escalation   | Handle failures, uncertainty, and blocked execution.         | Procedural                 |
| Communication and handoff | Define how results are reported or transferred.              | Usually durable preference |

A concern raised in the discussion is that Mixed is not a precise lifecycle.
Assurance can be split into durable standards and procedural verification. Recovery,
communication, and capability usage can also contain both durable facts and
procedural guidance; decomposition may be better than a mixed label.

### General composition

```text
domain knowledge
+ environment context
+ applicable preferences
+ task procedure
+ capability-use instructions
+ assurance
+ recovery and escalation
+ communication format
```

This is a dependency graph, not raw prompt concatenation.

### Domain knowledge skills

Purpose:

- business terminology;
- entity relationships;
- state machines;
- invariants;
- known exceptions;
- important business rules;
- authoritative references;
- ownership boundaries.

Examples:

- domain.booking-model
- domain.passenger-rights
- domain.payment-lifecycle
- domain.order-state-machine
- domain.incident-severity
- domain.service-ownership

They explain what data and processes mean, not how to call APIs.

### Environment and topology skills

Purpose:

- repository structure;
- service dependencies;
- environments;
- CI/CD topology;
- logging and tracing relationships;
- team ownership;
- dashboards and source-of-truth locations;
- legacy boundaries.

Examples:

- context.repository-map
- context.service-topology
- context.deployment-environments
- context.logging-topology
- context.team-ownership
- context.pipeline-structure

The same context skill can support implementation, tracing, incident diagnosis, and
release procedures.

### Preference and convention skills

Purpose:

- coding style;
- preferred libraries;
- architecture;
- dependency selection;
- testing and logging conventions;
- naming and error handling;
- documentation;
- pull-request conventions;
- definition of done.

Example specificity:

```text
company
  -> department
    -> team
      -> language or framework
        -> repository
          -> task-specific override
```

Only preferences should silently overlay, and narrower preferences should override
broader ones only when the override is explicit.

Domain and context conflicts are different. They should be resolved through
authority, freshness, scope, and evidence rather than a generic most-specific-wins
rule.

### Capability-use skills

Purpose:

- when to use a capability;
- required inputs;
- query or request construction;
- result interpretation;
- limitations and common errors;
- required evidence;
- when not to use it;
- effect verification.

Examples:

- capability-use.logs-search
- capability-use.trace-inspect
- capability-use.repository-edit
- capability-use.pipeline-retry
- capability-use.deployment-promote
- capability-use.pull-request-create

A capability-use skill explains reliable operation. It does not grant the
capability.

### Task procedure skills

Purpose:

- preconditions;
- inputs;
- supporting skill roles;
- required functional capabilities;
- ordered or partially ordered steps;
- decisions;
- outputs;
- completion criteria;
- failure and escalation conditions.

Examples:

- procedure.trace-request
- procedure.diagnose-pipeline-failure
- procedure.add-api-endpoint
- procedure.create-react-application
- procedure.migrate-express-to-hono
- procedure.investigate-incident
- procedure.prepare-release
- procedure.review-pull-request

Generic procedures should depend on semantic roles or functional capabilities where
possible. Exact skill identifiers are appropriate only for invariant dependencies.

### Assurance skills

Purpose:

- evidence quality;
- change impact;
- functional parity;
- test selection;
- security review;
- observability;
- release readiness;
- definition of done.

Examples:

- assurance.evidence
- assurance.change-impact
- assurance.functional-parity
- assurance.test-selection
- assurance.security-review
- assurance.observability
- assurance.release-readiness
- assurance.definition-of-done

Assurance should wrap or compose with procedures rather than be copied into each
procedure.

### Recovery and escalation skills

Purpose:

- retryable versus non-retryable failure;
- alternative evidence;
- compensation and rollback;
- unknown write outcomes;
- replanning;
- human approval or expertise;
- immediate termination conditions.

Examples:

- recovery.capability-timeout
- recovery.partial-write
- recovery.test-failure
- recovery.conflicting-evidence
- recovery.unexpected-repository-shape
- recovery.rollback
- escalation.request-approval
- escalation.request-domain-expert
- escalation.report-blocker

Runtime policy still enforces retry limits, approvals, budgets, and authority.

### Communication and handoff skills

Purpose:

- audience;
- required sections;
- terminology and technical depth;
- evidence and risk formatting;
- open questions;
- links;
- expected next actions.

Examples:

- communication.incident-report
- communication.pull-request-description
- communication.architecture-decision
- communication.change-summary
- communication.operator-handoff
- communication.executive-summary

The same investigation can therefore produce an operator handoff or an executive
summary without changing the investigation procedure.

### Composition operators

Overlay:

- Intended for explicit preference precedence.

Dependency:

- A skill requires another semantic role, skill, schema, or capability.

Sequence:

- One procedural stage follows another.
- Deterministic workflows should encode the sequence in runtime code or a validated
  workflow definition, not rely only on model prose.

Wrapper:

- Assurance or governance behavior surrounds a procedure.

Conditional composition:

- Context adds a skill or requirement, such as migration assurance after a schema
  change or rollback guidance for production.

### Runtime responsibilities that are not skills

Do not implement these as skills:

- authentication;
- authorization;
- credentials;
- capability grants;
- approval enforcement;
- runtime budgets;
- trigger validation;
- audit logging;
- retry limits;
- idempotency enforcement;
- sandbox isolation;
- mandatory policy;
- kill switches;
- skill conflict resolution.

The clean conceptual separation is:

```text
Domain skills explain what things mean.
Context skills explain where things are.
Preference skills explain how the organization works.
Procedure skills explain how to perform a task.
Capability-use skills explain how to operate mechanisms.
Assurance skills explain how to prove correctness.
Recovery skills explain how to handle failure.
Communication skills explain how to present the result.

The runtime composes them.
Capabilities perform effects.
Policy determines authority.
Evaluations test whether the composition works.
```

### Mapping workflow inputs to skill families

| Workflow concern                     | Likely skill families          |
| ------------------------------------ | ------------------------------ |
| Subject and semantic kind            | domain, context                |
| Repository, service, and environment | context, preferences           |
| Perspective                          | assurance, preferences, domain |
| Method                               | procedure                      |
| Provider and tool                    | capability usage               |
| Risk, uncertainty, or failure        | recovery and escalation        |
| Audience and destination             | communication and handoff      |

### Skill identity and metadata

The dotted names above are useful conceptual coordinates. Current Agent Skill names
use kebab-case and must match their parent directories.

If the taxonomy becomes operational, catalog metadata could carry:

- family;
- lifecycle;
- scope;
- authority and freshness metadata for durable knowledge;
- semantic requirements;
- provided perspectives or methods;
- supported capabilities;
- schemas produced;
- version.

Persisted operation results should record the exact skill identifiers and versions
selected, plus the selection reason. This makes cross-session execution
reproducible.

## Durable workflows across sessions

Chat history is not a reliable workflow store. A workflow that spans context windows,
days, or multiple agents needs durable provider-owned state.

Discussed direction:

- Use a native work record such as a tracker item, repository file, database row,
  pull request, or document.
- Keep provider-native identities rather than minting a mandatory global Xonovex
  object ID.
- Persist exact references and revisions.
- Persist selected methods, perspectives, criteria, skills, versions, and
  resolution reasons.
- Persist operation results, evidence, effects, unresolved questions, and retry
  boundaries.
- Resume from those exact records rather than relying on conversational memory.
- Use optimistic concurrency or revision pinning for parallel sessions.

Same-conversation shorthand may refer to the exact immediately preceding operation
result. Cross-session continuation must use a persisted result or an explicitly
supplied provider-native reference.

### Illustrative durable record

```yaml
work:
  provider: tracker
  reference: native-work-reference
  revision: exact-work-revision
parent:
  provider: tracker
  reference: native-parent-reference
inputs:
  - provider: repository
    reference: native-code-reference
    revision: exact-commit
resolution:
  method:
    value: story-splitting
    source: explicit
  perspectives:
    - value: user-value
      source: suggested
      reason: parent is a user story
  skills:
    - id: user-stories-guide
      version: exact-installed-version
criteria:
  - text: Every parent acceptance criterion is covered by a child.
    status: binding
    source: parent-artifact
latestResult:
  status: partial
  unresolved:
    - Ownership of the reporting slice is unknown.
```

Again, the structure is illustrative. The receiving model should design the actual
contract.

## Story decomposition

Story splitting is a useful test of whether the operation model composes.

### Story into smaller stories

One possible composition:

1. Create with the parent story as the exact subject.
2. Select a story-set result kind and story-splitting method.
3. Derive candidate criteria from the parent, INVEST, SPIDR, and vertical slicing.
4. Review the split for value, independence, coverage, gaps, and overlap.
5. Revise the candidate split.
6. Decide which split to use without treating the decision as publication authority.
7. Publish provider-native child stories and parent links.
8. Validate that the published children cover the parent criteria.

The split result should include:

- exact parent reference and revision;
- stable draft labels before provider IDs exist;
- child story cards and acceptance criteria;
- mapping from every parent criterion to one or more children;
- dependencies and recommended order;
- deferred or residual scope;
- detected gaps and overlaps;
- rationale for vertical slices.

Good child stories are independently valuable, small, and testable. Avoid horizontal
splits such as database, API, and UI stories when no slice delivers user value on its
own.

### Story into implementation subtasks

Subtasks are different from child stories:

- They are engineering steps under one valuable story.
- They may be horizontal.
- They need not independently deliver user value.
- They should be specific, observable, and bounded.
- Their completion should roll up to the story's acceptance criteria.

A procedure can therefore select result kind child-stories or subtasks without
inventing two unrelated command verbs.

### Parallel work

Each child story can run in a separate session or workspace using its exact native
reference and revision. A parent work record aggregates child state and coverage.

Before closing the parent, validation should check:

- all binding parent criteria are covered;
- all required children are completed or explicitly deferred;
- integration evidence exists;
- cross-child gaps and conflicts are resolved;
- the parent revision has not changed unnoticed.

## SDLC composition

The SDLC belongs one layer above the stable operations. A process or policy chooses
which operations are required, their order, evidence gates, loops, and authority.

Illustrative lifecycle:

```text
Discover -> Define -> Design -> Plan -> Build -> Verify -> Review -> Release -> Operate
              ^          |        |        |         |
              +----------+--------+---------+---------+
                         feedback and revision loops
```

Possible operation mappings:

| Stage              | Possible operations                 |
| ------------------ | ----------------------------------- |
| Discover           | create, review, revise              |
| Define             | create, review, revise, decide      |
| Design             | create, review, revise, decide      |
| Plan               | create, review, revise              |
| Build              | workspace-create, execute           |
| Verify             | validate                            |
| Review             | review, execute fixes, validate     |
| Integrate          | publish or workspace-merge          |
| Release            | validate, decide readiness, publish |
| Deploy and operate | execute, validate, review           |

Important separation:

- Operations produce results and evidence.
- SDLC policy requires particular results and evidence.
- Provider and organization policy grant authority.
- A ready decision does not itself release.
- An approved review does not itself merge.
- A passing validation does not itself deploy.

No universal lifecycle should be mandatory. A trivial fix, incident response,
documentation change, product discovery effort, and regulated release need different
graphs. The same operation vocabulary should support omission, repetition,
parallelism, and feedback loops.

## Slash-command experience

### Design principles

- Keep common cases short.
- Explain concepts in work language before exposing flags.
- Prefer repeatable perspectives.
- Infer implementation capabilities and report them.
- Make every external effect explicit or policy-controlled.
- Use exact resource bindings for advanced multi-provider work.
- Allow a request file when flags become cumbersome.
- Preserve a thin command that delegates semantics to skills or a workflow runtime.

### Simple examples

```text
/create "A traveller can change a refundable booking"

/review ./proposal.md

/review ./proposal.md --perspective accessibility --perspective security

/revise ./proposal.md --feedback "Clarify rollback and ownership"

/validate ./proposal.md --criteria "Every public API change is documented"

/decide ./options.md --outcome "Select option B"
```

These examples are illustrative. The receiving model should decide exact syntax.

### Simple provider-bound example

```text
/review PR-NATIVE-REFERENCE \
  --subject-provider github \
  --subject-revision EXACT-HEAD-REVISION \
  --perspective compatibility \
  --perspective security
```

### Advanced multi-provider example

A flat list of provider flags becomes ambiguous. A labeled request block or request
file is clearer:

```text
/review --request review-request.yaml
```

```yaml
subject:
  provider: github
  reference: native-pull-request-reference
  revision: exact-head-revision
evidence:
  - provider: datadog
    reference: native-trace-reference
  - provider: jira
    reference: native-story-reference
    revision: exact-story-revision
perspectives:
  - compatibility
  - reliability
effectMode: inspect
```

### Publish as a separate effect

If publish remains an operation:

```text
/review ./proposal.md --perspective security
/publish PRIOR-RESULT --destination-provider confluence \
  --destination-reference NATIVE-PAGE-REFERENCE \
  --effect preview
/publish PRIOR-RESULT --destination-provider confluence \
  --destination-reference NATIVE-PAGE-REFERENCE \
  --effect apply
```

Across sessions, PRIOR-RESULT must be replaced by a durable native reference or
result record rather than conversational shorthand.

### Criteria and perspective suggestions in the interaction

An assisted interaction might report:

```text
Selected perspectives
- compatibility: inferred from a public API change
- security: suggested because authorization behavior changed

Criteria
- Binding: existing story acceptance criteria at revision 17
- Binding: organization API compatibility policy
- Advisory: rollback rehearsal recommended by the reliability perspective

Proceeding with advisory checks. No advisory item will become a release gate without
explicit acceptance or policy authority.
```

### Why not expose capability as a normal flag?

The user asks for review, validation, or execution in terms of work. The runtime
should determine whether that needs repository read, logs search, trace inspection,
or a particular skill. Exposing capability too early makes users describe the
implementation instead of the desired outcome.

An advanced exact override can remain possible when deterministic selection is
required.

## Validation and evaluation architecture

Command validation parses frontmatter, argument declarations, argument hints,
semantic requirement declarations, and cross-package links. Requirement lines use
the same provision identifiers and SemVer ranges as the composition catalog, and
validation resolves them against the installed skill inventory.

Workflow request, result, and work-record fixtures are validated against executable
JSON Schemas. Requests carry semantic requirements and preference overlays; results
preserve selected identities, natural selection failures, overlay conflicts, and the
exact catalog identity. Fixture validation rejects stale catalog digests.

The canonical composition catalog is synchronized into the workflow skill as an
exact packaged snapshot and checked for byte identity. A shared runtime performs
exact selection, semantic selection, required-versus-preferred handling, and
deterministic broad-to-narrow preference overlay resolution.

Structural routing and output evaluation contracts run in normal package checks.
A scheduled canary additionally exercises the skill, command, and workflow guides
through both routing and output harnesses; it can also be dispatched against either
supported harness.

## Required evaluation scenarios

### Composition

- Review a GitHub pull request using evidence from Datadog and a story from Jira.
- Continue the review in a new session without relying on prior chat.
- Run the same operation with accessibility, security, and reliability
  perspectives.
- Resolve two skills that support the same perspective without duplicating checks.
- Handle an unavailable specialist skill visibly.
- Detect contradictory domain or context skills without using preference overlay.

### Criteria

- Inherit binding criteria from an exact parent-story revision.
- Suggest method criteria but keep them advisory until accepted.
- Apply mandatory policy criteria regardless of the selected method.
- Explain every inferred criterion and perspective.
- Refuse to invent a binding release gate.

### Effects and authority

- Inspect without acquiring write capabilities.
- Preview every proposed external write.
- Apply an authorized publish exactly once.
- Recover safely from an unknown publish outcome using idempotency state.
- Prevent review, decision, or validation language from granting merge or release
  authority.

### Story decomposition

- Split one story vertically into independently valuable child stories.
- Produce subtasks instead when independent value is not intended.
- Preserve full parent acceptance-criteria coverage.
- Coordinate parallel child sessions with revision checks.
- Detect overlap, gaps, and a changed parent revision.

### SDLC variability

- Express a small defect workflow with minimal ceremony.
- Express a regulated release with required evidence and approvals.
- Express incident response with rapid mitigation and later follow-up.
- Repeat or skip stages without changing operation semantics.

### Recovery

- Capability timeout before any effect.
- Partial write with a known successful subset.
- Unknown write outcome.
- Conflicting evidence.
- Missing logs due to retention.
- Repository topology different from durable context.
- Concurrent modification of a referenced artifact.

### Harness behavior

- Claude slash command delegates correctly.
- Codex skill invocation yields equivalent semantics without requiring a slash
  command namespace.
- Human output remains useful when structured output is unavailable.
- Structured output remains valid when the human summary changes.

## Design decisions to revisit explicitly

1. Are the eight core operations complete, redundant, or incorrectly factored?
2. Should publish remain an operation or become output binding?
3. Are workspace commands primitives, transactions, or convenience macros?
4. What is the minimum stable public composition contract?
5. Which information belongs in a request, result, event, or durable work record?
6. How are multiple providers and revisions represented?
7. Is perspective a first-class concept, a skill query, or both?
8. Are role lenses worth retaining?
9. How do suggested perspectives and criteria become accepted?
10. Which criteria can be binding, and who or what has authority to bind them?
11. How are preference overlays distinguished from fact conflicts?
12. Should skills declare exact dependencies or semantic requirements?
13. How are skills selected, versioned, evaluated, and recorded?
14. What happens when required skills or capabilities are unavailable?
15. Which workflow transitions must be deterministic code?
16. What does idempotency mean for each effectful operation?
17. How should simple slash commands expand into advanced requests?
18. What compatibility or migration promises are actually needed?

## Architecture documentation

Keep these artifacts current:

- a concise conceptual model;
- a glossary with non-overlapping definitions;
- a request contract;
- a result and evidence contract;
- a durable cross-session state model;
- a skill metadata and dependency model;
- a constraint-resolution algorithm;
- an effect, authority, and approval model;
- several slash-command examples;
- two or three SDLC workflow graphs;
- a story-splitting example;
- failure and recovery behavior;
- a compatibility policy for package evolution;
- semantic validators and evaluation cases;
- rejected alternatives and tradeoffs.

## Appendix A: complete current skill catalog

Current count: 92 skills.

The catalog below lists the skill directory name used by Agent Skills and its
present focus. Paths are under packages/skill/skill-*/.

| Skill                       | Present focus                                                 |
| --------------------------- | ------------------------------------------------------------- |
| accessibility-guide         | Web accessibility review and implementation guidance.         |
| adr-guide                   | Architecture decision records and decision documentation.     |
| agent-governance-guide      | Governance boundaries for agent systems.                      |
| ai-governance-guide         | Organizational governance for AI use and systems.             |
| android-analytics-guide     | Analytics instrumentation in Android applications.            |
| android-wcag-guide          | Accessibility and WCAG concerns for Android applications.     |
| asset-pipeline-guide        | Asset import, compilation, caching, and runtime formats.      |
| astro-guide                 | Astro sites, content collections, and islands architecture.   |
| atlassian-guide             | Atlassian product and workflow usage.                         |
| audio-guide                 | Low-level real-time audio and mixing systems.                 |
| aws-guide                   | AWS platform usage and architecture.                          |
| azure-devops-guide          | Azure DevOps repositories, pipelines, and work management.    |
| bdd-guide                   | Behavior-driven development and concrete examples.            |
| bitbucket-guide             | Bitbucket repository and pull-request workflows.              |
| bitrise-guide               | Bitrise CI/CD configuration and operation.                    |
| c99-game-opinionated-guide  | Opinionated C99 game-engine and runtime style.                |
| c99-opinionated-guide       | Opinionated data-oriented C99 systems style.                  |
| c99-guide                   | General-purpose C99 implementation and review.                |
| claude-code-guide           | Claude Code configuration and usage.                          |
| cmake-guide                 | CMake build configuration for C and C++.                      |
| code-quality-guide          | Read-only code-quality auditing and issue routing.            |
| code-review-guide           | Structured code-review feedback.                              |
| codex-guide                 | Codex configuration and usage.                                |
| command-guide               | Authoring and maintaining reusable slash commands.            |
| connascence-guide           | Coupling, cohesion, and connascence analysis.                 |
| content-guide               | Multilingual structured content and editorial guidance.       |
| copilot-guide               | GitHub Copilot configuration and usage.                       |
| credential-management-guide | Credential storage, rotation, and handling.                   |
| cross-platform-guide        | Native cross-platform abstraction and porting.                |
| data-model-guide            | Typed in-memory object and property models.                   |
| data-oriented-design-guide  | Cache-aware data layouts and processing.                      |
| datadog-guide               | Datadog observability, queries, and operations.               |
| ddd-guide                   | Domain-driven design and bounded contexts.                    |
| debugging-guide             | Native and low-level debugging procedures.                    |
| docker-guide                | Production Docker images and Compose configuration.           |
| ecs-guide                   | Data-oriented entity-component-system architecture.           |
| editor-viewport-guide       | Interactive 3D editor viewport behavior.                      |
| expressjs-guide             | Express 5 TypeScript API servers.                             |
| fdd-guide                   | Feature-Driven Development process scaffolding.               |
| figma-guide                 | Figma design-system and design workflow usage.                |
| fp-guide                    | Functional programming style and composition.                 |
| game-networking-guide       | Real-time multiplayer networking architecture.                |
| git-guide                   | Git operations, history, conflicts, and conventions.          |
| github-guide                | GitHub repositories, issues, pull requests, and workflows.    |
| gitlab-guide                | GitLab repositories, merge requests, and pipelines.           |
| gpu-rendering-vulkan-guide  | Concrete Vulkan renderer implementation.                      |
| gpu-rendering-guide         | Explicit GPU renderer architecture.                           |
| hexagonal-pattern-guide     | Ports-and-adapters and clean architecture.                    |
| hono-opinionated-guide      | Opinionated Hono API implementation style.                    |
| hono-guide                  | Hono TypeScript API servers.                                  |
| imgui-guide                 | Immediate-mode GUI architecture and implementation.           |
| instruction-guide           | AGENTS.md project-instruction lifecycle.                      |
| kiro-guide                  | Kiro configuration and usage.                                 |
| kubernetes-guide            | Kubernetes manifests in GitOps repositories.                  |
| llmstxt-guide               | llms.txt and Markdown content mirrors.                        |
| lock-free-guide             | Atomics and lock-free shared-memory concurrency.              |
| lua-opinionated-guide       | Performance-focused Lua and LuaJIT style.                     |
| lua-guide                   | General-purpose Lua 5.4 implementation.                       |
| memory-management-guide     | Ownership, allocation, and buffer-passing design.             |
| microkernel-pattern-guide   | Minimal cores and plug-in architectures.                      |
| moon-guide                  | moonrepo task and project configuration.                      |
| motion-guide                | React animation with Motion.                                  |
| node-graph-guide            | Visual node and typed data-flow graphs.                       |
| npm-guide                   | npm package and workspace management.                         |
| oop-guide                   | Object-oriented design and class modeling.                    |
| opencode-guide              | OpenCode configuration and usage.                             |
| orthogonal-pattern-guide    | Decomposition along independent variation points.             |
| pi-guide                    | Pi coding-agent configuration and usage.                      |
| plan-guide                  | Planning, research, revision, and validation.                 |
| presentation-guide          | Slide decks and visual codebase walkthroughs.                 |
| pull-request-guide          | Pull-request preparation and description.                     |
| python-guide                | Python 3.12 implementation and review.                        |
| react-guide                 | React 19 components, hooks, and routing.                      |
| reflect-guide               | Session reflection and reusable lesson extraction.            |
| reliability-guide           | Reliability engineering and operational review.               |
| remotion-guide              | Programmatic video with Remotion.                             |
| security-assurance-guide    | Security assurance and evidence.                              |
| shell-scripting-guide       | POSIX shell and Bash automation.                              |
| skill-guide                 | Authoring, reviewing, composing, and validating Agent Skills. |
| sql-postgresql-guide        | PostgreSQL schemas, migrations, and queries.                  |
| strudel-guide               | Browser-based live-coded music with Strudel.                  |
| tdd-guide                   | Test-driven development and red-green-refactor.               |
| terraform-guide             | Terraform infrastructure configuration.                       |
| testing-guide               | Framework-independent test design.                            |
| threejs-guide               | Three.js WebGL and WebGPU scenes.                             |
| typescript-to-lua-guide     | TypeScript compiled to Lua through TSTL.                      |
| typescript-guide            | TypeScript in Node.js ESM projects.                           |
| user-stories-guide          | User-story writing, evaluation, and splitting.                |
| versioning-guide            | Package versions, releases, and changelogs.                   |
| vitest-guide                | Vitest tests and configuration.                               |
| workflow-guide              | Generic workflow operations and composition guidance.         |
| zod-guide                   | Zod runtime validation schemas.                               |

## Appendix B: complete current slash-command catalog

Current count: 35 commands.

### Workflow commands

| Command            | Purpose                                            | Current argument shape                                                                                                                                                                                                                                                                                                                                                                              |
| ------------------ | -------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| /abandon           | Stop work on a subject and record a reason.        | &lt;subject&gt; --reason &lt;text&gt; [--reference &lt;reference&gt;...] [--revision &lt;revision&gt;] [--kind &lt;selection&gt;] [--perspective &lt;selection&gt;] [--criteria &lt;criteria&gt;...] [--method &lt;selection&gt;] [--capability &lt;selection&gt;...] [--provider &lt;selection&gt;] [--result &lt;destination-reference&gt;] [--cleanup &lt;selection&gt;] [--confirm] [--dry-run] |
| /create            | Create a work product.                             | &lt;subject&gt; [--reference &lt;reference&gt;...] [--revision &lt;revision&gt;] [--kind &lt;selection&gt;] [--perspective &lt;selection&gt;] [--criteria &lt;criteria&gt;...] [--method &lt;selection&gt;] [--capability &lt;selection&gt;...] [--provider &lt;selection&gt;] [--result &lt;destination-reference&gt;] [--dry-run]                                                                 |
| /decide            | Record a decision about a subject.                 | &lt;subject&gt; [--outcome &lt;text&gt;] [--reference &lt;reference&gt;...] [--revision &lt;revision&gt;] [--kind &lt;selection&gt;] [--perspective &lt;selection&gt;] [--criteria &lt;criteria&gt;...] [--method &lt;selection&gt;] [--capability &lt;selection&gt;...] [--provider &lt;selection&gt;] [--result &lt;destination-reference&gt;]                                                    |
| /execute           | Carry out work on a subject.                       | &lt;subject&gt; [--reference &lt;reference&gt;...] [--revision &lt;revision&gt;] [--kind &lt;selection&gt;] [--perspective &lt;selection&gt;] [--criteria &lt;criteria&gt;...] [--method &lt;selection&gt;] [--capability &lt;selection&gt;...] [--provider &lt;selection&gt;] [--result &lt;destination-reference&gt;] [--dry-run]                                                                 |
| /publish           | Persist a result to a destination.                 | &lt;subject&gt; --result &lt;destination-reference&gt; [--reference &lt;reference&gt;...] [--revision &lt;revision&gt;] [--kind &lt;selection&gt;] [--perspective &lt;selection&gt;] [--criteria &lt;criteria&gt;...] [--method &lt;selection&gt;] [--capability &lt;selection&gt;...] [--provider &lt;selection&gt;] [--confirm] [--dry-run]                                                       |
| /review            | Review a subject.                                  | &lt;subject&gt; [--reference &lt;reference&gt;...] [--revision &lt;revision&gt;] [--kind &lt;selection&gt;] [--perspective &lt;selection&gt;] [--criteria &lt;criteria&gt;...] [--method &lt;selection&gt;] [--capability &lt;selection&gt;...] [--provider &lt;selection&gt;] [--result &lt;destination-reference&gt;]                                                                             |
| /revise            | Revise a subject from one or more feedback inputs. | &lt;subject&gt; --feedback &lt;feedback&gt;... [--reference &lt;reference&gt;...] [--revision &lt;revision&gt;] [--kind &lt;selection&gt;] [--perspective &lt;selection&gt;] [--criteria &lt;criteria&gt;...] [--method &lt;selection&gt;] [--capability &lt;selection&gt;...] [--provider &lt;selection&gt;] [--result &lt;destination-reference&gt;] [--dry-run]                                  |
| /validate          | Validate a subject against one or more criteria.   | &lt;subject&gt; --criteria &lt;criteria&gt;... [--reference &lt;reference&gt;...] [--revision &lt;revision&gt;] [--kind &lt;selection&gt;] [--perspective &lt;selection&gt;] [--method &lt;selection&gt;] [--capability &lt;selection&gt;...] [--provider &lt;selection&gt;] [--result &lt;destination-reference&gt;]                                                                               |
| /workspace-abandon | Abandon an isolated workspace.                     | &lt;target&gt; --reason &lt;text&gt; [--revision &lt;revision&gt;] [--capability &lt;selection&gt;] [--provider &lt;selection&gt;] [--result &lt;destination-reference&gt;] [--snapshot] [--remove] [--remove-reference] [--confirm] [--dry-run]                                                                                                                                                    |
| /workspace-cleanup | Remove workspace artifacts.                        | &lt;target&gt;... [--capability &lt;selection&gt;] [--provider &lt;selection&gt;] [--remove-reference] [--prune] [--force] [--confirm] [--dry-run]                                                                                                                                                                                                                                                  |
| /workspace-create  | Create an isolated workspace.                      | &lt;target&gt; --source &lt;reference&gt; [--branch &lt;reference&gt;] [--capability &lt;selection&gt;] [--provider &lt;selection&gt;] [--dry-run]                                                                                                                                                                                                                                                  |
| /workspace-merge   | Merge workspace work into a destination.           | &lt;target&gt; --into &lt;destination-reference&gt; [--revision &lt;revision&gt;] [--criteria &lt;criteria&gt;...] [--method &lt;selection&gt;] [--capability &lt;selection&gt;...] [--provider &lt;selection&gt;] [--squash] [--remove] [--confirm] [--dry-run]                                                                                                                                    |

These are the current frontmatter argument hints. The files should still be treated
as authoritative because this handout questions how those concepts should be
factored.

### Utility commands

| Command                   | Purpose                                                       | Current argument hint                                                                                                                                                                 |
| ------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| /content-humanize         | Write or rewrite content in a human voice.                    | [text-or-file] [--tone formal, casual, or technical] [--in-place] [--audit]                                                                                                           |
| /content-news-add         | Curate recent bilingual news content.                         | [topic] [--path path] [--lang en,nl] [--days days] [--max max] [--slug slug]                                                                                                          |
| /content-travelguide-add  | Create a multilingual travel guide.                           | [topic] [subject] [--path path] [--lang en,nl] [--research-only] [--slug slug]                                                                                                        |
| /instructions-assimilate  | Merge selected instruction elements.                          | [target-instructions] [source-instructions] [--aspects aspects] [--percentage percent] [--interactive] [--dry-run]                                                                    |
| /instructions-consolidate | Remove redundant instruction files and standardize placement. | [--dry-run] [--path directory]                                                                                                                                                        |
| /instructions-init        | Create an AGENTS.md file.                                     | [directory] [--dry-run] [--recursive]                                                                                                                                                 |
| /instructions-simplify    | Simplify an AGENTS.md file.                                   | [instruction-file] [--dry-run] [--target-reduction percent]                                                                                                                           |
| /instructions-sync        | Synchronize AGENTS.md instructions.                           | [agents-file or --all] [--dry-run] [--update-workflows]                                                                                                                               |
| /reflect-extract          | Extract reusable lessons from a session.                      | [category] [--out-dir dir]                                                                                                                                                            |
| /reflect-to-instructions  | Apply session lessons to AGENTS.md.                           | [category] [--from-reflections] [--persist] [--dry-run] [--agents-file path]                                                                                                          |
| /reflect-to-skill         | Apply session lessons to a skill.                             | [category] [--from-reflections] [--persist] [--dry-run] [--force] [--output path]                                                                                                     |
| /skill-assimilate         | Merge selected elements from one skill into another.          | [target-skill] [source-skill] [--aspects aspects] [--percentage percent] [--interactive] [--dry-run]                                                                                  |
| /skill-create             | Create a skill from a document or URL.                        | [source] [--name name] [--dry-run]                                                                                                                                                    |
| /skill-decompose          | Split a multi-concern skill.                                  | [skill-file] [--into names] [--dry-run]                                                                                                                                               |
| /skill-evaluate           | Author or refresh skill evaluations.                          | [skill-file] [--count n]                                                                                                                                                              |
| /skill-extract            | Extract a skill from source code.                             | [skill-name] [source-path] [--update] [--interactive] [--dry-run]                                                                                                                     |
| /skill-optimize           | Optimize a skill for model-knowledge delta.                   | [skill-file or --all] [--model weakest] [--tier auto, aggressive, moderate, or conservative] [--dry-run] [--report-only]                                                              |
| /skill-simplify           | Condense a skill.                                             | [skill-file] [--dry-run] [--target-reduction percent]                                                                                                                                 |
| /slashcommand-assimilate  | Merge selected elements from one slash command into another.  | [target-command] [source-command] [--aspects aspects] [--percentage percent] [--interactive] [--dry-run]                                                                              |
| /slashcommand-create      | Create a slash command.                                       | [description] [--name name] [--interactive]                                                                                                                                           |
| /slashcommand-distill     | Reduce a command to a thin skill delegator.                   | [command-file] [--skill plugin] [--operation name] [--dry-run]                                                                                                                        |
| /slashcommand-simplify    | Simplify a command.                                           | [command-file] [--dry-run] [--target-reduction percent]                                                                                                                               |
| /version-bump             | Propagate versions and update changelogs.                     | [patch, minor, or major] [--type type] [--exact version] [--preid tag] [--dry-run] [--no-changelog] [--no-dependents] [--changelog-path file] [--git-base ref] [--include-types list] |

## Appendix C: proposal history in compact form

The conversation progressed through these questions:

1. What remains fixable in Xonovex, especially packages/command, composition,
   multi-axis behavior, and role lenses?
2. What should be done about eight identified concerns?
3. Why describe parameters in code rather than user language?
4. What exactly is a capability?
5. How would concerns one through seven work together?
6. What would the slash-command experience look like?
7. How could the same primitives express an SDLC?
8. Can criteria and perspectives be derived or auto-selected?
9. How does work continue when steps are not inline or span sessions?
10. How does a story split into smaller stories or subtasks?
11. Should perspectives have skills?
12. Can skills use independent lifecycle and functional-role taxonomies?

The resulting direction emphasized semantic clarity, composability, provenance,
durable references, explicit authority, and small skills. No final architecture was
accepted, and the receiving model is expected to revisit every decision.
