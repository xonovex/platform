package kube

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/events"
)

// EventRecorder is the event adapter of the reconciler event port.
type EventRecorder struct {
	recorder events.EventRecorder
}

// NewEventRecorder returns an EventRecorder over a client-go event recorder.
func NewEventRecorder(recorder events.EventRecorder) EventRecorder {
	return EventRecorder{recorder: recorder}
}

// Record records an event about object. The reason doubles as the action,
// because each operator event names one action.
func (e EventRecorder) Record(object runtime.Object, eventType, reason, message string) {
	e.recorder.Eventf(object, nil, eventType, reason, reason, "%s", message)
}
