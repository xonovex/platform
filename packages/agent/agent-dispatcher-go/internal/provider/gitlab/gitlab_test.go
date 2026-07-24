package gitlab

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

func TestBuildRequestsCoversGitLabEffects(t *testing.T) {
	contextRecord := testContext()
	tests := []struct {
		effect domain.Effect
		path   string
		count  int
	}{
		{testEffect(domain.EffectTicketCreate, domain.EffectTarget{Repository: "group/project"}, domain.EffectPayload{Title: "T", Assignees: []string{"7"}}), "/issues", 1},
		{testEffect(domain.EffectTicketUpdate, domain.EffectTarget{Repository: "group/project", Number: 2}, domain.EffectPayload{Body: "B", AddLabels: []string{"doing"}}), "/issues/2", 1},
		{testEffect(domain.EffectTicketState, domain.EffectTarget{Repository: "group/project", Number: 2}, domain.EffectPayload{State: "closed"}), "/issues/2", 1},
		{testEffect(domain.EffectKanbanStatus, domain.EffectTarget{Repository: "group/project", Number: 2}, domain.EffectPayload{AddLabels: []string{"doing"}, RemoveLabels: []string{"todo"}}), "/issues/2", 1},
		{testEffect(domain.EffectKanbanStatus, domain.EffectTarget{Repository: "group/project", WorkItemID: "gid://work", OptionID: "gid://status"}, domain.EffectPayload{}), "/api/graphql", 1},
		{testEffect(domain.EffectContextPublish, domain.EffectTarget{Repository: "group/project", SubjectKind: "issue", Number: 2}, domain.EffectPayload{Context: &contextRecord}), "/notes", 1},
		{testEffect(domain.EffectReviewPublish, domain.EffectTarget{Repository: "group/project", Number: 2}, domain.EffectPayload{Body: "Review", ReviewEvent: "APPROVE", CommitID: "abc"}), "/notes", 2},
		{testEffect(domain.EffectDeploymentCreate, domain.EffectTarget{Repository: "group/project"}, domain.EffectPayload{Ref: "main", SHA: "abc", Environment: "prod"}), "/deployments", 1},
		{testEffect(domain.EffectDeploymentStatus, domain.EffectTarget{Repository: "group/project", DeploymentID: 3}, domain.EffectPayload{Status: "success"}), "/deployments/3", 1},
		{testEffect(domain.EffectDeploymentApproval, domain.EffectTarget{Repository: "group/project", DeploymentID: 3}, domain.EffectPayload{Approval: "approved"}), "/approval", 1},
	}
	for _, test := range tests {
		t.Run(string(test.effect.Kind)+test.path, func(t *testing.T) {
			requests, err := buildRequests(test.effect)
			if err != nil {
				t.Fatal(err)
			}
			if len(requests) != test.count || !strings.Contains(requests[0].Path, test.path) {
				t.Fatalf("unexpected requests: %+v", requests)
			}
		})
	}
}

func TestApplyContextAndReviewReconcile(t *testing.T) {
	contextRecord := testContext()
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			body := contextRecord.Markdown()
			if strings.Contains(request.URL.Path, "merge_requests") {
				body = appendEffectMarker("Review", "key-1")
			}
			_ = json.NewEncoder(writer).Encode([]map[string]any{{"id": 4, "body": body}})
			return
		}
		posts++
		_ = json.NewEncoder(writer).Encode(map[string]any{"id": 4})
	}))
	defer server.Close()
	adapter := testAdapter(t, server)
	contextEffect := testEffect(
		domain.EffectContextPublish,
		domain.EffectTarget{Repository: "group/project", SubjectKind: "issue", Number: 2},
		domain.EffectPayload{Context: &contextRecord},
	)
	result, err := adapter.Apply(context.Background(), contextEffect)
	if err != nil || !result.Reconciled {
		t.Fatalf("context result=%+v err=%v", result, err)
	}
	reviewEffect := testEffect(
		domain.EffectReviewPublish,
		domain.EffectTarget{Repository: "group/project", Number: 2},
		domain.EffectPayload{Body: "Review", ReviewEvent: "COMMENT"},
	)
	result, err = adapter.Apply(context.Background(), reviewEffect)
	if err != nil || !result.Reconciled || posts != 0 {
		t.Fatalf("review result=%+v err=%v posts=%d", result, err, posts)
	}
}

func TestApplyNativeStatusGraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"data": map[string]any{"workItemUpdate": map[string]any{"errors": []string{"not allowed"}}},
		})
	}))
	defer server.Close()
	adapter := testAdapter(t, server)
	effect := testEffect(
		domain.EffectKanbanStatus,
		domain.EffectTarget{Repository: "group/project", WorkItemID: "gid://work", OptionID: "gid://status"},
		domain.EffectPayload{},
	)
	if _, err := adapter.Apply(context.Background(), effect); err == nil {
		t.Fatal("expected GraphQL mutation error")
	}
}

