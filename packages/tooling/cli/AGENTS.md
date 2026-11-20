# Agent Command Line Interface

Release each CLI and its platform binaries together, and resolve execution environments through the sandbox interfaces.

- Move each CLI release line in lockstep: the version and five `optionalDependencies` of `agent-cli-go`, `agent-validate-go`, `agent-audit-go` or `agent-eval-go`, its five `<cli>-<platform>` packages, and its `<cli>-github` platform refs, with the matching `## <version>` `CHANGELOG.md` section from conventional commits, or `github-publish` fails or ships stale notes. One invocation moves one whole line, never a loop over `<project>:version-bump`; for `agent-validate-go` it is `npx moon run workspace-config:version-bump-lockstep -- --lockstep agent-validate-go,agent-validate-go-darwin-arm64,agent-validate-go-darwin-x64,agent-validate-go-linux-arm64,agent-validate-go-linux-x64,agent-validate-go-win32-x64 --type minor`, and the other CLIs follow the same pattern, which writes every version and every exact `@xonovex/*` reference as one transaction and generates each `CHANGELOG.md` section; preview with `--dry-run` first.
- Resolve isolators and provisioners through the `internal/sandbox` interfaces, explicit `Registry`, and method-agnostic `sandbox.EnforcePolicy`; only `internal/sandbox/plugins.DefaultRegistry` imports concrete plugins. Add a plugin as a package plus one `Register` call, and build isolated registries in tests.
- Resolve plugin and marketplace sources through the `internal/origin` registry, and approve content a repository supplies by the digest of exactly what the user reviewed, through the one trust gate.
- Declare each configuration key as data with its merge rule and the layers that may set it, and let a project file set only keys that cannot widen access.
- Keep model ids and provider routes in data the user supplies, never in code.
- Declare each isolator's egress transport as data, and never let a credential store reveal a secret.
- List every file a Go binary embeds as an input of each platform `go-build` task, and make a launcher's `npm-publish` wait for its platform packages.
- Give every unit test an `internal/system/systemtest.Fake` host with its own home directory and environment: only `internal/system` touches the machine, so a test never binds the agent configuration of the developer who runs it, and `forbidigo` fails a test that reads the host home directly.
