// Package project reads the agentic.validate.json file of a repository and
// turns it, with the command line, into the plan of one run: which checks run,
// where each one looks, and how severe its findings stay.
package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// File is the default name of the project file.
const File = "agentic.validate.json"

// ErrUnknownCheck reports a check name that no registered check has.
var ErrUnknownCheck = errors.New("unknown check")

// Settings are the project settings of one check.
type Settings struct {
	Enabled *bool           `json:"enabled"`
	Paths   []string        `json:"paths"`
	Level   string          `json:"level"`
	Options json.RawMessage `json:"options"`
}

// Project is the parsed project file.
type Project struct {
	Version int                 `json:"version"`
	Skip    []string            `json:"skip"`
	Checks  map[string]Settings `json:"checks"`
}

// Load reads the project file. A file that does not exist is an empty project
// unless required is true. A file that breaks its schema is an error that
// lists every violation.
func Load(t tree.Tree, file string, validator jsonschema.Validator, required bool) (Project, error) {
	text, ok, err := t.Optional(file)
	if err != nil {
		return Project{}, err
	}
	if !ok {
		if required {
			return Project{}, fmt.Errorf("project file %s does not exist", file)
		}
		return Project{}, nil
	}
	violations, err := validator.Validate(schemas.ValidateProject, []byte(text))
	if err != nil {
		return Project{}, err
	}
	if len(violations) > 0 {
		parts := make([]string, len(violations))
		for i, v := range violations {
			parts[i] = v.Pointer + ": " + v.Message
		}
		return Project{}, fmt.Errorf("invalid project file %s: %s", file, strings.Join(parts, "; "))
	}
	var parsed Project
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return Project{}, fmt.Errorf("decode project file %s: %w", file, err)
	}
	return parsed, nil
}

// Step is one check of the plan, with the settings that apply to it.
type Step struct {
	Check   check.Check
	Paths   []string
	Options json.RawMessage
	// Cap is the most severe level the findings of the step keep.
	Cap findings.Level
}

// Plan chooses the checks of a run. A named check runs even when the project
// switches it off. With no name, every check the project does not switch off
// runs. A path from the command line replaces the paths of every chosen check.
func Plan(registered []check.Check, p Project, names, paths []string) ([]Step, error) {
	byName := map[string]check.Check{}
	for _, c := range registered {
		byName[c.Name()] = c
	}
	for name := range p.Checks {
		if _, ok := byName[name]; !ok {
			return nil, fmt.Errorf("%w %q in the project file; known checks: %s", ErrUnknownCheck, name, strings.Join(Names(registered), ", "))
		}
	}
	chosen := registered
	if len(names) > 0 {
		chosen = nil
		for _, name := range names {
			c, ok := byName[name]
			if !ok {
				return nil, fmt.Errorf("%w %q; known checks: %s", ErrUnknownCheck, name, strings.Join(Names(registered), ", "))
			}
			chosen = append(chosen, c)
		}
	}
	steps := make([]Step, 0, len(chosen))
	for _, c := range chosen {
		settings := p.Checks[c.Name()]
		if len(names) == 0 && settings.Enabled != nil && !*settings.Enabled {
			continue
		}
		step := Step{Check: c, Paths: c.DefaultPaths(), Options: settings.Options, Cap: findings.Fail}
		if len(settings.Paths) > 0 {
			step.Paths = settings.Paths
		}
		if len(paths) > 0 {
			step.Paths = paths
		}
		if settings.Level != "" {
			level, err := findings.ParseLevel(settings.Level)
			if err != nil {
				return nil, err
			}
			step.Cap = level
		}
		steps = append(steps, step)
	}
	return steps, nil
}

// Names lists the names of the checks, sorted.
func Names(registered []check.Check) []string {
	names := make([]string, len(registered))
	for i, c := range registered {
		names[i] = c.Name()
	}
	sort.Strings(names)
	return names
}
