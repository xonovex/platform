package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/builder"
)

// AgentWorkspaceReconciler reconciles an AgentWorkspace object
type AgentWorkspaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;delete

func (r *AgentWorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var ws agentv1alpha1.AgentWorkspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Skip if already in terminal phase
	if ws.Status.Phase == agentv1alpha1.AgentWorkspacePhaseReady ||
		ws.Status.Phase == agentv1alpha1.AgentWorkspacePhaseFailed {
		return ctrl.Result{}, nil
	}

	// 1. Create workspace PVC if needed
	workspacePVCName := fmt.Sprintf("%s-ws", ws.Name)
	if ws.Status.WorkspacePVC == "" {
		pvc := builder.BuildWorkspacePVC(workspacePVCName, &ws)
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
		pvc := builder.BuildSharedVolumePVC(sharedPVCName, &ws, vol)
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

	// 3. Create init Job if needed
	initJobName := fmt.Sprintf("%s-init", ws.Name)
	if ws.Status.InitJobName == "" {
		image := "alpine/git:latest"

		job := builder.BuildWorkspaceInitJob(&ws, workspacePVCName, image)
		if err := ctrl.SetControllerReference(&ws, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, job); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "failed to create workspace init job")
			return ctrl.Result{}, err
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

func (r *AgentWorkspaceReconciler) reconcileInitJobStatus(ctx context.Context, ws *agentv1alpha1.AgentWorkspace, job *batchv1.Job) (ctrl.Result, error) {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			return r.updateWorkspacePhase(ctx, ws, agentv1alpha1.AgentWorkspacePhaseReady, "")
		}
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
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

	if err := r.Status().Update(ctx, ws); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *AgentWorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1alpha1.AgentWorkspace{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
