// Package schema validates JSON documents against the schemas of the contract.
// A rule names a schema and the files it governs; a document that names a
// schema through its $schema member is checked against it as well.
package schema

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/glob"
)

// Name is the name of the check.
const Name = "schema"

// Rule binds a schema to the files it governs.
type Rule struct {
	// Schema is a schema id or its short name, for example "baseline".
	Schema string `json:"schema"`
	// Files are glob patterns of the documents the schema governs.
	Files []string `json:"files"`
}

// Options tune the check.
type Options struct {
	// Rules bind schemas to files.
	Rules []Rule `json:"rules"`
	// Detect checks every JSON document whose $schema names a known schema.
	Detect bool `json:"detect"`
	// Ignore names schemas, by id or short name, that detection leaves to the
	// check that owns them.
	Ignore []string `json:"ignore"`
}

func defaults() Options {
	return Options{
		Detect: true,
		Ignore: []string{
			schemas.Sources, schemas.EvalTrigger, schemas.EvalOutput, schemas.EvalInstruction,
		},
	}
}

var jsonFile = regexp.MustCompile(`\.json$`)

// Check is the schema check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string { return "JSON documents against the contract schemas" }

// DefaultPaths returns the repository root.
func (Check) DefaultPaths() []string { return []string{"."} }

// resolve returns the id of a schema named by id or by short name.
func resolve(name string) (string, bool) {
	for _, id := range schemas.IDs() {
		if id == name || strings.TrimSuffix(path.Base(id), ".schema.json") == name {
			return id, true
		}
	}
	return "", false
}

type binding struct {
	id      string
	pattern *regexp.Regexp
}

// Run checks the JSON documents below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	var bindings []binding
	for _, rule := range options.Rules {
		id, ok := resolve(rule.Schema)
		if !ok {
			return nil, fmt.Errorf("unknown schema %q", rule.Schema)
		}
		for _, pattern := range rule.Files {
			re, err := glob.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("pattern %q: %w", pattern, err)
			}
			bindings = append(bindings, binding{id, re})
		}
	}
	ignored := map[string]bool{}
	for _, name := range options.Ignore {
		id, ok := resolve(name)
		if !ok {
			return nil, fmt.Errorf("unknown schema %q", name)
		}
		ignored[id] = true
	}
	report := check.NewReporter(Name)
	files, err := ctx.Tree.WalkAll(ctx.Paths, func(p string) bool { return jsonFile.MatchString(p) })
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		ids := idsFor(file, bindings)
		if len(ids) == 0 && !options.Detect {
			continue
		}
		text, err := ctx.Tree.Read(file)
		if err != nil {
			return nil, err
		}
		if options.Detect {
			if id, ok := declared(text); ok && !ignored[id] && !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
		for _, id := range ids {
			if err := validate(ctx, report, file, text, id); err != nil {
				return nil, err
			}
		}
	}
	return report.Findings(), nil
}

func idsFor(file string, bindings []binding) []string {
	var ids []string
	for _, b := range bindings {
		if b.pattern.MatchString(file) && !slices.Contains(ids, b.id) {
			ids = append(ids, b.id)
		}
	}
	return ids
}

// declared returns the contract schema that a document names, if it names one.
func declared(text string) (string, bool) {
	var head struct {
		Schema string `json:"$schema"`
	}
	if json.Unmarshal([]byte(text), &head) != nil {
		return "", false
	}
	return resolve(head.Schema)
}

func validate(ctx check.Context, report *check.Reporter, file, text, id string) error {
	if !json.Valid([]byte(text)) {
		report.Fail("schema.json", file, 1, "%s is not valid JSON", file)
		return nil
	}
	violations, err := ctx.Validator.Validate(id, []byte(text))
	if err != nil {
		return err
	}
	for _, v := range violations {
		report.Fail("schema.invalid", file, 1, "%s%s does not match %s: %s", path.Base(file), v.Pointer, path.Base(id), v.Message)
	}
	return nil
}
