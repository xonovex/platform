# @xonovex/moon-nix-extension

A Moon extension that maps a task's detected native toolchains to curated Nix
environment components and runs the task in the resulting workspace devShell.

Version 0.2.0 is qualified against Moon 2.4.5 and Moon PDK 2.0.4. The Moon WASM
extension interface remains experimental, so compatibility with newer Moon
versions is recorded separately before the declared support version changes.

## Choose one Nix plugin

Every workspace must use exactly one Xonovex Nix plugin. Neither replaces the
other: both answer the same question — run selected tasks inside a flake
devShell — with different trade-offs, and both are maintained.

| Aspect         | `moon_nix_extension`                                                                                                                                            | `moon_nix_toolchain`                                                                                                     |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Task selection | Detected native toolchains mapped to components, plus typed per-project and per-task overrides, all in one validated file                                        | Explicit `nix` toolchain selection, routed by task, toolchain, tag, and language selectors                               |
| Environment    | Centrally composed components through `lib.mkMoonShell`, or an explicit installable (`path:` self-contained copy, `dir:` resolved through the workspace git tree) | Workspace devShell, or a project's own `flake.nix` discovered automatically; the devShell picked by selector             |
| Cache contract | Consumer declares central and project Nix inputs by hand — a Moon extension cannot contribute task hash contents, so a missed input is a silently stale cache    | Plugin folds the resolved flake root, devShell, and `flake.lock` into every task hash automatically and precisely        |
| Realization    | Lazy, when an active task runs                                                                                                                                   | Eager pre-build through `setup_environment`, so the first wrapped task is warm                                           |
| New projects   | Must be added to the config; an unlisted project silently runs on host tools                                                                                     | Wrap by tag or language; adding the tag is the whole opt-in                                                              |
| Fail-closed    | One global `failClosed` flag                                                                                                                                     | Opt-in per tag and language allowlists                                                                                   |

Pick the extension when the workspace wants every wrapped project named
explicitly in one typed file, environments composed from a central component
registry, and no `nix` entry in any task's toolchain list — it models Nix as an
environment concern rather than as a toolchain of the task, and it costs
nothing until an active task runs.

Pick the toolchain plugin when the workspace wants automatic, exact cache
invalidation (the resolved shell is hashed for you, and an edit no shell uses
invalidates nothing), projects that ship and compose their own flakes with zero
per-project config, or tag and language routing. Arbitrary peer command
replacement is also unsupported by the extension.

Do not configure both plugins, even temporarily; switching between them
replaces the whole configuration atomically in one reviewed PR.

## Configure the extension

Register the exact release locator in `.moon/extensions.yml`. The GitHub
release contains `moon_nix_extension.wasm` and a
`moon_nix_extension.wasm.sha256` sidecar. The sidecar supports independent
download verification; Moon's `github://` locator is not known to verify it
automatically.

```yaml
# .moon/extensions.yml
nix-environment:
  plugin: "github://xonovex/platform/moon_nix_extension@moon_nix_extension-v0.2.0"
  baseComponents: [general]
  environmentByToolchain:
    javascript: node
    node: node
    npm: node
    go: go
    system: general
  environmentByProject:
    api:
      components: [postgresql]
      mode: append
    special-compiler:
      installable: "path:./projects/special-compiler#moon"
  environmentByTask:
    "api:release":
      components: [release]
      mode: replace
    "special-compiler:bootstrap":
      installable: "path:./projects/special-compiler#bootstrap"
  failClosed: true
```

Projects keep their semantic Moon toolchains. When Nix owns a native binary,
set its Moon version to `null` so Proto does not install a second version.

```yaml
# .moon/toolchains.yml
$schema: https://moonrepo.dev/schemas/toolchain.json
go:
  version: null
npm:
  version: null
node:
  version: null
javascript:
  packageManager: npm
  syncPackageManagerField: false
  syncProjectWorkspaceDependencies: false
```

Do not add a `nix` entry to `.moon/toolchains.yml` or select `nix` in a
project's `toolchains`. Moon detects the native toolchains on each task while
Nix supplies their executables after the extension composes their components.

Set `GITHUB_TOKEN` in CI so Moon's GitHub API requests are authenticated and
do not depend on the anonymous rate limit:

