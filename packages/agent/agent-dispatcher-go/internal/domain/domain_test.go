package domain

import (
	"strings"
	"testing"
)

func TestContextRecordValidateAndMarkdown(t *testing.T) {
	record := validContext()
	if err := record.Validate(); err != nil {
		t.Fatalf("validate context: %v", err)
	}
	markdown := record.Markdown()
	for _, value := range []string{
		"<!-- xonovex-context id=decision.retry version=1 digest=",
		"Context ID: decision.retry",
		"Summary: Use a durable marker.",
		"Visibility: provider-visible",
	} {
		if !strings.Contains(markdown, value) {
			t.Errorf("markdown missing %q", value)
		}
	}
	if !strings.HasSuffix(markdown, "\n") {
		t.Error("markdown does not have a trailing newline")
	}
}

func TestContextRecordRejectsInvalidData(t *testing.T) {
	tests := []struct {
		name   string
		change func(*ContextRecord)
	}{
		{"id", func(record *ContextRecord) { record.ID = "UPPER" }},
		{"version", func(record *ContextRecord) { record.Version = 0 }},
		{"supersedes", func(record *ContextRecord) { record.Version = 2 }},
		{"digest syntax", func(record *ContextRecord) { record.Digest = "sha256:no" }},
		{"digest mismatch", func(record *ContextRecord) { record.Summary = "changed" }},
		{"missing field", func(record *ContextRecord) { record.Source = "" }},
		{"multiline", func(record *ContextRecord) {
			record.Summary = "one\ntwo"
			record.Digest = Digest(record.SemanticPayload())
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record := validContext()
			test.change(&record)
			if err := record.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestEffectValidateAcceptedKinds(t *testing.T) {
	contextRecord := validContext()
	milestone := int64(4)
	tests := []Effect{
		validEffect(EffectTicketCreate, EffectTarget{Repository: "owner/repo"}, EffectPayload{Title: "Ticket", Milestone: &milestone}),
		validEffect(EffectTicketUpdate, EffectTarget{Repository: "owner/repo", Number: 1}, EffectPayload{Title: "Changed"}),
		validEffect(EffectTicketState, EffectTarget{Repository: "owner/repo", Number: 1}, EffectPayload{State: "closed"}),
		validEffect(EffectKanbanAdd, EffectTarget{Repository: "owner/repo", ProjectID: "P", ContentID: "C"}, EffectPayload{}),
		validEffect(EffectKanbanStatus, EffectTarget{Repository: "owner/repo", ProjectID: "P", ItemID: "I", FieldID: "F", OptionID: "O"}, EffectPayload{}),
		validEffect(EffectKanbanArchive, EffectTarget{Repository: "owner/repo", ProjectID: "P", ItemID: "I"}, EffectPayload{}),
		validEffect(EffectContextPublish, EffectTarget{Repository: "owner/repo", Number: 1}, EffectPayload{Context: &contextRecord}),
		validEffect(EffectReviewPublish, EffectTarget{Repository: "owner/repo", Number: 1}, EffectPayload{Body: "Review"}),
		validEffect(EffectDeploymentCreate, EffectTarget{Repository: "owner/repo"}, EffectPayload{Ref: "main", SHA: "abc", Environment: "prod"}),
		validEffect(EffectDeploymentStatus, EffectTarget{Repository: "owner/repo", DeploymentID: 1}, EffectPayload{Status: "success"}),
	}
	for _, effect := range tests {
		t.Run(string(effect.Kind), func(t *testing.T) {
			if err := effect.Validate(); err != nil {
				t.Fatalf("validate effect: %v", err)
			}
			if effect.ResourceKey() == "" {
				t.Error("resource key is empty")
			}
		})
	}

	gitlabStatus := validEffect(
		EffectKanbanStatus,
		EffectTarget{Repository: "group/project", WorkItemID: "gid://work-item", OptionID: "gid://status"},
		EffectPayload{},
	)
	gitlabStatus.Provider = ProviderGitLab
	if err := gitlabStatus.Validate(); err != nil {
		t.Fatalf("validate native GitLab status: %v", err)
	}
	gitlabLabel := validEffect(
		EffectKanbanStatus,
		EffectTarget{Repository: "group/project", Number: 2},
		EffectPayload{AddLabels: []string{"workflow::doing"}},
	)
	gitlabLabel.Provider = ProviderGitLab
	if err := gitlabLabel.Validate(); err != nil {
		t.Fatalf("validate label GitLab status: %v", err)
	}
	approval := validEffect(
		EffectDeploymentApproval,
		EffectTarget{Repository: "group/project", DeploymentID: 2},
		EffectPayload{Approval: "approved"},
	)
	approval.Provider = ProviderGitLab
	if err := approval.Validate(); err != nil {
		t.Fatalf("validate GitLab deployment approval: %v", err)
	}
	gitlabContext := validEffect(
		EffectContextPublish,
		EffectTarget{Repository: "group/project", SubjectKind: "merge_request", Number: 2},
		EffectPayload{Context: &contextRecord},
	)
	gitlabContext.Provider = ProviderGitLab
	if err := gitlabContext.Validate(); err != nil {
		t.Fatalf("validate GitLab context: %v", err)
	}
}

func TestEffectValidateRejectsUnsafeInputs(t *testing.T) {
	tests := []Effect{
		{},
		{ID: "bad id", CorrelationID: "c", IdempotencyKey: "i", Provider: ProviderGitHub, Tenant: "t", Kind: EffectTicketCreate, Mode: EffectModePreview, Target: EffectTarget{Repository: "a/b"}, Payload: EffectPayload{Title: "x"}},
		validEffect(EffectTicketCreate, EffectTarget{Repository: "a/b"}, EffectPayload{}),
		validEffect(EffectTicketUpdate, EffectTarget{Repository: "a/b"}, EffectPayload{}),
		validEffect(EffectKanbanAdd, EffectTarget{Repository: "a/b"}, EffectPayload{}),
		validEffect(EffectReviewPublish, EffectTarget{Repository: "a/b"}, EffectPayload{}),
		validEffect(EffectDeploymentCreate, EffectTarget{Repository: "a/b"}, EffectPayload{}),
	}
	for index, effect := range tests {
		if err := effect.Validate(); err == nil {
			t.Errorf("case %d: expected validation error", index)
		}
	}
	record := validContext()
	record.Visibility = "internal"
	record.Digest = Digest(record.SemanticPayload())
	effect := validEffect(
		EffectContextPublish,
		EffectTarget{Repository: "a/b", Number: 1},
		EffectPayload{Context: &record},
	)
	if err := effect.Validate(); err == nil {
		t.Error("expected internal context publication rejection")
	}
}

func TestNewIDAndDigest(t *testing.T) {
	first, err := NewID("test")
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewID("test")
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasPrefix(first, "test_") {
		t.Fatalf("unexpected generated ids %q and %q", first, second)
	}
	if Digest([]byte("payload")) != "sha256:239f59ed55e737c77147cf55ad0c1b030b6d7ee748a7426952f9b852d5a935e5" {
		t.Error("unexpected digest")
	}
}

func validContext() ContextRecord {
	record := ContextRecord{
		ID:           "decision.retry",
		Version:      1,
		Type:         "decision",
		Summary:      "Use a durable marker.",
		Rationale:    "It makes retries converge.",
		Alternatives: "An in-memory key is lost on restart.",
		Tradeoffs:    "Markers require retention.",
		AppliesTo:    "owner/repo@abc",
		Source:       "issue:1",
		Status:       "active",
		Audience:     "implementers and reviewers",
		Visibility:   "provider-visible",
	}
	record.Digest = Digest(record.SemanticPayload())
	return record
}

func validEffect(kind EffectKind, target EffectTarget, payload EffectPayload) Effect {
	return Effect{
		ID:             "effect-1",
		CorrelationID:  "correlation-1",
		IdempotencyKey: "idempotency-1",
		Provider:       ProviderGitHub,
		Tenant:         "tenant",
		Kind:           kind,
		Mode:           EffectModePreview,
		Target:         target,
		Payload:        payload,
	}
}
