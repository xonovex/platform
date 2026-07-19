package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	runtel "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/telemetry"
)

func TestAgentTriggerRemediationRouter_PostsAuthenticatedTrigger(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		called = true
		if request.Method != http.MethodPost || request.Header.Get("Authorization") != "Bearer remediation-token" {
			t.Fatalf("request = %s, authorization = %q", request.Method, request.Header.Get("Authorization"))
		}
		if request.Header.Get("Idempotency-Key") != "correlation-1:drift-remediation" {
			t.Fatalf("idempotency key = %q", request.Header.Get("Idempotency-Key"))
		}
		response.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	run := &agentv1alpha1.AgentRun{ObjectMeta: metav1.ObjectMeta{
		Name: "run", Namespace: "test",
		Annotations: map[string]string{governanceCorrelationAnnotation: "correlation-1"},
	}}
	router := NewAgentTriggerRemediationRouter(server.URL, "remediation-token")
	if err := router.Raise(context.Background(), run, runtel.DriftAssessment{Detected: true, FailureCodes: []string{"oversight-degraded"}}); err != nil {
		t.Fatalf("raise remediation: %v", err)
	}
	if !called {
		t.Fatal("AgentTrigger endpoint was not called")
	}
}
