# worktree-create: Create Feature Worktree

**Guideline:** Create worktree with feature branch for isolated development without affecting current worktree.

**Rationale:** Enables parallel feature development in separate directories, avoiding context switching and stash complications that plague single-worktree workflows.

**Example:**

```bash
# From services directory with master branch
cd /home/user/project/services

# Create worktree for auth-flow feature
# Creates: ../services-feature-auth-flow directory
# Creates: services/feature/auth-flow branch (off master)
git worktree add ../services-feature-auth-flow -b services/feature/auth-flow master

# Config stores the merge target
git config branch.services/feature/auth-flow.mergeBackTo master

# Navigate to new worktree and start development
cd ../services-feature-auth-flow
npm install
npm run typecheck

# Commit work independently without affecting main worktree
git add services/auth/login.ts
git commit -m "feat: Add LoginFlow component"

# Later, merge back to source
cd ../services
git merge services/feature/auth-flow --no-ff
git push origin master
```

**Techniques:**
- Detect worktree name from directory basename
- Get source branch from arg or `git branch --show-current`
- Sanitize feature name to kebab-case format
- Create directory `../<worktree>-feature-<name>` and branch `<worktree>/feature/<name>`
- Create worktree: `git worktree add <dir> -b <branch> <source-branch>`
- Store merge target in git config: `git config branch.<branch>.mergeBackTo <source-branch>`
