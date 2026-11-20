// Package agent checks the frontmatter and prompt of agent definition files.
package agent

import (
	"path"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/component"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
)

// Name is the name of the check.
const Name = "agent"

var identifier = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var spec = component.Spec{
	Prefix: "agent",
	Noun:   "an agent",
	Known: []string{
		"background", "color", "description", "disallowedTools", "effort", "hooks",
		"initialPrompt", "isolation", "maxTurns", "mcpServers", "memory", "model",
		"name", "permissionMode", "skills", "tools",
	},
	Required: []string{"name", "description"},
	Body:     "agent prompt",
	Fields: []component.Field{
		{Key: "tools", Valid: component.IsStrings, Message: "tools must be a non-empty string or list of strings"},
		{Key: "disallowedTools", Valid: component.IsStrings, Message: "disallowedTools must be a non-empty string or list of strings"},
		{Key: "skills", Valid: component.IsStrings, Message: "skills must be a non-empty string or list of strings"},
		{Key: "model", Valid: component.IsString, Message: "model must be a string"},
		{Key: "permissionMode", Valid: component.OneOf("default", "acceptEdits", "auto", "dontAsk", "bypassPermissions", "plan"), Message: "permissionMode has an unsupported value"},
		{Key: "maxTurns", Valid: component.IsPositiveInteger, Message: "maxTurns must be a positive integer"},
		{Key: "background", Valid: component.IsBool, Message: "background must be true or false"},
		{Key: "memory", Valid: component.OneOf("user", "project", "local"), Message: "memory must be user, project or local"},
		{Key: "effort", Valid: component.OneOf("low", "medium", "high", "xhigh", "max"), Message: "effort has an unsupported value"},
		{Key: "isolation", Valid: component.OneOf("worktree"), Message: "isolation must be worktree"},
		{Key: "color", Valid: component.OneOf("red", "blue", "green", "yellow", "purple", "orange", "pink", "cyan"), Message: "color has an unsupported value"},
		{Key: "initialPrompt", Valid: component.IsString, Message: "initialPrompt must be a string"},
		{Key: "mcpServers", Valid: component.IsListOrMapping, Message: "mcpServers must be a list or mapping"},
		{Key: "hooks", Valid: component.IsMapping, Message: "hooks must be a mapping"},
	},
}

// Check is the agent check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string { return "agent frontmatter and prompt bodies" }

// DefaultPaths returns the repository root.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks the agent files below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	if err := check.DecodeOptions(ctx.Options, &struct{}{}); err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	files, err := component.Files(ctx.Tree, ctx.Paths, "agents")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		source, err := ctx.Tree.Read(file)
		if err != nil {
			return nil, err
		}
		doc, ok := spec.Check(report, file, source)
		if !ok {
			continue
		}
		name, isText := doc.String("name")
		if !isText || strings.TrimSpace(name) == "" {
			continue
		}
		if !identifier.MatchString(name) {
			report.Fail("agent.name-format", file, document.KeyLine(doc.Source, "name"), "name must contain lowercase letters, digits and single hyphens")
		} else if stem := strings.TrimSuffix(path.Base(file), ".md"); stem != name {
			report.Fail("agent.name-file", file, document.KeyLine(doc.Source, "name"), "name '%s' differs from the file name '%s': the harness addresses the agent by name", name, stem)
		}
	}
	return report.Findings(), nil
}
