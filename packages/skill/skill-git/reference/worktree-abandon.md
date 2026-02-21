# worktree-abandon: Document and Abandon Feature

**Guideline:** Document why feature is abandoned and clean up worktree to preserve learnings for future decisions.

**Rationale:** Recording abandonment reasons and learnings prevents repeated failed attempts and provides valuable context for future development decisions.

**Example:**

```bash
# Document abandonment in insights/abandoned-features/
cat > insights/abandoned-features/auth-redesign-2026-01.md << 'EOF'
## Feature: Authentication UI Redesign

**Goal:** Redesign login form with multi-step experience

**Work Completed:**
- Step 1: Email validation form (80% complete)
- Step 2: Password reset flow (30% complete)
- Custom TOTP implementation (blocked by deps)

**Reason for Abandonment:**
Third-party auth provider API changed requirements. Multi-step approach conflicts with new spec.

**Alternative:** Use updated provider's OAuth 2.0 flow, leverage built-in MFA.

**Learnings:** Don't commit to custom flow when third-party alternatives exist. Test assumptions with provider early.

**Time Invested:** 12 hours across 5 commits
**Date:** 2026-01-15
**Branch:** auth/redesign-feature
EOF

# Clean up worktree and branch
git worktree remove ../services-feature-auth-redesign
git branch -D auth/redesign-feature

# Push archived branch to preserve history
git tag abandoned/auth-redesign-2026-01 origin/auth/redesign-feature
git push origin abandoned/auth-redesign-2026-01
```

**Techniques:**

- Verify in feature worktree
- Gather context: feature name, plan, commits, work completed
- Prompt for abandonment reason
- Document in `insights/abandoned-features/<feature>-<date>.md`
- Clean up: `git worktree remove <path>` and `git branch -D <branch>`
- Clean git config entries
