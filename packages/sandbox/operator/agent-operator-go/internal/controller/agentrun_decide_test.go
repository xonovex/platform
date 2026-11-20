package controller

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provider"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
)

// A retried Job leaves several pods behind, so the newest one is the run's.
func TestObservePod_ObservesTheNewestPod(t *testing.T) {
	status := agentv1alpha1.AgentRunStatus{}
	pods := []corev1.Pod{
		*agentPod("solo-newer", testEpoch.Add(time.Minute), nil),
		*agentPod("solo-older", testEpoch, nil),
	}

	if !observePod(&status, pods) {
		t.Error("observePod() = false, want it to report the newly observed pod")
	}
	if status.PodName != "solo-newer" {
		t.Errorf("pod = %q, want the newest", status.PodName)
	}
}

// Identical creation timestamps are common at second granularity, so the name
// breaks the tie and keeps the choice deterministic.
func TestObservePod_BreaksTimestampTiesByName(t *testing.T) {
	status := agentv1alpha1.AgentRunStatus{}

	observePod(&status, []corev1.Pod{*agentPod("solo-zzz", testEpoch, nil), *agentPod("solo-aaa", testEpoch, nil)})

	if status.PodName != "solo-zzz" {
		t.Errorf("pod = %q, want the tie broken toward solo-zzz", status.PodName)
	}
}

func TestObservePod_ReportsNoChangeWhenAlreadyObserved(t *testing.T) {
	exitCode := int32(0)
	pods := []corev1.Pod{*agentPod("solo-pod", testEpoch, &exitCode)}
	status := agentv1alpha1.AgentRunStatus{}
	observePod(&status, pods)

	if observePod(&status, pods) {
		t.Error("observePod() = true on unchanged state, want false")
	}
}

func TestObservePod_ReportsNoChangeWithoutPods(t *testing.T) {
	status := agentv1alpha1.AgentRunStatus{}

	if observePod(&status, nil) {
		t.Error("observePod() = true with no pods, want false")
	}
}

// Only the agent container's exit code is the run's.
func TestObservePod_IgnoresOtherContainers(t *testing.T) {
	pod := agentPod("solo-pod", testEpoch, nil)
	pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
		Name:  "sidecar",
		State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 3}},
	}}
	status := agentv1alpha1.AgentRunStatus{}

	observePod(&status, []corev1.Pod{*pod})

	if status.ExitCode != nil {
		t.Errorf("ExitCode = %d, want none from a sidecar", *status.ExitCode)
	}
}

func TestObserveRunJob_LeavesTheStoredRunUnchanged(t *testing.T) {
	run := runningRun("solo")
	run.Status.Conditions = nil

	observation := observeRunJob(run, runJob(run, failedJobStatus("boom")), nil, testEpoch)

	if len(run.Status.Conditions) != 0 || run.Status.Phase != agentv1alpha1.AgentRunPhaseRunning {
		t.Errorf("run status = %+v, want it untouched", run.Status)
	}
	if observation.status.Phase != agentv1alpha1.AgentRunPhaseFailed {
		t.Errorf("observed phase = %q, want Failed", observation.status.Phase)
	}
}

func TestPlanExecution_DefaultsToClaudeWithoutAHarness(t *testing.T) {
	execution, err := planExecution(runnableAgentRun("solo"), nil, providerless(), nil)
	if err != nil {
		t.Fatalf("planExecution() error = %v", err)
	}
	if execution.agentType != agentv1alpha1.AgentTypeClaude {
		t.Errorf("agent type = %q, want claude", execution.agentType)
	}
}

func TestPlanExecution_KeepsTheRunImageWhenTheToolchainHasNone(t *testing.T) {
	run := runnableAgentRun("solo")

	execution, err := planExecution(run, nil, providerless(), stubToolchain{})
	if err != nil {
		t.Fatalf("planExecution() error = %v", err)
	}
	if execution.image != run.Spec.Image {
		t.Errorf("image = %q, want the run image", execution.image)
	}
}

func TestStandaloneVolume_DefaultsToAGitWorkspaceOfTenGigabytes(t *testing.T) {
	run := runnableAgentRun("solo")
	run.Spec.Workspace = nil

	volume := standaloneVolume(run)

	if volume.workspaceType != agentv1alpha1.WorkspaceTypeGit {
		t.Errorf("workspace type = %q, want git", volume.workspaceType)
	}
	if got := volume.claim.Spec.Resources.Requests.Storage().String(); got != defaultVolumeSize {
		t.Errorf("storage size = %q, want %q", got, defaultVolumeSize)
	}
}

func TestStandaloneVolume_CreatesNoClaimOnceRecorded(t *testing.T) {
	run := runnableAgentRun("solo")
	run.Status.WorkspacePVC = "solo-workspace"

	if volume := standaloneVolume(run); volume.claim != nil {
		t.Errorf("claim = %v, want none once the run records its PVC", volume.claim.Name)
	}
}

func TestSharedWorkspaceVolume_SelectsTheWorktreeBranches(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*agentv1alpha1.AgentWorkspace)
		branch string
		source string
		vcs    agentv1alpha1.WorkspaceType
	}{
		{name: "defaults", mutate: func(*agentv1alpha1.AgentWorkspace) {}, branch: "attached", source: "HEAD", vcs: agentv1alpha1.WorkspaceTypeGit},
		{
			name: "git worktree",
			mutate: func(workspace *agentv1alpha1.AgentWorkspace) {
				workspace.Spec.Git = &agentv1alpha1.GitWorkspaceConfig{Worktree: &agentv1alpha1.WorktreeSpec{Branch: "work", SourceBranch: "main"}}
			},
			branch: "work", source: "main", vcs: agentv1alpha1.WorkspaceTypeGit,
		},
		{
			name: "jj revision",
			mutate: func(workspace *agentv1alpha1.AgentWorkspace) {
				workspace.Spec.Type = agentv1alpha1.WorkspaceTypeJujutsu
				workspace.Spec.Jj = &agentv1alpha1.JujutsuWorkspaceConfig{Revision: "trunk()"}
			},
			branch: "attached", source: "trunk()", vcs: agentv1alpha1.WorkspaceTypeJujutsu,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspace := testutil.NewAgentWorkspace("default", "shared")
			workspace.Status.WorkspacePVC = "shared-ws"
			test.mutate(workspace)

			volume := sharedWorkspaceVolume(runnableAgentRun("attached", testutil.WithWorkspaceRef("shared")), workspace)

			if volume.pvcName != "shared-ws" || volume.workspaceType != test.vcs {
				t.Errorf("volume = %s/%s, want shared-ws/%s", volume.pvcName, volume.workspaceType, test.vcs)
			}
			if volume.binding.WorktreeBranch != test.branch || volume.binding.SourceBranch != test.source {
				t.Errorf("branches = %s from %s, want %s from %s",
					volume.binding.WorktreeBranch, volume.binding.SourceBranch, test.branch, test.source)
			}
		})
	}
}

func providerless() provider.ResolvedProvider { return provider.ResolvedProvider{} }

func failedJobStatus(message string) batchv1.JobStatus {
	return batchv1.JobStatus{Conditions: []batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: message}}}
}
