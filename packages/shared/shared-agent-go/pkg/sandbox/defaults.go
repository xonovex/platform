package sandbox

// DefaultContainerImage is the default container image for running agents.
const DefaultContainerImage = "node:trixie-slim"

// UserConfigPaths lists home-relative paths that should be bind-mounted
// into sandboxed environments so agents can access user configuration.
var UserConfigPaths = []string{
	".claude",
	".claude.json",
	".gitconfig",
	".gitignore_global",
	".ssh",
	".config",
	".npmrc",
	".npm",
	".npm-global",
	".cargo",
	".rustup",
	".local",
	".cache",
}
