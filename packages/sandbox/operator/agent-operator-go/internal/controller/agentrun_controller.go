package controller

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provider"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/resolver"
)

// AgentRunDependencies are the ports and settings of the AgentRun reconciler.
type AgentRunDependencies struct {
	Store      RunStore
	Toolchains ToolchainResolver
	Harnesses  HarnessResolver
	VCS        VCSResolver
	Events     EventRecorder
	Clock      Clock
	JobPolicy  isoshared.JobPolicy
}

// AgentRunReconciler reconciles an AgentRun object
type AgentRunReconciler struct {
	deps AgentRunDependencies
}

// NewAgentRunReconciler returns an AgentRun reconciler over its ports.
func NewAgentRunReconciler(deps AgentRunDependencies) *AgentRunReconciler {
	return &AgentRunReconciler{deps: deps}
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
	run, err := r.deps.Store.GetAgentRun(ctx, req.NamespacedName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if runSettled(run) {
		return ctrl.Result{}, nil
	}
	if run.Spec.WorkspaceRef != "" {
		return r.reconcileWithWorkspace(ctx, run)
	}
	return r.reconcileStandalone(ctx, run)
}

func (r *AgentRunReconciler) reconcileStandalone(ctx context.Context, run *agentv1alpha1.AgentRun) (ctrl.Result, error) {
	execution, err := r.resolveExecution(ctx, run)
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to resolve execution inputs", "agentRun", run.Name)
		return r.setPhase(ctx, run, agentv1alpha1.AgentRunPhaseFailed, err.Error())
	}

	volume := standaloneVolume(run)
	if volume.claim != nil {
		if err := ensureClaim(ctx, r.deps.Store, volume.claim); err != nil {
			return ctrl.Result{}, err
		}
		run.Status.WorkspacePVC = volume.pvcName
		if _, err := r.setPhase(ctx, run, agentv1alpha1.AgentRunPhaseInitializing, ""); err != nil {
			return ctrl.Result{}, err
		}
	}
	return r.reconcileExecution(ctx, run, execution, volume)
}

func (r *AgentRunReconciler) reconcileWithWorkspace(ctx context.Context, run *agentv1alpha1.AgentRun) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("agentRun", run.Name, "namespace", run.Namespace)

	workspace, err := r.deps.Store.GetAgentWorkspace(ctx, types.NamespacedName{Namespace: run.Namespace, Name: run.Spec.WorkspaceRef})
	if err != nil {
		logger.Error(err, "failed to resolve workspace", "workspaceRef", run.Spec.WorkspaceRef)
		return r.setPhase(ctx, run, agentv1alpha1.AgentRunPhaseFailed,
			fmt.Sprintf("WorkspaceResolutionFailed: failed to get workspace %s: %v", run.Spec.WorkspaceRef, err))
	}
	if workspace.Status.Phase != agentv1alpha1.AgentWorkspacePhaseReady {
		logger.Info("workspace not ready, requeuing", "workspace", workspace.Name, "phase", workspace.Status.Phase)
		return ctrl.Result{RequeueAfter: pendingRequeue}, nil
	}
	if run.Status.WorkspacePVC == "" {
		run.Status.WorkspacePVC = workspace.Status.WorkspacePVC
		if _, err := r.setPhase(ctx, run, agentv1alpha1.AgentRunPhaseInitializing, ""); err != nil {
			return ctrl.Result{}, err
		}
	}

	execution, err := r.resolveExecution(ctx, run)
	if err != nil {
		logger.Error(err, "failed to resolve execution inputs")
		return r.setPhase(ctx, run, agentv1alpha1.AgentRunPhaseFailed, err.Error())
	}
	return r.reconcileExecution(ctx, run, execution, sharedWorkspaceVolume(run, workspace))
}

// resolveExecution reads the harness, provider and toolchain a run names and
// plans its execution. The error message starts with the failed resolution.
func (r *AgentRunReconciler) resolveExecution(ctx context.Context, run *agentv1alpha1.AgentRun) (*runExecution, error) {
	harness, err := resolver.ResolveHarness(ctx, r.deps.Store, run.Namespace, run.Spec.HarnessRef, run.Spec.Harness)
	if err != nil {
		return nil, fmt.Errorf("HarnessResolutionFailed: %w", err)
	}
	defaultProvider := ""
	if harness != nil {
		defaultProvider = harness.Spec.DefaultProvider
	}
	resolvedProvider, err := provider.ResolveProvider(ctx, r.deps.Store, run, defaultProvider)
	if err != nil {
		return nil, fmt.Errorf("ProviderResolutionFailed: %w", err)
	}
	toolchain, err := resolver.ResolveToolchain(ctx, r.deps.Store, run.Namespace, run.Spec.ToolchainRef, run.Spec.Toolchain)
	if err != nil {
		return nil, fmt.Errorf("ToolchainResolutionFailed: %w", err)
	}
	return planExecution(run, harness, resolvedProvider, r.deps.Toolchains.ResolveToolchain(toolchain))
}

