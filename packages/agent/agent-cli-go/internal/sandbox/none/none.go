package none

import (
	"fmt"
	"os"
	"strings"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandboxutil"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
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
func (i *Isolator) Run(cfg *types.SandboxConfig, c types.Contribution) (int, error) {
	if cfg.Network != types.NetworkHost {
		return 1, fmt.Errorf("isolation=none cannot restrict network egress (network=%q); use bwrap or docker", cfg.Network)
	}
	cmd, env := i.hostCommand(cfg, c)
	return sandboxutil.SpawnSandboxInDir(cmd[0], cmd[1:], env, cfg.WorkDir, "host execution", cfg.Verbose)
}

// Command returns the host command (for display / terminal wrappers).
func (i *Isolator) Command(cfg *types.SandboxConfig, c types.Contribution) []string {
	cmd, _ := i.hostCommand(cfg, c)
	return cmd
}

// hostCommand builds the host command and environment, applying the Contribution.
func (i *Isolator) hostCommand(cfg *types.SandboxConfig, c types.Contribution) ([]string, []string) {
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
	cmd = sandboxutil.WrapWithInitCommands(cmd, c.InitCommands)

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
