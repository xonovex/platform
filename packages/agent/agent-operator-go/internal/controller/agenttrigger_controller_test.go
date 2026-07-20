package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestParseTriggerPath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		wantNamespace string
		wantEndpoint  string
		wantMatched   bool
	}{
		{name: "namespaced endpoint", path: "/v1/triggers/team/review", wantNamespace: "team", wantEndpoint: "review", wantMatched: true},
		{name: "wrong prefix", path: "/triggers/team/review"},
		{name: "missing endpoint", path: "/v1/triggers/team/"},
		{name: "extra segment", path: "/v1/triggers/team/review/extra"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			namespace, endpoint, matched := parseTriggerPath(test.path)
			if namespace != test.wantNamespace || endpoint != test.wantEndpoint || matched != test.wantMatched {
				t.Fatalf("parseTriggerPath(%q) = (%q, %q, %t), want (%q, %q, %t)", test.path, namespace, endpoint, matched, test.wantNamespace, test.wantEndpoint, test.wantMatched)
			}
		})
	}
}

func TestBuildTriggeredRun(t *testing.T) {
	templateRun := testutil.NewAgentRun("team", "template", testutil.WithPrompt("review the change"))
	template := agentv1alpha1.AgentRunTemplate{
		Metadata: metav1.ObjectMeta{
			Labels:      map[string]string{"workflow": "review"},
			Annotations: map[string]string{"example.com/source": "test"},
		},
		Spec: templateRun.Spec,
	}

	run := buildTriggeredRun(template, "team", "review", "AgentTrigger", "request-1")

	if run.Name != "review-fda135ad8395" {
		t.Fatalf("name = %q, want deterministic trigger name", run.Name)
	}
	if run.Namespace != "team" || run.Spec.Prompt != "review the change" {
		t.Fatalf("run = %#v, want copied namespace and spec", run)
	}
	if run.Labels["workflow"] != "review" || run.Annotations["example.com/source"] != "test" {
		t.Fatalf("metadata = %#v, want template metadata", run.ObjectMeta)
	}
	if run.Annotations[agentv1alpha1.TriggeredByKindAnnotation] != "AgentTrigger" || run.Annotations[agentv1alpha1.TriggeredByNameAnnotation] != "review" {
		t.Fatalf("annotations = %#v, want trigger provenance", run.Annotations)
	}
}

func TestReadSecretToken(t *testing.T) {
	ctx := context.Background()
	kubeClient := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "token", Namespace: "team"}, Data: map[string][]byte{"value": []byte("secret")}},
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "empty", Namespace: "team"}, Data: map[string][]byte{"value": {}}},
		).
		Build()

	tests := []struct {
		name      string
		reference agentv1alpha1.SecretKeyRef
		want      string
		wantError bool
	}{
		{name: "token", reference: agentv1alpha1.SecretKeyRef{Name: "token", Key: "value"}, want: "secret"},
		{name: "missing reference", reference: agentv1alpha1.SecretKeyRef{}, wantError: true},
		{name: "missing secret", reference: agentv1alpha1.SecretKeyRef{Name: "missing", Key: "value"}, wantError: true},
		{name: "empty value", reference: agentv1alpha1.SecretKeyRef{Name: "empty", Key: "value"}, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token, err := readSecretToken(ctx, kubeClient, "team", test.reference)
			if (err != nil) != test.wantError {
				t.Fatalf("readSecretToken() error = %v, wantError %t", err, test.wantError)
			}
			if string(token) != test.want {
				t.Fatalf("token = %q, want %q", token, test.want)
			}
		})
	}
}

func TestAgentTriggerReconcilerReportsTokenReadiness(t *testing.T) {
	tests := []struct {
		name       string
		withSecret bool
		wantStatus metav1.ConditionStatus
		wantReason string
	}{
		{name: "available", withSecret: true, wantStatus: metav1.ConditionTrue, wantReason: "TokenAvailable"},
		{name: "unavailable", wantStatus: metav1.ConditionFalse, wantReason: "TokenUnavailable"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trigger := newAgentTrigger()
			objects := []client.Object{trigger}
			if test.withSecret {
				objects = append(objects, newTriggerSecret())
			}
			kubeClient := fake.NewClientBuilder().
				WithScheme(testutil.NewScheme()).
				WithStatusSubresource(&agentv1alpha1.AgentTrigger{}).
				WithObjects(objects...).
				Build()
			reconciler := &AgentTriggerReconciler{Client: kubeClient, Scheme: testutil.NewScheme()}

			if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(trigger)}); err != nil {
				t.Fatalf("Reconcile() error = %v", err)
			}

			var updated agentv1alpha1.AgentTrigger
			if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(trigger), &updated); err != nil {
				t.Fatalf("get AgentTrigger: %v", err)
			}
			condition := apimeta.FindStatusCondition(updated.Status.Conditions, "Ready")
			if condition == nil || condition.Status != test.wantStatus || condition.Reason != test.wantReason {
				t.Fatalf("Ready condition = %#v, want status %q and reason %q", condition, test.wantStatus, test.wantReason)
			}
		})
	}
}

