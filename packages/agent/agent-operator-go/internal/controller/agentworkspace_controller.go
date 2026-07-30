package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/network/shared"
	wsshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/shared"
	agentvalidation "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"
)

// DefaultWorkspaceInitImage is pinned to the multi-architecture OCI index for
// alpine/git 2.54.0. Operators may configure another digest-pinned image.
const DefaultWorkspaceInitImage = "docker.io/alpine/git:2.54.0@sha256:697cb1c85aefc5724febaec2202a974e0d66f6abb6be91a9a86d0c8757af692a"

// AgentWorkspaceReconciler reconciles an AgentWorkspace object
type AgentWorkspaceReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Recorder           events.EventRecorder
	WorkspaceInitImage string
}

// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

func (r *AgentWorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues(
		"workspace", req.Name,
		"namespace", req.Namespace,
	)

	var ws agentv1alpha1.AgentWorkspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !ws.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// Skip if already in terminal phase
	if ws.Status.Phase == agentv1alpha1.AgentWorkspacePhaseReady ||
		ws.Status.Phase == agentv1alpha1.AgentWorkspacePhaseFailed {
		return ctrl.Result{}, nil
	}

	// 1. Create workspace PVC if needed
	workspacePVCName := fmt.Sprintf("%s-ws", ws.Name)
	if ws.Status.WorkspacePVC == "" {
		pvc := wsshared.BuildWorkspacePVC(workspacePVCName, &ws)
		if err := r.Create(ctx, pvc); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "failed to create workspace PVC")
			return ctrl.Result{}, err
		}

		ws.Status.WorkspacePVC = workspacePVCName
	}

	// 2. Create shared volume PVCs if needed
	if ws.Status.SharedVolumePVCs == nil {
		ws.Status.SharedVolumePVCs = make(map[string]string)
	}

	for _, vol := range ws.Spec.SharedVolumes {
		if _, exists := ws.Status.SharedVolumePVCs[vol.Name]; exists {
			continue
		}

		sharedPVCName := fmt.Sprintf("%s-%s", ws.Name, vol.Name)
		pvc := wsshared.BuildSharedVolumePVC(sharedPVCName, &ws, vol)
		if err := r.Create(ctx, pvc); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "failed to create shared volume PVC", "volume", vol.Name)
			return ctrl.Result{}, err
		}

		ws.Status.SharedVolumePVCs[vol.Name] = sharedPVCName
	}

	// Update status to Pending after PVCs are created
	if ws.Status.Phase == "" {
		if _, err := r.updateWorkspacePhase(ctx, &ws, agentv1alpha1.AgentWorkspacePhasePending, ""); err != nil {
			return ctrl.Result{}, err
		}
	}
	if err := ensureAgentServiceAccount(ctx, r.Client, ws.Namespace); err != nil {
		return ctrl.Result{}, fmt.Errorf("create workspace ServiceAccount: %w", err)
	}
	if err := r.ensureWorkspaceInitNetworkPolicy(ctx, &ws); err != nil {
		return ctrl.Result{}, err
	}

	// 3. Create init Job if needed
	initJobName := fmt.Sprintf("%s-init", ws.Name)
	if ws.Status.InitJobName == "" {
		image := r.WorkspaceInitImage
		if image == "" {
			image = DefaultWorkspaceInitImage
		}

		job, err := isoshared.BuildWorkspaceInitJob(&ws, workspacePVCName, image, ws.Spec.RuntimeClassName)
		if err != nil {
			log.Error(err, "failed to build workspace init job")
			return r.updateWorkspacePhase(ctx, &ws, agentv1alpha1.AgentWorkspacePhaseFailed, fmt.Sprintf("WorkspaceInitBuildFailed: %v", err))
		}
		if err := ctrl.SetControllerReference(&ws, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, job); err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Error(err, "failed to create workspace init job")
				return ctrl.Result{}, err
			}
			var existing batchv1.Job
			if err := r.Get(ctx, client.ObjectKeyFromObject(job), &existing); err != nil {
				return ctrl.Result{}, fmt.Errorf("get existing workspace init Job: %w", err)
			}
			if err := verifyControlledResource(&ws, &existing, job.Spec, existing.Spec); err != nil {
				return ctrl.Result{}, fmt.Errorf("existing workspace init Job %q: %w", job.Name, err)
			}
		} else {
			repository, err := agentvalidation.ParseRepositoryURL(ws.Spec.Repository.URL)
			if err != nil {
				return r.updateWorkspacePhase(ctx, &ws, agentv1alpha1.AgentWorkspacePhaseFailed, "WorkspaceRepositoryInvalid")
			}
			r.Recorder.Eventf(&ws, nil, corev1.EventTypeNormal, "WorkspaceInitStarted", "WorkspaceInitStarted",
				"Created init Job %s to clone %s", initJobName, repository.Display())
		}

		ws.Status.InitJobName = initJobName
		if _, err := r.updateWorkspacePhase(ctx, &ws, agentv1alpha1.AgentWorkspacePhaseInitializing, ""); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 4. Watch init Job status
	var job batchv1.Job
	if err := r.Get(ctx, types.NamespacedName{Name: initJobName, Namespace: ws.Namespace}, &job); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	return r.reconcileInitJobStatus(ctx, &ws, &job)
}

