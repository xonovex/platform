package sandbox

import (
	"fmt"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/bwrap"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/compose"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/docker"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/nix"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/nixflake"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/none"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
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
	case types.SandboxNixFlake:
		return nixflake.NewExecutor(), nil
	default:
		return nil, fmt.Errorf("unknown sandbox method: %s", method)
	}
}

// Isolation classifies a tier's host-tool reachability guarantee.
type Isolation int

const (
	// IsolationHostToolsLeaked: the tier ro-binds host /usr,/lib,/bin and/or
	// appends the host PATH, so host binaries stay reachable (none, bwrap).
	IsolationHostToolsLeaked Isolation = iota
	// IsolationContainerPinned: host tools are unreachable only when a concrete
	// --image/pin is supplied; an empty image resolves host-equivalent tools.
	IsolationContainerPinned
	// IsolationHostToolsUnreachable: the tier exposes only a nix-built closure or
	// a flake.lock devShell and binds no host /usr,/lib,/bin (nix, nixflake).
	IsolationHostToolsUnreachable
)

// tierIsolation maps a method to its guarantee. Pure: depends only on method.
// It is the single source of truth for the per-tier guarantee, so adding a tier
// forces a classification decision here.
func tierIsolation(method types.SandboxMethod) Isolation {
	switch method {
	case types.SandboxNix, types.SandboxNixFlake:
		return IsolationHostToolsUnreachable
	case types.SandboxDocker, types.SandboxCompose:
		return IsolationContainerPinned
	default: // SandboxNone, SandboxBwrap
		return IsolationHostToolsLeaked
	}
}

// SelectExecutor resolves an executor for method under policy. When the policy
// requires a pinned toolchain, leaky tiers and image-less container tiers are
// rejected before construction; otherwise it behaves like GetExecutor. The
// returned method echoes the requested one so callers can record what was used.
func SelectExecutor(method types.SandboxMethod, image string, policy types.SandboxPolicy) (types.SandboxExecutor, types.SandboxMethod, error) {
	if policy.RequirePinnedProvisioning || policy.RequireHostToolsUnreachable {
		if err := enforcePinnedToolchain(method, image); err != nil {
			return nil, "", err
		}
	}
	exec, err := GetExecutor(method)
	if err != nil {
		return nil, "", err
	}
	return exec, method, nil
}

// enforcePinnedToolchain rejects any method that cannot guarantee unreachable
// host tools under a pinned-toolchain policy.
func enforcePinnedToolchain(method types.SandboxMethod, image string) error {
	switch tierIsolation(method) {
	case IsolationHostToolsUnreachable:
		return nil
	case IsolationContainerPinned:
		if image == "" {
			return fmt.Errorf("sandbox %q needs a pinned --image under require-pinned-toolchain: an image-less container resolves host-equivalent tools", method)
		}
		return nil
	default:
		return fmt.Errorf("sandbox %q is rejected by require-pinned-toolchain: it ro-binds host /usr,/lib,/bin and appends the host PATH, so host tools stay reachable — use the nix or nixflake tier", method)
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
		types.SandboxNixFlake,
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
