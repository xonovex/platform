// Package spec declares the keys of the agentic configuration as data: the
// section and JSON Pointer of each key, the rule that combines the layers that
// set it, and the layers that may set it.
package spec

import "strings"

// Layer names one source of configuration.
type Layer string

const (
	// Managed is the system-wide file an administrator controls.
	Managed Layer = "managed"
	// User is the file of the user.
	User Layer = "user"
	// Project is the agentic.json of the project. A project can be hostile, so
	// it may set only keys that cannot widen access, and a gated section of it
	// takes effect only after the user approves its digest.
	Project Layer = "project"
	// Flags is the command line of this run.
	Flags Layer = "flags"
	// Default is the built-in value of a key that no layer sets.
	Default Layer = "default"
)

// Rule is how the layers that set one key combine.
type Rule string

const (
	// Replace gives the key one value. The managed file wins, then the command
	// line, then the project, then the user. A command line value that differs
	// from a managed one is an error, not a silent override.
	Replace Rule = "replace"
	// Append gives the key the items of every layer, the managed ones last so that
	// a later item of the same name still comes from the administrator.
	Append Rule = "append"
	// Require gives the key true when any layer sets it true. A guarantee that one
	// layer demands stays demanded.
	Require Rule = "require"
	// Cap gives the key a ceiling. The user or the managed file sets it, a
	// project may only lower it, the command line may choose any value up to the
	// managed one, and the managed value bounds every other layer.
	Cap Rule = "cap"
	// Merge combines the entries of a map: each entry comes whole from the layer
	// that Replace would choose for it.
	Merge Rule = "merge"
)

var (
	// EveryFile lets every file set a key that only narrows what a run may do,
	// selects among fixed choices, or tunes a tool.
	EveryFile = []Layer{Managed, User, Project}
	// TrustedFiles keeps a project file from setting a key that widens access.
	TrustedFiles = []Layer{Managed, User}
)

// Key declares one setting.
type Key struct {
	// Section is the top-level member of agentic.json that holds the key.
	Section string
	// Pointer is the JSON Pointer of the key inside its section, empty for a key
	// that is the whole section.
	Pointer string
	Rule    Rule
	// Layers are the files that may set the key. The command line may always
	// set it.
	Layers []Layer
	Doc    string
	// Ceiling is the value of a Cap key that no layer sets.
	Ceiling float64
}

// Allows reports whether layer may set the key.
func (k Key) Allows(layer Layer) bool {
	if layer == Flags {
		return true
	}
	for _, allowed := range k.Layers {
		if allowed == layer {
			return true
		}
	}
	return false
}

// Name is the dotted name of the key, such as eval.budgets.maxModelRuns.
func (k Key) Name() string {
	if k.Pointer == "" {
		return k.Section
	}
	return k.Section + strings.ReplaceAll(k.Pointer, "/", ".")
}

// Section declares one top-level member of agentic.json and the tool keys in
// it.
type Section struct {
	Name string
	// SchemaID is the id of the schema that the owning tool checks the section
	// against.
	SchemaID string
	// Gated sections of a project file take effect only after the user approves
	// the digest of the file.
	Gated bool
	Keys  []Key
}

// Key returns the key at pointer.
func (s Section) Key(pointer string) (Key, bool) {
	for _, key := range s.Keys {
		if key.Pointer == pointer {
			return key, true
		}
	}
	return Key{}, false
}
