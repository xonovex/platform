// Package osscratch is the scratch adapter that makes real temporary
// directories. Only the composition root imports it.
package osscratch

import "os"

// Scratch makes directories below the operating system temporary directory.
type Scratch struct{}

// New returns a Scratch.
func New() Scratch { return Scratch{} }

// Make creates a temporary directory.
func (Scratch) Make(prefix string) (string, error) { return os.MkdirTemp("", prefix) }

// Remove deletes a directory tree.
func (Scratch) Remove(dir string) error { return os.RemoveAll(dir) }
