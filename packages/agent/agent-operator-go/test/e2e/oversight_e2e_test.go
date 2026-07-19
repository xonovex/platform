//go:build e2e

package e2e

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestE2E_OversightDriftDemotesAndContainsRun(t *testing.T) {
	namespace := createNamespace(t, "e2e-oversight")
	options := append([]testutil.AgentRunOption{testutil.WithImage("busybox:1.37")}, testutil.E2ESecurityOverrides()...)
	run := testutil.NewAgentRun(namespace, "drift-run", options...)
	run.Annotations = map[string]string{
		"governance.xonovex.com/correlation-id": "e2e-drift-correlation",
	}
	run.Spec.AccountableOwner = "team:e2e"
	run.Spec.Provenance = &agentv1alpha1.AgentRunProvenance{
		Model: "test-model", Provider: "test-provider", PromptReference: "prompt://e2e/1",
		Tools: []string{"Read"}, GrantedPermissions: []string{"repository:read"},
	}
	run.Spec.Autonomy = &agentv1alpha1.AgentAutonomySpec{
		Level: agentv1alpha1.AutonomyLevelUnattended, ProtectedTargets: []string{"repository:main"},
		EscalationRoute: &agentv1alpha1.AgentEscalationRoute{
			Recipient: "human:e2e-on-call", Window: metav1.Duration{Duration: time.Minute}, SafeDefault: agentv1alpha1.EscalationSafeDefaultPause,
		},
	}
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("create A3 run: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var current agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &current); err != nil {
			return false
		}
		return current.Status.Journal != nil && current.Status.JobName != ""
	})

	var current agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &current); err != nil {
		t.Fatalf("get live run: %v", err)
	}
	jobName := current.Status.JobName
	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	current.Annotations["governance.xonovex.com/oversight-state"] = "degraded"
	if err := k8sClient.Update(ctx, &current); err != nil {
		t.Fatalf("induce oversight drift: %v", err)
	}

	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhasePaused, 60*time.Second)
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &current); err != nil {
		t.Fatalf("get contained run: %v", err)
	}
	if current.Status.EffectiveAutonomy != agentv1alpha1.AutonomyLevelSupervised {
		t.Fatalf("effective autonomy = %q, want A2", current.Status.EffectiveAutonomy)
	}
	if current.Status.Containment == nil || current.Status.Containment.Action != "kill-and-pause" {
		t.Fatalf("containment evidence = %#v", current.Status.Containment)
	}

	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: jobName}, &batchv1.Job{})
		return apierrors.IsNotFound(err)
	})
}
