# Findings schema — the review's data contract

The single canonical shape for review findings, shared by every stage that produces, refines, or publishes them (analyze → refine → post → resolve). Hold it in session between same-session stages, or persist it to a JSON file to cross sessions or hand-edit.

## Shape

```jsonc
{
  "summary": "Markdown body. Lead with positives. Number the priority (blocking) points so cross-links can attach.",
  "findings": [
    {
      "path": "packages/.../module.ts", // repo-relative path in the new file version
      "line": 420, // new-file line number present in the diff
      "lineType": "ADDED", // ADDED | CONTEXT
      "label": "issue", // praise | nitpick | suggestion | issue | question | thought | chore
      "decoration": "blocking", // blocking | non-blocking | if-minor
      "blocking": true, // must-fix-before-merge
      "body": "Self-contained markdown: the problem, why, and a suggested fix.",
      "status": "new", // new | recurring — only when comparing against prior findings
      "commentId": 101, // written back by the post stage; carried on a recurring finding
    },
  ],
}
```

- `summary` is required; lead it with what works, then number the blocking points.
- Every finding needs `path`, `line`, `lineType`, `label`, `decoration`, `blocking`, and `body`.
- `status` is set only when comparing against a prior findings set; `commentId` is written by the post stage and carried forward on a recurring finding so threads stay linked.
- `label` is a Conventional Comments label; `decoration` and `blocking` must agree — dropping `blocking` flips the decoration off `(blocking)`. See [conventional-comments.md](conventional-comments.md).
- Bodies are self-contained — no "see the other comment" — since findings may be posted, edited, or resolved independently.

## Anchoring to new-file lines

Anchors are **new-file** line numbers — the side a host's inline comment attaches to — never old-file or absolute:

- Parse each hunk header `@@ -a,b +c,d @@` and walk the hunk: `+` lines and context lines advance the new-file counter; `-` lines do not.
- An added (`+`) line is `lineType: ADDED`; an unchanged in-hunk line is `lineType: CONTEXT`.
- A finding whose `path` / `line` is not a real `ADDED` / `CONTEXT` diff line orphans when posted — re-anchor it to one in a changed hunk.

## Building and editing the JSON

- Build and edit findings with a serializer (`python3` + `json`), never hand-escape — bodies carry backticks, quotes, and newlines that break hand-patched JSON.
- Read, mutate, and re-serialize the whole document; do not splice strings into it.
