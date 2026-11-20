# Pi Onboarding

Applies the lifecycle of **configuring-harness-integrations** to Pi extensions and packages.

## Discover

- Probe `pi --version`, interactive/non-interactive mode, user settings, project settings, saved trust decision, local extensions, installed packages, skills, prompts, themes, and CLI `-e` resources.
- Inspect package sources and pins, npm/git/local identities, filters, enabled resources, extension imports, lifecycle handlers, custom tools, filesystem/network/secret access, and child-process behavior.
- Record that full user permissions and no built-in sandbox are the baseline.

## Preview

Show the exact settings/package/resource change, scope, source/version/digest, project trust consequence, enabled resource filters, extension code paths, npm install behavior, side effects, and idempotency.

Trying a package with a one-run extension flag is still code execution, not a safe preview. A preview reads and reviews the source without loading it.

## Apply and verify

Require project trust where applicable, then use native install/configuration or reviewed local paths. Verify list/config state, then probe trust rejection, event mapping, tool blocking, input mutation, result middleware, context filtering, compaction, cancellation, and output bounds.

## Lifecycle

- Disable resources with native configuration filters or remove only the owned path/package entry.
- Roll back settings and the pinned package or local extension digest; verify resources no longer load.
- Drift includes trust decision, package identity/ref/digest, enabled filters, extension code/imports, and child process behavior.

Non-interactive project trust defaults can skip or accept project resources without a prompt. Treat that mode and its explicit approval flags as part of the authorization record.
