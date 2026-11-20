package launch

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// flagInput is the command line of this run as the top configuration layer. A
// scalar counts only when the user chose it, because a flag that holds its default
// must not shadow the value of a file.
func flagInput(options Options) []merge.Flag {
	var flags []merge.Flag
	for _, text := range []struct {
		key   runconfig.Key
		value string
	}{
		{runconfig.KeyAgent, options.Agent},
		{runconfig.KeyIsolation, options.Isolation},
		{runconfig.KeyProvision, options.Provision},
		{runconfig.KeyNetwork, options.Network},
	} {
		if options.Explicit.Has(text.key) && text.value != "" {
			flags = append(flags, runconfig.Text(text.key, text.value))
		}
	}
	for _, text := range []struct {
		key   runconfig.Key
		value string
	}{
		{runconfig.KeyProvider, options.Provider},
		{runconfig.KeyImage, options.Image},
	} {
		if text.value != "" {
			flags = append(flags, runconfig.Text(text.key, text.value))
		}
	}
	for _, list := range []struct {
		key   runconfig.Key
		items []string
	}{
		{runconfig.KeyModels, options.Models},
		{runconfig.KeyBindPaths, options.BindPaths},
		{runconfig.KeyRoBindPaths, options.RoBindPaths},
		{runconfig.KeyCustomEnv, options.CustomEnv},
		{runconfig.KeyAllowHosts, options.AllowHosts},
		{runconfig.KeyCredentials, options.Credentials},
	} {
		if len(list.items) > 0 {
			flags = append(flags, runconfig.List(list.key, list.items))
		}
	}
	for _, demand := range []struct {
		key runconfig.Key
		on  bool
	}{
		{runconfig.KeyRequirePinnedProvision, options.Policy.RequirePinnedProvision},
		{runconfig.KeyRequireHostToolsUnreachable, options.Policy.RequireHostToolsUnreachable},
		{runconfig.KeyRequireEgressRestricted, options.Policy.RequireEgressRestricted},
		{runconfig.KeyRequireKernelIsolation, options.Policy.RequireKernelIsolation},
	} {
		if demand.on {
			flags = append(flags, runconfig.Demand(demand.key))
		}
	}
	return flags
}

// applyResolved folds the merged configuration into the options of the run. It
// replaces each list with the merged one, which already holds the entries of the
// command line, and records whether a layer chose an isolation or provision
// axis, so the pinned default yields to it.
func applyResolved(options Options, resolved runconfig.Resolved) Options {
	for key, target := range map[runconfig.Key]*string{
		runconfig.KeyAgent:     &options.Agent,
		runconfig.KeyProvider:  &options.Provider,
		runconfig.KeyIsolation: &options.Isolation,
		runconfig.KeyProvision: &options.Provision,
		runconfig.KeyNetwork:   &options.Network,
		runconfig.KeyImage:     &options.Image,
		runconfig.KeyHomeDir:   &options.HomeDir,
	} {
		if resolved.Has(key) {
			*target = resolved.Text(key)
		}
	}
	for _, key := range []runconfig.Key{runconfig.KeyIsolation, runconfig.KeyProvision} {
		if options.Explicit.Has(key) || resolved.Has(key) {
			options.ExplicitAxes = true
		}
	}
	options.Models = resolved.Items(runconfig.KeyModels)
	options.BindPaths = resolved.Items(runconfig.KeyBindPaths)
	options.RoBindPaths = resolved.Items(runconfig.KeyRoBindPaths)
	options.CustomEnv = resolved.Items(runconfig.KeyCustomEnv)
	options.AllowHosts = resolved.Items(runconfig.KeyAllowHosts)
	options.Credentials = resolved.Items(runconfig.KeyCredentials)
	options.Policy.RequirePinnedProvision = resolved.On(runconfig.KeyRequirePinnedProvision)
	options.Policy.RequireHostToolsUnreachable = resolved.On(runconfig.KeyRequireHostToolsUnreachable)
	options.Policy.RequireEgressRestricted = resolved.On(runconfig.KeyRequireEgressRestricted)
	options.Policy.RequireKernelIsolation = resolved.On(runconfig.KeyRequireKernelIsolation)
	return options
}

// resolution is a run whose options carry every configuration layer and whose plan
// is built from them.
type resolution struct {
	config  runconfig.Resolved
	plan    Plan
	planErr error
}

// configRequest locates the configuration of the options.
func configRequest(options Options) runconfig.Request {
	return runconfig.Request{
		Start: options.WorkDir, Project: options.Project,
		ApproveDigest: options.ApproveDigest, Flags: flagInput(options),
	}
}

// resolve loads the configuration layers, lets the project content through the
// trust gate, merges the layers with the command line and builds the plan. A
// configuration error is returned; a plan error is returned as well unless
// advisory is set, which builds the plan without demanding the guarantees of
// the policy.
func resolve(deps Dependencies, options Options, advisory bool) (resolution, error) {
	cwd, err := deps.Host.Getwd()
	if err != nil {
		return resolution{}, fmt.Errorf("resolve current working directory: %w", err)
	}
	loaded, err := LoadConfig(deps, options)
	if err != nil {
		return resolution{}, fmt.Errorf("resolve configuration: %w", err)
	}
	effective := applyResolved(options, loaded.Resolved)
	effective.Catalog = providers.NewCatalog(loaded.Providers...)
	effective.Harnesses = loaded.Harnesses
	effective.UserProviderDir = loaded.UserProviderDir()
	if effective.ModelOverrides, err = parseModelOverrides(loaded.Resolved.Sourced(runconfig.KeyModels)); err != nil {
		return resolution{}, fmt.Errorf("resolve configuration: %w", err)
	}
	home := effective.HomeDir
	if home == "" {
		if home, err = deps.Host.UserHomeDir(); err != nil {
			return resolution{}, fmt.Errorf("resolve user home directory: %w", err)
		}
	}
	build := NewPlan
	if advisory {
		build = advisoryPlan
	}
	plan, err := build(effective, cwd, home, deps.Sandboxes)
	return resolution{config: loaded.Resolved, plan: plan, planErr: err}, nil
}