func (r *AgentWorkspaceReconciler) ensureWorkspaceInitNetworkPolicy(ctx context.Context, workspace *agentv1alpha1.AgentWorkspace) error {
	networkPolicy, err := netshared.BuildWorkspaceInitNetworkPolicy(workspace)
	if err != nil {
		return fmt.Errorf("build workspace NetworkPolicy: %w", err)
	}
	if err := ctrl.SetControllerReference(workspace, networkPolicy, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, networkPolicy); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("create workspace NetworkPolicy: %w", err)
		}
		var existing networkingv1.NetworkPolicy
		if err := r.Get(ctx, client.ObjectKeyFromObject(networkPolicy), &existing); err != nil {
			return fmt.Errorf("get existing workspace NetworkPolicy: %w", err)
		}
		if err := verifyControlledResource(workspace, &existing, networkPolicy.Spec, existing.Spec); err != nil {
			return fmt.Errorf("existing NetworkPolicy %q: %w", networkPolicy.Name, err)
		}
	}
	return nil
}

func (r *AgentWorkspaceReconciler) reconcileInitJobStatus(ctx context.Context, ws *agentv1alpha1.AgentWorkspace, job *batchv1.Job) (ctrl.Result, error) {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			r.Recorder.Eventf(ws, nil, corev1.EventTypeNormal, "WorkspaceReady", "WorkspaceReady",
				"Repository cloned successfully, workspace is ready")
			return r.updateWorkspacePhase(ctx, ws, agentv1alpha1.AgentWorkspacePhaseReady, "")
		}
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			r.Recorder.Eventf(ws, nil, corev1.EventTypeWarning, "WorkspaceFailed", "WorkspaceFailed",
				"Init Job failed: %s", condition.Message)
			return r.updateWorkspacePhase(ctx, ws, agentv1alpha1.AgentWorkspacePhaseFailed, condition.Message)
		}
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *AgentWorkspaceReconciler) updateWorkspacePhase(ctx context.Context, ws *agentv1alpha1.AgentWorkspace, phase agentv1alpha1.AgentWorkspacePhase, message string) (ctrl.Result, error) {
	ws.Status.Phase = phase

	if message != "" {
		condition := metav1.Condition{
			Type:               string(phase),
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             string(phase),
			Message:            message,
		}
		ws.Status.Conditions = append(ws.Status.Conditions, condition)
	}

	if err := r.updateWorkspaceStatus(ctx, ws); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *AgentWorkspaceReconciler) updateWorkspaceStatus(ctx context.Context, workspace *agentv1alpha1.AgentWorkspace) error {
	desired := workspace.DeepCopy().Status
	key := client.ObjectKeyFromObject(workspace)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var current agentv1alpha1.AgentWorkspace
		if err := r.Get(ctx, key, &current); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}
		current.Status = desired
		if err := r.Status().Update(ctx, &current); err != nil {
			return err
		}
		*workspace = current
		return nil
	})
}

func (r *AgentWorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1alpha1.AgentWorkspace{}).
		Owns(&batchv1.Job{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Complete(r)
}
