package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/network/shared"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/plugins"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/provider"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/resolver"
	runtel "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/telemetry"
	wsshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/shared"
)

const (
	governanceCorrelationAnnotation   = "governance.xonovex.com/correlation-id"
	oversightStateAnnotation          = "governance.xonovex.com/oversight-state"
	appliedPolicyReferenceAnnotation  = "governance.xonovex.com/applied-policy-reference"
	observedPolicyReferenceAnnotation = "governance.xonovex.com/observed-policy-reference"
)

// AgentRunReconciler reconciles an AgentRun object
type AgentRunReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	Recorder          events.EventRecorder
	Now               func() time.Time
	Telemetry         runtel.Sink
	RemediationRouter RemediationRouter
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

	// Skip if already completed
	if agentRun.Status.Phase == agentv1alpha1.AgentRunPhaseSucceeded ||
		agentRun.Status.Phase == agentv1alpha1.AgentRunPhaseFailed ||
		agentRun.Status.Phase == agentv1alpha1.AgentRunPhaseTimedOut ||
		agentRun.Status.Phase == agentv1alpha1.AgentRunPhasePaused {
		return ctrl.Result{}, nil
	}

	if result, handled, err := r.reconcileOversight(ctx, &agentRun); handled || err != nil {
		return result, err
	}

	// Branch based on workspace mode
	if agentRun.Spec.WorkspaceRef != "" {
		return r.reconcileWithWorkspace(ctx, &agentRun)
	}

	return r.reconcileStandalone(ctx, &agentRun)
}

func (r *AgentRunReconciler) reconcileOversight(ctx context.Context, run *agentv1alpha1.AgentRun) (ctrl.Result, bool, error) {
	requestedLevel := agentv1alpha1.AutonomyLevelManual
	if run.Spec.Autonomy != nil && run.Spec.Autonomy.Level != "" {
		requestedLevel = run.Spec.Autonomy.Level
	}
	if run.Spec.Provenance == nil {
		if requestedLevel != agentv1alpha1.AutonomyLevelUnattended {
			return ctrl.Result{}, false, nil
		}
		run.Status.EffectiveAutonomy = agentv1alpha1.AutonomyLevelSupervised
		return r.updatePhaseHandled(ctx, run, agentv1alpha1.AgentRunPhaseFailed, "A3ProvenanceMissing: unattended execution requires a provenance journal")
	}

	if run.Status.Journal == nil {
		if run.Status.JobName != "" {
			return r.containOversightDrift(ctx, run, "A3ProvenanceDrift: provenance journal disappeared after execution started")
		}
		now := metav1.NewTime(r.currentTime())
		provenance := run.Spec.Provenance
		run.Status.Journal = &agentv1alpha1.AgentRunJournal{
			RecordedAt:         now,
			Generation:         run.Generation,
			AccountableOwner:   run.Spec.AccountableOwner,
			Model:              provenance.Model,
			Provider:           provenance.Provider,
			PromptReference:    provenance.PromptReference,
			Tools:              append([]string(nil), provenance.Tools...),
			GrantedPermissions: append([]string(nil), provenance.GrantedPermissions...),
		}
		run.Status.EffectiveAutonomy = requestedLevel
		if err := r.Status().Update(ctx, run); err != nil {
			return ctrl.Result{}, true, err
		}
		r.recordSignal(ctx, run, "provenance.recorded", "provenance", "healthy")
		return ctrl.Result{Requeue: true}, true, nil
	}

	if requestedLevel == agentv1alpha1.AutonomyLevelUnattended {
		assessment := r.assessA3Oversight(ctx, run)
		if assessment.Detected {
			if r.RemediationRouter != nil {
				if err := r.RemediationRouter.Raise(ctx, run, assessment); err != nil {
					r.recordSignal(ctx, run, "remediation.trigger", "agent-trigger", "failed")
				} else {
					r.recordSignal(ctx, run, "remediation.trigger", "agent-trigger", "created")
				}
			}
			return r.containOversightDrift(ctx, run, strings.Join(assessment.FailureCodes, ","))
		}
	}

	if requestedLevel != agentv1alpha1.AutonomyLevelUnattended || run.Spec.Autonomy == nil || !run.Spec.Autonomy.NeedsHuman {
		return ctrl.Result{}, false, nil
	}
	return r.reconcileEscalation(ctx, run)
}

