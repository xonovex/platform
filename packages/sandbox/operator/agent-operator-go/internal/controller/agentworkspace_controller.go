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

	agentvalidation "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/validation"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/network/shared"
)

// AgentWorkspaceDependencies are the ports and settings of the AgentWorkspace
// reconciler. WorkspaceInitImage is the digest-pinned image of the init Job,
// which the deployment of the operator supplies: the operator ships no image.
type AgentWorkspaceDependencies struct {
	Store              WorkspaceStore
	VCS                VCSResolver
	Events             EventRecorder
	Clock              Clock
	JobPolicy          isoshared.JobPolicy
	WorkspaceInitImage string
}

// AgentWorkspaceReconciler reconciles an AgentWorkspace object
type AgentWorkspaceReconciler struct {
	deps AgentWorkspaceDependencies
}

// NewAgentWorkspaceReconciler returns an AgentWorkspace reconciler over its
// ports.
func NewAgentWorkspaceReconciler(deps AgentWorkspaceDependencies) *AgentWorkspaceReconciler {
	return &AgentWorkspaceReconciler{deps: deps}
}

// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentworkspaces/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

func (r *AgentWorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("workspace", req.Name, "namespace", req.Namespace)

	workspace, err := r.deps.Store.GetAgentWorkspace(ctx, req.NamespacedName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if workspaceSettled(workspace) {
		return ctrl.Result{}, nil
	}

	claims, status := workspaceClaims(workspace)
	for _, claim := range claims {
		if err := ensureClaim(ctx, r.deps.Store, claim); err != nil {
			logger.Error(err, "failed to create workspace PVC", "pvc", claim.Name)
			return ctrl.Result{}, err
		}
	}
	workspace.Status = status
	if workspace.Status.Phase == "" {
		if _, err := r.setPhase(ctx, workspace, agentv1alpha1.AgentWorkspacePhasePending, ""); err != nil {
			return ctrl.Result{}, err
		}
	}
	if err := ensureAgentServiceAccount(ctx, r.deps.Store, workspace.Namespace); err != nil {
		return ctrl.Result{}, fmt.Errorf("create workspace ServiceAccount: %w", err)
	}
	networkPolicy, err := netshared.BuildWorkspaceInitNetworkPolicy(workspace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("build workspace NetworkPolicy: %w", err)
	}
	setController(workspace, agentWorkspaceKind, networkPolicy)
	if _, err := ensureNetworkPolicy(ctx, r.deps.Store, workspace, agentWorkspaceKind, networkPolicy); err != nil {
		return ctrl.Result{}, err
	}

	if workspace.Status.InitJobName == "" {
		job, err := r.buildInitJob(workspace)
		if err != nil {
			logger.Error(err, "failed to build workspace init job")
			return r.setPhase(ctx, workspace, agentv1alpha1.AgentWorkspacePhaseFailed, fmt.Sprintf("WorkspaceInitBuildFailed: %v", err))
		}
		created, err := ensureJob(ctx, r.deps.Store, workspace, agentWorkspaceKind, job)
		if err != nil {
			logger.Error(err, "failed to create workspace init job")
			return ctrl.Result{}, err
		}
		if created {
			repository, err := agentvalidation.ParseRepositoryURL(workspace.Spec.Repository.URL)
			if err != nil {
				return r.setPhase(ctx, workspace, agentv1alpha1.AgentWorkspacePhaseFailed, "WorkspaceRepositoryInvalid")
			}
			r.record(workspace, normalEvent("WorkspaceInitStarted",
				fmt.Sprintf("Created init Job %s to clone %s", job.Name, repository.Display())))
		}
		workspace.Status.InitJobName = job.Name
		if _, err := r.setPhase(ctx, workspace, agentv1alpha1.AgentWorkspacePhaseInitializing, ""); err != nil {
			return ctrl.Result{}, err
		}
	}

	job, err := r.deps.Store.GetJob(ctx, types.NamespacedName{Namespace: workspace.Namespace, Name: workspaceInitJobName(workspace)})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: pendingRequeue}, nil
		}
		return ctrl.Result{}, err
	}
	observation := observeInitJob(workspace, job, r.deps.Clock.Now())
	if observation.write {
		workspace.Status = observation.status
		if err := r.deps.Store.UpdateAgentWorkspaceStatus(ctx, workspace); err != nil {
			return ctrl.Result{}, err
		}
	}
	for _, e := range observation.events {
		r.record(workspace, e)
	}
	return observation.result, nil
}

// buildInitJob resolves the VCS leaf of the workspace and builds its init Job.
func (r *AgentWorkspaceReconciler) buildInitJob(workspace *agentv1alpha1.AgentWorkspace) (*batchv1.Job, error) {
	vcs, err := r.deps.VCS.ResolveVCSStrategy(workspace.Spec.Type)
	if err != nil {
		return nil, fmt.Errorf("resolve VCS strategy for workspace type %q: %w", workspace.Spec.Type, err)
	}
	job, err := isoshared.BuildWorkspaceInitJob(workspace, workspacePVCName(workspace), r.deps.WorkspaceInitImage,
		workspace.Spec.RuntimeClassName, vcs, r.deps.JobPolicy)
	if err != nil {
		return nil, err
	}
	setController(workspace, agentWorkspaceKind, job)
	return job, nil
}

func (r *AgentWorkspaceReconciler) setPhase(ctx context.Context, workspace *agentv1alpha1.AgentWorkspace, phase agentv1alpha1.AgentWorkspacePhase, message string) (ctrl.Result, error) {
	workspace.Status = withWorkspacePhase(workspace.Status, phase, message, r.deps.Clock.Now())
	return ctrl.Result{}, r.deps.Store.UpdateAgentWorkspaceStatus(ctx, workspace)
}

func (r *AgentWorkspaceReconciler) record(workspace *agentv1alpha1.AgentWorkspace, e event) {
	r.deps.Events.Record(workspace, e.eventType, e.reason, e.message)
}

// SetupWithManager registers the reconciler with the manager.
func (r *AgentWorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1alpha1.AgentWorkspace{}).
		Owns(&batchv1.Job{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Complete(r)
}
