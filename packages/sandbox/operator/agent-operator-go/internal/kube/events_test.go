package kube

import (
	"testing"

	"k8s.io/client-go/tools/events"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

func TestEventRecorder_RecordsTheReasonAndTheMessage(t *testing.T) {
	fake := events.NewFakeRecorder(1)
	recorder := NewEventRecorder(fake)

	recorder.Record(&agentv1alpha1.AgentRun{ObjectMeta: meta("ns", "run")}, "Warning", "AgentRunFailed", "100% failed")

	if got, want := <-fake.Events, "Warning AgentRunFailed 100% failed"; got != want {
		t.Errorf("event = %q, want %q", got, want)
	}
}
