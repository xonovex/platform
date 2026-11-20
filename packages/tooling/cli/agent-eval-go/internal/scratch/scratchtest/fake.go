// Package scratchtest holds the scratch double that the unit tests use. It
// hands out numbered paths and records what was removed; the files below those
// paths live in whatever in-memory file system the test passes to the code.
package scratchtest

import (
	"fmt"
	"sync"
)

// Fake is a Scratch that touches no disk.
type Fake struct {
	mu      sync.Mutex
	made    []string
	removed []string
	// MakeErr makes Make fail.
	MakeErr error
}

// New returns an empty Fake.
func New() *Fake { return &Fake{} }

// Make returns /scratch/<prefix><n>.
func (f *Fake) Make(prefix string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.MakeErr != nil {
		return "", f.MakeErr
	}
	dir := fmt.Sprintf("/scratch/%s%d", prefix, len(f.made)+1)
	f.made = append(f.made, dir)
	return dir, nil
}

// Remove records the directory.
func (f *Fake) Remove(dir string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removed = append(f.removed, dir)
	return nil
}

// Made lists the directories handed out.
func (f *Fake) Made() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.made...)
}

// Removed lists the directories removed.
func (f *Fake) Removed() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.removed...)
}
