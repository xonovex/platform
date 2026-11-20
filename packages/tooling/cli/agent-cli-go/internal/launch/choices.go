package launch

import (
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

// AgentNames lists the registered agent types in sorted order.
func AgentNames() []string {
	types := agents.GetAgentTypes()
	names := make([]string, 0, len(types))
	for _, agentType := range types {
		names = append(names, string(agentType))
	}
	return names
}

// NetworkNames lists the network modes a run may select.
func NetworkNames() []string {
	return []string{string(network.NetworkHost), string(network.NetworkNone), string(network.NetworkManaged)}
}

// IsolationNames lists the isolations the build registers.
func IsolationNames(deps Dependencies) []string {
	if deps.Sandboxes == nil {
		return nil
	}
	var names []string
	for _, method := range deps.Sandboxes.IsolationMethods() {
		names = append(names, string(method))
	}
	return names
}

// TerminalNames lists the terminal wrappers the build registers, in sorted
// order, and the automatic selection when the build offers a wrapper for it.
func TerminalNames(deps Dependencies) []string {
	names := make([]string, 0, len(deps.Terminals)+1)
	for name := range deps.Terminals {
		names = append(names, string(name))
	}
	if len(deps.AutoTerminals) > 0 {
		names = append(names, string(termshared.TerminalAuto))
	}
	sort.Strings(names)
	return names
}

// CredentialCompletions completes a credential binding NAME=scheme:locator
// with each scheme the build has a store for. It offers nothing until typed
// holds the variable name and the equals sign, because the variable name is the
// user's own.
func CredentialCompletions(deps Dependencies, typed string) []string {
	if deps.Credentials == nil {
		return nil
	}
	name, _, hasEquals := strings.Cut(typed, "=")
	if !hasEquals {
		return nil
	}
	var candidates []string
	for _, scheme := range deps.Credentials.Schemes() {
		candidates = append(candidates, name+"="+scheme+":")
	}
	return candidates
}

// ProviderNames names the providers that the configuration of options defines
// for the agent type. A failure to read them names none, because a completion
// has no place to report it.
func ProviderNames(deps Dependencies, options Options, agent string) []string {
	listing, err := ListProviders(deps, options)
	if err != nil {
		return nil
	}
	var names []string
	for _, row := range listing.Rows {
		if string(row.Agent) == agent {
			names = append(names, row.Name)
		}
	}
	return names
}
