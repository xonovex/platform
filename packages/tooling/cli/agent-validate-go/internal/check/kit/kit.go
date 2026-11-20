// Package kit checks Docker Sandbox kit v2 specifications: the grammar through
// the kit schema, and the rules between fields that a schema cannot state.
package kit

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
)

// Name is the name of the check.
const Name = "kit"

var specFiles = []string{"spec.yaml", "spec.yml"}

var (
	argumentReference = regexp.MustCompile(`\$\{\{\s*kit\.args\.([A-Za-z_][A-Za-z0-9_-]*)\s*\}\}`)
	quotedName        = regexp.MustCompile(`'([^']+)'`)
	yamlLine          = regexp.MustCompile(`^yaml: line (\d+):`)
)

// Check is the kit check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string { return "Docker Sandbox kit v2 grammar and semantics" }

// DefaultPaths returns the repository root.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks the kit specifications below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	if err := check.DecodeOptions(ctx.Options, &struct{}{}); err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	files, err := ctx.Tree.WalkAll(ctx.Paths, func(p string) bool { return slices.Contains(specFiles, path.Base(p)) })
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		text, err := ctx.Tree.Read(file)
		if err != nil {
			return nil, err
		}
		if err := validate(ctx, report, file, text); err != nil {
			return nil, err
		}
	}
	return report.Findings(), nil
}

func validate(ctx check.Context, report *check.Reporter, file, text string) error {
	var value any
	if err := yaml.Unmarshal([]byte(text), &value); err != nil {
		report.Fail("kit.yaml", file, errorLine(err), "invalid YAML: %v", err)
		return nil
	}
	document, err := json.Marshal(value)
	if err != nil {
		report.Fail("kit.yaml", file, 1, "specification is not representable as JSON: %v", err)
		return nil
	}
	violations, err := ctx.Validator.Validate(schemas.Kit, document)
	if err != nil {
		return err
	}
	for _, v := range violations {
		pointer := v.Pointer
		if pointer == "" {
			pointer = "/"
		}
		report.Fail("kit.schema", file, lineOfKey(text, violationKey(pointer, v.Message)), "kit schema %s %s", pointer, v.Message)
	}
	if len(violations) == 0 {
		if root, ok := value.(map[string]any); ok {
			semantics(report, file, text, root)
		}
	}
	return nil
}

func errorLine(err error) int {
	m := yamlLine.FindStringSubmatch(err.Error())
	if m == nil {
		return 1
	}
	var line int
	_, _ = fmt.Sscanf(m[1], "%d", &line)
	return line
}

// lineOfKey finds the first line that opens the key, or the first line.
func lineOfKey(text, key string) int {
	for i, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), key+":") {
			return i + 1
		}
	}
	return 1
}

// violationKey names the key that a violation concerns: the property a message
// reports as missing, else the last segment of its pointer.
func violationKey(pointer, message string) string {
	if strings.Contains(message, "missing propert") {
		if m := quotedName.FindStringSubmatch(message); m != nil {
			return m[1]
		}
	}
	segments := strings.FieldsFunc(pointer, func(r rune) bool { return r == '/' })
	if len(segments) == 0 {
		return "schemaVersion"
	}
	return strings.NewReplacer("~1", "/", "~0", "~").Replace(segments[len(segments)-1])
}

func record(value any) map[string]any {
	m, _ := value.(map[string]any)
	return m
}

func strs(value any) []string {
	var out []string
	items, _ := value.([]any)
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// A credential reaches a domain only through the network policy, and a format
// carries the secret in exactly one place.
func semantics(report *check.Reporter, file, text string, root map[string]any) {
	allowed := strs(record(record(root["permissions"])["network"])["allow"])
	credentials, _ := root["credentials"].([]any)
	for _, entry := range credentials {
		injections, _ := record(record(entry)["apiKey"])["inject"].([]any)
		for _, injected := range injections {
			fields := record(injected)
			if domain, ok := fields["domain"].(string); ok && !slices.Contains(allowed, domain) {
				report.Fail("kit.domain", file, lineOfKey(text, "domain"), "apiKey injection domain %s must appear in permissions.network.allow", domain)
			}
			if format, ok := fields["format"].(string); ok && strings.Count(format, "%s") != 1 {
				report.Fail("kit.format", file, lineOfKey(text, "format"), "apiKey injection format must contain exactly one %%s")
			}
		}
	}
	args := record(root["args"])
	reported := map[string]bool{}
	for _, m := range argumentReference.FindAllStringSubmatch(text, -1) {
		if _, declared := args[m[1]]; !declared && !reported[m[1]] {
			reported[m[1]] = true
			report.Fail("kit.argument", file, lineOfKey(text, "args"), "kit argument %s is referenced but not declared", m[1])
		}
	}
}
