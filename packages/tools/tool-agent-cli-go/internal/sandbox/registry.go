package sandbox

import (
	"fmt"

	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/sandbox/bwrap"
	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/sandbox/compose"
	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/sandbox/docker"
	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/sandbox/nix"
	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/sandbox/none"
	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/types"
)

// GetExecutor returns a sandbox executor for the specified method
func GetExecutor(method types.SandboxMethod) (types.SandboxExecutor, error) {
	switch method {
	case types.SandboxNone:
		return none.NewExecutor(), nil
	case types.SandboxBwrap:
		return bwrap.NewExecutor(), nil
	case types.SandboxDocker:
		return docker.NewExecutor(), nil
	case types.SandboxCompose:
		return compose.NewExecutor(), nil
	case types.SandboxNix:
		return nix.NewExecutor(), nil
	default:
		return nil, fmt.Errorf("unknown sandbox method: %s", method)
	}
}

// GetAvailableMethods returns all sandbox methods that are currently available
func GetAvailableMethods() []types.SandboxMethod {
	allMethods := []types.SandboxMethod{
		types.SandboxNone,
		types.SandboxBwrap,
		types.SandboxDocker,
		types.SandboxCompose,
		types.SandboxNix,
	}

	available := make([]types.SandboxMethod, 0, len(allMethods))
	for _, method := range allMethods {
		executor, err := GetExecutor(method)
		if err != nil {
			continue
		}
		isAvailable, err := executor.IsAvailable()
		if err == nil && isAvailable {
			available = append(available, method)
		}
	}

	return available
}
