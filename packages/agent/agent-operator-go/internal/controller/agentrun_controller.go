package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
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
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/plugins"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/provider"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/resolver"
	wsshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/shared"
)

// AgentRunReconciler reconciles an AgentRun object
type AgentRunReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder events.EventRecorder
}

// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentruns/finalizers,verbs=update
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentharnesses,verbs=get;list;watch
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agenttoolchains,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *AgentRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// 1. Fetch AgentRun
	var agentRun agentv1alpha1.AgentRun
	if err := r.Get(ctx, req.NamespacedName, &agentRun); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !agentRun.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
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

type resolvedRunExecution struct {
	agentType       agentv1alpha1.AgentType
	providerEnv     []corev1.EnvVar
	providerCliArgs []string
	toolchain       *agentv1alpha1.ToolchainSpec
	defaults        resolver.ResolvedDefaults
	image           string
}

func (r *AgentRunReconciler) resolveRunExecution(ctx context.Context, run *agentv1alpha1.AgentRun) (*resolvedRunExecution, error) {
	harness, err := resolver.ResolveHarness(ctx, r.Client, run.Namespace, run.Spec.HarnessRef, run.Spec.Harness)
	if err != nil {
		return nil, fmt.Errorf("HarnessResolutionFailed: %w", err)
	}

	agentType := agentv1alpha1.AgentTypeClaude
	defaultProvider := ""
	if harness != nil {
		agentType = harness.Spec.Type
		defaultProvider = harness.Spec.DefaultProvider
	}
	resolvedProvider, err := provider.ResolveProvider(ctx, r.Client, run, defaultProvider)
	if err != nil {
		return nil, fmt.Errorf("ProviderResolutionFailed: %w", err)
	}
	toolchain, err := resolver.ResolveToolchain(ctx, r.Client, run.Namespace, run.Spec.ToolchainRef, run.Spec.Toolchain)
	if err != nil {
		return nil, fmt.Errorf("ToolchainResolutionFailed: %w", err)
	}

	defaults := resolver.ApplyHarnessDefaults(run, harness)
	image := defaults.Image
	if resolvedToolchain := plugins.ResolveToolchain(toolchain); resolvedToolchain != nil && resolvedToolchain.Image() != "" {
		image = resolvedToolchain.Image()
	}
	if image == "" {
		return nil, fmt.Errorf("ImageResolutionFailed: admission did not resolve an agent execution image")
	}
	if run.Spec.RuntimeClassName == nil || strings.TrimSpace(*run.Spec.RuntimeClassName) == "" {
		return nil, fmt.Errorf("RuntimeClassResolutionFailed: admission did not resolve a sandboxed runtimeClassName")
	}

	return &resolvedRunExecution{
		agentType:       agentType,
		providerEnv:     resolvedProvider.Environment,
		providerCliArgs: resolvedProvider.CliArgs,
		toolchain:       toolchain,
		defaults:        defaults,
		image:           image,
	}, nil
}

