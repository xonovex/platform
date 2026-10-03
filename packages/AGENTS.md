# Packages

- Use `packages/<kind>/<package>/` for libraries, configuration, and assets. Use `packages/<group>/<kind>/<package>/` for runtime components, development tools, and execution environments.
- Keep a package basename and its published name stable when only its parent directory changes. After moving code, update relative paths, workspace discovery, task inputs, and documentation links.

## Agent execution policy

- Model the sandbox as independent `Isolation {none,bwrap,docker}`, `Provision {none,nix,command}`, and `Network {host,none,proxy}` axes plus `hostPassthrough`; never fuse them into one method.
- `bwrap` and default-runc Docker reduce attack surface but are not kernel trust boundaries. `nix` mounts only `flake.lock`/rev-pinned closure requisites from `nix path-info -r`, read-only; never mount all of `/nix/store` or the Nix daemon socket. `command` runs initialization commands before the agent.
- Apply network policy explicitly: `host` is unrestricted and does not satisfy `RequireEgressRestricted`; `none` uses `--unshare-net`/`--network none`; `proxy` is reserved and must fail closed until a transport prevents direct sockets from bypassing an HTTP(S) allowlist.
- Keep `hostPassthrough` off by default so host tools stay off PATH and are not bind-reachable; enabling it exposes host/base-image fallback tools and forfeits that guarantee. Default to `host`; callers requiring restricted egress must select `none`.
- Enforce guarantees independently and fail closed: `RequirePinnedProvision` needs Nix or a pinned image with `--frozen`/`--no-write-lock-file`; `RequireHostToolsUnreachable` needs host tools off PATH, closure-only binds, no host `$HOME`, and a pinned Docker image; `RequireEgressRestricted` needs `none`; `RequireKernelIsolation` needs runsc/gVisor or a sandboxed gVisor/Kata/kata-cc `runtimeClass`, never bwrap or default runc.
- Treat model-generated code as untrusted and subject to indirect prompt injection. Host tools unreachable does not mean host unreachable; filesystem and network guarantees remain separate, and unmet requested guarantees must stop execution.
- Use the shared `flake.lock`-pinned `nix/agent-env.nix` plus vendored `nix/mkAgentImage.nix`: the CLI bind-mounts closure requisites and the operator builds the same store paths into `dockerTools.streamLayeredImage`. Verify with `nix path-info -r`, not byte-identical layers; keep `maxLayers=100` and epoch `created`, never `now`. Use `nix2container` only when pushing every small change.
- Keep the adapted `nothingnesses/agent-images` uid-1000 passwd/group, `/workspace`, and XDG setup. `numtide/llm-agents.nix` is packaging only and pinned by `flake.lock`, and enabling `cache.numtide.com` expands trust.
