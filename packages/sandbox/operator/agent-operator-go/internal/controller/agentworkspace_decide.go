package controller

import (
	"fmt"
	"maps"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

const agentWorkspaceKind = "AgentWorkspace"

// workspaceSettled reports whether a workspace needs no reconcile: it is being
// deleted or reached a terminal phase.
func workspaceSettled(workspace *agentv1alpha1.AgentWorkspace) bool {
	if !workspace.DeletionTimestamp.IsZero() {
		return true
	}
	return workspace.Status.Phase == agentv1alpha1.AgentWorkspacePhaseReady ||
		workspace.Status.Phase == agentv1alpha1.AgentWorkspacePhaseFailed
}

func workspacePVCName(workspace *agentv1alpha1.AgentWorkspace) string {
	return workspace.Name + "-ws"
}

func workspaceInitJobName(workspace *agentv1alpha1.AgentWorkspace) string {
	return workspace.Name + "-init"
}

// workspaceClaims decides the PVCs a workspace still lacks: the workspace PVC
// and one PVC per shared volume. It returns the claims to create and the status
// that records them.
func workspaceClaims(workspace *agentv1alpha1.AgentWorkspace) ([]*corev1.PersistentVolumeClaim, agentv1alpha1.AgentWorkspaceStatus) {
	status := *workspace.Status.DeepCopy()
	var claims []*corev1.PersistentVolumeClaim
	if status.WorkspacePVC == "" {
		claims = append(claims, wsshared.BuildWorkspacePVC(workspacePVCName(workspace), workspace))
		status.WorkspacePVC = workspacePVCName(workspace)
	}
	status.SharedVolumePVCs = maps.Clone(status.SharedVolumePVCs)
	if status.SharedVolumePVCs == nil {
		status.SharedVolumePVCs = map[string]string{}
	}
	for _, volume := range workspace.Spec.SharedVolumes {
		if _, exists := status.SharedVolumePVCs[volume.Name]; exists {
			continue
		}
		name := fmt.Sprintf("%s-%s", workspace.Name, volume.Name)
		claims = append(claims, wsshared.BuildSharedVolumePVC(name, workspace, volume))
		status.SharedVolumePVCs[volume.Name] = name
	}
	return claims, status
}

// withWorkspacePhase sets the phase of a workspace status; a non-empty message
// adds a condition of the same name.
func withWorkspacePhase(status agentv1alpha1.AgentWorkspaceStatus, phase agentv1alpha1.AgentWorkspacePhase, message string, now time.Time) agentv1alpha1.AgentWorkspaceStatus {
	status.Phase = phase
	if message != "" {
		status.Conditions = append(status.Conditions, phaseCondition(string(phase), message, now))
	}
	return status
}

// workspaceObservation is the status, event and requeue the init Job of a
// workspace decides. write reports whether the status changed.
type workspaceObservation struct {
	status agentv1alpha1.AgentWorkspaceStatus
	events []event
	result ctrl.Result
	write  bool
}

// observeInitJob maps the init Job of a workspace to the workspace status.
func observeInitJob(workspace *agentv1alpha1.AgentWorkspace, job *batchv1.Job, now time.Time) workspaceObservation {
	complete, failed := jobOutcome(job)
	switch {
	case complete:
		return workspaceObservation{
			status: withWorkspacePhase(*workspace.Status.DeepCopy(), agentv1alpha1.AgentWorkspacePhaseReady, "", now),
			events: []event{normalEvent("WorkspaceReady", "Repository cloned successfully, workspace is ready")},
			write:  true,
		}
	case failed != nil:
		return workspaceObservation{
			status: withWorkspacePhase(*workspace.Status.DeepCopy(), agentv1alpha1.AgentWorkspacePhaseFailed, failed.Message, now),
			events: []event{warningEvent("WorkspaceFailed", "Init Job failed: "+failed.Message)},
			write:  true,
		}
	}
	return workspaceObservation{status: workspace.Status, result: ctrl.Result{RequeueAfter: pendingRequeue}}
}
