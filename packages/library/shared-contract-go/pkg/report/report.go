// Package report names the output formats of the validate, audit and eval
// commands and reads one from a flag. Each command supports a subset of them.
package report

import (
	"errors"
	"fmt"
	"strings"
)

// Format is an output format.
type Format string

// The output formats.
const (
	Text  Format = "text"
	JSON  Format = "json"
	SARIF Format = "sarif"
)

// ErrUnknownFormat is the error for a format name a command does not support.
var ErrUnknownFormat = errors.New("unknown output format")

// Parse returns the format a name stands for. A name outside supported is an
// error that lists the supported names.
func Parse(name string, supported ...Format) (Format, error) {
	for _, format := range supported {
		if string(format) == name {
			return format, nil
		}
	}
	return "", fmt.Errorf("%w %q: use %s", ErrUnknownFormat, name, Names(supported...))
}

// Names renders formats as a list for a person, such as "text, json or sarif".
func Names(formats ...Format) string {
	names := make([]string, len(formats))
	for i, format := range formats {
		names[i] = string(format)
	}
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1]
}
