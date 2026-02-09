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
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/resolver"
)

// AgentRunReconciler reconciles an AgentRun object
type AgentRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentruns/finalizers,verbs=update
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *AgentRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// 1. Fetch AgentRun
	var agentRun agentv1alpha1.AgentRun
	if err := r.Get(ctx, req.NamespacedName, &agentRun); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Skip if already completed
	if agentRun.Status.Phase == agentv1alpha1.AgentRunPhaseSucceeded ||
		agentRun.Status.Phase == agentv1alpha1.AgentRunPhaseFailed ||
		agentRun.Status.Phase == agentv1alpha1.AgentRunPhaseTimedOut {
		return ctrl.Result{}, nil
	}

	// Branch based on workspace mode
	if agentRun.Spec.WorkspaceRef != "" {
		return r.reconcileWithWorkspace(ctx, &agentRun)
	}

	return r.reconcileStandalone(ctx, &agentRun)
}

func (r *AgentRunReconciler) reconcileStandalone(ctx context.Context, agentRun *agentv1alpha1.AgentRun) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 2. Resolve AgentConfig defaults
	agentConfig, err := resolver.ResolveConfig(ctx, r.Client, agentRun.Namespace)
	if err != nil {
		log.Error(err, "failed to resolve agent config")
	}

	// 3. Resolve provider
	providerEnv, err := resolver.ResolveProvider(ctx, r.Client, agentRun, agentConfig)
	if err != nil {
		log.Error(err, "failed to resolve provider")
		return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseFailed, fmt.Sprintf("ProviderResolutionFailed: %v", err))
	}

	// 4. Create workspace PVC if needed
	pvcName := fmt.Sprintf("%s-workspace", agentRun.Name)
	if agentRun.Status.WorkspacePVC == "" {
		storageClass := ""
		storageSize := "10Gi"
		if agentConfig != nil {
			if agentConfig.Spec.StorageClass != "" {
				storageClass = agentConfig.Spec.StorageClass
			}
			if agentConfig.Spec.StorageSize != "" {
				storageSize = agentConfig.Spec.StorageSize
			}
		}

		pvc := builder.BuildPVC(pvcName, agentRun.Namespace, storageClass, storageSize, agentRun)
		if err := r.Create(ctx, pvc); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "failed to create workspace PVC")
			return ctrl.Result{}, err
		}

		agentRun.Status.WorkspacePVC = pvcName
		if _, err := r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseInitializing, ""); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 5. Create Job if needed
	jobName := agentRun.Name
	if agentRun.Status.JobName == "" {
		defaultImage := "node:trixie-slim"
		defaultTimeout := time.Hour
		if agentConfig != nil {
			if agentConfig.Spec.DefaultImage != "" {
				defaultImage = agentConfig.Spec.DefaultImage
			}
			if agentConfig.Spec.DefaultTimeout != nil {
				defaultTimeout = agentConfig.Spec.DefaultTimeout.Duration
			}
		}

		job := builder.BuildJob(agentRun, providerEnv, pvcName, defaultImage, defaultTimeout)
		if err := ctrl.SetControllerReference(agentRun, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, job); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "failed to create job")
			return ctrl.Result{}, err
		}

		agentRun.Status.JobName = jobName
		if err := r.Status().Update(ctx, agentRun); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 6. Watch Job status
	var job batchv1.Job
	if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: agentRun.Namespace}, &job); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	return r.reconcileJobStatus(ctx, agentRun, &job)
}

func (r *AgentRunReconciler) reconcileWithWorkspace(ctx context.Context, agentRun *agentv1alpha1.AgentRun) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. Resolve AgentWorkspace
	ws, err := resolver.ResolveWorkspace(ctx, r.Client, agentRun.Namespace, agentRun.Spec.WorkspaceRef)
	if err != nil {
		log.Error(err, "failed to resolve workspace")
		return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseFailed, fmt.Sprintf("WorkspaceResolutionFailed: %v", err))
	}

	// Check workspace is ready
	if ws.Status.Phase != agentv1alpha1.AgentWorkspacePhaseReady {
		log.Info("workspace not ready, requeuing", "workspace", ws.Name, "phase", ws.Status.Phase)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Record workspace PVC in status
	if agentRun.Status.WorkspacePVC == "" {
		agentRun.Status.WorkspacePVC = ws.Status.WorkspacePVC
		if _, err := r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseInitializing, ""); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 2. Resolve AgentConfig defaults
	agentConfig, err := resolver.ResolveConfig(ctx, r.Client, agentRun.Namespace)
	if err != nil {
		log.Error(err, "failed to resolve agent config")
	}

	// 3. Resolve provider
	providerEnv, err := resolver.ResolveProvider(ctx, r.Client, agentRun, agentConfig)
	if err != nil {
		log.Error(err, "failed to resolve provider")
		return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseFailed, fmt.Sprintf("ProviderResolutionFailed: %v", err))
	}

	// 4. Create Job if needed
	jobName := agentRun.Name
	if agentRun.Status.JobName == "" {
		defaultImage := "node:trixie-slim"
		defaultTimeout := time.Hour
		if agentConfig != nil {
			if agentConfig.Spec.DefaultImage != "" {
				defaultImage = agentConfig.Spec.DefaultImage
			}
			if agentConfig.Spec.DefaultTimeout != nil {
				defaultTimeout = agentConfig.Spec.DefaultTimeout.Duration
			}
		}

		job := builder.BuildWorkspaceJob(agentRun, providerEnv, ws.Status.WorkspacePVC, ws.Spec.SharedVolumes, ws.Status.SharedVolumePVCs, defaultImage, defaultTimeout)
		if err := ctrl.SetControllerReference(agentRun, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, job); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "failed to create workspace job")
			return ctrl.Result{}, err
		}

		agentRun.Status.JobName = jobName
		if err := r.Status().Update(ctx, agentRun); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 5. Watch Job status
	var job batchv1.Job
	if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: agentRun.Namespace}, &job); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	return r.reconcileJobStatus(ctx, agentRun, &job)
}

func (r *AgentRunReconciler) reconcileJobStatus(ctx context.Context, agentRun *agentv1alpha1.AgentRun, job *batchv1.Job) (ctrl.Result, error) {
	now := metav1.Now()

	// Check for completion
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			agentRun.Status.CompletionTime = &now
			return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseSucceeded, "")
		}
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			agentRun.Status.CompletionTime = &now
			return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseFailed, condition.Message)
		}
	}

	// Job is still running
	if job.Status.Active > 0 {
		if agentRun.Status.Phase != agentv1alpha1.AgentRunPhaseRunning {
			agentRun.Status.StartTime = &now
			return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseRunning, "")
		}

		// Check timeout
		if agentRun.Spec.Timeout != nil && agentRun.Status.StartTime != nil {
			elapsed := time.Since(agentRun.Status.StartTime.Time)
			if elapsed > agentRun.Spec.Timeout.Duration {
				agentRun.Status.CompletionTime = &now
				return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseTimedOut, "agent run exceeded timeout")
			}
		}
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *AgentRunReconciler) updatePhase(ctx context.Context, agentRun *agentv1alpha1.AgentRun, phase agentv1alpha1.AgentRunPhase, message string) (ctrl.Result, error) {
	agentRun.Status.Phase = phase

	if message != "" {
		condition := metav1.Condition{
			Type:               string(phase),
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             string(phase),
			Message:            message,
		}
		agentRun.Status.Conditions = append(agentRun.Status.Conditions, condition)
	}

	if err := r.Status().Update(ctx, agentRun); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *AgentRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1alpha1.AgentRun{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
