package executor

import (
	"os"
	"os/exec"
	"strings"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/agents"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/providers"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/types"
	"github.com/xonovex/platform/packages/lib/core-go/pkg/scriptlib"
)

// Options for agent execution
type Options struct {
	Agent     *types.AgentConfig
	Provider  *types.ModelProvider
	Args      []string
	Cwd       string
	Verbose   bool
	Sandbox   bool
	CustomEnv []string
}

// Execute runs an agent with the specified options
func Execute(opts Options) (int, error) {
	if opts.Verbose {
		scriptlib.LogInfo("Using agent: " + opts.Agent.DisplayName)
		if opts.Provider != nil {
			scriptlib.LogInfo("Using provider: " + opts.Provider.DisplayName)
		}
	}

	// Build provider environment
	var providerEnv map[string]string
	var providerCliArgs []string

	if opts.Provider != nil {
		var err error
		providerEnv, err = providers.BuildProviderEnv(opts.Provider)
		if err != nil {
			return 1, err
		}
		providerCliArgs = providers.GetProviderCliArgs(opts.Provider)
	}

	// Build agent args and env
	execOpts := types.AgentExecOptions{
		Sandbox:         opts.Sandbox,
		ProviderCliArgs: providerCliArgs,
	}

	var agentArgs []string
	var agentEnv map[string]string

	switch opts.Agent.Type {
	case types.AgentClaude:
		agentArgs = agents.BuildClaudeArgs(opts.Args, execOpts)
		agentEnv = agents.BuildClaudeEnv(providerEnv)
	case types.AgentOpencode:
		agentArgs = agents.BuildOpencodeArgs(opts.Args, execOpts)
		agentEnv = agents.BuildOpencodeEnv(providerEnv)
	}

	// Merge environment
	env := os.Environ()
	for k, v := range agentEnv {
		env = append(env, k+"="+v)
	}
	env = append(env, opts.CustomEnv...)

	// Build command
	cmd := exec.Command(opts.Agent.Binary, agentArgs...)
	cmd.Dir = opts.Cwd
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if opts.Verbose {
		argsStr := ""
		if len(agentArgs) > 0 {
			argsStr = " " + strings.Join(agentArgs, " ")
		}
		scriptlib.LogInfo("Executing: " + opts.Agent.Binary + argsStr)
	}

	// Execute
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}

	return 0, nil
}
