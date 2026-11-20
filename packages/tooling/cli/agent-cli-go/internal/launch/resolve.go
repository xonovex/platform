package launch

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
)

// flagsOrigin is the origin of the command line in the provenance.
const flagsOrigin = "command line"

// flagInput is the command line of this run as the top configuration layer. A
// scalar counts only when the user chose it, because a flag that holds its default
// must not shadow the value of a file.
func flagInput(options Options) config.Input {
	entries := config.Entries{}
	for key, value := range map[config.Key]string{
		config.KeyAgent:     options.Agent,
		config.KeyIsolation: options.Isolation,
		config.KeyProvision: options.Provision,
		config.KeyNetwork:   options.Network,
	} {
		if options.Explicit.Has(key) && value != "" {
			entries[key] = config.Value{Text: value}
		}
	}
	if options.Provider != "" {
		entries[config.KeyProvider] = config.Value{Text: options.Provider}
	}
	for key, items := range map[config.Key][]string{
		config.KeyModels:      options.Models,
		config.KeyBindPaths:   options.BindPaths,
		config.KeyRoBindPaths: options.RoBindPaths,
		config.KeyCustomEnv:   options.CustomEnv,
		config.KeyAllowHosts:  options.AllowHosts,
		config.KeyCredentials: options.Credentials,
	} {
		if len(items) > 0 {
			entries[key] = config.Value{Items: items}
		}
	}
	for key, on := range map[config.Key]bool{
		config.KeyRequirePinnedProvision:      options.Policy.RequirePinnedProvision,
		config.KeyRequireHostToolsUnreachable: options.Policy.RequireHostToolsUnreachable,
		config.KeyRequireEgressRestricted:     options.Policy.RequireEgressRestricted,
		config.KeyRequireKernelIsolation:      options.Policy.RequireKernelIsolation,
	} {
		if on {
			entries[key] = config.Value{On: true}
		}
	}
	return config.Input{Layer: config.LayerFlags, Origin: flagsOrigin, Entries: entries}
}

// applyResolved folds the merged configuration into the options of the run. It
// replaces each list with the merged one, which already holds the entries of the
// command line, and records whether a layer chose an isolation or provision
// axis, so the pinned default yields to it.
func applyResolved(options Options, resolved config.Resolved) Options {
	for key, target := range map[config.Key]*string{
		config.KeyAgent:     &options.Agent,
		config.KeyProvider:  &options.Provider,
		config.KeyIsolation: &options.Isolation,
		config.KeyProvision: &options.Provision,
		config.KeyNetwork:   &options.Network,
		config.KeyHomeDir:   &options.HomeDir,
	} {
		if resolved.Has(key) {
			*target = resolved.Text(key)
		}
	}
	for _, key := range []config.Key{config.KeyIsolation, config.KeyProvision} {
		if options.Explicit.Has(key) || resolved.Has(key) {
			options.ExplicitAxes = true
		}
	}
	options.Models = resolved.Items(config.KeyModels)
	options.BindPaths = resolved.Items(config.KeyBindPaths)
	options.RoBindPaths = resolved.Items(config.KeyRoBindPaths)
	options.CustomEnv = resolved.Items(config.KeyCustomEnv)
	options.AllowHosts = resolved.Items(config.KeyAllowHosts)
	options.Credentials = resolved.Items(config.KeyCredentials)
	options.Policy.RequirePinnedProvision = resolved.On(config.KeyRequirePinnedProvision)
	options.Policy.RequireHostToolsUnreachable = resolved.On(config.KeyRequireHostToolsUnreachable)
	options.Policy.RequireEgressRestricted = resolved.On(config.KeyRequireEgressRestricted)
	options.Policy.RequireKernelIsolation = resolved.On(config.KeyRequireKernelIsolation)
	return options
}

// resolution is a run whose options carry every configuration layer and whose plan
// is built from them.
type resolution struct {
	config  config.Resolved
	plan    Plan
	planErr error
}

// resolve loads the configuration layers, merges them with the command line and
// builds the plan. A configuration error is returned; a plan error is returned
// as well unless advisory is set, which builds the plan without demanding the
// guarantees of the policy.
func resolve(deps Dependencies, options Options, advisory bool) (resolution, error) {
	cwd, err := deps.Host.Getwd()
	if err != nil {
		return resolution{}, fmt.Errorf("resolve current working directory: %w", err)
	}
	inputs, err := config.LoadLayers(deps.Host, options.Config, absolute(cwd, options.WorkDir))
	if err != nil {
		return resolution{}, fmt.Errorf("failed to load config: %w", err)
	}
	merged, err := config.Merge(append(inputs, flagInput(options)))
	if err != nil {
		return resolution{}, fmt.Errorf("resolve configuration: %w", err)
	}
	effective := applyResolved(options, merged)
	set, err := config.LoadProviders(deps.Host, options.Config, absolute(cwd, options.WorkDir))
	if err != nil {
		return resolution{}, fmt.Errorf("failed to load providers: %w", err)
	}
	effective.Catalog = providers.NewCatalog(set.Documents...)
	effective.UserProviderDir = userProviderDir(set.Dirs)
	if effective.ModelOverrides, err = parseModelOverrides(merged.Sourced(config.KeyModels)); err != nil {
		return resolution{}, fmt.Errorf("resolve configuration: %w", err)
	}
	material := providers.Material{Documents: set.Project, Models: config.ProjectModels(inputs)}
	if needsApproval(effective, material) {
		approved, err := approvedDigest(deps.Approvals, material.Digest())
		if err != nil {
			return resolution{}, fmt.Errorf("read the approvals: %w", err)
		}
		if !approved {
			return resolution{}, &ApprovalRequiredError{Material: material}
		}
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
	return resolution{config: merged, plan: plan, planErr: err}, nil
}

// userProviderDir is the provider directory of the user layer, or empty when
// no user layer has one.
func userProviderDir(dirs []config.ProviderDir) string {
	path := ""
	for _, dir := range dirs {
		if dir.Layer == config.LayerUser {
			path = dir.Path
		}
	}
	return path
}

// approvedDigest asks the approvals whether the digest is approved; a build
// without approvals approves nothing.
func approvedDigest(approvals Approvals, digest string) (bool, error) {
	if approvals == nil {
		return false, nil
	}
	return approvals.Approved(ApprovalKind, digest)
}
