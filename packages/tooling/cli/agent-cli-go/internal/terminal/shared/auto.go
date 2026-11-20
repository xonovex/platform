package shared

import "errors"

// ErrNotInside marks an automatic selection that finds no wrapper whose session
// the CLI runs in.
var ErrNotInside = errors.New("agent-cli runs inside no terminal session that --terminal auto can use")

// SelectAuto returns the first candidate, which is the wrapper of highest
// priority among those whose session the CLI runs in: a session nests inside
// another, so the first candidate is the innermost one the user can act in. No
// candidate fails closed.
func SelectAuto(inside []TerminalType) (TerminalType, error) {
	if len(inside) == 0 {
		return TerminalNone, ErrNotInside
	}
	return inside[0], nil
}
