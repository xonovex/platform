// Package repository checks the files that every repository of the platform
// keeps: the AGENTS.md and CLAUDE.md pairs, the claims of the README, and the
// safety of the release workflow.
package repository

import (
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
)

// Name is the name of the check.
const Name = "repository"

// ClaudePointer is the line that a CLAUDE.md holds to load its sibling.
const ClaudePointer = "See @AGENTS.md for complete documentation."

// Options tune the check.
type Options struct {
	// Packages is the root of the package tree whose groups each need an
	// AGENTS.md. An empty value switches the rule off.
	Packages string `json:"packages"`
	// Fixtures names directories of test data, such as eval fixtures, whose
	// AGENTS.md files are inputs to a test and need no CLAUDE.md.
	Fixtures []string `json:"fixtures"`
	// Readme is the README whose claims are checked. An empty value switches the
	// rule off.
	Readme string `json:"readme"`
	// ForbiddenClaims are case-insensitive patterns the README must not match.
	ForbiddenClaims []string `json:"forbiddenClaims"`
	// Workflow is the release workflow. An empty value switches the rule off.
	Workflow string `json:"workflow"`
}

func defaults() Options {
	return Options{
		Packages: "packages",
		Fixtures: []string{"evals"},
		Readme:   "README.md",
		ForbiddenClaims: []string{
			`all harnesses (?:have|support|provide)`,
			`skills? (?:are|is|provide|provides) enforcement`,
			`install(?:ing|ed)? (?:a )?skills? (?:enforces|proves)`,
			`(?:provides?|ensures?|establishes?|achieves?) automatic compliance`,
			`silently launch(?:es|ing)? (?:a |an )?(?:child )?agent`,
		},
		Workflow: ".github/workflows/release.yml",
	}
}

var (
	pushTrigger   = regexp.MustCompile(`(?m)^\s{2}push:\s*$`)
	dispatchInput = regexp.MustCompile(`inputs\.dry_run\s*!=\s*true`)
	stepEnd       = regexp.MustCompile(`\n\s+- (?:name:|uses:)`)
	publishStart  = regexp.MustCompile(`\n\s+- name: Publish\n`)
	dryRunStart   = regexp.MustCompile(`\n\s+- name: Publish \(dry run\)\n`)
	pullRequestIf = regexp.MustCompile(`if:\s*github\.event_name == 'pull_request'`)
	dispatchIf    = regexp.MustCompile(`if:\s*github\.event_name == 'workflow_dispatch'`)
)

// Check is the repository check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "AGENTS.md and CLAUDE.md pairs, README claims and release workflow safety"
}

// DefaultPaths returns the paths the AGENTS.md pairing rule walks.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks the repository.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	if err := checkInstructionPairs(ctx, report, options); err != nil {
		return nil, err
	}
	if options.Readme != "" {
		if err := checkClaims(ctx, report, options); err != nil {
			return nil, err
		}
	}
	if options.Workflow != "" {
		if err := checkWorkflow(ctx, report, options); err != nil {
			return nil, err
		}
	}
	return report.Findings(), nil
}

