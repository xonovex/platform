# Markdown Handoffs

Use one Markdown request when an operation needs cross-role traceability, several
evidence inputs, explicit capability criticality, or provider-native relationships.
Omit sections that do not apply; never invent a reference or revision.

```markdown
# Workflow request

## Subject

Reference: <inline content, path, or opaque native reference>
Revision: <exact native revision when available>

## Relationships

Parent: <optional opaque reference and revision>
Supersedes: <optional opaque reference and revision>
Related:

- <optional opaque reference and revision>

## Operation inputs

Method:
Perspectives:

- <review or validation lens>

Criteria:

- <binding or advisory criterion, labelled clearly>
  Feedback:
- <revision feedback>

Options:

- <decision option>

## Required capabilities

- <exact installed guide name or capability description>

## Preferred capabilities

- <exact installed guide name or capability description>

## Evidence

- Source: <opaque evidence reference>
  Subject revision: <revision the evidence assessed>
  Perspective: <originating role or lens>
  Criterion: <criterion evaluated>
  Outcome: <pass, fail, blocked, finding, or observation>
  Freshness: <native revision or observation time>
  Limitations: <known coverage limits>

## Effects

Mode: <inline, inspect, preview, or apply>
Expected destination revision: <when applicable>
Idempotency key: <when applicable>

## Constraints

- <scope, authority, retention, recovery, or timing constraint>
```

## Input Rules

- `--request` and shorthand arguments are mutually exclusive.
- An exact guide name requests that installed guide; a capability description permits
  the narrowest unambiguous installed match.
- Required capabilities block before effects when unavailable. Preferred capabilities
  may degrade only when a safe baseline remains useful.
- A provider-native subject keeps its opaque reference and native revision. Do not
  translate it into a local filename or invent a revision scheme.
- A protected operation requires a revision when the provider exposes one. When it
  does not, state that concurrency protection is unavailable.
- Evidence remains attached to the revision it assessed. A later subject revision
  makes that evidence stale until its criterion explicitly allows reuse.

## Result Handoff

Return only material sections, but preserve:

- operation status and exact subject identity;
- produced result and successor or destination revision when one exists;
- parent, supersedes, and related references supplied by the caller;
- criterion-level evidence with origin, freshness, and limitations;
- selected, missing required, and unavailable preferred capabilities;
- planned, applied, failed, and unknown effects;
- unresolved blockers and the safe retry boundary.

Separate independent reviewers or validators produce separate evidence entries. A
later Decide operation may compare or summarize them, but it never erases dissent,
changes criterion authority, or converts descriptive evidence into approval.
