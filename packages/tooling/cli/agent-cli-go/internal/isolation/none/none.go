// Package none is the isolation=none leaf: direct host execution, no namespace
// boundary. It is the single source of the direct-exec command/env.
package none

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agentcmd"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
)

// Isolator runs the agent directly on the host — no namespace boundary. It
// applies the provisioner's Contribution (PATH/env/init) but cannot hide host
// tools or restrict network egress. The agent runs WITHOUT the sandbox
// permission bypass: there is no isolation boundary, so it keeps its normal host
// permission prompts.
type Isolator struct {
	host isoshared.Host
	log  logging.Logger
}

// NewIsolator creates a host (no-isolation) isolator over host, writing through log.
func NewIsolator(host isoshared.Host, log logging.Logger) *Isolator {
	return &Isolator{host: host, log: log}
}

// Available always reports true — direct host execution is always possible.
func (i *Isolator) Available() (bool, error) { return true, nil }

// HidesHost reports false: host execution always exposes the host.
func (i *Isolator) HidesHost(bool, string) bool { return false }

// PinnedProvision mirrors the selected provisioner's immutability because host
// execution has no image input.
func (i *Isolator) PinnedProvision(_ provision.ProvisionMethod, provisionerPinned bool, _ string) bool {
	return provisionerPinned
}

// KernelIsolated reports false: there is no namespace or kernel boundary.
func (i *Isolator) KernelIsolated(string) bool { return false }

// ConfinesNetwork reports false: with no namespace, egress stays unrestricted.
func (i *Isolator) ConfinesNetwork() bool { return false }

// CheckAgent requires the agent executable on the PATH of the host.
func (i *Isolator) CheckAgent(cfg isoshared.RunConfig) error {
	if _, err := i.host.LookPath(cfg.Agent.Binary); err != nil {
		return fmt.Errorf("agent executable %q is not available on host PATH: %w", cfg.Agent.Binary, err)
	}
	return nil
}

// Run executes the agent on the host. It fails CLOSED when a network restriction
// is requested: with no namespace there is no way to enforce restricted egress.
func (i *Isolator) Run(cfg isoshared.RunConfig, c provision.Contribution) (int, error) {
	cmd, env, err := i.hostCommand(cfg, c)
	if err != nil {
		return 1, err
	}
	return isoshared.SpawnSandbox(i.host, i.log, cmd[0], cmd[1:], env, cfg.WorkDir, "host execution")
}

// Command returns the host command (for display / terminal wrappers).
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	cmd, _, err := i.hostCommand(cfg, c)
	return cmd, err
}

// TerminalCommand returns the host command AND its full resolved environment
// (provider tokens, custom env, and the provisioner's PATH/env). Unlike
// bwrap/docker, host execution bakes nothing into the command, so the wrapper
// must carry this environment for the agent to see its toolchain.
func (i *Isolator) TerminalCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	return i.hostCommand(cfg, c)
}

// hostCommand builds the host command and environment, applying the Contribution.
func (i *Isolator) hostCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	if cfg.Network != netshared.ModeHost {
		return nil, nil, fmt.Errorf("isolation=none cannot restrict network egress (network=%q); use bwrap or docker", cfg.Network)
	}
	if _, err := isoshared.ResolveDirectory(i.host, cfg.WorkDir, "work directory"); err != nil {
		return nil, nil, err
	}
	behavior, err := agents.Lookup(cfg.Agent.Type)
	if err != nil {
		return nil, nil, err
	}
	var providerCliArgs []string
	if cfg.Provider != nil {
		providerCliArgs = providers.GetProviderCliArgs(cfg.Provider)
	}
	agentEnv, err := agentcmd.BuildProviderEnv(cfg.Agent, cfg.Provider, i.host.Getenv)
	if err != nil {
		return nil, nil, err
	}
	agentArgs := behavior.BuildArgs(cfg.AgentArgs, types.AgentExecOptions{Sandbox: false, ProviderCliArgs: providerCliArgs})

	cmd := append([]string{cfg.Agent.Binary}, agentArgs...)
	cmd = isoshared.WrapWithInitCommands(cmd, c.InitCommands)

	customEnv, err := envutil.ParseCustomEnv(cfg.CustomEnv)
	if err != nil {
		return nil, nil, err
	}
	env := envutil.MergeEnvMaps(envutil.ParseEnv(i.host.Environ()), agentEnv, customEnv, c.Env)
	// Prepend the contribution's tool dirs; host execution always exposes the
	// host PATH (it cannot hide host tools), so HostPassthrough is implied.
	if len(c.PathEntries) > 0 {
		path := strings.Join(c.PathEntries, ":")
		if host := i.host.Getenv("PATH"); host != "" {
			path += ":" + host
		}
		env["PATH"] = path
	}

	return cmd, envutil.EnvMapToSlice(env), nil
}