func TestInvalidAssigneeAndSubject(t *testing.T) {
	effect := testEffect(domain.EffectTicketCreate, domain.EffectTarget{Repository: "group/project"}, domain.EffectPayload{Title: "T", Assignees: []string{"name"}})
	if _, err := buildRequests(effect); err == nil {
		t.Fatal("expected numeric assignee error")
	}
	effect = testEffect(domain.EffectContextPublish, domain.EffectTarget{Repository: "group/project", SubjectKind: "other", Number: 1}, domain.EffectPayload{Context: pointer(testContext())})
	if _, err := buildRequests(effect); err == nil {
		t.Fatal("expected note subject error")
	}
}

func TestPreviewTicketAndDeploymentApply(t *testing.T) {
	var issuePosts int
	var deploymentPosts int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/issues"):
			_ = json.NewEncoder(writer).Encode([]any{})
		case request.Method == http.MethodPost && strings.Contains(request.URL.Path, "/issues"):
			issuePosts++
			_ = json.NewEncoder(writer).Encode(map[string]any{"id": 10, "web_url": "https://gitlab.test/issues/10"})
		case request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/deployments"):
			_ = json.NewEncoder(writer).Encode([]any{})
		case request.Method == http.MethodPost && strings.Contains(request.URL.Path, "/deployments"):
			deploymentPosts++
			_ = json.NewEncoder(writer).Encode(map[string]any{"id": 11})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	adapter := testAdapter(t, server)
	ticket := testEffect(
		domain.EffectTicketCreate,
		domain.EffectTarget{Repository: "group/project"},
		domain.EffectPayload{Title: "Ticket"},
	)
	preview, err := adapter.Preview(ticket)
	if err != nil || len(preview.Requests) != 1 {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	if result, err := adapter.Apply(context.Background(), ticket); err != nil || result.NativeReference == "" {
		t.Fatalf("ticket result=%+v err=%v", result, err)
	}
	deployment := testEffect(
		domain.EffectDeploymentCreate,
		domain.EffectTarget{Repository: "group/project"},
		domain.EffectPayload{Ref: "main", SHA: "abc", Environment: "prod"},
	)
	if result, err := adapter.Apply(context.Background(), deployment); err != nil || result.NativeReference == "" {
		t.Fatalf("deployment result=%+v err=%v", result, err)
	}
	if issuePosts != 1 || deploymentPosts != 1 {
		t.Fatalf("issue posts=%d deployment posts=%d", issuePosts, deploymentPosts)
	}
}

func TestGitLabPreconditionAndApprovedReview(t *testing.T) {
	var approvalPosts int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/merge_requests/2"):
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"state": "opened", "diff_refs": map[string]string{"head_sha": "abc"},
			})
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/notes"):
			_ = json.NewEncoder(writer).Encode([]map[string]any{{
				"id": 4, "body": appendEffectMarker("Review", "key-1"),
			}})
		case request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/approve"):
			approvalPosts++
			_ = json.NewEncoder(writer).Encode(map[string]any{"id": 2})
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()
	adapter := testAdapter(t, server)
	effect := testEffect(
		domain.EffectReviewPublish,
		domain.EffectTarget{Repository: "group/project", Number: 2},
		domain.EffectPayload{Body: "Review", ReviewEvent: "APPROVE", CommitID: "abc"},
	)
	effect.Preconditions = domain.Preconditions{Revision: "abc", State: "opened"}
	result, err := adapter.Apply(context.Background(), effect)
	if err != nil || approvalPosts != 1 || !result.Reconciled {
		t.Fatalf("result=%+v err=%v approvals=%d", result, err, approvalPosts)
	}
}

func testAdapter(t *testing.T, server *httptest.Server) *Adapter {
	t.Helper()
	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	return New(provider.NewClientWithHTTP(baseURL, "token", provider.AuthenticationGitLab, &http.Client{Timeout: time.Second}))
}

func testEffect(kind domain.EffectKind, target domain.EffectTarget, payload domain.EffectPayload) domain.Effect {
	return domain.Effect{
		ID: "effect-1", CorrelationID: "correlation-1", IdempotencyKey: "key-1",
		Provider: domain.ProviderGitLab, Tenant: "tenant", Kind: kind, Mode: domain.EffectModeApply,
		Target: target, Payload: payload,
	}
}

func testContext() domain.ContextRecord {
	record := domain.ContextRecord{
		ID: "decision", Version: 1, Type: "decision", Summary: "Use markers.",
		Rationale: "Retries converge.", AppliesTo: "group/project@abc", Source: "issue:1",
		Status: "active", Audience: "reviewers", Visibility: "provider-visible",
	}
	record.Digest = domain.Digest(record.SemanticPayload())
	return record
}

func pointer[T any](value T) *T {
	return &value
}
