# develop-abandon: Abandon a Development Assignment

## Core workflow

1. Resolve the assignment's Planning and current Development/workspace references and revisions.
2. Stop or cancel active execution using the declared kill behavior, then observe whether child work or provider side effects remain in flight.
3. Capture the reason, completed and incomplete scope, partial changes, validation state, native evidence, unresolved side effects, secrets or temporary resources requiring cleanup, and useful learning.
4. Apply only explicitly authorized cleanup through the workspace provider. Preserve the workspace or snapshot when removal would destroy needed evidence.
5. Publish an abandoned Development result with exact references/revisions, actor or executor origin, replacement/retry eligibility, cleanup outcome, and limitations.

Abandonment is a terminal state for that assignment revision. A new attempt uses a new assignment or native revision; it does not overwrite the abandoned result.
