# Kiro Onboarding

Phases follow **configuring-harness-integrations**; this file holds the Kiro specifics.

## Discover

- Probe `kiro-cli --version`, active IDE/CLI surface, v3 mode where relevant, workspace roots, and hook schema.
- Inspect every `.kiro/hooks/*.json`, the unified Agent Hooks panel, enabled flags, duplicate matchers, command paths, agent prompts, timeouts, and root ownership.
- In multi-root workspaces, preserve the root that owns each hook and test file events only within that root.

## Preview

Preview a complete standalone hook file or a focused edit to an owned file. Include schema version, name, description, trigger, action type, command/prompt, enabled state, prompt/tool content exposure, expected stdout/stderr, exit behavior, test cases, and deletion.

When Kiro generates a hook from natural language, treat the result as untrusted proposed configuration. Review the command and prompt, narrow matchers, restore a finite timeout, and compare the exact JSON before Save Hook.

## Apply and verify

Verify the hook appears under the correct workspace root and remains enabled. Exercise success, exit `2`, other non-zero exit, malformed input, narrow and broad matchers, command context, and agent-action budget behavior. Test command logic manually before relying on the event.

## Probe runbook

Run only with a sign-in that already exists; acquire no credential. All steps are read-only until the matrix is edited.

1. Version: `kiro-cli --version` (or the desktop surface's about dialog). If no binary is on PATH, record "not installed" and leave the capability matrix at candidate status.
2. Credential presence: an AWS Builder ID sign-in must already exist on the host (do not create one). Without it, record a version-only observation: "CLI version observed; contract not exercised, so hook semantics remain documentation-derived."
3. Guard contract: with a credentialed install, execute the deterministic guard contract locally (JSON event on stdin, exit 0 allow / exit 2 deny) using walking-skeleton guard assets; native hook registration stays unexercised.
4. Rerun trigger: installed version differs from the observed one, hook schema change, or 90 days.

## Lifecycle

- Disable with the native enabled flag or panel toggle; verify the hook no longer fires.
- Roll back by restoring the verified prior file or deleting the newly owned file, then reload and probe.
- A change to command, prompt, matcher, timeout, action type, or workspace root is an update.
- Drift includes file digest, enabled flag, surface, and workspace root.
