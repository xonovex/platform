package launch

import (
	"errors"
	"slices"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

func TestDefaultOptionsStartTheClaudeAgentWithoutABoundary(t *testing.T) {
	got := DefaultOptions()
	if got.Agent != "claude" || got.Isolation != "none" || got.Provision != "none" || got.Network != "host" ||
		got.NixSource != "packages" || got.NixShell != "default" || got.VCS != "git" {
		t.Errorf("DefaultOptions() = %+v", got)
	}
}

func TestExplicitKeysMarkOnlyTheAxisFlagsTheUserSet(t *testing.T) {
	changed := func(flag string) bool { return flag == "agent" || flag == "network" || flag == "work-dir" }

	got := ExplicitKeys(changed)

	if len(got) != 2 || !got.Has(runconfig.KeyAgent) || !got.Has(runconfig.KeyNetwork) {
		t.Errorf("ExplicitKeys() = %v, want the agent and the network", got)
	}
	if none := ExplicitKeys(func(string) bool { return false }); len(none) != 0 {
		t.Errorf("ExplicitKeys() without a changed flag = %v", none)
	}
}

func TestAgentAndNetworkNamesListEveryChoice(t *testing.T) {
	names := AgentNames()
	if len(names) != len(agents.GetAgentTypes()) || !slices.Contains(names, "claude") {
		t.Errorf("AgentNames() = %v", names)
	}
	if got := NetworkNames(); !slices.Equal(got, []string{"host", "none", "managed"}) {
		t.Errorf("NetworkNames() = %v", got)
	}
}

func TestIsolationNamesListTheRegisteredIsolations(t *testing.T) {
	if got := IsolationNames(newExecuteFixture().deps()); !slices.Equal(got, []string{"bwrap", "docker", "none", "openshell"}) {
		t.Errorf("IsolationNames() = %v", got)
	}
	if got := IsolationNames(Dependencies{}); got != nil {
		t.Errorf("a build without isolations offers %v", got)
	}
}

func TestTerminalNamesOfferAutoOnlyWithACandidate(t *testing.T) {
	terminals := map[termshared.TerminalType]termshared.TerminalExecutor{"tmux": nil, "herdr-core": nil}
	with := TerminalNames(Dependencies{Terminals: terminals, AutoTerminals: []termshared.TerminalType{"herdr-core"}})
	if !slices.Equal(with, []string{"auto", "herdr-core", "tmux"}) {
		t.Errorf("TerminalNames() = %v, want auto and every wrapper", with)
	}
	if without := TerminalNames(Dependencies{Terminals: terminals}); slices.Contains(without, "auto") {
		t.Errorf("a build without automatic candidates offers auto: %v", without)
	}
}

func TestCredentialCompletionsOfferTheSchemesAfterTheVariableName(t *testing.T) {
	deps := Dependencies{Credentials: credshared.NewRegistry().
		Register("env", func() credshared.Store { return nil }).
		Register("pass", func() credshared.Store { return nil })}

	if got := CredentialCompletions(deps, "API_KEY="); !slices.Equal(got, []string{"API_KEY=env:", "API_KEY=pass:"}) {
		t.Errorf("CredentialCompletions() = %v", got)
	}
	if got := CredentialCompletions(deps, "API"); got != nil {
		t.Errorf("before the equals sign the user names the variable: %v", got)
	}
	if got := CredentialCompletions(Dependencies{}, "API_KEY="); got != nil {
		t.Errorf("a build without stores offers %v", got)
	}
}

func TestProviderNamesListTheProvidersOfOneAgent(t *testing.T) {
	f := newExecuteFixture()
	if got := ProviderNames(f.deps(), Options{}, "claude"); !slices.Equal(got, []string{"alpha", "beta"}) {
		t.Errorf("ProviderNames(claude) = %v", got)
	}
	if got := ProviderNames(f.deps(), Options{}, "codex"); got != nil {
		t.Errorf("ProviderNames(codex) = %v, want none", got)
	}
	f.host.HomeErr = errors.New("no home")
	if got := ProviderNames(f.deps(), Options{}, "claude"); got != nil {
		t.Errorf("ProviderNames() that cannot read the providers = %v, want none", got)
	}
}
