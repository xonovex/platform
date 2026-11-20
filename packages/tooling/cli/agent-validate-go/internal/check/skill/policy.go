package skill

import (
	"path"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

type harnessPattern struct {
	pattern *regexp.Regexp
	label   string
}

// A skill runs under every agent harness, so the names that one vendor's
// harness gives its tools, models, paths and files do not belong in it.
var harnessPatterns = []harnessPattern{
	{regexp.MustCompile(`\bEnterPlanMode\b`), "Claude Code tool name"},
	{regexp.MustCompile(`\bExitPlanMode\b`), "Claude Code tool name"},
	{regexp.MustCompile(`\bAskUserQuestion\b`), "Claude Code tool name"},
	{regexp.MustCompile(`\bTodoWrite\b`), "Claude Code tool name"},
	{regexp.MustCompile(`\bsubagent_type\b`), "Claude Code parameter"},
	{regexp.MustCompile(`\bmodel=haiku\b`), "Vendor model id"},
	{regexp.MustCompile(`\bmodel=sonnet\b`), "Vendor model id"},
	{regexp.MustCompile(`\bmodel=opus\b`), "Vendor model id"},
	{regexp.MustCompile(`\.claude/skills/`), "Vendor-namespaced path"},
	{regexp.MustCompile(`\.claude/commands/`), "Vendor-namespaced path"},
	{regexp.MustCompile(`\bclaude:`), "Vendor-prefixed frontmatter key"},
	{regexp.MustCompile(`\bCLAUDE\.md\b`), "Vendor-specific filename (use AGENTS.md)"},
}

func checkHarnessNeutrality(report *check.Reporter, doc document.Document, file string) {
	for _, hp := range harnessPatterns {
		for i, line := range markdown.Lines(doc.Body) {
			if hp.pattern.MatchString(line) {
				report.Fail("skill.harness-neutrality", file, doc.BodyLine+i, "harness-neutrality: %s: %s", hp.label, excerpt(line))
			}
		}
	}
}

func excerpt(line string) string {
	line = strings.TrimSpace(line)
	if runes := []rune(line); len(runes) > 80 {
		return string(runes[:80])
	}
	return line
}

type credentialPattern struct {
	pattern *regexp.Regexp
	reason  string
}

// A reader copies an example as written, so an example that leaks a secret into
// the shell history or a file teaches the leak.
var credentialPatterns = []credentialPattern{
	{regexp.MustCompile(`(?i)\becho\s+(?:["']?\$[A-Z_]*(?:TOKEN|KEY|PAT|SECRET|PASSWORD)["']?|<token>)\s*\|`), "pipes a secret through echo"},
	{regexp.MustCompile(`(?i)<\s*token\.txt\b`), "reads a secret from token.txt"},
	{regexp.MustCompile(`(?i)\bexport\s+[A-Z_]*(?:TOKEN|KEY|PAT|SECRET|PASSWORD)\s*=\s*["']?[^$"'\s]\S*`), "assigns a secret in an export command"},
}

func checkCredentials(report *check.Reporter, file, text string, references []reference) {
	scan := func(label, at, body string) {
		for _, cp := range credentialPatterns {
			if loc := cp.pattern.FindStringIndex(body); loc != nil {
				report.Fail("skill.credentials", at, markdown.LineOf(body, loc[0]), "credentials: %s %s", label, cp.reason)
			}
		}
	}
	scan("SKILL.md", file, text)
	for _, r := range topLevel(references) {
		scan("references/"+r.name, r.path, r.text)
	}
}

func checkImplementationFiles(ctx check.Context, report *check.Reporter, options Options, doc document.Document, file, dir string) error {
	forbidden, err := ctx.Tree.Walk(dir, func(p string) bool {
		for _, part := range strings.Split(strings.TrimPrefix(p, dir+"/"), "/") {
			if strings.HasPrefix(part, ".") || part == "__pycache__" {
				return false
			}
		}
		for _, ext := range options.ForbiddenExtensions {
			if strings.HasSuffix(p, ext) {
				return true
			}
		}
		return false
	})
	if err != nil {
		return err
	}
	if len(forbidden) > 0 {
		names := make([]string, len(forbidden))
		for i, f := range forbidden {
			names[i] = strings.TrimPrefix(f, dir+"/")
		}
		report.Fail("skill.implementation-files", file, 1, "catalog: skills must not contain TypeScript or MJS implementation files: %s", strings.Join(names, ", "))
	}
	scripts, err := ctx.Tree.Walk(path.Join(dir, "scripts"), func(p string) bool {
		for _, part := range strings.Split(strings.TrimPrefix(p, path.Join(dir, "scripts")+"/"), "/") {
			if strings.HasPrefix(part, ".") || part == "__pycache__" {
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}
	if len(scripts) == 0 {
		return nil
	}
	if text, _ := doc.String("compatibility"); strings.TrimSpace(text) == "" {
		report.Fail("skill.scripts-compatibility", file, 1, "catalog: scripted skills need compatibility runtime and network metadata")
	}
	if text, _ := doc.String("allowed-tools"); strings.TrimSpace(text) == "" {
		report.Fail("skill.scripts-allowed-tools", file, 1, "catalog: scripted skills need a non-empty allowed-tools policy")
	}
	return nil
}
