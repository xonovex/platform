package shared

import (
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agentcmd"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
)

// HostExecution builds the command and the environment of an agent that runs on
// the host, under the wrapper of the isolator or none at all: the host
// environment, the provider environment, the custom environment and credential
// values, and the provisioner's environment and tool directories. sandbox tells
// the agent that a boundary surrounds it, which lets it skip its own permission
// prompts, so only an isolator with a boundary passes true.
func HostExecution(host Host, cfg RunConfig, c provision.Contribution, sandbox bool) ([]string, []string, error) {
	if _, err := agents.GetAgent(cfg.Agent.Type); err != nil {
		return nil, nil, err
	}
	var providerCliArgs []string
	if cfg.Provider != nil {
		providerCliArgs = providers.GetProviderCliArgs(cfg.Provider)
	}
	agentEnv, err := agentcmd.BuildProviderEnv(cfg.Agent, cfg.Provider, host.Getenv)
	if err != nil {
		return nil, nil, err
	}
	agentArgs := agents.BuildArgs(cfg.Agent, cfg.AgentArgs, types.AgentExecOptions{Sandbox: sandbox, ProviderCliArgs: providerCliArgs})

	cmd := append([]string{cfg.Agent.Binary}, agentArgs...)
	cmd = WrapWithInitCommands(cmd, c.InitCommands)

	customEnv, err := RunEnv(cfg)
	if err != nil {
		return nil, nil, err
	}
	env := envutil.MergeEnvMaps(envutil.ParseEnv(host.Environ()), agentEnv, customEnv, c.Env)
	// Prepend the contribution's tool dirs; the host PATH stays behind them.
	if len(c.PathEntries) > 0 {
		path := strings.Join(c.PathEntries, ":")
		if hostPath := host.Getenv("PATH"); hostPath != "" {
			path += ":" + hostPath
		}
		env["PATH"] = path
	}
	return cmd, envutil.EnvMapToSlice(env), nil
}
