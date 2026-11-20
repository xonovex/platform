package cmd

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/resolve"
	evalsection "github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/eval"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/shard"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// resolvePath makes a path absolute against the working directory.
func resolvePath(dir, p string) string {
	if strings.HasPrefix(p, "/") {
		return path.Clean(p)
	}
	return path.Join(dir, p)
}

func (d Dependencies) reader() document.Reader {
	return document.Reader{FS: d.FS, Validator: d.Validator}
}

// configRequest names what the eval reads of agentic.json: the eval section,
// from the project file the flag names or the nearest one, with the budget
// flags as the flags layer.
type configRequest struct {
	project, approveDigest string
	maxModelRuns           *int
	maxCostUSD             *float64
}

// resolveConfig merges the eval section of every layer. A project file that
// sets the section passes the trust gate first.
func (d Dependencies) resolveConfig(r configRequest) (merge.Resolved, error) {
	var flags []merge.Flag
	if r.maxModelRuns != nil {
		flags = append(flags, merge.Flag{Section: evalsection.Name, Pointer: "/budgets/maxModelRuns", Value: json.RawMessage(strconv.Itoa(*r.maxModelRuns))})
	}
	if r.maxCostUSD != nil {
		flags = append(flags, merge.Flag{Section: evalsection.Name, Pointer: "/budgets/maxCostUsd", Value: json.RawMessage(strconv.FormatFloat(*r.maxCostUSD, 'f', -1, 64))})
	}
	env := resolve.Env{
		Files: d.ConfigFiles, Places: d.Places, Validator: d.Validator,
		Approvals: d.approvals(), Prompter: d.Prompter, Out: d.Err,
	}
	return resolve.Resolve(env, resolve.Request{
		Sections: []spec.Section{evalsection.Section()}, Start: d.Dir,
		Project: r.project, ApproveDigest: r.approveDigest, Flags: flags,
	})
}

func (d Dependencies) approvals() trust.Ledger {
	return trust.Ledger{Store: d.ConfigFiles, StateDir: d.Places.StateDir}
}

// sourceHarness is the harness whose plugin layout the plugin sources follow:
// the plugins of a repository ship the Claude Code layout, and each leaf stages
// the skills it offers in the layout of its own harness.
const sourceHarness = types.AgentClaude

// loadHarnesses returns the facts of the harnesses with the override files of
// every layer applied: user, then project, then managed. A project override
// passes the trust gate first, because it can change the program a run starts.
func (d Dependencies) loadHarnesses(r configRequest) (agents.Catalog, error) {
	base := agents.Embedded()
	var names []string
	for _, agentType := range base.Types() {
		names = append(names, string(agentType))
	}
	env := shard.Env{Files: d.ConfigFiles, Places: d.Places, Approvals: d.approvals(), Prompter: d.Prompter, Out: d.Err}
	shards, err := shard.Read(env, shard.Request{
		Dir: agents.OverrideDir, Names: names, Kind: agents.OverrideKind,
		Start: d.Dir, Project: r.project, ApproveDigest: r.approveDigest,
	})
	if err != nil {
		return agents.Catalog{}, err
	}
	overrides := make([]agents.Override, len(shards))
	for i, s := range shards {
		overrides[i] = agents.Override{Name: types.AgentType(s.Name), Origin: s.Path, Data: s.Data}
	}
	return base.Apply(d.Validator, overrides...)
}

// loadProject decodes the merged eval section and checks that its default
// profile exists. It returns the directory that relative plugin directories
// resolve against: the directory of the file that set them, else the working
// directory.
func (d Dependencies) loadProject(r configRequest) (model.Project, string, error) {
	resolved, err := d.resolveConfig(r)
	if err != nil {
		return model.Project{}, "", err
	}
	var project model.Project
	if err := json.Unmarshal(resolved.Section(evalsection.Name), &project); err != nil {
		return model.Project{}, "", fmt.Errorf("decode the eval section: %w", err)
	}
	if project.DefaultProfile != "" {
		if _, ok := project.Profiles[project.DefaultProfile]; !ok {
			return model.Project{}, "", fmt.Errorf("the eval section: defaultProfile %s names no profile", project.DefaultProfile)
		}
	}
	dir := d.Dir
	if value, ok := resolved.Source(evalsection.Name, "/pluginDirs"); ok && len(value.Origins) > 0 && value.Origins[0].Path != "" {
		dir = filepath.ToSlash(filepath.Dir(value.Origins[0].Path))
	}
	return project, dir, nil
}

// loadCatalog discovers the plugins in the project's plugin directories and the
// directories the flags add. With neither, it looks in the working directory.
func (d Dependencies) loadCatalog(facts agents.Catalog, project model.Project, projectDir string, extra []string) (catalog.Catalog, error) {
	source, err := facts.Get(sourceHarness)
	if err != nil {
		return catalog.Catalog{}, err
	}
	var roots []string
	for _, dir := range project.PluginDirs {
		roots = append(roots, resolvePath(projectDir, dir))
	}
	for _, dir := range extra {
		roots = append(roots, resolvePath(d.Dir, dir))
	}
	if len(roots) == 0 {
		roots = []string{d.Dir}
	}
	return catalog.Load(d.FS, source.Plugin, roots)
}
