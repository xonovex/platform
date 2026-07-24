# Findings schema — the review's data contract

The canonical inline shape returned by Review analysis and refinement. A later Publish
or Execute operation may use this shape as its subject, but Review never persists or
delivers it.

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

`decoration` and `blocking` must agree — dropping `blocking` flips the decoration off `(blocking)`. Bodies are self-contained because later delivery may create independent threads.

Anchors are **new-file** line numbers (the side a host inline comment attaches to), never old-file or absolute: a `+` line is `lineType: ADDED`, an unchanged in-hunk line is `CONTEXT`; a `path` / `line` that is not a real ADDED/CONTEXT diff line orphans when posted.

Return valid JSON inline. Use a serializer when one is already available in the
calling environment; never hand-splice escaped strings.