func TestAgentTriggerReceiver(t *testing.T) {
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		method        string
		path          string
		authorization string
		withSecret    bool
		wantStatus    int
	}{
		{name: "creates run", method: http.MethodPost, path: "/v1/triggers/team/review", authorization: "Bearer secret", withSecret: true, wantStatus: http.StatusCreated},
		{name: "rejects method", method: http.MethodGet, path: "/v1/triggers/team/review", withSecret: true, wantStatus: http.StatusMethodNotAllowed},
		{name: "rejects malformed path", method: http.MethodPost, path: "/v1/triggers/team", withSecret: true, wantStatus: http.StatusNotFound},
		{name: "rejects unknown endpoint", method: http.MethodPost, path: "/v1/triggers/team/unknown", withSecret: true, wantStatus: http.StatusNotFound},
		{name: "reports missing token", method: http.MethodPost, path: "/v1/triggers/team/review", authorization: "Bearer secret", wantStatus: http.StatusServiceUnavailable},
		{name: "rejects token", method: http.MethodPost, path: "/v1/triggers/team/review", authorization: "Bearer wrong", withSecret: true, wantStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trigger := newAgentTrigger()
			objects := []client.Object{trigger}
			if test.withSecret {
				objects = append(objects, newTriggerSecret())
			}
			scheme := testutil.NewScheme()
			kubeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&agentv1alpha1.AgentTrigger{}, &agentv1alpha1.AgentRun{}).
				WithObjects(objects...).
				Build()
			receiver := &AgentTriggerReceiver{Client: kubeClient, Scheme: scheme, Now: func() time.Time { return now }}
			request := httptest.NewRequest(test.method, test.path, nil)
			request.Header.Set("Authorization", test.authorization)
			request.Header.Set("Idempotency-Key", "request-1")
			response := httptest.NewRecorder()

			receiver.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			if test.wantStatus != http.StatusCreated {
				return
			}
			var updated agentv1alpha1.AgentTrigger
			if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(trigger), &updated); err != nil {
				t.Fatalf("get AgentTrigger: %v", err)
			}
			if updated.Status.LastRunName == "" || updated.Status.LastTriggeredTime == nil || !updated.Status.LastTriggeredTime.Time.Equal(now) {
				t.Fatalf("status = %#v, want created run at %s", updated.Status, now)
			}
			var run agentv1alpha1.AgentRun
			if err := kubeClient.Get(context.Background(), types.NamespacedName{Namespace: "team", Name: updated.Status.LastRunName}, &run); err != nil {
				t.Fatalf("get created AgentRun: %v", err)
			}
			if run.Spec.Prompt != "review the change" || len(run.OwnerReferences) != 1 || run.OwnerReferences[0].Name != trigger.Name {
				t.Fatalf("created run = %#v, want trigger template and owner", run)
			}
		})
	}
}

func TestAgentTriggerReceiverDoesNotNeedLeaderElection(t *testing.T) {
	if (&AgentTriggerReceiver{}).NeedLeaderElection() {
		t.Fatal("NeedLeaderElection() = true, want false")
	}
}

func newAgentTrigger() *agentv1alpha1.AgentTrigger {
	templateRun := testutil.NewAgentRun("team", "template", testutil.WithPrompt("review the change"))
	return &agentv1alpha1.AgentTrigger{
		ObjectMeta: metav1.ObjectMeta{Name: "review-trigger", Namespace: "team", UID: types.UID("trigger-uid")},
		Spec: agentv1alpha1.AgentTriggerSpec{
			Endpoint:       "review",
			TokenSecretRef: agentv1alpha1.SecretKeyRef{Name: "trigger-token", Key: "token"},
			Template:       agentv1alpha1.AgentRunTemplate{Spec: templateRun.Spec},
		},
	}
}

func newTriggerSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "trigger-token", Namespace: "team"},
		Data:       map[string][]byte{"token": []byte("secret")},
	}
}
