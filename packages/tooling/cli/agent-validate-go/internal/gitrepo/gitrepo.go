// Package gitrepo answers the git questions of the checks: which files are
// committed and which paths are ignored. It reads the answers from a command
// runner that the composition root binds to the git program.
package gitrepo

import (
	"fmt"
	"strings"
)

// Runner runs git with the arguments in the directory and returns its output.
type Runner func(dir string, args ...string) (string, error)

// Repository reads the git state of the directory.
type Repository struct {
	Dir string
	Run Runner
}

// TrackedFiles lists the committed files, as slash paths relative to Dir.
func (r Repository) TrackedFiles() ([]string, error) {
	return r.lines("ls-files")
}

// IgnoredPaths lists what git ignores. A directory that is ignored as a whole
// ends in a slash and is not listed file by file.
func (r Repository) IgnoredPaths() ([]string, error) {
	return r.lines("ls-files", "--others", "--ignored", "--exclude-standard", "--directory")
}

func (r Repository) lines(args ...string) ([]string, error) {
	out, err := r.Run(r.Dir, args...)
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimRight(line, "\r"); line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}
