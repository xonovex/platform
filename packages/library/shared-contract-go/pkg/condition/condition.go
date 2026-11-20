// Package condition names the three states of an instruction under evaluation:
// present in the context, absent from it, or ablated by removing one part.
package condition

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Condition is the state of an instruction during one eval run.
type Condition string

const (
	// Present keeps the instruction in the context unchanged.
	Present Condition = "present"
	// Absent removes the instruction from the context.
	Absent Condition = "absent"
	// Ablated removes one named part of the instruction and keeps the rest.
	Ablated Condition = "ablated"
)

// ErrUnknown reports a condition name that is not one of All.
var ErrUnknown = errors.New("unknown condition")

// All lists every condition.
func All() []Condition { return []Condition{Present, Absent, Ablated} }

// Parse returns the condition a name denotes.
func Parse(name string) (Condition, error) {
	for _, c := range All() {
		if string(c) == name {
			return c, nil
		}
	}
	return "", fmt.Errorf("%w %q", ErrUnknown, name)
}

// Valid reports whether c is one of All.
func (c Condition) Valid() bool {
	_, err := Parse(string(c))
	return err == nil
}

// KeepsInstruction reports whether any of the instruction stays in the
// context: all of it when present, the rest of it when ablated.
func (c Condition) KeepsInstruction() bool { return c == Present || c == Ablated }

// UnmarshalJSON rejects a name that is not a condition.
func (c *Condition) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	parsed, err := Parse(name)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}