```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Expose central components

The workspace flake must expose `lib.mkMoonShell = system: names: ...`. A small
central registry keeps component names stable and validates unknown names.

```nix
# flake.nix
{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/<locked-revision>";

  outputs = { nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
      componentsFor = system: import ./nix/components.nix {
        pkgs = nixpkgs.legacyPackages.${system};
      };
      mkMoonShell = system: names:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          components = componentsFor system;
          missing = builtins.filter (name: !(builtins.hasAttr name components)) names;
        in
        if missing != [ ] then
          throw "unknown Moon component(s): ${builtins.concatStringsSep ", " missing}"
        else
          pkgs.mkShell {
            inputsFrom = map (name: components.${name}) names;
          };
    in
    {
      lib.mkMoonShell = mkMoonShell;
      devShells = forAllSystems (system: {
        default = mkMoonShell system [ "general" ];
      });
    };
}
```

```nix
# nix/components.nix
{ pkgs }:
{
  general = pkgs.mkShell { packages = [ pkgs.bash ]; };
  node = pkgs.mkShell { packages = [ pkgs.nodejs_24 ]; };
  go = pkgs.mkShell { packages = [ pkgs.go ]; };
  postgresql = pkgs.mkShell { packages = [ pkgs.postgresql ]; };
  release = pkgs.mkShell { packages = [ pkgs.gh ]; };
}
```

`baseComponents` are added only when a detected toolchain mapping or scoped
project/task override activates the extension. They do not activate an
otherwise unmapped task.

## Use scoped overrides and locked project flakes

Component overrides use `append` or `replace`. A project override applies
before a full-target task override. An installable selects a named devShell
from a locked, workspace-contained project flake:

```nix
# projects/special-compiler/flake.nix
{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/<locked-revision>";

  outputs = { nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
    in
    {
      devShells = nixpkgs.lib.genAttrs systems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          moon = pkgs.mkShell { packages = [ pkgs.clang ]; };
          bootstrap = pkgs.mkShell { packages = [ pkgs.cmake ]; };
        });
    };
}
```

Commit both `projects/special-compiler/flake.nix` and its generated
`projects/special-compiler/flake.lock`. Installables use
`path:./workspace-relative#devShell` or `dir:./workspace-relative#devShell`;
remote locators, inline Nix expressions, missing attributes, missing locks, and
paths that resolve outside the workspace are rejected.

The two schemes differ in how nix is handed the directory. A `path:`
installable is copied into the store alone, so the flake must be
self-contained: every input has to be fetchable from its lock, and a relative
input escaping the directory cannot resolve. A `dir:` installable is a bare
directory reference that nix resolves in place — in a git workspace, through
the enclosing git tree — so the project flake may compose from shared workspace
files through relative inputs (for example `path:../../nix`). With `dir:` only
tracked files exist to the evaluation in a git workspace: a file that is not in
the index does not exist to the build, and the whole tracked tree is hashed per
evaluation, which a large workspace pays on every wrapped task run. Choose
`path:` for a self-contained flake and `dir:` for one that composes from shared
workspace files.

Resolution is deterministic:

1. Sort and deduplicate `baseComponents` and mapped toolchain components.
2. Apply the project override.
3. Apply the full-target task override.

`append` adds components to the current component set and `replace` replaces
it. A task `replace` or installable can supersede a project installable, but
appending components to an installable is rejected.

## Declare cache inputs

Moon extensions cannot contribute task hash contents. Consumers must inherit
the central Nix and extension configuration inputs into every task:

```yaml
# .moon/tasks/nix.yml
$schema: https://moonrepo.dev/schemas/tasks.json
implicitInputs:
  - /flake.nix
  - /flake.lock
  - /nix/**/*
  - /.moon/extensions.yml
```

Tasks using a project installable must also declare that project's flake and
component inputs:

```yaml
# projects/special-compiler/moon.yml
$schema: https://moonrepo.dev/schemas/project.json
language: system
tasks:
  compile:
    command: clang
    args: [src/main.c, -o, build/app]
    inputs:
      - src/**/*
      - flake.nix
      - flake.lock
      - nix/**/*
```

Only declared inputs invalidate Moon's task cache. The extension does not
discover central or project flake files for the hasher.

## Execution constraints

- Another extension that replaces command or script text on the same mapped
  task is unsupported because two arbitrary replacements cannot be composed
  safely. Environment-only extensions remain compatible.
- `IN_NIX_SHELL` and `MOON_NIX_WRAPPED=1` prevent nested wrapping.
- With `failClosed: true`, an active task fails with an actionable error when
  `nix` is unavailable. With `false`, the active task remains unchanged.
- The extension realizes Nix lazily when an active task runs; it has no eager
  `setup_environment` hook.

## Switch from the toolchain plugin

Switching in either direction is a deliberate trade (see the comparison above),
not an upgrade. When a workspace does switch to the extension, make it atomic
in one reviewed PR:

1. Remove the `nix` locator and explicit `nix` task/project selections.
2. Add the pinned extension locator, native toolchain mappings, scoped
   overrides, and explicit central/project cache inputs.
3. Set every Nix-owned native toolchain version to `null` and run the full
   workspace CI before removing the old pin.

Do not change unrelated workspaces or consumers as part of a migration. Each
workspace owns its own locator, cache contract, and validation evidence.

## License

MIT
