# @xonovex/agent-cli-go

## 0.3.0

### Breaking changes

These changes need an action when you upgrade from 0.2.0.

- `-c/--config <file>` is removed. agent-cli reads the `run` and `providers` sections of three `agentic.json` layers: the managed file (`/etc/agentic/agentic.json` on Linux), the user file `$XDG_CONFIG_HOME/agentic/agentic.json`, and the nearest project `agentic.json` or the file that `--project <file>` names. A project file takes effect only after you approve its digest with `agent-cli providers approve` or `--approve-config-digest <digest>`.
- The 0.2.0 configuration file `~/.config/sandboxed-claude/config` is no longer read, and no migration command exists. Write its settings by hand in the `run` section of the user `agentic.json`.
- No provider ships. The `gemini`, `gemini-claude`, `glm` and `gpt5-codex` presets of Claude Code and the `gemini` preset of OpenCode are removed. Define each provider in the `providers` section of an `agentic.json`, or as `providers.d/<name>.json` beside it. `-p <name>` fails for a name that no layer defines.
- The `docker` isolation has no default image. Set `run.image` in a user or managed `agentic.json`, or pass `--image`. A run without an image stops before it starts. The new `openshell` isolation also needs an image.

### Minor Changes

- [`23fd761`](https://github.com/xonovex/platform/commit/23fd761c55cc05f1b6bcee7f1001c67fa7558c2a) [@Deorder](https://github.com/Deorder)! - remove the overlay of a herdr-core run that waited for the agent
- [`5574c82`](https://github.com/xonovex/platform/commit/5574c82b56a57e77d4b0c23e471124d7d683b69c) [@Deorder](https://github.com/Deorder)! - run the agent in a herdr-core terminal and choose the herdr wrapper automatically
- [`417386b`](https://github.com/xonovex/platform/commit/417386b39a472cb772a17ea38fab78f0c1b6ed9f) [@Deorder](https://github.com/Deorder)! - **Breaking:** ship no container image
- [`18187d3`](https://github.com/xonovex/platform/commit/18187d381ab22e55602b7f6f384833a669cce2fd) [@Deorder](https://github.com/Deorder)! - **Breaking:** name a docker container and say how to stop a sandbox
- [`fe12d0f`](https://github.com/xonovex/platform/commit/fe12d0f04e248b440c771d6c3c5c8b1c2d6654d0) [@Deorder](https://github.com/Deorder)! - **Breaking:** attach a terminal only to an interactive run
- [`0449ad9`](https://github.com/xonovex/platform/commit/0449ad9075a0498b111bf86394f9d18ba9966b4e) [@Deorder](https://github.com/Deorder)! - **Breaking:** plan the provider environment before the credential is read
- [`4e6891a`](https://github.com/xonovex/platform/commit/4e6891a763de57188d78372d710fee77669dc611) [@Deorder](https://github.com/Deorder)! - **Breaking:** keep the run records and the plugin store in agentic
- [`d8d653a`](https://github.com/xonovex/platform/commit/d8d653aa00545928992c7932cf14fcda339ce874) [@Deorder](https://github.com/Deorder)! - refuse a managed agentic.json others can write
- [`ca8c611`](https://github.com/xonovex/platform/commit/ca8c611ae8d096fcfc0083a0582ef02ae3ac7914) [@Deorder](https://github.com/Deorder)! - **Breaking:** install a plugin through the shared trust gate
- [`d805682`](https://github.com/xonovex/platform/commit/d8056828d1851e614a34e2e87095e8d01f46f364) [@Deorder](https://github.com/Deorder)! - apply the harness overrides of every layer
- [`be45f17`](https://github.com/xonovex/platform/commit/be45f173ad61000a9213a8b0a76fffbc470dcc29) [@Deorder](https://github.com/Deorder)! - **Breaking:** read the run and providers sections of agentic.json
- [`e9246fc`](https://github.com/xonovex/platform/commit/e9246fc8ded978bb84005b7fa20aea55b7bd4d47) [@Deorder](https://github.com/Deorder)! - own the run record schema under the shared schema namespace
- [`f29294f`](https://github.com/xonovex/platform/commit/f29294fbfcd170ca9e7747bf0cc70fa27cdba267) [@Deorder](https://github.com/Deorder)! - move the isolation, provision, egress and credential leaves and the OS adapter out of agent-cli-go
- [`d1afd11`](https://github.com/xonovex/platform/commit/d1afd114d2343fe7be3168e07f98367efa1be061) [@Deorder](https://github.com/Deorder)! - plan an openshell run without a host bind
- [`6cb8e80`](https://github.com/xonovex/platform/commit/6cb8e80293f898138b4c9e214f52ea4f87b8b0c6) [@Deorder](https://github.com/Deorder)! - move the run kernel out of agent-cli-go
- [`addcc74`](https://github.com/xonovex/platform/commit/addcc74d36a0d584cb4232b9b1fa8cf0ef90e6fb) [@Deorder](https://github.com/Deorder)! - plan the agent process once in a portable spec
- [`02c037d`](https://github.com/xonovex/platform/commit/02c037d7e8fc3b3bfb4d3f698cf829a13c54822c) [@Deorder](https://github.com/Deorder)! - ship the harness facts as embedded data
- [`fe2b7c8`](https://github.com/xonovex/platform/commit/fe2b7c84ee48ee5165b8de9c727917ffcfdc2dd8) [@Deorder](https://github.com/Deorder)! - own the network method parsing and checks
- [`f75b993`](https://github.com/xonovex/platform/commit/f75b993f6e6a6e4d27eb3a79c5337e83cccab214) [@Deorder](https://github.com/Deorder)! - define providers and models in user data with a digest approval gate for project files
- [`fcbe4ae`](https://github.com/xonovex/platform/commit/fcbe4aec1ec028502ec7fc10aba8c87dec5279c9) [@Deorder](https://github.com/Deorder)! - ask for plugin digest approval on a terminal and publish platform packages first
- [`ebefe47`](https://github.com/xonovex/platform/commit/ebefe4717737f7d827cd9d4658a608c266c360b9) [@Deorder](https://github.com/Deorder)! - layer configuration, record runs, add preflight, schema and herdr
- [`90a20d8`](https://github.com/xonovex/platform/commit/90a20d809598e0d5514873108329492404372273) [@Deorder](https://github.com/Deorder)! - add seatbelt, Docker Sandboxes and OpenShell isolations, managed egress and credential stores
- [`30580f8`](https://github.com/xonovex/platform/commit/30580f816321577b379f255e6f1f94af95d78e39) [@Deorder](https://github.com/Deorder)! - install approved plugins and materialize a per-run overlay
- [`8123747`](https://github.com/xonovex/platform/commit/812374732bc2ff5d23f6ed3e176dcee9197ef825) [@Deorder](https://github.com/Deorder)! - launch Codex, Pi, Copilot and Kiro from declared agent facts
- [`c920dae`](https://github.com/xonovex/platform/commit/c920dae3bede745dba61c5c435af878705cc15b8) [@Deorder](https://github.com/Deorder)! - split run into a pure planner and an executor over ports
- [`0065a3a`](https://github.com/xonovex/platform/commit/0065a3aa0128b3097ce05da0e368185b14851d84) [@Deorder](https://github.com/Deorder)! - pass the logger and the environment reader as explicit inputs
- [`f725ddc`](https://github.com/xonovex/platform/commit/f725ddc88f3794ff54ed55a739d52385230fbfba) [@Deorder](https://github.com/Deorder)! - inject the package layout and signal numbers into the launcher
- [`891a19f`](https://github.com/xonovex/platform/commit/891a19f05bbcc348716d573a86fe7d3b07f554c7) [@Deorder](https://github.com/Deorder)! - organize packages and group agent plugins
- Updated dependency `@xonovex/agent-cli-go-linux-x64` to `0.3.0`
- Updated dependency `@xonovex/agent-cli-go-linux-arm64` to `0.3.0`
- Updated dependency `@xonovex/agent-cli-go-darwin-x64` to `0.3.0`
- Updated dependency `@xonovex/agent-cli-go-darwin-arm64` to `0.3.0`
- Updated dependency `@xonovex/agent-cli-go-win32-x64` to `0.3.0`

## 0.2.0

### Minor Changes

- [`a102cfb`](https://github.com/xonovex/platform/commit/a102cfb6484fe253cf9f84e40efcc057acf89b88) [@Deorder](https://github.com/Deorder)! - close the migration with docs, assets and the registrations
- [`ac40476`](https://github.com/xonovex/platform/commit/ac40476f6bbc1bfe5a4c9b489db345b0505585e8) [@Deorder](https://github.com/Deorder)! - land the shared libraries and agent tooling
- [`81123cc`](https://github.com/xonovex/platform/commit/81123cc9c5100d01102d8abf7bedbd49262a9397) [@Deorder](https://github.com/Deorder)! - make Go nix flakes multi-platform
- Updated dependency `@xonovex/agent-cli-go-linux-x64` to `0.2.0`
- Updated dependency `@xonovex/agent-cli-go-linux-arm64` to `0.2.0`
- Updated dependency `@xonovex/agent-cli-go-darwin-x64` to `0.2.0`
- Updated dependency `@xonovex/agent-cli-go-darwin-arm64` to `0.2.0`
- Updated dependency `@xonovex/agent-cli-go-win32-x64` to `0.2.0`

## 0.1.31

### Patch Changes

- [`ec791d7`](https://github.com/xonovex/platform/commit/ec791d7511bb05ab27eba0ad3c2140d3a1fe81bb) [@Deorder](https://github.com/Deorder)! - track CLI and shared-lib Go sources in each platform binary's build-cache inputs so a release recompiles the binary instead of republishing a stale cached one (#31)
- Updated dependency `@xonovex/agent-cli-go-linux-arm64` to `0.1.31`
- Updated dependency `@xonovex/agent-cli-go-linux-x64` to `0.1.31`
- Updated dependency `@xonovex/agent-cli-go-darwin-arm64` to `0.1.31`
- Updated dependency `@xonovex/agent-cli-go-darwin-x64` to `0.1.31`
- Updated dependency `@xonovex/agent-cli-go-win32-x64` to `0.1.31`

## 0.1.30

### Patch Changes

- [`11dea8c`](https://github.com/xonovex/platform/commit/11dea8ccde78ba31bfa330d73c97b49d2d7cd84f) [@Deorder](https://github.com/Deorder)! - reorganize CLI/shared into orthogonal confinement axes
- [`01a38ee`](https://github.com/xonovex/platform/commit/01a38eec436112efb0d80836ecc187569b188581) [@Deorder](https://github.com/Deorder)! - prefix tag tasks per language and aggregate via ci-check
- [`12e5a13`](https://github.com/xonovex/platform/commit/12e5a1382afb4f4139c12e29ff05f9ea6724cac5) [@Deorder](https://github.com/Deorder)! - add nix provisioner and operator image realizer, drop compose and docker-agent
- [`2d30102`](https://github.com/xonovex/platform/commit/2d3010210913b74edca492c6ea356e7f885ddd48) [@Deorder](https://github.com/Deorder)! - split isolation/provisioning/network axes and hoist shared helpers
- [`fa6e337`](https://github.com/xonovex/platform/commit/fa6e3376197ec32f11e849decad1d12f05172783) [@Deorder](https://github.com/Deorder)! - add missing 0.1.24 CHANGELOG entry; publish version-specific notes
- [`ffa01e5`](https://github.com/xonovex/platform/commit/ffa01e5da8a6d5eb551945247629161840d0c415) [@Deorder](https://github.com/Deorder)! - expose named go devShell in Go project flakes
- [`cdb0fd2`](https://github.com/xonovex/platform/commit/cdb0fd21ebcee694fbaf6b88ebe5b56a4f7d7938) [@Deorder](https://github.com/Deorder)! - nixflake sandbox tier + pinned-toolchain deny-default policy (#25)
- [`52a05be`](https://github.com/xonovex/platform/commit/52a05be3f4378a729ddf9fab5fb6d38f9e6d08f1) [@Deorder](https://github.com/Deorder)! - go.nix carries go_1_26; all Go projects on 1.26; docker-agent composes from nix/
- [`e7f1cb6`](https://github.com/xonovex/platform/commit/e7f1cb61dba7072025c710c242cc6540e43af58e) [@Deorder](https://github.com/Deorder)! - compose flakes from per-tool nix/\*.nix modules
- Updated dependency `@xonovex/agent-cli-go-linux-arm64` to `0.1.30`
- Updated dependency `@xonovex/agent-cli-go-linux-x64` to `0.1.30`
- Updated dependency `@xonovex/agent-cli-go-darwin-arm64` to `0.1.30`
- Updated dependency `@xonovex/agent-cli-go-darwin-x64` to `0.1.30`
- Updated dependency `@xonovex/agent-cli-go-win32-x64` to `0.1.30`

## 0.1.24

### Patch Changes

- [`ffa01e5`](https://github.com/xonovex/platform/commit/ffa01e5da8a6d5eb551945247629161840d0c415) [@Deorder](https://github.com/Deorder)! - expose named go devShell in Go project flakes
- [`cdb0fd2`](https://github.com/xonovex/platform/commit/cdb0fd21ebcee694fbaf6b88ebe5b56a4f7d7938) [@Deorder](https://github.com/Deorder)! - nixflake sandbox tier + pinned-toolchain deny-default policy (#25)
- [`52a05be`](https://github.com/xonovex/platform/commit/52a05be3f4378a729ddf9fab5fb6d38f9e6d08f1) [@Deorder](https://github.com/Deorder)! - go.nix carries go_1_26; all Go projects on 1.26; docker-agent composes from nix/
- [`e7f1cb6`](https://github.com/xonovex/platform/commit/e7f1cb61dba7072025c710c242cc6540e43af58e) [@Deorder](https://github.com/Deorder)! - compose flakes from per-tool nix/\*.nix modules
- Updated dependency `@xonovex/agent-cli-go-linux-arm64` to `0.1.24`
- Updated dependency `@xonovex/agent-cli-go-linux-x64` to `0.1.24`
- Updated dependency `@xonovex/agent-cli-go-darwin-arm64` to `0.1.24`
- Updated dependency `@xonovex/agent-cli-go-darwin-x64` to `0.1.24`
- Updated dependency `@xonovex/agent-cli-go-win32-x64` to `0.1.24`

## 0.1.23

### Patch Changes

- [`1e775f8`](https://github.com/xonovex/platform/commit/1e775f8) - refactor: migrate agent-cli from TypeScript to Go with shared code extraction
- [`49b95b7`](https://github.com/xonovex/platform/commit/49b95b7) - refactor: extract agent sandbox implementations to shared-agent-go
- Updated dependency `@xonovex/agent-cli-go-linux-arm64` to `0.1.23`
- Updated dependency `@xonovex/agent-cli-go-linux-x64` to `0.1.23`
- Updated dependency `@xonovex/agent-cli-go-darwin-arm64` to `0.1.23`
- Updated dependency `@xonovex/agent-cli-go-darwin-x64` to `0.1.23`
- Updated dependency `@xonovex/agent-cli-go-win32-x64` to `0.1.23`

## 0.1.20

### Patch Changes

- [`d6d5985`](https://github.com/xonovex/platform/commit/d6d59851cb653cdbf6136c93101b63b6efa3c61e) [@Deorder](https://github.com/Deorder)! - update plugin metadata and descriptions across multiple plugins
- [`c7f41f4`](https://github.com/xonovex/platform/commit/c7f41f4e03e493bf25556ec592b74150f9e4547e) [@Deorder](https://github.com/Deorder)! - add TypeScript configuration and Vitest setup
- Updated dependency `@xonovex/agent-cli-go-linux-arm64` to `0.1.20`
- Updated dependency `@xonovex/agent-cli-go-linux-x64` to `0.1.20`
- Updated dependency `@xonovex/agent-cli-go-win32-x64` to `0.1.20`

## 0.1.13

### Patch Changes

- [`e1c5d24`](https://github.com/xonovex/platform/commit/e1c5d24c2713015f833d1c6d7ac5944056b1b5de) Thanks [@deorder](https://github.com/deorder)! - Simplify GitHub Actions workflows with composite actions

## 0.1.12

### Patch Changes

- [`1baa44d`](https://github.com/xonovex/platform/commit/1baa44d839c9d73f3c85a69e2ffa3c14410f0f6a) Thanks [@deorder](https://github.com/deorder)! - chore: bump all packages

## 0.1.11

### Patch Changes

- [`50e70c9`](https://github.com/xonovex/platform/commit/50e70c914227ce406bdcd29d0b9892407c317e19) Thanks [@deorder](https://github.com/deorder)! - chore: bump all packages

## 0.1.10

### Patch Changes

- [`afcdbad`](https://github.com/xonovex/platform/commit/afcdbad99e619fe22f85b1c26589e79df62118a5) Thanks [@deorder](https://github.com/deorder)! - Bump versions past provenance conflicts (final)

## 0.1.9

### Patch Changes

- [`a1d4f78`](https://github.com/xonovex/platform/commit/a1d4f78f30f1b7cb97e42937badfb5ff40a27797) Thanks [@deorder](https://github.com/deorder)! - Bump versions past provenance conflicts (provenance now disabled)

## 0.1.8

### Patch Changes

- [`220a4cb`](https://github.com/xonovex/platform/commit/220a4cb2324f51cdaa9e1702e2eac219919749e5) Thanks [@deorder](https://github.com/deorder)! - Bump version to resolve npm provenance conflict

## 0.1.7

### Patch Changes

- [`b183356`](https://github.com/xonovex/platform/commit/b1833560a6d057832e260d93cbca2de431aec925) Thanks [@deorder](https://github.com/deorder)! - Add README documentation to all packages

## 0.1.6

### Patch Changes

- [`217896d`](https://github.com/xonovex/platform/commit/217896d9c1848fa330cdc03540da579ee02f93df) Thanks [@deorder](https://github.com/deorder)! - Test publish order fix

## 0.1.5

### Patch Changes

- [`e2d1d83`](https://github.com/xonovex/platform/commit/e2d1d8386010c6bd14e93957f873dd5cdb0d59c9) Thanks [@deorder](https://github.com/deorder)! - Test fixed package group release

## 0.1.4

### Patch Changes

- [`c73f44f`](https://github.com/xonovex/platform/commit/c73f44f7d1ff73fb64041ded61f4d6d6b648bcb3) Thanks [@deorder](https://github.com/deorder)! - Test release workflow for Go packages
