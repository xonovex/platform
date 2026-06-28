# @xonovex/agent-cli-go

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
