# Agent

CLI tools (`agent-cli-go`) and the Kubernetes operator (`agent-operator-go`).

## Version bump — agent-cli-go

`agent-cli-go` versions independently (currently `0.1.x`), in lockstep across its own package and the per-platform binary packages. Bump the **same** version in:

- `agent-cli-go/package.json` — `version` and the 5 `optionalDependencies` (`@xonovex/agent-cli-go-<platform>`);
- each of the 5 platform packages — `agent-cli-go-{linux-x64,linux-arm64,darwin-x64,darwin-arm64,win32-x64}/package.json` `version`;
- `agent-cli-go-github/package.json` — its 5 platform `optionalDependencies` (leave its own `version` at `0.0.0`).

Then run `npm install` to refresh `package-lock.json`. Publish via a `version packages` PR (root `AGENTS.md` Release rule): `release.yml` runs `agent-cli-go:ci-publish` = npm publish (`@xonovex/agent-cli-go` + the 5 platform packages) **and** the `agent-cli-go-v<version>` GitHub release (binaries rebuilt in CI from `main`). `moon-npm-publish` is version-gated, so a publish only happens for a version not already on npm — the bump is what triggers it.

## Version bump — agent-operator-go

The operator ships as a container image, not an npm package. Bump `agent-operator-go-docker/package.json` `version`; publish with `npx moon run agent-operator-go-docker:docker-publish`.
