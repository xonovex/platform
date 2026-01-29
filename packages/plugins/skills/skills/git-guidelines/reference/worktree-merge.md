# worktree-merge: Merge Feature Worktree Back to Source

**Guideline:** Merge feature worktree branch back to source branch after validation passes to complete the feature development cycle.

**Rationale:** Completing the merge integrates changes to main codebase and enables cleanup of the temporary feature worktree while maintaining code quality.

**Example:**

```bash
# In feature worktree after validation passes
cd ../services-feature-auth-flow
npm run typecheck && npm run lint && npm run build && npm run test
# All pass ✓

# Merge to source worktree
MERGE_SOURCE=$(git config branch.services/feature/auth-flow.mergeBackTo)
# MERGE_SOURCE = master

cd ../services
git pull origin master
git merge services/feature/auth-flow --no-ff
# Merge commit created

# Validate merged result
npm run typecheck && npm run lint
# All pass ✓

# Push to remote
git push origin master

# Optional: Clean up feature worktree
cd ../services
git worktree remove ../services-feature-auth-flow
git branch -d services/feature/auth-flow
```

**Techniques:**

- Verify validation passed
- Read source: `git config branch.<branch>.mergeBackTo`
- Navigate to source worktree directory
- Update source: `git pull origin <source-branch>`
- Merge: `git merge <feature-branch> --no-ff`
- Handle conflicts (auto-resolve simple, manual review complex)
- Validate merged result: `npm run typecheck && npm run lint`
- Push: `git push origin <source-branch>`
