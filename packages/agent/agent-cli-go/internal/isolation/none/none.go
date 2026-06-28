// Package none is the isolation=none leaf: direct host execution, no namespace
// boundary. It is the single source of the direct-exec command/env.
package none

import (
	"fmt"
	"os"
	"strings"

	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// Isolator runs the agent directly on the host — no namespace boundary. It
// applies the provisioner's Contribution (PATH/env/init) but cannot hide host
// tools or restrict network egress. The agent runs WITHOUT the sandbox
// permission bypass: there is no isolation boundary, so it keeps its normal host
// permission prompts.
type Isolator struct{}

// NewIsolator creates a host (no-isolation) isolator.
func NewIsolator() *Isolator { return &Isolator{} }

// Available always reports true — direct host execution is always possible.
func (i *Isolator) Available() (bool, error) { return true, nil }

// HidesHost reports false: host execution always exposes the host.
func (i *Isolator) HidesHost(bool, string) bool { return false }

// KernelIsolated reports false: there is no namespace or kernel boundary.
func (i *Isolator) KernelIsolated(string) bool { return false }

// Run executes the agent on the host. It fails CLOSED when a network restriction
// is requested: with no namespace there is no way to unshare or proxy egress.
func (i *Isolator) Run(cfg isoshared.RunConfig, c provision.Contribution) (int, error) {
	if cfg.Network != netshared.ModeHost {
		return 1, fmt.Errorf("isolation=none cannot restrict network egress (network=%q); use bwrap or docker", cfg.Network)
	}
	cmd, env := i.hostCommand(cfg, c)
	return isoshared.SpawnSandboxInDir(cmd[0], cmd[1:], env, cfg.WorkDir, "host execution", cfg.Verbose)
}

// Command returns the host command (for display / terminal wrappers).
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) []string {
	cmd, _ := i.hostCommand(cfg, c)
	return cmd
}

// TerminalCommand returns the host command AND its full resolved environment
// (provider tokens, custom env, and the provisioner's PATH/env). Unlike
// bwrap/docker, host execution bakes nothing into the command, so the wrapper
// must carry this environment for the agent to see its toolchain.
func (i *Isolator) TerminalCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string) {
	return i.hostCommand(cfg, c)
}

// hostCommand builds the host command and environment, applying the Contribution.
func (i *Isolator) hostCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string) {
	var providerCliArgs []string
	var providerEnv map[string]string
	if cfg.Provider != nil {
		providerCliArgs = providers.GetProviderCliArgs(cfg.Provider)
		providerEnv, _ = providers.BuildProviderEnv(cfg.Provider)
	}
	execOpts := types.AgentExecOptions{Sandbox: false, ProviderCliArgs: providerCliArgs}

	var agentArgs []string
	var agentEnv map[string]string
	switch cfg.Agent.Type {
	case types.AgentClaude:
		agentArgs = agents.BuildClaudeArgs(cfg.AgentArgs, execOpts)
		agentEnv = agents.BuildClaudeEnv(providerEnv)
	case types.AgentOpencode:
		agentArgs = agents.BuildOpencodeArgs(cfg.AgentArgs, execOpts)
		agentEnv = agents.BuildOpencodeEnv(providerEnv)
	}

	cmd := append([]string{cfg.Agent.Binary}, agentArgs...)
	cmd = isoshared.WrapWithInitCommands(cmd, c.InitCommands)

	env := os.Environ()
	for k, v := range agentEnv {
		env = append(env, k+"="+v)
	}
	env = append(env, cfg.CustomEnv...)
	for k, v := range c.Env {
		env = append(env, k+"="+v)
	}
	// Prepend the contribution's tool dirs; host execution always exposes the
	// host PATH (it cannot hide host tools), so HostPassthrough is implied.
	if len(c.PathEntries) > 0 {
		path := strings.Join(c.PathEntries, ":")
		if host := os.Getenv("PATH"); host != "" {
			path += ":" + host
		}
		env = append(env, "PATH="+path)
	}

	return cmd, env
}
