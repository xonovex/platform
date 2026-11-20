// Package registry is the only place that names the concrete leaves: the
// harnesses, the protocols and the predicate evaluators. Adding a variant is one
// package and one line here; nothing else changes.
package registry

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness/claude"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness/codex"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
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
