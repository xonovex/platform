# worktree-validate: Pre-Merge Validation Checkpoint

**Guideline:** Validate feature worktree before merge by running typecheck, lint, build, and tests to ensure stability.

**Rationale:** Running validation before merge catches issues early, ensures quality standards are met, and verifies the feature aligns with plan criteria when configured.

**Example:**

```bash
# In feature worktree
cd ../services-feature-auth-flow

# Run validation checks
npm run typecheck
# ✓ PASS (0 errors)

npm run lint
# ✓ PASS (0 errors)

npm run build
# ✓ PASS (bundle 245 KB)

npm run test
# ✓ PASS (42 tests, 0 failures, 95% coverage)

# Check plan criteria if configured
git config branch.services/feature/auth-flow.plan
# "Implement email+password flow with TOTP and backup codes"

# Verify plan success criteria met:
# ✓ Email validation (part of LoginFlow)
# ✓ TOTP verification (part of LoginFlow)
# ✓ Backup codes support (in VerificationFlow)

# Overall status: READY TO MERGE
# All validation passed, all plan criteria met
```

**Techniques:**

- Verify in feature worktree
- Check commits complete
- Run validation in sequence: `npm run typecheck`, `npm run lint`, `npm run build`, `npm run test`
- Check plan criteria if associated
- Report results
