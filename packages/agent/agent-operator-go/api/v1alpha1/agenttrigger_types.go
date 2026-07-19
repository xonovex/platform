package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	TriggeredByKindAnnotation = "agent.xonovex.com/trigger-kind"
	TriggeredByNameAnnotation = "agent.xonovex.com/trigger-name"
)

// AgentRunTemplate defines the metadata and immutable execution specification
// copied into a run created by a non-human trigger.
type AgentRunTemplate struct {
	Metadata metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec     AgentRunSpec      `json:"spec"`
}

// AgentScheduleConcurrencyPolicy controls overlap between scheduled runs.
// +kubebuilder:validation:Enum=Allow;Forbid;Replace
type AgentScheduleConcurrencyPolicy string

const (
	AgentScheduleConcurrencyAllow   AgentScheduleConcurrencyPolicy = "Allow"
	AgentScheduleConcurrencyForbid  AgentScheduleConcurrencyPolicy = "Forbid"
	AgentScheduleConcurrencyReplace AgentScheduleConcurrencyPolicy = "Replace"
)

// AgentScheduleSpec defines a cron-driven AgentRun trigger.
type AgentScheduleSpec struct {
	Schedule          string                         `json:"schedule"`
	Suspend           bool                           `json:"suspend,omitempty"`
	ConcurrencyPolicy AgentScheduleConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`
	Template          AgentRunTemplate               `json:"template"`
}

// AgentScheduleStatus records the last observed schedule and active run.
type AgentScheduleStatus struct {
	Conditions       []metav1.Condition `json:"conditions,omitempty"`
	LastScheduleTime *metav1.Time       `json:"lastScheduleTime,omitempty"`
	ActiveRunName    string             `json:"activeRunName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`
// +kubebuilder:printcolumn:name="Suspend",type=boolean,JSONPath=`.spec.suspend`

// AgentSchedule creates AgentRuns at declared cron times.
type AgentSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentScheduleSpec   `json:"spec,omitempty"`
	Status AgentScheduleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentScheduleList contains AgentSchedule resources.
type AgentScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentSchedule `json:"items"`
}

// AgentTriggerSpec defines an authenticated HTTP trigger for AgentRuns.
type AgentTriggerSpec struct {
	Endpoint       string           `json:"endpoint"`
	TokenSecretRef SecretKeyRef     `json:"tokenSecretRef"`
	Template       AgentRunTemplate `json:"template"`
}

// AgentTriggerStatus records receiver readiness and the most recently created run.
type AgentTriggerStatus struct {
	Conditions        []metav1.Condition `json:"conditions,omitempty"`
	LastTriggeredTime *metav1.Time       `json:"lastTriggeredTime,omitempty"`
	LastRunName       string             `json:"lastRunName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.spec.endpoint`

// AgentTrigger creates AgentRuns from authenticated HTTP events.
type AgentTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentTriggerSpec   `json:"spec,omitempty"`
	Status AgentTriggerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentTriggerList contains AgentTrigger resources.
type AgentTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentTrigger `json:"items"`
}

func init() {
	registerTypes(&AgentSchedule{}, &AgentScheduleList{}, &AgentTrigger{}, &AgentTriggerList{})
}