func (r *AgentRunReconciler) reconcileEscalation(ctx context.Context, run *agentv1alpha1.AgentRun) (ctrl.Result, bool, error) {
	route := run.Spec.Autonomy.EscalationRoute
	if route == nil || route.Recipient == "" || route.Window.Duration <= 0 {
		run.Status.EffectiveAutonomy = agentv1alpha1.AutonomyLevelSupervised
		return r.updatePhaseHandled(ctx, run, agentv1alpha1.AgentRunPhaseFailed, "A3EscalationRouteUnavailable: unattended execution requires a bounded accountable recipient")
	}

	now := r.currentTime()
	if run.Status.Escalation == nil {
		requestedAt := metav1.NewTime(now)
		expiresAt := metav1.NewTime(now.Add(route.Window.Duration))
		run.Status.Escalation = &agentv1alpha1.AgentEscalationStatus{
			Recipient:   route.Recipient,
			RequestedAt: requestedAt,
			ExpiresAt:   expiresAt,
			SafeDefault: route.SafeDefault,
			Outcome:     agentv1alpha1.EscalationOutcomePending,
		}
		apimeta.SetStatusCondition(&run.Status.Conditions, metav1.Condition{
			Type:               "Escalation",
			Status:             metav1.ConditionTrue,
			Reason:             "AwaitingRecipient",
			Message:            "execution is paused pending an accountable response",
			ObservedGeneration: run.Generation,
		})
		if err := r.Status().Update(ctx, run); err != nil {
			return ctrl.Result{}, true, err
		}
		r.recordSignal(ctx, run, "escalation.requested", "escalation-route", "pending")
		return ctrl.Result{RequeueAfter: route.Window.Duration}, true, nil
	}
	if now.Before(run.Status.Escalation.ExpiresAt.Time) {
		return ctrl.Result{RequeueAfter: run.Status.Escalation.ExpiresAt.Sub(now)}, true, nil
	}

	run.Status.CompletionTime = &metav1.Time{Time: now}
	if run.Status.Escalation.SafeDefault == agentv1alpha1.EscalationSafeDefaultAbandon {
		run.Status.Escalation.Outcome = agentv1alpha1.EscalationOutcomeAbandoned
		r.recordSignal(ctx, run, "escalation.expired", "escalation-route", "abandoned")
		return r.updatePhaseHandled(ctx, run, agentv1alpha1.AgentRunPhaseFailed, "EscalationExpired: safe default abandoned the run")
	}
	run.Status.Escalation.Outcome = agentv1alpha1.EscalationOutcomePaused
	r.recordSignal(ctx, run, "escalation.expired", "escalation-route", "paused")
	return r.updatePhaseHandled(ctx, run, agentv1alpha1.AgentRunPhasePaused, "EscalationExpired: safe default paused the run")
}

func (r *AgentRunReconciler) assessA3Oversight(ctx context.Context, run *agentv1alpha1.AgentRun) runtel.DriftAssessment {
	correlationPresent := run.Annotations[governanceCorrelationAnnotation] != ""
	forcedDegradation := run.Annotations[oversightStateAnnotation] == "degraded"
	signals := []runtel.Signal{
		r.oversightSignal(run, "governance-verdict", correlationPresent && !forcedDegradation),
		r.oversightSignal(run, "protected-target", run.Spec.Autonomy != nil && len(run.Spec.Autonomy.ProtectedTargets) > 0),
		r.oversightSignal(run, "escalation-route", run.Spec.Autonomy != nil && run.Spec.Autonomy.EscalationRoute != nil && run.Spec.Autonomy.EscalationRoute.Recipient != ""),
		r.oversightSignal(run, "provenance", run.Status.Journal != nil),
	}
	required := []string{"governance-verdict", "protected-target", "escalation-route", "provenance"}
	if applied := run.Annotations[appliedPolicyReferenceAnnotation]; applied != "" {
		signals = append(signals, r.oversightSignal(run, "applied-reference", run.Annotations[observedPolicyReferenceAnnotation] == applied))
		required = append(required, "applied-reference")
	}
	for _, signal := range signals {
		if r.Telemetry != nil {
			r.Telemetry.Record(ctx, signal)
		}
	}
	assessment := runtel.AssessA3(signals, required)
	if assessment.Detected {
		r.recordSignal(ctx, run, "drift.detected", "oversight", "degraded")
	}
	return assessment
}

func (r *AgentRunReconciler) oversightSignal(run *agentv1alpha1.AgentRun, control string, healthy bool) runtel.Signal {
	outcome := "degraded"
	if healthy {
		outcome = "healthy"
	}
	return runtel.Signal{
		CorrelationID:     runCorrelationID(run),
		RunReference:      fmt.Sprintf("agentrun:%s/%s", run.Namespace, run.Name),
		Kind:              "oversight.control",
		Control:           control,
		Outcome:           outcome,
		EffectiveAutonomy: string(run.Status.EffectiveAutonomy),
	}
}

