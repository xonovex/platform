# Distill Sources: Per-Source Mining Recipes

The invariant procedure for any log store: find it, shortlist by mtime against the work's
timeline, establish the format with a `head`, then sample - never read wholesale.

## Git history

```bash
git merge-base <base> <branch>                      # the starting state the suite will target
git log --oneline <merge-base>..<base> | wc -l      # base drift; zero means the branch diff IS the work
git log --oneline <base>..<branch>                  # read every subject; count the commits
git show --stat --format='%s%n%b' <sha>             # per commit: body + file list
```

Commit bodies on a well-run branch are themselves a distilled record; read them all before any
diff. The subjects give the tier partition for the reconstruction fan-out. Representative hunks
(`git show <sha> -- <path>`) supply the before/after examples and 10-30 line snippets the skills
need inline. A change that appears, disappears, and reappears differently is usually where a rule
was learned - reconstruct the rule, not just the final state.

Traps: content that exists only at a commit (find candidates with
`git log --diff-filter=D --name-only` and "folded/replaced" language in bodies); the final diff
hides the ordering, reverts, and corrections that carry the lessons; verify base-branch existence
with `git cat-file -e <base>:<path>` rather than memory; commit ids referenced by the suite
survive only while a ref reaches them, so suggest a tag when archaeology value is high - but the
suite must never depend on it.

## Claude Code session logs

- `~/.claude/projects/<encoded-cwd>/*.jsonl` - one file per session; the encoding replaces `/`
  with `-`. Sibling `<uuid>/` directories hold that session's subagent transcripts (delegated
  runs live here; a small supervisor session can hide large subagent work). Watch for
  near-duplicate project directories from old checkout paths or worktrees; merge on the timeline.
- Format: JSONL, one event per line, `type` (`user`, `assistant`, ...) and `message.content` as a
  string or structured array. Files run to tens of MB.

```bash
ls -lat ~/.claude/projects/<dir>/ | head -30    # shortlist by mtime
# opening prompts / operator steering (maps sessions to phases):
jq -r 'select(.type=="user" and (.message.content|type=="string")) | .message.content' f.jsonl | head -c 3000
# failure vocabulary across a session:
grep -oiE 'revert|failed|blocker|flaky|attempt|wrong|regression|rollback|<domain terms>' f.jsonl | sort | uniq -c | sort -rn
```

For heavier slicing (assistant text near keywords, attempt sequences), drop to `python3` over the
JSONL. These logs are uniquely good for: actual execution order and parallelism of a delegated
run (subagent metadata), operator steering that changed direction, and supervisor
rejections/re-validations that never appear in git.

## Other sources

- **Other agent harnesses**: each keeps session state in its own store, usually a dot-directory
  in `$HOME` or the workspace. If unknown, search:
  `grep -rl "<distinctive phrase from the work>" ~/.<harness> 2>/dev/null`, or list
  dot-directories by mtime in the work's window. Structured formats get jq/python keyed on
  role/type fields; plain text gets `grep -n` with context. When work spans harnesses, mine each
  and merge on the timeline.
- **The plan record**: retrospectives, blockers, decision registers, merge briefs. Highest
  density; read first, personally. If the record does not travel with the repository, everything
  durable in it must end up inside the suite.
- **Non-agent records**: CI runs (failed pipelines, retried jobs), PR review threads (pushback
  and "looks wrong but is not" explanations are ready-made pitfall entries), issue trackers and
  team chat (decisions made outside the repo, human gates no repo artifact can show).
