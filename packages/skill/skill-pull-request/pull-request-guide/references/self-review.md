# self-review: Review Your Own Diff First

Read your full diff file by file as the reviewer would before assigning: remove debug logging, commented-out code, unrelated changes, and stray files; confirm the diff still matches the description and update it if it drifted. Keep the PR a draft until CI is green, then flip to ready and assign - reviewers should not start on a red CI.

## Related

[size-and-atomicity.md](./size-and-atomicity.md)
