# worktree-commit: Auto-Commit in Worktree with Plan Context

**Guideline:** Commit worktree changes with auto-generated messages enhanced by plan context to align with implementation objectives.

**Rationale:** Extending auto-commit with plan awareness generates contextual messages that document not just what changed, but how it aligns with the overall feature plan.

**Example:**

```bash
# In feature worktree: services-feature-auth-flow
cd ../services-feature-auth-flow

# Add implementation and tests
# services/auth/flow.ts (220 lines)
# services/auth/__tests__/flow.test.ts (180 lines)

# Worktree detects changes and generates message with plan context
git status --porcelain
# M  services/auth/flow.ts
# A  services/auth/__tests__/flow.test.ts

# Plan stored in git config was: "Implement email+password flow with TOTP"
# Auto-generated message:
# "feat: Implement email+password flow with TOTP
#
# Add LoginFlow component with email validation and TOTP verification.
# Test coverage: happy path, invalid email, expired TOTP."

git add -A
git commit -m "feat: Implement email+password flow with TOTP

Add LoginFlow component with email validation and TOTP verification.
Test coverage: happy path, invalid email, expired TOTP."

# Optional: push with CI skip
git push -o ci.skip origin HEAD:services/feature/auth-flow
```

**Techniques:**

- Detect worktree name from directory
- Check changes: `git status --porcelain`
- Analyze changed files: test, docs, source, config
- Read plan: `git config branch.<branch>.plan`
- Generate message using file patterns, diff stats, plan context
- Commit: `git add -A && git commit -m "<type>: <description>"`
- Optionally push: `git push -o ci.skip origin HEAD:<branch>`
