package shared

import (
	"errors"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
)

// HostExecution builds the command and the environment of an agent process
// that runs on the host, under the wrapper of the isolator or none at all: the
// host environment, the provider environment, the environment of the process
// and credential values, and the provisioner's environment and tool directories.
func HostExecution(host Host, cfg RunConfig, c provision.Contribution) ([]string, []string, error) {
	if len(cfg.Process.Argv) == 0 {
		return nil, nil, errors.New("the agent process has no command")
	}
	cmd := WrapWithInitCommands(cfg.Process.Argv, c.InitCommands)

	customEnv, err := RunEnv(cfg)
	if err != nil {
		return nil, nil, err
	}
	env := envutil.MergeEnvMaps(envutil.ParseEnv(host.Environ()), cfg.ProviderEnv, customEnv, c.Env)
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
