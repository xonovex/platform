// Package registry is the only place that names the concrete leaves: the
// harnesses, the protocols and the predicate evaluators. Adding a variant is one
// package and one line here; nothing else changes.
package registry

import (
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

// Harnesses returns the registry of agent harnesses.
func Harnesses() *harness.Registry {
	r := harness.NewRegistry()
	r.Register(claude.New())
	r.Register(codex.New())
	return r
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
