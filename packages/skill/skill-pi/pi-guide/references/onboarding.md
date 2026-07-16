# Pi Onboarding

## Discover

- Probe `pi --version`, interactive/non-interactive mode, user settings, project settings, saved trust decision, local extensions, installed packages, skills, prompts, themes, and CLI `-e` resources.
- Inspect package sources and pins, npm/git/local identities, filters, enabled resources, extension imports, lifecycle handlers, custom tools, filesystem/network/secret access, and child-process behavior.
- Record that full user permissions and no built-in sandbox are the baseline.

## Preview

Show the exact settings/package/resource change, scope, source/version/digest, project trust consequence, enabled resource filters, extension code paths, npm install behavior, permissions, files, network, secrets, model/context content, side effects, concurrency, idempotency, evidence, verification, disable, removal, update, and rollback.

Trying a package with a one-run extension flag is still code execution, not a safe preview. A preview reads and reviews the source without loading it.

## Apply and verify

After explicit consent and project trust where applicable, use native install/configuration or reviewed local paths. Verify list/config state and then probe trust rejection, event mapping, tool blocking, input mutation, result middleware, concurrent siblings, context filtering, compaction, cancellation, and output bounds.

## Lifecycle

- Disable resources with native configuration filters or remove only the owned path/package entry.
- Roll back settings and the pinned package or local extension digest; verify resources no longer load.
- Update pinned sources only through a new preview and permission/trust review.
- Drift includes trust decision, settings precedence, package identity/ref/digest, enabled filters, extension code/imports, full permissions, child process behavior, context/data access, and runtime version.

Non-interactive project trust defaults can skip or accept project resources without a prompt. Treat that mode and its explicit approval flags as part of the authorization record.