func (r *AgentRunReconciler) reconcileExecutionResources(
	ctx context.Context,
	run *agentv1alpha1.AgentRun,
	execution *resolvedRunExecution,
	pvcName string,
	workspaceType agentv1alpha1.WorkspaceType,
	binding *isoshared.WorkspaceBinding,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("agentRun", run.Name, "namespace", run.Namespace)
	networkPolicy := execution.defaults.NetworkPolicy
	if networkPolicy == nil || !networkPolicy.Disabled {
		resource, err := netshared.BuildNetworkPolicy(run, networkPolicy)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err := ctrl.SetControllerReference(run, resource, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, resource); err != nil {
			if !errors.IsAlreadyExists(err) {
				return ctrl.Result{}, fmt.Errorf("create network policy: %w", err)
			}
			var existing networkingv1.NetworkPolicy
			if err := r.Get(ctx, client.ObjectKeyFromObject(resource), &existing); err != nil {
				return ctrl.Result{}, fmt.Errorf("get existing network policy: %w", err)
			}
			if err := verifyControlledResource(run, &existing, resource.Spec, existing.Spec); err != nil {
				return ctrl.Result{}, fmt.Errorf("existing NetworkPolicy %q: %w", resource.Name, err)
			}
		} else {
			r.Recorder.Eventf(run, nil, corev1.EventTypeNormal, "NetworkPolicyCreated", "NetworkPolicyCreated",
				"Created NetworkPolicy %s", resource.Name)
		}
	}

	jobName := run.Name
	if run.Status.JobName == "" {
		if err := r.ensureAgentServiceAccount(ctx, run.Namespace); err != nil {
			return ctrl.Result{}, err
		}
		job, err := isoshared.BuildJob(
			run,
			execution.providerEnv,
			execution.providerCliArgs,
			pvcName,
			execution.image,
			execution.defaults.Timeout,
			execution.agentType,
			workspaceType,
			execution.toolchain,
			execution.defaults.TTL,
			binding,
		)
		if err != nil {
			logger.Error(err, "failed to build agent Job")
			return r.updatePhase(ctx, run, agentv1alpha1.AgentRunPhaseFailed, fmt.Sprintf("JobBuildFailed: %v", err))
		}
		if err := ctrl.SetControllerReference(run, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, job); err != nil {
			if !errors.IsAlreadyExists(err) {
				return ctrl.Result{}, fmt.Errorf("create agent Job: %w", err)
			}
			var existing batchv1.Job
			if err := r.Get(ctx, client.ObjectKeyFromObject(job), &existing); err != nil {
				return ctrl.Result{}, fmt.Errorf("get existing agent Job: %w", err)
			}
			if err := verifyControlledResource(run, &existing, job.Spec, existing.Spec); err != nil {
				return ctrl.Result{}, fmt.Errorf("existing Job %q: %w", job.Name, err)
			}
		} else {
			logger.Info("created Job", "jobName", jobName, "agentType", execution.agentType,
				"runtimeClass", ptrOrEmpty(run.Spec.RuntimeClassName))
			r.Recorder.Eventf(run, nil, corev1.EventTypeNormal, "AgentRunStarted", "AgentRunStarted",
				"Created Job %s (agent=%s, provider=%s, runtimeClass=%s)",
				jobName, string(execution.agentType), run.Spec.ProviderRef, ptrOrEmpty(run.Spec.RuntimeClassName))
		}

		run.Status.JobName = jobName
		if err := r.updateStatus(ctx, run); err != nil {
			return ctrl.Result{}, err
		}
	}

	var job batchv1.Job
	if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: run.Namespace}, &job); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	return r.reconcileJobStatus(ctx, run, &job)
}

func verifyControlledResource(run *agentv1alpha1.AgentRun, existing client.Object, desiredSpec, existingSpec any) error {
	if !metav1.IsControlledBy(existing, run) {
		return fmt.Errorf("resource is not controlled by AgentRun UID %q", run.UID)
	}
	if !equality.Semantic.DeepDerivative(desiredSpec, existingSpec) {
		return fmt.Errorf("resource specification differs from the admitted execution")
	}
	return nil
}

func (r *AgentRunReconciler) reconcileStandalone(ctx context.Context, agentRun *agentv1alpha1.AgentRun) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues(
		"agentRun", agentRun.Name,
		"namespace", agentRun.Namespace,
	)

	execution, err := r.resolveRunExecution(ctx, agentRun)
	if err != nil {
		log.Error(err, "failed to resolve execution inputs")
		return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseFailed, err.Error())
	}

	// Get workspace config
	wsType := agentv1alpha1.WorkspaceTypeGit
	storageClass := ""
	storageSize := "10Gi"
	if agentRun.Spec.Workspace != nil {
		if agentRun.Spec.Workspace.Type != "" {
			wsType = agentRun.Spec.Workspace.Type
		}
		if agentRun.Spec.Workspace.StorageClass != "" {
			storageClass = agentRun.Spec.Workspace.StorageClass
		}
		if agentRun.Spec.Workspace.StorageSize != "" {
			storageSize = agentRun.Spec.Workspace.StorageSize
		}
	}

	// Create workspace PVC if needed
	pvcName := fmt.Sprintf("%s-workspace", agentRun.Name)
	if agentRun.Status.WorkspacePVC == "" {
		pvc := wsshared.BuildPVC(pvcName, agentRun.Namespace, storageClass, storageSize, agentRun)
		if err := r.Create(ctx, pvc); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "failed to create workspace PVC", "pvcName", pvcName)
			return ctrl.Result{}, err
		}

		agentRun.Status.WorkspacePVC = pvcName
		if _, err := r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseInitializing, ""); err != nil {
			return ctrl.Result{}, err
		}
	}

	return r.reconcileExecutionResources(ctx, agentRun, execution, pvcName, wsType, nil)
}

