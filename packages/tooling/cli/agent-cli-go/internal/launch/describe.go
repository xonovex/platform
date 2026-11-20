package launch

import (
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// describeProtocol names the protocol the agent declares, or none.
func describeProtocol(agent *types.AgentConfig) string {
	if agent.Protocol == types.ProtocolNone {
		return "none"
	}
	return string(agent.Protocol)
}

// describeAccess names how the agent obtains model access: its own login, or the
// routes through which a provider reaches it.
func describeAccess(agent *types.AgentConfig) string {
	if agent.Access == types.AccessLogin {
		return "native login"
	}
	var routes []string
	if agent.ProviderEnv {
		routes = append(routes, "environment")
	}
	if agent.ProviderArgs {
		routes = append(routes, "arguments")
	}
	return "provider (" + strings.Join(routes, ", ") + ")"
}

// describeVersionProbe is the command line that prints the agent version.
func describeVersionProbe(agent *types.AgentConfig) string {
	return strings.Join(append([]string{agent.Binary}, agent.VersionArgs...), " ")
}

// describeSlot names a model slot with its id, its effort and the layer or file
// that decided the id.
func describeSlot(slot types.ModelSlot) string {
	text := slot.Name + ": " + slot.ID
	if slot.Effort != "" {
		text += " (effort " + slot.Effort + ")"
	}
	return text + " (" + slot.Source + ")"
}
