package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestHTTPEscalationRouter_DeliversExactRevisionAndResponseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer router-token" {
			t.Errorf("authorization = %q", request.Header.Get("Authorization"))
		}
		if request.Header.Get("Idempotency-Key") != "correlation-1:escalation:spec-sha256:exact" {
			t.Errorf("idempotency key = %q", request.Header.Get("Idempotency-Key"))
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if payload["subjectRevision"] != "spec-sha256:exact" || payload["responseUrl"] != "https://operator.example/v1/escalations/test/run" {
			t.Errorf("payload = %#v", payload)
		}
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusAccepted)
		_, _ = response.Write([]byte(`{"deliveryReference":"pagerduty:incident-1"}`))
	}))
	defer server.Close()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "run", Namespace: "test",
			Annotations: map[string]string{governanceCorrelationAnnotation: "correlation-1"},
		},
	}
	escalation := &agentv1alpha1.AgentEscalationStatus{
		Recipient:       "human:on-call",
		SubjectRevision: "spec-sha256:exact",
		RequestedAt:     metav1.NewTime(time.Now().UTC()),
		ExpiresAt:       metav1.NewTime(time.Now().UTC().Add(time.Minute)),
		SafeDefault:     agentv1alpha1.EscalationSafeDefaultPause,
		Outcome:         agentv1alpha1.EscalationOutcomePending,
	}
	router := NewHTTPEscalationRouter(server.URL, "router-token", "https://operator.example")

	reference, err := router.Raise(context.Background(), run, escalation)
	if err != nil {
		t.Fatalf("Raise() error = %v", err)
	}
	if reference != "pagerduty:incident-1" {
		t.Fatalf("reference = %q", reference)
	}
}
