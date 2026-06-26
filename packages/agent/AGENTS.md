# Agent

- Version bump: run `npx moon-version-bump <patch|minor|major>` from `packages/agent/agent-cli-go` — it bumps the version, lockstep-updates the five `agent-cli-go-<platform>` binary packages and the `agent-cli-go-github` refs, and generates the `CHANGELOG.md` entry. Don't hand-edit versions: that skips the changelog, and `github-publish` fails (or ships stale notes) without a matching `## <version>` section.
