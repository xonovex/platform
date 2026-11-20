// Package registry lists the checks of the validator. It is the only package
// that imports a concrete check: a new check is one line here, and the runner,
// the plan and the command tree see only the check.Check port.
package registry

import (
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/agent"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/command"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/drift"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/hook"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/instructions"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/kit"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/links"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/llms"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/marketplace"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/outputstyle"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/plugin"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/repository"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/routing"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/schema"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/skill"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/text"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/workspace"
)

// All returns every check, in the order a full run executes them.
func All() []check.Check {
	return []check.Check{
		skill.Check{},
		sources.Check{},
		drift.Check{},
		links.Check{},
		routing.Check{},
		plugin.Check{},
		marketplace.Check{},
		instructions.Check{},
		command.Check{},
		agent.Check{},
		hook.Check{},
		outputstyle.Check{},
		kit.Check{},
		llms.Check{},
		schema.Check{},
		repository.Check{},
		workspace.Check{},
		text.Check{},
	}
}