// Every AGENTS.md carries a CLAUDE.md beside it, because Claude Code reads only
// CLAUDE.md; every package group documents itself.
func checkInstructionPairs(ctx check.Context, report *check.Reporter, options Options) error {
	files, err := ctx.Tree.WalkAll(ctx.Paths, func(p string) bool {
		if path.Base(p) != "AGENTS.md" {
			return false
		}
		for _, segment := range strings.Split(path.Dir(p), "/") {
			if slices.Contains(options.Fixtures, segment) {
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}
	for _, file := range files {
		dir := path.Dir(file)
		claude := path.Join(dir, "CLAUDE.md")
		text, found, err := ctx.Tree.Optional(claude)
		if err != nil {
			return err
		}
		switch {
		case !found:
			report.Fail("repository.claude-missing", file, 1, "%s has no CLAUDE.md beside it: Claude Code reads only CLAUDE.md, so the guidance never loads", file)
		case !strings.Contains(text, "@AGENTS.md"):
			report.Fail("repository.claude-pointer", claude, 1, "%s must point at its sibling with \"%s\"", claude, ClaudePointer)
		}
	}
	if options.Packages == "" {
		return nil
	}
	groups, err := groups(ctx, options.Packages)
	if err != nil {
		return err
	}
	for _, group := range groups {
		if !ctx.Tree.IsFile(path.Join(group, "AGENTS.md")) {
			report.Fail("repository.group-undocumented", group, 1, "%s must document the group in AGENTS.md", group)
		}
	}
	return nil
}

// groups returns the directories of the package tree that hold packages rather
// than being one. A directory with a package.json or a moon.yml is a package.
func groups(ctx check.Context, root string) ([]string, error) {
	var out []string
	var visit func(dir string) error
	visit = func(dir string) error {
		if ctx.Tree.IsFile(path.Join(dir, "package.json")) || ctx.Tree.IsFile(path.Join(dir, "moon.yml")) {
			return nil
		}
		entries, err := ctx.Tree.List(dir)
		if err != nil {
			return err
		}
		out = append(out, dir)
		for _, e := range entries {
			if e.IsDir && e.Name != "node_modules" && !strings.HasPrefix(e.Name, ".") {
				if err := visit(path.Join(dir, e.Name)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if !ctx.Tree.IsDir(root) {
		return nil, nil
	}
	if err := visit(root); err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

func checkClaims(ctx check.Context, report *check.Reporter, options Options) error {
	text, found, err := ctx.Tree.Optional(options.Readme)
	if err != nil {
		return err
	}
	if !found {
		report.Fail("repository.readme-missing", options.Readme, 1, "%s does not exist", options.Readme)
		return nil
	}
	for _, claim := range options.ForbiddenClaims {
		pattern, err := regexp.Compile("(?i)" + claim)
		if err != nil {
			return err
		}
		if loc := pattern.FindStringIndex(text); loc != nil {
			report.Fail("repository.readme-claim", options.Readme, strings.Count(text[:loc[0]], "\n")+1, "root README rejects /%s/i", claim)
		}
	}
	return nil
}

// stepBody returns the text of the workflow step that starts at the pattern, up
// to the next step.
func stepBody(text string, start *regexp.Regexp) (string, bool) {
	loc := start.FindStringIndex(text)
	if loc == nil {
		return "", false
	}
	rest := text[loc[1]:]
	if end := stepEnd.FindStringIndex(rest); end != nil {
		rest = rest[:end[0]]
	}
	return rest, true
}

// The release publishes from a reviewed pull request only: never from a pushed
// tag, and never on the authority of a dispatch input.
func checkWorkflow(ctx check.Context, report *check.Reporter, options Options) error {
	text, found, err := ctx.Tree.Optional(options.Workflow)
	if err != nil {
		return err
	}
	if !found {
		report.Fail("repository.workflow-missing", options.Workflow, 1, "%s does not exist", options.Workflow)
		return nil
	}
	if pushTrigger.MatchString(text) || strings.Contains(text, "refs/tags/") {
		report.Fail("repository.workflow-tags", options.Workflow, 1, "release workflow must not publish from pushed tags")
	}
	if dispatchInput.MatchString(text) {
		report.Fail("repository.workflow-dispatch-input", options.Workflow, 1, "release workflow must not derive publish authority from a dispatch input")
	}
	if body, ok := stepBody(text, publishStart); !ok || !pullRequestIf.MatchString(body) {
		report.Fail("repository.workflow-publish", options.Workflow, 1, "publish step must be restricted to the pull_request event")
	}
	if body, ok := stepBody(text, dryRunStart); !ok || !dispatchIf.MatchString(body) {
		report.Fail("repository.workflow-dry-run", options.Workflow, 1, "manual dispatch must be restricted to the dry-run publish step")
	}
	return nil
}
