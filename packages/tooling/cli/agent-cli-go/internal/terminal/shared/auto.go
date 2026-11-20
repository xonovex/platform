package shared

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotInside marks an automatic selection that finds no wrapper whose session
// the CLI runs in.
var ErrNotInside = errors.New("agent-cli runs inside no terminal session that --terminal auto can use")

// SelectAuto returns the one wrapper whose session the CLI runs in, given the
// candidates that report it runs inside them. It never guesses: no candidate
// fails closed, and more than one asks the user to name the wrapper.
func SelectAuto(inside []TerminalType) (TerminalType, error) {
	switch len(inside) {
	case 0:
		return TerminalNone, ErrNotInside
	case 1:
		return inside[0], nil
	default:
		names := make([]string, len(inside))
		for i, name := range inside {
			names[i] = string(name)
		}
		return TerminalNone, fmt.Errorf("agent-cli runs inside %s at once; name the wrapper with --terminal %s", strings.Join(names, " and "), strings.Join(names, " or --terminal "))
	}
}
