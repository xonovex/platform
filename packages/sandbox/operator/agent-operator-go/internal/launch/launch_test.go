package launch

import (
	"errors"
	"slices"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	netshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/network/shared"
)

// podDouble is a confinement that bounds the agent like a pod and approves
// the runtime class gvisor.
type podDouble struct{}

func (podDouble) PinnedProvision(_ provision.ProvisionMethod, _ bool, image string) bool {
	return strings.Contains(image, "@sha256:")
}
func (podDouble) HidesHost(bool, string) bool              { return true }
func (podDouble) KernelIsolated(runtime string) bool       { return runtime == "gvisor" }
func (podDouble) ConfinesNetwork() bool                    { return true }
func (podDouble) EgressTransport() network.EgressTransport { return network.TransportNone }
func (podDouble) Bounded() bool                            { return true }
func (podDouble) RequiresImage() bool                      { return true }
func (podDouble) AcceptsHostBinds() bool                   { return false }

// imageDouble is a provisioner that leaves every tool to the image.
type imageDouble struct{}

func (imageDouble) Contribute(provshared.Input) (provision.Contribution, error) {
	return provision.Contribution{}, nil
}
func (imageDouble) Pinned() bool { return false }
func (imageDouble) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return false, nil
}

// facts resolves a harness from the shipped agent facts.
type facts struct{}

func (facts) Harness(agent agentv1alpha1.AgentType) (*types.AgentConfig, error) {
	return agents.GetAgent(types.AgentType(agent))
}

func pod() Sandbox {
	registry := sandbox.NewRegistry[isoshared.Confinement]().
		RegisterIsolator("pod", func() isoshared.Confinement { return podDouble{} }).
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return imageDouble{} })
	return Sandbox{Registry: registry, Isolation: "pod", Provision: provision.ProvisionNone}
}

const pinned = "ghcr.io/xonovex/agent@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func agentRun(prompt string) *agentv1alpha1.AgentRun {
	gvisor := "gvisor"
	return &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{Prompt: prompt, RuntimeClassName: &gvisor},
	}
}

func TestPlanBuildsTheProcessOfAPromptedRun(t *testing.T) {
	spec, err := Plan(pod(), facts{}, agentRun("Fix the tests"), Execution{AgentType: agentv1alpha1.AgentTypeClaude, Image: pinned})

	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if want := []string{"claude", "--permission-mode", "bypassPermissions", "-p", "Fix the tests"}; !slices.Equal(spec.Run.Process.Argv, want) {
		t.Errorf("argv = %v, want %v", spec.Run.Process.Argv, want)
	}
	if spec.Run.Process.Dir != "/workspace" || spec.Run.Image != pinned || spec.Run.Runtime != "gvisor" {
		t.Errorf("run = dir %q, image %q, runtime %q, want /workspace, the image and gvisor", spec.Run.Process.Dir, spec.Run.Image, spec.Run.Runtime)
	}
	caps := spec.Capabilities
	if !caps.Pinned || !caps.KernelIsolated || !caps.EgressRestricted || !caps.HostToolsUnreachable {
		t.Errorf("capabilities = %+v, want every guarantee of the pod", caps)
	}
}

func TestPlanStartsASharedWorkspaceRunInItsWorktree(t *testing.T) {
	run := agentRun("")
	run.Spec.WorkspaceRef = "shared"

	spec, err := Plan(pod(), facts{}, run, Execution{AgentType: "codex", Image: pinned})

	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if spec.Run.Process.Dir != "/workspace-wt/run" || !slices.Equal(spec.Run.Process.Argv, []string{"codex", "--sandbox", "danger-full-access"}) {
		t.Errorf("process = %+v, want codex with its sandbox arguments in the worktree of the run", spec.Run.Process)
	}
}

func TestPlanPassesTheProviderArgumentsOfARoute(t *testing.T) {
	route := &types.ModelProvider{Name: "gemini", CliArgs: []string{"--model", "google/gemini-2.5-pro"}}

	spec, err := Plan(pod(), facts{}, agentRun(""), Execution{AgentType: agentv1alpha1.AgentTypeOpencode, Image: pinned, Provider: route})

	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	argv := spec.Run.Process.Argv
	if at := slices.Index(argv, "--model"); at < 0 || at+1 >= len(argv) || argv[at+1] != "google/gemini-2.5-pro" {
		t.Errorf("argv = %v, want the provider arguments --model google/gemini-2.5-pro", argv)
	}
}

func TestPlanReportsOpenEgressAndAnUnapprovedRuntime(t *testing.T) {
	run := agentRun("")
	run.Spec.Network = agentv1alpha1.NetworkModeHost
	runc := "runc"
	run.Spec.RuntimeClassName = &runc

	spec, err := Plan(pod(), facts{}, run, Execution{AgentType: agentv1alpha1.AgentTypeClaude, Image: "node:latest"})

	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if c := spec.Capabilities; c.EgressRestricted || c.KernelIsolated || c.Pinned {
		t.Errorf("capabilities = %+v, want open egress, no kernel isolation and no pin", c)
	}
}

func TestPlanRefusesARunItCannotPlan(t *testing.T) {
	proxied := agentRun("")
	proxied.Spec.Network = agentv1alpha1.NetworkModeProxy
	cases := map[string]struct {
		run       *agentv1alpha1.AgentRun
		execution Execution
		want      string
	}{
		"unknown agent":           {run: agentRun(""), execution: Execution{AgentType: "mystery"}, want: "unknown agent type"},
		"route the agent ignores": {run: agentRun(""), execution: Execution{AgentType: agentv1alpha1.AgentTypeClaude, Provider: &types.ModelProvider{Name: "p", CliArgs: []string{"--x"}}}, want: "does not take"},
		"login agent":             {run: agentRun(""), execution: Execution{AgentType: "copilot", Provider: &types.ModelProvider{Name: "p"}}, want: "own login"},
		"interactive agent":       {run: agentRun("go"), execution: Execution{AgentType: agentv1alpha1.AgentTypeOpencode}, want: "non-interactive"},
		"prompt read as option":   {run: agentRun("-x"), execution: Execution{AgentType: agentv1alpha1.AgentTypeClaude}, want: "dash"},
		"proxy":                   {run: proxied, execution: Execution{AgentType: agentv1alpha1.AgentTypeClaude}, want: "plan network"},
	}
	for name, tc := range cases {
		_, err := Plan(pod(), facts{}, tc.run, tc.execution)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: Plan() err = %v, want it to contain %q", name, err, tc.want)
		}
	}
	if _, err := Plan(pod(), facts{}, proxied, Execution{AgentType: agentv1alpha1.AgentTypeClaude}); !errors.Is(err, netshared.ErrProxyEnforcementUnavailable) {
		t.Errorf("Plan(proxy) err = %v, want ErrProxyEnforcementUnavailable", err)
	}
}

func TestPlanWrapsAFailureOfTheSharedPlanner(t *testing.T) {
	box := pod()
	box.Isolation = "absent"

	if _, err := Plan(box, facts{}, agentRun(""), Execution{AgentType: agentv1alpha1.AgentTypeClaude}); err == nil || !strings.HasPrefix(err.Error(), "plan the run: ") {
		t.Fatalf("Plan() err = %v, want the failure of the shared planner", err)
	}
}
