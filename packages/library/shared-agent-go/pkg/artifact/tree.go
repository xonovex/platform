package artifact

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
)

// Tree is the regular files of a source directory, sorted by path with no
// duplicate. It is the unit a digest covers.
type Tree []File

// NewTree validates files and returns them as a sorted Tree. It fails closed on
// a path that leaves its directory, a duplicate path, a path that is also the
// directory of another, and two paths that differ only in case.
func NewTree(files []File) (Tree, error) {
	tree := slices.Clone(files)
	slices.SortFunc(tree, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	folded := map[string]string{}
	directories := map[string]bool{}
	for _, file := range tree {
		if err := ConfinedPath(file.Path); err != nil {
			return nil, err
		}
		key := strings.ToLower(file.Path)
		if other, found := folded[key]; found {
			return nil, fmt.Errorf("paths %q and %q collide", other, file.Path)
		}
		folded[key] = file.Path
		for index, character := range file.Path {
			if character == '/' {
				directories[file.Path[:index]] = true
			}
		}
	}
	for _, file := range tree {
		if directories[file.Path] {
			return nil, fmt.Errorf("path %q is both a file and a directory", file.Path)
		}
	}
	return tree, nil
}

// Get returns the file at path.
func (t Tree) Get(path string) (File, bool) {
	index, found := slices.BinarySearchFunc(t, path, func(file File, target string) int {
		return strings.Compare(file.Path, target)
	})
	if !found {
		return File{}, false
	}
	return t[index], true
}

// Under returns the files below directory, with the directory taken off their
// paths.
func (t Tree) Under(directory string) Tree {
	prefix := directory + "/"
	var under Tree
	for _, file := range t {
		if rest, ok := strings.CutPrefix(file.Path, prefix); ok {
			file.Path = rest
			under = append(under, file)
		}
	}
	return under
}

// DigestPrefix starts every digest.
const DigestPrefix = "sha256:"

// Digest returns the content address of the tree: the SHA-256 of one record per
// file with its path, whether it is executable and the SHA-256 of its content.
// Two trees have one digest only when every path, mode bit and byte is equal.
func (t Tree) Digest() string {
	hash := sha256.New()
	for _, file := range t {
		mode := "-"
		if file.Executable {
			mode = "x"
		}
		sum := sha256.Sum256(file.Content)
		_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\n", file.Path, mode, hex.EncodeToString(sum[:]))
	}
	return DigestPrefix + hex.EncodeToString(hash.Sum(nil))
}

// ValidDigest reports whether digest is a sha256 digest in the form Digest
// returns.
func ValidDigest(digest string) bool {
	hexPart, ok := strings.CutPrefix(digest, DigestPrefix)
	if !ok || len(hexPart) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(hexPart)
	return err == nil && hexPart == strings.ToLower(hexPart)
}
