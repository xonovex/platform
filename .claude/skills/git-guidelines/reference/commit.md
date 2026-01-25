# commit: Auto-Commit with Smart Messages

**Guideline:** Commit with conventional format `<type>: <description>` by auto-detecting type from file changes and generating descriptive messages.

**Rationale:** Auto-generating messages from file changes maintains consistent commit style with minimal effort while accurately reflecting what changed in each commit.

**Example:**

```bash
# Changes in 3 test files
$ git commit
test: add validation tests for user registration

# Changes in documentation
$ git commit
docs: update authentication flow diagram

# Bug fix in service layer
$ git commit
fix: handle null pointer in user service

# New feature implementation
$ git commit
feat: add OAuth2 provider integration
```

**Techniques:**
- Detect changes via `git status --porcelain` and `git diff --stat`
- Auto-detect type from file patterns: test (*.test.ts), docs (*.md), ci, chore (package.json), feat (new), fix (small), refactor (large)
- Generate message from file patterns and diff stats
- Commit with `git add -A && git commit -m "<type>: <description>"`
- Optionally push with `git push -o ci.skip origin HEAD:<branch>`
