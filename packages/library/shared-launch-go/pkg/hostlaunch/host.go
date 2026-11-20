// Package hostlaunch holds what every isolation leaf that starts the agent on
// this machine shares: the Host port, the resolution of the host paths a
// sandbox binds, the environment and the init commands of the agent process,
// and the start of the sandbox program. It reaches the machine only through
// Host.
package hostlaunch

import "io/fs"

// FileSystem is the read-only view of the host file system that an isolator
// resolves its mounts against.
type FileSystem interface {
	Stat(path string) (fs.FileInfo, error)
	EvalSymlinks(path string) (string, error)
}

// Host is the isolation axis port to the machine that launches the sandbox: the
// files an isolator mounts, the environment it launches under, and the
// processes it probes and starts.
type Host interface {
	FileSystem
	Environ() []string
	Getenv(key string) string
	LookPath(file string) (string, error)
	// Capture runs name in dir and returns its trimmed standard output.
	Capture(name string, args []string, dir string) (string, error)
	// Run starts name attached to the terminal of this process and returns its
	// exit code; a nil env inherits the environment of this process.
	Run(name string, args, env []string, dir string) (int, error)
}