func (r *AgentRunReconciler) containOversightDrift(ctx context.Context, run *agentv1alpha1.AgentRun, message string) (ctrl.Result, bool, error) {
	if run.Status.JobName != "" {
		job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: run.Status.JobName, Namespace: run.Namespace}}
		if err := r.Delete(ctx, job); err != nil && !errors.IsNotFound(err) {
			return ctrl.Result{}, true, err
		}
	}
	run.Status.EffectiveAutonomy = agentv1alpha1.AutonomyLevelSupervised
	run.Status.CompletionTime = &metav1.Time{Time: r.currentTime()}
	run.Status.Containment = &agentv1alpha1.AgentContainmentStatus{
		RecordedAt:    metav1.Time{Time: r.currentTime()},
		CorrelationID: runCorrelationID(run),
		Reason:        message,
		Action:        "kill-and-pause",
		DemotedTo:     agentv1alpha1.AutonomyLevelSupervised,
	}
	r.recordSignal(ctx, run, "incident.containment", "kill-switch", "paused")
	return r.updatePhaseHandled(ctx, run, agentv1alpha1.AgentRunPhasePaused, message)
}

func (r *AgentRunReconciler) updatePhaseHandled(ctx context.Context, run *agentv1alpha1.AgentRun, phase agentv1alpha1.AgentRunPhase, message string) (ctrl.Result, bool, error) {
	result, err := r.updatePhase(ctx, run, phase, message)
	return result, true, err
}

func (r *AgentRunReconciler) currentTime() time.Time {
	if r.Now != nil {
		return r.Now().UTC()
	}
	return time.Now().UTC()
}

func (r *AgentRunReconciler) recordSignal(ctx context.Context, run *agentv1alpha1.AgentRun, kind, control, outcome string) {
	if r.Telemetry == nil {
		return
	}
	r.Telemetry.Record(ctx, runtel.Signal{
		CorrelationID:     runCorrelationID(run),
		RunReference:      fmt.Sprintf("agentrun:%s/%s", run.Namespace, run.Name),
		Kind:              kind,
		Control:           control,
		Outcome:           outcome,
		EffectiveAutonomy: string(run.Status.EffectiveAutonomy),
	})
}

func runCorrelationID(run *agentv1alpha1.AgentRun) string {
	if correlationID := run.Annotations[governanceCorrelationAnnotation]; correlationID != "" {
		return correlationID
	}
	return fmt.Sprintf("agentrun:%s/%s:generation:%d", run.Namespace, run.Name, run.Generation)
}

type resolvedRunExecution struct {
	agentType   agentv1alpha1.AgentType
	providerEnv map[string]string
	toolchain   *agentv1alpha1.ToolchainSpec
	defaults    resolver.ResolvedDefaults
	image       string
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
	providerEnv, err := provider.ResolveProvider(ctx, r.Client, run, defaultProvider)
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
		agentType:   agentType,
		providerEnv: providerEnv,
		toolchain:   toolchain,
		defaults:    defaults,
		image:       image,
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
		resource := netshared.BuildNetworkPolicy(run, networkPolicy)
		if err := ctrl.SetControllerReference(run, resource, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, resource); err != nil && !errors.IsAlreadyExists(err) {
			return ctrl.Result{}, fmt.Errorf("create network policy: %w", err)
		}
		r.Recorder.Eventf(run, nil, corev1.EventTypeNormal, "NetworkPolicyCreated", "NetworkPolicyCreated",
			"Created NetworkPolicy %s", resource.Name)
	}

	jobName := run.Name
	if run.Status.JobName == "" {
		if err := r.ensureAgentServiceAccount(ctx, run.Namespace); err != nil {
			return ctrl.Result{}, err
		}
		job := isoshared.BuildJob(
			run,
			execution.providerEnv,
			pvcName,
			execution.image,
			execution.defaults.Timeout,
			execution.agentType,
			workspaceType,
			execution.toolchain,
			execution.defaults.TTL,
			binding,
		)
		if err := ctrl.SetControllerReference(run, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, job); err != nil && !errors.IsAlreadyExists(err) {
			return ctrl.Result{}, fmt.Errorf("create agent Job: %w", err)
		}

		logger.Info("creating Job", "jobName", jobName, "agentType", execution.agentType,
			"runtimeClass", ptrOrEmpty(run.Spec.RuntimeClassName))
		r.Recorder.Eventf(run, nil, corev1.EventTypeNormal, "AgentRunStarted", "AgentRunStarted",
			"Created Job %s (agent=%s, provider=%s, runtimeClass=%s)",
			jobName, string(execution.agentType), run.Spec.ProviderRef, ptrOrEmpty(run.Spec.RuntimeClassName))

		run.Status.JobName = jobName
		if err := r.Status().Update(ctx, run); err != nil {
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

	if err := r.Status().Update(ctx, agentRun); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
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
