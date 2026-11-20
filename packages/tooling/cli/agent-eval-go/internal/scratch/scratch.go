// Package scratch is the port for a private working directory. A harness run
// stages skills, fixtures and an isolated home in one, so a run reads nothing
// of the developer's machine and leaves nothing behind.
package scratch

// Scratch makes and removes private directories.
type Scratch interface {
	// Make creates an empty directory whose name starts with prefix and returns
	// its absolute path.
	Make(prefix string) (string, error)
	// Remove deletes a directory made by Make and everything below it.
	Remove(dir string) error
}
