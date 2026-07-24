package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/provider"
)

func TestBuildRequestCoversGitHubEffects(t *testing.T) {
	contextRecord := testContext()
	milestone := int64(3)
	tests := []struct {
		effect domain.Effect
		method string
		path   string
	}{
		{testEffect(domain.EffectTicketCreate, domain.EffectTarget{Repository: "owner/repo"}, domain.EffectPayload{Title: "T", Labels: []string{"bug"}, Assignees: []string{"octo"}, Milestone: &milestone}), "POST", "/issues"},
		{testEffect(domain.EffectTicketUpdate, domain.EffectTarget{Repository: "owner/repo", Number: 2}, domain.EffectPayload{Body: "B"}), "PATCH", "/issues/2"},
		{testEffect(domain.EffectTicketState, domain.EffectTarget{Repository: "owner/repo", Number: 2}, domain.EffectPayload{State: "closed", Status: "completed"}), "PATCH", "/issues/2"},
		{testEffect(domain.EffectKanbanAdd, domain.EffectTarget{Repository: "owner/repo", ProjectID: "P", ContentID: "C"}, domain.EffectPayload{}), "POST", "/graphql"},
		{testEffect(domain.EffectKanbanStatus, domain.EffectTarget{Repository: "owner/repo", ProjectID: "P", ItemID: "I", FieldID: "F", OptionID: "O"}, domain.EffectPayload{}), "POST", "/graphql"},
		{testEffect(domain.EffectKanbanArchive, domain.EffectTarget{Repository: "owner/repo", ProjectID: "P", ItemID: "I"}, domain.EffectPayload{}), "POST", "/graphql"},
		{testEffect(domain.EffectContextPublish, domain.EffectTarget{Repository: "owner/repo", Number: 2}, domain.EffectPayload{Context: &contextRecord}), "POST", "/comments"},
		{testEffect(domain.EffectReviewPublish, domain.EffectTarget{Repository: "owner/repo", Number: 2}, domain.EffectPayload{Body: "review", ReviewEvent: "APPROVE", CommitID: "abc"}), "POST", "/reviews"},
		{testEffect(domain.EffectDeploymentCreate, domain.EffectTarget{Repository: "owner/repo"}, domain.EffectPayload{Ref: "main", SHA: "abc", Environment: "prod"}), "POST", "/deployments"},
		{testEffect(domain.EffectDeploymentStatus, domain.EffectTarget{Repository: "owner/repo", DeploymentID: 4}, domain.EffectPayload{Status: "success", LogURL: "https://logs"}), "POST", "/statuses"},
	}
	for _, test := range tests {
		t.Run(string(test.effect.Kind), func(t *testing.T) {
			request, err := buildRequest(test.effect)
			if err != nil {
				t.Fatal(err)
			}
			if request.Method != test.method || !strings.Contains(request.Path, test.path) {
				t.Fatalf("unexpected request: %+v", request)
			}
			if request.Body == nil {
				t.Fatal("request body is empty")
			}
		})
	}
}

func TestApplyContextReconcilesWithoutWriting(t *testing.T) {
	contextRecord := testContext()
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(writer).Encode([]map[string]any{{
				"body": contextRecord.Markdown(), "html_url": "https://github.test/comment/1",
			}})
		case http.MethodPost:
			writes++
			writer.WriteHeader(http.StatusCreated)
		}
	}))
	defer server.Close()
	adapter := testAdapter(t, server)
	effect := testEffect(
		domain.EffectContextPublish,
		domain.EffectTarget{Repository: "owner/repo", Number: 2},
		domain.EffectPayload{Context: &contextRecord},
	)
	result, err := adapter.Apply(context.Background(), effect)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Reconciled || writes != 0 || result.NativeReference == "" {
		t.Fatalf("unexpected context result: %+v writes=%d", result, writes)
	}
}

func TestApplyTicketCreateAndGraphQLFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/search/issues" {
			_ = json.NewEncoder(writer).Encode(map[string]any{"items": []any{}})
			return
		}
		if request.URL.Path == "/graphql" {
			_ = json.NewEncoder(writer).Encode(map[string]any{"errors": []map[string]string{{"message": "bad field"}}})
			return
		}
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(map[string]any{"id": 9, "html_url": "https://github.test/issues/9"})
	}))
	defer server.Close()
	adapter := testAdapter(t, server)
	effect := testEffect(domain.EffectTicketCreate, domain.EffectTarget{Repository: "owner/repo"}, domain.EffectPayload{Title: "Ticket"})
	result, err := adapter.Apply(context.Background(), effect)
	if err != nil || result.NativeReference == "" {
		t.Fatalf("apply ticket: result=%+v err=%v", result, err)
	}
	graph := testEffect(domain.EffectKanbanStatus, domain.EffectTarget{
		Repository: "owner/repo", ProjectID: "P", ItemID: "I", FieldID: "F", OptionID: "O",
	}, domain.EffectPayload{})
	if _, err := adapter.Apply(context.Background(), graph); err == nil {
		t.Fatal("expected GraphQL error")
	}
}

func TestGitHubPreconditionConflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(writer).Encode(map[string]any{"state": "open", "updated_at": "new"})
	}))
	defer server.Close()
	adapter := testAdapter(t, server)
	effect := testEffect(domain.EffectTicketUpdate, domain.EffectTarget{Repository: "owner/repo", Number: 2}, domain.EffectPayload{Title: "x"})
	effect.Preconditions.Revision = "old"
	if _, err := adapter.Apply(context.Background(), effect); err == nil {
		t.Fatal("expected precondition conflict")
	}
}

func TestPreviewReviewAndDeploymentApply(t *testing.T) {
	var reviewPosts int
	var deploymentPosts int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/reviews"):
			_ = json.NewEncoder(writer).Encode([]any{})
		case request.Method == http.MethodPost && strings.Contains(request.URL.Path, "/reviews"):
			reviewPosts++
			_ = json.NewEncoder(writer).Encode(map[string]any{"id": 10, "html_url": "https://github.test/review/10"})
		case request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/deployments"):
			_ = json.NewEncoder(writer).Encode([]any{})
		case request.Method == http.MethodPost && strings.Contains(request.URL.Path, "/deployments"):
			deploymentPosts++
			_ = json.NewEncoder(writer).Encode(map[string]any{"id": 11, "url": "https://api.github.test/deployments/11"})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	adapter := testAdapter(t, server)
	review := testEffect(
		domain.EffectReviewPublish,
		domain.EffectTarget{Repository: "owner/repo", Number: 2},
		domain.EffectPayload{Body: "Review", ReviewEvent: "COMMENT"},
	)
	preview, err := adapter.Preview(review)
	if err != nil || len(preview.Requests) != 1 {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	if _, err := adapter.Apply(context.Background(), review); err != nil {
		t.Fatal(err)
	}
	deployment := testEffect(
		domain.EffectDeploymentCreate,
		domain.EffectTarget{Repository: "owner/repo"},
		domain.EffectPayload{Ref: "main", SHA: "abc", Environment: "prod"},
	)
	if _, err := adapter.Apply(context.Background(), deployment); err != nil {
		t.Fatal(err)
	}
	if reviewPosts != 1 || deploymentPosts != 1 {
		t.Fatalf("review posts=%d deployment posts=%d", reviewPosts, deploymentPosts)
	}
}

func testAdapter(t *testing.T, server *httptest.Server) *Adapter {
	t.Helper()
	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	return New(provider.NewClientWithHTTP(baseURL, "token", provider.AuthenticationGitHub, &http.Client{Timeout: time.Second}))
}

func testEffect(kind domain.EffectKind, target domain.EffectTarget, payload domain.EffectPayload) domain.Effect {
	return domain.Effect{
		ID: "effect-1", CorrelationID: "correlation-1", IdempotencyKey: "key-1",
		Provider: domain.ProviderGitHub, Tenant: "tenant", Kind: kind, Mode: domain.EffectModeApply,
		Target: target, Payload: payload,
	}
}

func testContext() domain.ContextRecord {
	record := domain.ContextRecord{
		ID: "decision", Version: 1, Type: "decision", Summary: "Use markers.",
		Rationale: "Retries converge.", AppliesTo: "owner/repo@abc", Source: "issue:1",
		Status: "active", Audience: "reviewers", Visibility: "provider-visible",
	}
	record.Digest = domain.Digest(record.SemanticPayload())
	return record
}
