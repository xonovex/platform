# Story Decomposition

## Choose the result kind

Use Create with method `story-splitting` and output kind `child-story-set` when each
child should independently deliver user value. Use output kind `subtask-set` when the
items are engineering steps beneath one valuable story.

The [story-split request](../assets/examples/story-split-request.json) pins the parent
story and requests a vertical child-story set without publishing tracker items.

## Child stories

- [ ] Preserve the exact parent reference and revision.
- [ ] Give drafts stable local labels before provider IDs exist.
- [ ] Produce a valuable, testable outcome and acceptance criteria per child.
- [ ] Map every binding parent criterion to one or more children.
- [ ] Record dependencies, recommended order, residual scope, gaps, and overlaps.
- [ ] Prefer vertical value slices over database/API/UI layers.
- [ ] Review, revise, and decide on the inline draft set as separate operations.
- [ ] Use Publish to create provider-native children and links.
- [ ] Validate published coverage against the unchanged parent revision.

## Implementation subtasks

Subtasks may be horizontal, need not deliver independent value, and roll up to the
parent story's acceptance criteria. Do not call them child stories or require an
unrelated operation verb.

## Parallel continuation

Give each published child an exact native work binding and administrative child
record. Parallel sessions pin child, repository, and parent revisions independently.
The
[continuation request](../assets/examples/child-story-continuation-request.json)
demonstrates that baseline. Before completing the parent, validate coverage,
completion or explicit deferral, integration evidence, cross-child conflicts, and
unchanged parent revision.
