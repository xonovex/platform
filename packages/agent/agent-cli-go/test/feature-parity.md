# Feature Parity Checklist: TypeScript vs Go

This document tracks feature parity between the TypeScript `script-agent` and Go `script-agent-go` implementations.

## Core Commands

### `run` command

- [x] Agent selection (`-a`/`--agent`)
- [x] Provider selection (`-p`/`--provider`)
- [x] Sandbox methods (`-s`/`--sandbox`): none, bwrap, docker, compose, nix
- [x] Work directory (`-w`/`--work-dir`)
- [x] Worktree support (`--worktree-branch`, `--worktree-source-branch`, `--worktree-dir`)
- [x] Config file (`-c`/`--config`)
- [x] Network control (`-N`/`--network`)
- [x] Bind mounts (`--bind`, `--ro-bind`)
- [x] Environment variables (`--env`)
- [x] Verbose/debug flags (`-v`, `-d`)
- [x] Container image (`--image`)
- [x] Compose file (`--compose-file`)
- [x] Service selection (`--service`)

### `sync` command

- [x] Target selection (`-t`/`--target`): claude, gemini, codex, opencode, copilot, kiro, all
- [x] Entity filtering (`-e`/`--entity`): command, skill, agent, instructions
- [x] File filtering (`-f`/`--file`)
- [x] Dry run (`-n`/`--noop`)
- [x] Overwrite control (`--overwrite`)
- [x] Delete sync (`--sync-delete`)
- [x] Directory overrides (all `--*-dir` flags)

### `completion` command

- [x] Bash completion
- [x] Zsh completion
- [x] Fish completion
- [x] PowerShell completion

## Agents

### Claude agent

- [x] Arg building
- [x] Env building
- [x] Command resolution

### OpenCode agent

- [x] Arg building
- [x] Env building
- [x] Command resolution

## Providers

- [x] Gemini (Claude API key)
- [x] GLM (Claude API)
- [x] GPT-5 Codex (Claude API)
- [x] Gemini (OpenCode API)

## Sandbox Methods

### None (direct execution)

- [x] Direct process execution
- [x] Environment passing
- [x] Working directory support

### Bubblewrap

- [x] Namespace isolation
- [x] Network control
- [x] Read-write bind mounts
- [x] Read-only bind mounts
- [x] Environment passing

### Docker

- [x] Image selection
- [x] Volume mounts
- [x] Network control
- [x] Environment passing
- [x] Interactive mode

### Docker Compose

- [x] Compose file support
- [x] Service selection
- [x] Environment passing

### Nix

- [x] Preset support (fullstack, backend, frontend)
- [x] Package set selection
- [x] Custom packages
- [x] Network control
- [x] Environment passing

## Sync Converters

### Claude

- [x] Command converter
- [x] Skill converter
- [x] Agent converter

### Gemini

- [x] Command converter (TOML format)
- [x] Skill converter
- [x] Agent converter (TOML format)

### Codex

- [x] Command converter

### OpenCode

- [x] Command converter
- [x] Skill converter
- [x] Agent converter (JSON format)

### Copilot

- [x] Command converter (.prompt.md format)
- [x] Skill converter
- [x] Agent converter

### Kiro

- [x] Command converter
- [x] Skill converter
- [x] Agent converter (JSON format)
- [x] Power converter

## Configuration

- [x] YAML config file loading
- [x] TOML config file loading
- [x] Config merge (CLI overrides file)
- [x] Environment variable support

## Utilities (script-lib-go)

- [x] Colors (with NO_COLOR support)
- [x] Logging (info, error, warning, success, debug)
- [x] Platform detection (linux, darwin, windows)
- [x] Error handling (require_command, require_file, etc.)
- [x] Path utilities (expand_home, resolve, etc.)
- [x] Git utilities (get_root, current_branch, etc.)

## Tests

- [x] Unit tests for config loading
- [x] Unit tests for colors
- [x] Unit tests for logging
- [x] Unit tests for path utilities
- [x] Unit tests for git utilities
- [x] Unit tests for error handling
- [x] Unit tests for platform detection
- [x] Integration tests for run command
- [x] Integration tests for sync command
- [x] Integration tests for completion command

## Not Ported (Intentionally)

The following features from TypeScript were not ported due to Go implementation differences:

- Init commands (`--init-command`, `--sandbox-init-command`) - Simplified in Go version
- Tmux integration - Not needed for Go binary
- ID generation utilities - Simplified in Go version
