package controller

import (
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	sharedlaunch "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/launch"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/network/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provider"
	provshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/resolver"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

const (
	agentRunKind      = "AgentRun"
	pendingRequeue    = 5 * time.Second
	runningRequeue    = 10 * time.Second
	defaultVolumeSize = "10Gi"
)

// runSettled reports whether a run needs no reconcile: it is being deleted or
// reached a terminal phase.
func runSettled(run *agentv1alpha1.AgentRun) bool {
	if !run.DeletionTimestamp.IsZero() {
		return true
	}
	switch run.Status.Phase {
	case agentv1alpha1.AgentRunPhaseSucceeded, agentv1alpha1.AgentRunPhaseFailed, agentv1alpha1.AgentRunPhaseTimedOut:
		return true
	}
	return false
}

// runExecution is the resolved input of the agent Job of one run. spec is the
// shared plan of the run, set once the run is planned.
type runExecution struct {
	agentType agentv1alpha1.AgentType
	route     provider.Route
	defaults  resolver.ResolvedDefaults
	image     string
	spec      sharedlaunch.Spec
}

// planExecution applies the harness defaults to run and decides the image and
// agent type its Job uses. toolchain is the provision leaf of the run's
// toolchain spec, or nil.
func planExecution(run *agentv1alpha1.AgentRun, harness *agentv1alpha1.AgentHarness, route provider.Route, toolchain provshared.Toolchain) (*runExecution, error) {
	agentType := agentv1alpha1.AgentTypeClaude
	if harness != nil {
		agentType = harness.Spec.Type
	}
	defaults := resolver.ApplyHarnessDefaults(run, harness)
	image := defaults.Image
	if toolchain != nil && toolchain.Image() != "" {
		image = toolchain.Image()
	}
	if image == "" {
		return nil, fmt.Errorf("ImageResolutionFailed: admission did not resolve an agent execution image")
	}
	if run.Spec.RuntimeClassName == nil || strings.TrimSpace(*run.Spec.RuntimeClassName) == "" {
		return nil, fmt.Errorf("RuntimeClassResolutionFailed: admission did not resolve a sandboxed runtimeClassName")
	}
	return &runExecution{
		agentType: agentType,
		route:     route,
		defaults:  defaults,
		image:     image,
	}, nil
}

// runVolume is the workspace volume of a run and the workspace type that
// selects its VCS leaf. claim is the PVC to create, or nil when the run already
// records one or mounts a shared workspace.
type runVolume struct {
	pvcName       string
	workspaceType agentv1alpha1.WorkspaceType
	claim         *corev1.PersistentVolumeClaim
	binding       *isoshared.WorkspaceBinding
}

// standaloneVolume decides the run-owned workspace PVC of a standalone run.
func standaloneVolume(run *agentv1alpha1.AgentRun) runVolume {
	volume := runVolume{
		pvcName:       run.Name + "-workspace",
		workspaceType: agentv1alpha1.WorkspaceTypeGit,
	}
	storageClass := ""
	storageSize := defaultVolumeSize
	if spec := run.Spec.Workspace; spec != nil {
		if spec.Type != "" {
			volume.workspaceType = spec.Type
		}
		storageClass = spec.StorageClass
		if spec.StorageSize != "" {
			storageSize = spec.StorageSize
		}
	}
	if run.Status.WorkspacePVC == "" {
		volume.claim = wsshared.BuildPVC(volume.pvcName, run.Namespace, storageClass, storageSize, run)
	}
	return volume
}

// sharedWorkspaceVolume decides how a run mounts a ready shared workspace: the
// worktree branch defaults to the run name and the source to HEAD, and a jj
// revision overrides the source.
func sharedWorkspaceVolume(run *agentv1alpha1.AgentRun, workspace *agentv1alpha1.AgentWorkspace) runVolume {
	workspaceType := workspace.Spec.Type
	if workspaceType == "" {
		workspaceType = agentv1alpha1.WorkspaceTypeGit
	}
	worktreeBranch := run.Name
	sourceBranch := "HEAD"
	if worktree := workspace.Spec.Git; worktree != nil && worktree.Worktree != nil {
		if worktree.Worktree.Branch != "" {
			worktreeBranch = worktree.Worktree.Branch
		}
		if worktree.Worktree.SourceBranch != "" {
			sourceBranch = worktree.Worktree.SourceBranch
		}
	}
	if workspace.Spec.Jj != nil && workspace.Spec.Jj.Revision != "" {
		sourceBranch = workspace.Spec.Jj.Revision
	}
	return runVolume{
		pvcName:       workspace.Status.WorkspacePVC,
		workspaceType: workspaceType,
		binding: &isoshared.WorkspaceBinding{
			SharedVolumes:    workspace.Spec.SharedVolumes,
			SharedVolumePVCs: workspace.Status.SharedVolumePVCs,
			WorktreeBranch:   worktreeBranch,
			SourceBranch:     sourceBranch,
			WorkspaceRef:     run.Spec.WorkspaceRef,
		},
	}
}

// runNetworkPolicy returns the NetworkPolicy a run owns, or nil when the run
// disables it.
func runNetworkPolicy(run *agentv1alpha1.AgentRun, spec *agentv1alpha1.AgentNetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	if spec != nil && spec.Disabled {
		return nil, nil
	}
	policy, err := netshared.BuildNetworkPolicy(run, spec)
	if err != nil {
		return nil, err
	}
	setController(run, agentRunKind, policy)
	return policy, nil
}

// withRunPhase sets the phase of a run status; a non-empty message adds a
// condition of the same name.
func withRunPhase(status agentv1alpha1.AgentRunStatus, phase agentv1alpha1.AgentRunPhase, message string, now time.Time) agentv1alpha1.AgentRunStatus {
	status.Phase = phase
	if message != "" {
		status.Conditions = append(status.Conditions, phaseCondition(string(phase), message, now))
	}
	return status
}

// runObservation is the status, events and requeue a run's Job state decides.
// write reports whether the status differs from the stored one.
type runObservation struct {
	status agentv1alpha1.AgentRunStatus
	events []event
	result ctrl.Result
	write  bool
}

// observeRunJob maps the Job of a run and its pods to the run status.
func observeRunJob(run *agentv1alpha1.AgentRun, job *batchv1.Job, pods []corev1.Pod, now time.Time) runObservation {
	status := *run.Status.DeepCopy()
	stamp := metav1.NewTime(now)
	podChanged := observePod(&status, pods)

	complete, failed := jobOutcome(job)
	switch {
	case complete:
		status.CompletionTime = &stamp
		return runObservation{
			status: withRunPhase(status, agentv1alpha1.AgentRunPhaseSucceeded, "", now),
			events: []event{normalEvent("AgentRunSucceeded", "Agent Job completed successfully")},
			write:  true,
		}
	case failed != nil:
		status.CompletionTime = &stamp
		return runObservation{
			status: withRunPhase(status, agentv1alpha1.AgentRunPhaseFailed, failed.Message, now),
			events: []event{warningEvent("AgentRunFailed", "Agent Job failed: "+failed.Message)},
			write:  true,
		}
	}

	if job.Status.Active > 0 {
		if status.Phase != agentv1alpha1.AgentRunPhaseRunning {
			status.StartTime = &stamp
			return runObservation{status: withRunPhase(status, agentv1alpha1.AgentRunPhaseRunning, "", now), write: true}
		}
		if run.Spec.Timeout != nil && status.StartTime != nil && now.Sub(status.StartTime.Time) > run.Spec.Timeout.Duration {
			status.CompletionTime = &stamp
			return runObservation{
				status: withRunPhase(status, agentv1alpha1.AgentRunPhaseTimedOut, "agent run exceeded timeout", now),
				events: []event{warningEvent("AgentRunTimedOut", fmt.Sprintf("Agent Job exceeded timeout of %v", run.Spec.Timeout.Duration))},
				write:  true,
			}
		}
	}
	return runObservation{status: status, result: ctrl.Result{RequeueAfter: runningRequeue}, write: podChanged}
}

// observePod records the newest pod of the Job and the exit code of its agent
// container, and reports whether the status changed. Ties on creation time
// resolve to the larger name so the choice is stable.
func observePod(status *agentv1alpha1.AgentRunStatus, pods []corev1.Pod) bool {
	if len(pods) == 0 {
		return false
	}
	observed := &pods[0]
	for index := 1; index < len(pods); index++ {
		candidate := &pods[index]
		if candidate.CreationTimestamp.After(observed.CreationTimestamp.Time) ||
			(candidate.CreationTimestamp.Equal(&observed.CreationTimestamp) && candidate.Name > observed.Name) {
			observed = candidate
		}
	}
	changed := status.PodName != observed.Name
	status.PodName = observed.Name
	for _, containerStatus := range observed.Status.ContainerStatuses {
		if containerStatus.Name != "agent" || containerStatus.State.Terminated == nil {
			continue
		}
		exitCode := containerStatus.State.Terminated.ExitCode
		if status.ExitCode == nil || *status.ExitCode != exitCode {
			status.ExitCode = &exitCode
			changed = true
		}
		break
	}
	return changed
}

func ptrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
