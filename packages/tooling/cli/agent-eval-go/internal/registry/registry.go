// Package registry is the only place that names the concrete leaves: the
// harnesses, the protocols, the predicate evaluators, the confinements and the
// isolators, provisioners and credential stores of shared-launch-go. Adding a
// variant is one package and one line here; nothing else changes.
package registry

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	credenv "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential/env"
	credfile "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential/file"
	credkeychain "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential/keychain"
	credpass "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential/pass"
	credsecretservice "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential/secretservice"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/egress/managed"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation/bwrap"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation/docker"
	isonone "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation/none"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation/seatbelt"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision/command"
	provnix "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision/nix"
	provnone "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision/none"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine"
	confinelaunch "github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine/launch"
	confinenone "github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine/none"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness/claude"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness/codex"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/filesystem"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/llm"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/session"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/instruction"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/orchestration"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/output"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/routing"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/trigger"
)

// Facts returns the facts of a harness.
type Facts func(types.AgentType) (*types.AgentConfig, error)

// leaf builds one harness leaf over its facts.
type leaf func(Facts) (harness.Harness, error)

func over[H harness.Harness](agentType types.AgentType, build func(*types.AgentConfig) (H, error)) leaf {
	return func(facts Facts) (harness.Harness, error) {
		agent, err := facts(agentType)
		if err != nil {
			return nil, err
		}
		return build(agent)
	}
}

// Harnesses returns the registry of agent harnesses, each leaf over its facts.
func Harnesses(facts Facts) (*harness.Registry, error) {
	r := harness.NewRegistry()
	for _, build := range []leaf{
		over(types.AgentClaude, claude.New),
		over(types.AgentCodex, codex.New),
	} {
		h, err := build(facts)
		if err != nil {
			return nil, err
		}
		r.Register(h)
	}
	return r, nil
}

// Protocols returns the registry of eval protocols.
func Protocols() *protocol.Registry {
	r := protocol.NewRegistry()
	r.Register(trigger.New())
	r.Register(routing.New())
	r.Register(output.New())
	r.Register(instruction.New())
	r.Register(orchestration.New())
	return r
}

// Predicates returns the registry of predicate evaluators. The model evaluator
// grades through the executor.
func Predicates(exec harness.Executor) *predicate.Registry {
	r := predicate.NewRegistry()
	r.Register(filesystem.New())
	r.Register(session.New())
	r.Register(llm.New(judge.Grader{Executor: exec}))
	return r
}

// Host is the union of the ports the isolators, the provisioners, the
// credential stores and the egress proxy of shared-launch-go consume.
type Host interface {
	docker.Host
	provnix.Host
	credfile.Host
	credenv.Host
	managed.Host
}

// Confiner returns the confinement of the sandbox: the one that confines
// nothing when the sandbox asks for nothing, else the shared isolator over
// host. The isolations that keep a sandbox alive apart from one command,
// docker-sandboxes and openshell, are not offered: a harness run is one
// process the executor starts and stops.
func Confiner(box model.Sandbox, facts agents.Catalog, host Host, log logging.Logger) (confine.Confiner, error) {
	if box.Unconfined() {
		return confinenone.New(), nil
	}
	home, err := host.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return confinelaunch.New(confinelaunch.Ports{
		Sandboxes:   Sandboxes(host, log),
		Credentials: Credentials(host),
		Egress:      managed.Service{Host: host, Log: log},
		Harnesses:   facts,
		Home:        home,
	}, box), nil
}

// Bridge returns the relay a sandbox with network managed runs the harness
// under.
func Bridge(host managed.BridgeHost) managed.Bridge { return managed.Bridge{Host: host} }

// Sandboxes returns the isolators and provisioners a harness run may use.
func Sandboxes(host Host, log logging.Logger) *sandbox.Registry[isoshared.Isolator] {
	return sandbox.NewRegistry[isoshared.Isolator]().
		RegisterIsolator(isolation.IsolationNone, func() isoshared.Isolator { return isonone.NewIsolator(host, log) }).
		RegisterIsolator(isolation.IsolationBwrap, func() isoshared.Isolator { return bwrap.NewIsolator(host, log) }).
		RegisterIsolator(isolation.IsolationDocker, func() isoshared.Isolator { return docker.NewIsolator(host, log) }).
		RegisterIsolator(isolation.IsolationSeatbelt, func() isoshared.Isolator { return seatbelt.NewIsolator(host, log) }).
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return provnone.New() }).
		RegisterProvisioner(provision.ProvisionCommand, func() provshared.Provisioner { return command.New() }).
		RegisterProvisioner(provision.ProvisionNix, func() provshared.Provisioner { return provnix.New(host) })
}

// Credentials returns the credential stores, by the scheme of a binding.
func Credentials(host Host) *credshared.Registry {
	return credshared.NewRegistry().
		Register("env", func() credshared.Store { return credenv.New(host) }).
		Register("file", func() credshared.Store { return credfile.New(host) }).
		Register("keychain", func() credshared.Store { return credkeychain.New(host) }).
		Register("secret-service", func() credshared.Store { return credsecretservice.New(host) }).
		Register("pass", func() credshared.Store { return credpass.New(host) })
}
