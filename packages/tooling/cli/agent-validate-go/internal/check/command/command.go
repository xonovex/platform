// Package command checks the frontmatter, argument hints and body length of
// slash command files.
package command

import (
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/component"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

// Name is the name of the check.
const Name = "command"

// Options tune the check.
type Options struct {
	// MaxBodyLines is the body size budget of one command. A command is an entry
	// point: it states the argument contract and hands off to a skill or an
	// agent that holds the procedure, so a body past the budget holds steps that
	// belong in a skill.
	MaxBodyLines int `json:"maxBodyLines"`
}

func defaults() Options { return Options{MaxBodyLines: 80} }

var spec = component.Spec{
	Prefix:   "command",
	Noun:     "a command",
	Known:    []string{"allowed-tools", "argument-hint", "description", "disable-model-invocation", "model"},
	Required: []string{"description"},
	Body:     "body: the command tells the model nothing",
	Fields: []component.Field{
		{Key: "model", Valid: component.IsString, Message: "model must be a string"},
		{Key: "disable-model-invocation", Valid: component.IsBool, Message: "disable-model-invocation must be true or false"},
		{Key: "argument-hint", Valid: nonEmptyString, Message: "argument-hint must be a non-empty string"},
	},
}

func nonEmptyString(value any) bool {
	s, ok := value.(string)
	return ok && strings.TrimSpace(s) != ""
}

var (
	toolName     = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(?:\(.*\))?$`)
	hintToken    = regexp.MustCompile(`<[^<>]+>|\[[^\[\]]+\]`)
	positional   = regexp.MustCompile(`\$(\d+)`)
	allArguments = regexp.MustCompile(`\$ARGUMENTS`)
)

// Check is the command check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string { return "command frontmatter, argument hints and body length" }

// DefaultPaths returns the repository root.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks the command files below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	files, err := component.Files(ctx.Tree, ctx.Paths, "commands")
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
		checkAllowedTools(report, file, doc)
		checkArguments(report, file, doc)
		checkBodyLength(report, file, doc, options.MaxBodyLines)
	}
	return report.Findings(), nil
}

func checkAllowedTools(report *check.Reporter, file string, doc document.Document) {
	value, present := doc.Fields["allowed-tools"]
	if !present {
		return
	}
	line := document.KeyLine(doc.Source, "allowed-tools")
	if !component.IsStrings(value) {
		report.Fail("command.allowed-tools", file, line, "allowed-tools must be a non-empty string or list of strings")
		return
	}
	seen := map[string]bool{}
	for _, tool := range component.Words(value) {
		switch {
		case tool == "":
			report.Fail("command.allowed-tools", file, line, "allowed-tools has an empty entry")
		case seen[tool]:
			report.Fail("command.allowed-tools", file, line, "allowed-tools names '%s' more than once", tool)
		case strings.Count(tool, "(") != strings.Count(tool, ")"):
			report.Fail("command.allowed-tools", file, line, "allowed-tools entry '%s' has an unbalanced '('", tool)
		case !toolName.MatchString(tool):
			report.Fail("command.allowed-tools", file, line, "allowed-tools entry '%s' is not a tool name, expected Tool or Tool(specifier)", tool)
		}
		seen[tool] = true
	}
}

// The hint promises arguments and the body must use them: a hint with no
// reference leaves the typed arguments unread, and a reference past the hint
// reads an argument nobody was told to give.
func checkArguments(report *check.Reporter, file string, doc document.Document) {
	hint, ok := doc.String("argument-hint")
	if !ok || strings.TrimSpace(hint) == "" {
		return
	}
	tokens := len(hintToken.FindAllString(hint, -1))
	if tokens == 0 {
		return
	}
	line := document.KeyLine(doc.Source, "argument-hint")
	body := markdown.MaskFences(doc.Body)
	referenced := allArguments.MatchString(body)
	highest := 0
	for _, m := range positional.FindAllStringSubmatch(body, -1) {
		referenced = true
		if n, err := strconv.Atoi(m[1]); err == nil && n > highest {
			highest = n
		}
	}
	if !referenced {
		report.Fail("command.argument-unused", file, line, "argument-hint declares %d argument token(s) but the body never refers to it", tokens)
	}
	if highest > tokens {
		report.Fail("command.argument-excess", file, line, "body refers to $%d but argument-hint describes at most %d argument token(s)", highest, tokens)
	}
}

func checkBodyLength(report *check.Reporter, file string, doc document.Document, budget int) {
	lines := strings.Split(strings.TrimRight(doc.Body, "\n"), "\n")
	if len(lines) > budget {
		report.Fail("command.body-length", file, doc.BodyLine+budget, "%s body is %d lines, over the %d line budget: a command is an entry point, so move the steps into a skill and name it",
			path.Base(file), len(lines), budget)
	}
}
