// Package launch plans the run of an AgentRun with the shared planner. It maps
// the run onto a shared launch request, and the shared planner computes the
// agent process from the facts of the harness and the guarantees of the
// confinement. The package names no confinement: the Sandbox the composition
// root builds holds it.
package launch

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	sharedlaunch "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/launch"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	podshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/network/shared"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

// Sandbox is the isolation and the provisioning every run of the operator
// selects, and the registry that holds them.
type Sandbox struct {
	Registry  *sandbox.Registry[isoshared.Confinement]
	Isolation isolation.IsolationMethod
	Provision provision.ProvisionMethod
}

// Sandboxes builds the Sandbox of a namespace whose AgentPolicy approves
// runtimeClasses as sandboxed runtimes.
type Sandboxes interface {
	Sandbox(runtimeClasses []string) Sandbox
}

// Harnesses resolves the facts of an agent type.
type Harnesses interface {
	Harness(agent agentv1alpha1.AgentType) (*types.AgentConfig, error)
}

// Execution is what a run executes after every default is applied: the agent
// type, the pod image, which a toolchain may supply, and the model provider,
// nil for none.
type Execution struct {
	AgentType agentv1alpha1.AgentType
	Image     string
	Provider  *types.ModelProvider
}

// Plan plans run in sandbox. The agent starts in the work directory of the run
// with the home directory of the pod; a prompt makes the run non-interactive.
// It fails when the provider sets a route the agent does not read, when the
// harness declares no non-interactive run for a prompt, and when the network
// mode has no enforceable backend. It enforces no policy: the caller evaluates
// the capabilities of the plan.
func Plan(box Sandbox, harnesses Harnesses, run *agentv1alpha1.AgentRun, execution Execution) (sharedlaunch.Spec, error) {
	agent, err := harnesses.Harness(execution.AgentType)
	if err != nil {
		return sharedlaunch.Spec{}, err
	}
	if err := agents.CheckRoutes(agent, execution.Provider); err != nil {
		return sharedlaunch.Spec{}, err
	}
	var args []string
	if run.Spec.Prompt != "" {
		if args, err = agents.PromptArgs(agent, run.Spec.Prompt); err != nil {
			return sharedlaunch.Spec{}, err
		}
	}
	mode, err := netshared.Method(run)
	if err != nil {
		return sharedlaunch.Spec{}, err
	}
	runtimeClass := ""
	if run.Spec.RuntimeClassName != nil {
		runtimeClass = *run.Spec.RuntimeClassName
	}
	request := sharedlaunch.Request{
		Agent:        string(agent.Type),
		Isolation:    string(box.Isolation),
		Provision:    string(box.Provision),
		Network:      string(mode),
		ExplicitAxes: true,
		// The runtime of a pod is its runtime class.
		IsolationDockerRuntime: runtimeClass,
		Image:                  execution.Image,
		AgentArgs:              args,
	}
	workDir := wsshared.WorkDir(run)
	spec, _, err := sharedlaunch.PlanSpec(request, agent, execution.Provider, workDir, workDir, podshared.AgentHome, box.Registry, false)
	if err != nil {
		return sharedlaunch.Spec{}, fmt.Errorf("plan the run: %w", err)
	}
	return spec, nil
}
