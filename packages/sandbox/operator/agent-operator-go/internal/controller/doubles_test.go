package controller

import (
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	launchiso "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	launchprov "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/launch"
	provshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provision/shared"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
)

// errInjected stands in for any API failure that is neither NotFound nor
// AlreadyExists, so the controller must surface it rather than absorb it.
var errInjected = errors.New("injected API failure")

// testEpoch is the start time of every fake clock.
var testEpoch = time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

// fakeClock returns a time that a test moves forward explicitly.
type fakeClock struct{ now time.Time }

func newFakeClock() *fakeClock { return &fakeClock{now: testEpoch} }

func (c *fakeClock) Now() time.Time { return c.now }

func (c *fakeClock) advance(step time.Duration) { c.now = c.now.Add(step) }

type recordedEvent struct {
	eventType string
	reason    string
	message   string
}

// eventLog keeps every recorded event in order.
type eventLog struct{ events []recordedEvent }

func (l *eventLog) Record(_ runtime.Object, eventType, reason, message string) {
	l.events = append(l.events, recordedEvent{eventType: eventType, reason: reason, message: message})
}

func (l *eventLog) reasons() []string {
	reasons := make([]string, 0, len(l.events))
	for _, e := range l.events {
		reasons = append(reasons, e.reason)
	}
	return reasons
}

// podDouble is a confinement that bounds the agent like a pod.
type podDouble struct{}

func (podDouble) PinnedProvision(provision.ProvisionMethod, bool, string) bool { return true }
func (podDouble) HidesHost(bool, string) bool                                  { return true }
func (podDouble) KernelIsolated(string) bool                                   { return true }
func (podDouble) ConfinesNetwork() bool                                        { return true }
func (podDouble) EgressTransport() network.EgressTransport                     { return network.TransportNone }
func (podDouble) Bounded() bool                                                { return true }
func (podDouble) RequiresImage() bool                                          { return true }
func (podDouble) AcceptsHostBinds() bool                                       { return false }

// imageDouble is a provisioner that leaves every tool to the image.
type imageDouble struct{}

func (imageDouble) Contribute(launchprov.Input) (provision.Contribution, error) {
	return provision.Contribution{}, nil
}
func (imageDouble) Pinned() bool { return false }
func (imageDouble) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return false, nil
}

// stubVCS is a VCS strategy whose scripts carry the workspace type.
type stubVCS struct{ workspace agentv1alpha1.WorkspaceType }

func (v stubVCS) PostCloneScript() string { return "post-clone " + string(v.workspace) + "\n" }

func (v stubVCS) WorktreeScript(path, branch, sourceBranch string) string {
	return fmt.Sprintf("%s worktree %s %s %s\n", v.workspace, path, branch, sourceBranch)
}

func (v stubVCS) InitContainerName() string { return string(v.workspace) + "-worktree" }

// stubToolchain is a provision leaf with a fixed image.
type stubToolchain struct{ image string }

func (t stubToolchain) Image() string { return t.image }

func (stubToolchain) Pinned() bool { return true }

// stubRegistry registers the harnesses of the shipped agent facts, a pod
// sandbox, the git and jj workspaces and an image-only toolchain.
type stubRegistry struct{}

func (stubRegistry) Harness(agent agentv1alpha1.AgentType) (*types.AgentConfig, error) {
	return agents.GetAgent(types.AgentType(agent))
}

func (stubRegistry) Sandbox([]string) launch.Sandbox {
	registry := sandbox.NewRegistry[launchiso.Confinement]().
		RegisterIsolator("pod", func() launchiso.Confinement { return podDouble{} }).
		RegisterProvisioner(provision.ProvisionNone, func() launchprov.Provisioner { return imageDouble{} })
	return launch.Sandbox{Registry: registry, Isolation: "pod", Provision: provision.ProvisionNone}
}

func (stubRegistry) ResolveToolchain(spec *agentv1alpha1.ToolchainSpec) provshared.Toolchain {
	if spec == nil || spec.Nix == nil {
		return nil
	}
	return stubToolchain{image: spec.Nix.Image}
}

func (stubRegistry) ResolveVCSStrategy(wsType agentv1alpha1.WorkspaceType) (wsshared.VCSStrategy, error) {
	switch wsType {
	case "", agentv1alpha1.WorkspaceTypeGit, agentv1alpha1.WorkspaceTypeJujutsu:
		return stubVCS{workspace: wsType}, nil
	}
	return nil, fmt.Errorf("unknown workspace type %q", wsType)
}

// runHarness is an AgentRun reconciler over a memory store, a fake clock and
// an event log.
type runHarness struct {
	reconciler *AgentRunReconciler
	store      *testutil.MemoryStore
	clock      *fakeClock
	events     *eventLog
}

func newRunHarness(objects ...any) runHarness {
	h := runHarness{store: testutil.NewMemoryStore(objects...), clock: newFakeClock(), events: &eventLog{}}
	h.reconciler = NewAgentRunReconciler(AgentRunDependencies{
		Store:      h.store,
		Toolchains: stubRegistry{},
		Harnesses:  stubRegistry{},
		Sandboxes:  stubRegistry{},
		VCS:        stubRegistry{},
		Events:     h.events,
		Clock:      h.clock,
		JobPolicy:  isoshared.DefaultJobPolicy(),
	})
	return h
}

// workspaceHarness is an AgentWorkspace reconciler over a memory store, a fake
// clock and an event log.
type workspaceHarness struct {
	reconciler *AgentWorkspaceReconciler
	store      *testutil.MemoryStore
	clock      *fakeClock
	events     *eventLog
}

const workspaceInitImage = "registry.example/git@sha256:dddd"

func newWorkspaceHarness(objects ...any) workspaceHarness {
	h := workspaceHarness{store: testutil.NewMemoryStore(objects...), clock: newFakeClock(), events: &eventLog{}}
	h.reconciler = NewAgentWorkspaceReconciler(AgentWorkspaceDependencies{
		Store:     h.store,
		VCS:       stubRegistry{},
		Events:    h.events,
		Clock:     h.clock,
		JobPolicy: isoshared.DefaultJobPolicy(),
		// The image of the init Job is configuration of the deployment.
		WorkspaceInitImage: workspaceInitImage,
	})
	return h
}

// providerHarness is an AgentProvider reconciler over a memory store, a fake
// clock and an event log.
type providerHarness struct {
	reconciler *AgentProviderReconciler
	store      *testutil.MemoryStore
	clock      *fakeClock
	events     *eventLog
}

func newProviderHarness(objects ...any) providerHarness {
	h := providerHarness{store: testutil.NewMemoryStore(objects...), clock: newFakeClock(), events: &eventLog{}}
	h.reconciler = NewAgentProviderReconciler(AgentProviderDependencies{Store: h.store, Events: h.events, Clock: h.clock})
	return h
}
