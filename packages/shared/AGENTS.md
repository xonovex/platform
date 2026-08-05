# Shared Libraries

- Shared sits in the middle of `config -> shared -> agent`: these libraries consume `config` and are consumed by `agent`, and they are `layer: library`.
- No package under `packages/script/` may depend on `shared-core`. `shared-core` is tagged `npm`, so its `npm-check`, publish, and version tasks already depend on the script packages; the reverse edge makes moon reject the project graph with `would_cycle`. Duplicate the helper into `script-moon-common` instead.
- The TypeScript and Go cores are separate packages with no bridge between them: `shared-core` for TypeScript, `shared-core-go` for Go. `shared-agent-go` holds the agent, provider, policy, and provisioning types and depends on `shared-core-go`.
- `shared-core` is tagged `npm` and versions with the `config` packages, a line separate from the plugin packages and the agent CLI. Move that whole line with one `workspace-config:version-bump-lockstep` invocation naming the ten `config` packages plus `shared-core`, never a loop over `<project>:version-bump`. The Go packages publish nothing: `packages/agent/` consumes them from source through `replace` directives in its `go.mod`.
