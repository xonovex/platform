# Markdown Handoffs

A handoff exists only at a cold boundary — a session or role ends and context is not
retained. Operations that chain inside one warm session pass results directly and
need none of this. A handoff is placement into a native system the receiving side
already reads: a file, ticket, comment, or pull request.

Use these headings. Omit any that does not apply; never invent a reference, revision,
or anchor.

```markdown
## Subject

Reference: <inline content, path, or opaque native reference>
Revision: <exact native revision when the provider exposes one>

## What was done

<what changed, in the receiving role's terms>

## Decisions

- What: <the choice made>
  Why: <the reason it was made>
  Where: <file:line or symbol anchor>

## References and links

- <path, opaque native reference, or URL>

## Open issues

- <unresolved question, blocker, or known limitation>
```

## Rules

- Anchor every decision to a code location: `file:line` or a symbol name. A decision
  a later session cannot locate is not a handoff, it is a note.
- Links are deferred context. Resolve them on demand through the skill that matches
  the linked service, and report anything unresolvable instead of dropping it.
- A revision is copied from the provider, never invented. When the provider exposes
  none, say so rather than substituting a local scheme.
- Keep provider-native references opaque and pass them to the selected provider
  capability instead of parsing their shape.
- Authority, evidence, and fetched-content rules live in
  [governance.md](governance.md).

## Result Handoffs

A returned result carries the same headings plus the operation's status and its
effects. Separate reviewers or validators produce separate entries; a later Decide
operation may summarize them but never erases dissent.
