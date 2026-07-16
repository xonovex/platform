# Findings schema — the review's data contract

The single canonical shape shared by every stage (analyze → refine → post → resolve). Hold in session between same-session stages, or persist to JSON to cross sessions.

## Shape

```jsonc
{
  "summary": "Markdown. Lead with positives; number the blocking points so cross-links can attach.",
  "findings": [
    {
      "path": "packages/.../module.ts", // repo-relative path in the new file version
      "line": 420, // new-file line number present in the diff
      "lineType": "ADDED", // ADDED | CONTEXT
      "label": "issue", // praise | nitpick | suggestion | issue | question | thought | chore
      "decoration": "blocking", // blocking | non-blocking | if-minor
      "blocking": true, // must-fix-before-merge; must agree with decoration
      "body": "Self-contained markdown: problem, why, suggested fix.",
      "status": "new", // new | recurring — only when comparing against prior findings
      "commentId": 101, // written by the post stage; carried on a recurring finding to keep threads linked
    },
  ],
}
```

`decoration` and `blocking` must agree — dropping `blocking` flips the decoration off `(blocking)`. Bodies are self-contained (no "see the other comment") since findings post/edit/resolve independently.

Anchors are **new-file** line numbers (the side a host inline comment attaches to), never old-file or absolute: a `+` line is `lineType: ADDED`, an unchanged in-hunk line is `CONTEXT`; a `path` / `line` that is not a real ADDED/CONTEXT diff line orphans when posted.

## Building the JSON

Build/edit with a serializer (`python3` + `json`), never hand-escape — bodies carry backticks, quotes, newlines. Read, mutate, re-serialize the whole document; do not splice strings.