// ensureAgentServiceAccount creates the dedicated zero-RBAC ServiceAccount agent
// pods bind to, idempotently. It is a shared namespace resource, so it carries no
// owner reference (it outlives any single AgentRun).
func (r *AgentRunReconciler) ensureAgentServiceAccount(ctx context.Context, namespace string) error {
	sa := isoshared.BuildAgentServiceAccount(namespace)
	if err := r.Create(ctx, sa); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *AgentRunReconciler) reconcileWithWorkspace(ctx context.Context, agentRun *agentv1alpha1.AgentRun) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues(
		"agentRun", agentRun.Name,
		"namespace", agentRun.Namespace,
	)

	// Resolve workspace
	ws, err := resolver.ResolveWorkspace(ctx, r.Client, agentRun.Namespace, agentRun.Spec.WorkspaceRef)
	if err != nil {
		log.Error(err, "failed to resolve workspace", "workspaceRef", agentRun.Spec.WorkspaceRef)
		return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseFailed, fmt.Sprintf("WorkspaceResolutionFailed: %v", err))
	}
	if ws.Status.Phase != agentv1alpha1.AgentWorkspacePhaseReady {
		log.Info("workspace not ready, requeuing", "workspace", ws.Name, "phase", ws.Status.Phase)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	if agentRun.Status.WorkspacePVC == "" {
		agentRun.Status.WorkspacePVC = ws.Status.WorkspacePVC
		if _, err := r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseInitializing, ""); err != nil {
			return ctrl.Result{}, err
		}
	}

	execution, err := r.resolveRunExecution(ctx, agentRun)
	if err != nil {
		log.Error(err, "failed to resolve execution inputs")
		return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseFailed, err.Error())
	}

	// Get workspace type and worktree info from AgentWorkspace CRD
	wsType := ws.Spec.Type
	if wsType == "" {
		wsType = agentv1alpha1.WorkspaceTypeGit
	}

	// Get worktree branch/source from workspace CRD's type-specific config
	worktreeBranch := agentRun.Name // default worktree branch to run name
	sourceBranch := "HEAD"
	if ws.Spec.Git != nil && ws.Spec.Git.Worktree != nil {
		if ws.Spec.Git.Worktree.Branch != "" {
			worktreeBranch = ws.Spec.Git.Worktree.Branch
		}
		if ws.Spec.Git.Worktree.SourceBranch != "" {
			sourceBranch = ws.Spec.Git.Worktree.SourceBranch
		}
	}
	if ws.Spec.Jj != nil && ws.Spec.Jj.Revision != "" {
		sourceBranch = ws.Spec.Jj.Revision
	}

	binding := &isoshared.WorkspaceBinding{
		SharedVolumes:    ws.Spec.SharedVolumes,
		SharedVolumePVCs: ws.Status.SharedVolumePVCs,
		WorktreeBranch:   worktreeBranch,
		SourceBranch:     sourceBranch,
		WorkspaceRef:     agentRun.Spec.WorkspaceRef,
	}
	return r.reconcileExecutionResources(ctx, agentRun, execution, ws.Status.WorkspacePVC, wsType, binding)
}

func (r *AgentRunReconciler) reconcileJobStatus(ctx context.Context, agentRun *agentv1alpha1.AgentRun, job *batchv1.Job) (ctrl.Result, error) {
	now := metav1.Now()

	// Check for completion
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			agentRun.Status.CompletionTime = &now
			r.Recorder.Eventf(agentRun, nil, corev1.EventTypeNormal, "AgentRunSucceeded", "AgentRunSucceeded",
				"Agent Job completed successfully")
			return r.updatePhase(ctx, agentRun, agentv1alpha1.AgentRunPhaseSucceeded, "")
		}
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			agentRun.Status.CompletionTime = &now
			r.Recorder.Eventf(agentRun, nil, corev1.EventTypeWarning, "AgentRunFailed", "AgentRunFailed",
				"Agent Job failed: %s", condition.Message)
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
				r.Recorder.Eventf(agentRun, nil, corev1.EventTypeWarning, "AgentRunTimedOut", "AgentRunTimedOut",
					"Agent Job exceeded timeout of %v", agentRun.Spec.Timeout.Duration)
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

	if err := r.updateStatus(ctx, agentRun); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *AgentRunReconciler) updateStatus(ctx context.Context, run *agentv1alpha1.AgentRun) error {
	desired := run.DeepCopy().Status
	key := client.ObjectKeyFromObject(run)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var current agentv1alpha1.AgentRun
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
		*run = current
		return nil
	})
}

func (r *AgentRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1alpha1.AgentRun{}).
		Owns(&batchv1.Job{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Complete(r)
}

func ptrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