func (r *AgentRunReconciler) reconcileExecution(ctx context.Context, run *agentv1alpha1.AgentRun, execution *runExecution, volume runVolume) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("agentRun", run.Name, "namespace", run.Namespace)

	networkPolicy, err := runNetworkPolicy(run, execution.defaults.NetworkPolicy)
	if err != nil {
		return ctrl.Result{}, err
	}
	if networkPolicy != nil {
		created, err := ensureNetworkPolicy(ctx, r.deps.Store, run, agentRunKind, networkPolicy)
		if err != nil {
			return ctrl.Result{}, err
		}
		if created {
			r.record(run, normalEvent("NetworkPolicyCreated", "Created NetworkPolicy "+networkPolicy.Name))
		}
	}

	if run.Status.JobName == "" {
		if err := ensureAgentServiceAccount(ctx, r.deps.Store, run.Namespace); err != nil {
			return ctrl.Result{}, err
		}
		job, err := r.buildJob(run, execution, volume)
		if err != nil {
			logger.Error(err, "failed to build agent Job")
			return r.setPhase(ctx, run, agentv1alpha1.AgentRunPhaseFailed, fmt.Sprintf("JobBuildFailed: %v", err))
		}
		created, err := ensureJob(ctx, r.deps.Store, run, agentRunKind, job)
		if err != nil {
			return ctrl.Result{}, err
		}
		if created {
			logger.Info("created Job", "jobName", job.Name, "agentType", execution.agentType,
				"runtimeClass", ptrOrEmpty(run.Spec.RuntimeClassName))
			r.record(run, normalEvent("AgentRunStarted", fmt.Sprintf(
				"Created Job %s (agent=%s, provider=%s, runtimeClass=%s)",
				job.Name, execution.agentType, run.Spec.ProviderRef, ptrOrEmpty(run.Spec.RuntimeClassName))))
		}
		run.Status.JobName = job.Name
		if err := r.deps.Store.UpdateAgentRunStatus(ctx, run); err != nil {
			return ctrl.Result{}, err
		}
	}

	job, err := r.deps.Store.GetJob(ctx, types.NamespacedName{Namespace: run.Namespace, Name: run.Name})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: pendingRequeue}, nil
		}
		return ctrl.Result{}, err
	}
	pods, err := r.deps.Store.ListJobPods(ctx, job.Namespace, job.Name)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("list pods for Job %q: %w", job.Name, err)
	}
	observation := observeRunJob(run, job, pods, r.deps.Clock.Now())
	if observation.write {
		run.Status = observation.status
		if err := r.deps.Store.UpdateAgentRunStatus(ctx, run); err != nil {
			return ctrl.Result{}, err
		}
	}
	for _, e := range observation.events {
		r.record(run, e)
	}
	return observation.result, nil
}

// buildJob resolves the harness and VCS leaves of the run and builds its Job.
func (r *AgentRunReconciler) buildJob(run *agentv1alpha1.AgentRun, execution *runExecution, volume runVolume) (*batchv1.Job, error) {
	harness, err := r.deps.Harnesses.ResolveHarnessCommand(execution.agentType)
	if err != nil {
		return nil, fmt.Errorf("resolve harness command for agent type %q: %w", execution.agentType, err)
	}
	vcs, err := r.deps.VCS.ResolveVCSStrategy(volume.workspaceType)
	if err != nil {
		return nil, fmt.Errorf("resolve VCS strategy for workspace type %q: %w", volume.workspaceType, err)
	}
	job, err := isoshared.BuildJob(run, isoshared.RunJob{
		ProviderEnv:     execution.providerEnv,
		ProviderCliArgs: execution.providerCliArgs,
		PVCName:         volume.pvcName,
		Image:           execution.image,
		Timeout:         execution.defaults.Timeout,
		AgentType:       execution.agentType,
		Harness:         harness,
		VCS:             vcs,
		TTL:             execution.defaults.TTL,
		Workspace:       volume.binding,
	}, r.deps.JobPolicy)
	if err != nil {
		return nil, err
	}
	setController(run, agentRunKind, job)
	return job, nil
}

func (r *AgentRunReconciler) setPhase(ctx context.Context, run *agentv1alpha1.AgentRun, phase agentv1alpha1.AgentRunPhase, message string) (ctrl.Result, error) {
	run.Status = withRunPhase(run.Status, phase, message, r.deps.Clock.Now())
	return ctrl.Result{}, r.deps.Store.UpdateAgentRunStatus(ctx, run)
}

func (r *AgentRunReconciler) record(run *agentv1alpha1.AgentRun, e event) {
	r.deps.Events.Record(run, e.eventType, e.reason, e.message)
}

// SetupWithManager registers the reconciler with the manager.
func (r *AgentRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1alpha1.AgentRun{}).
		Owns(&batchv1.Job{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Complete(r)
}
