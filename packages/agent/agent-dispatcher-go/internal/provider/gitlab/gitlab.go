package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/provider"
)

const maxReconciliationPages = 100

type Adapter struct {
	client *provider.Client
}

func New(client *provider.Client) *Adapter {
	return &Adapter{client: client}
}

func (adapter *Adapter) Preview(effect domain.Effect) (domain.RequestPreview, error) {
	requests, err := buildRequests(effect)
	if err != nil {
		return domain.RequestPreview{}, err
	}
	preview := domain.RequestPreview{Requests: make([]domain.ProviderRequestPreview, 0, len(requests))}
	for _, request := range requests {
		step, err := adapter.client.Preview(request)
		if err != nil {
			return domain.RequestPreview{}, err
		}
		preview.Requests = append(preview.Requests, step.Requests...)
	}
	return preview, nil
}

func (adapter *Adapter) Apply(ctx context.Context, effect domain.Effect) (domain.EffectResult, error) {
	if err := adapter.checkPreconditions(ctx, effect); err != nil {
		return domain.EffectResult{}, err
	}
	switch effect.Kind {
	case domain.EffectContextPublish:
		return adapter.publishContext(ctx, effect)
	case domain.EffectReviewPublish:
		return adapter.publishReview(ctx, effect)
	case domain.EffectTicketCreate:
		if result, found, err := adapter.reconcileTicketCreate(ctx, effect); err != nil || found {
			return result, err
		}
	case domain.EffectDeploymentCreate:
		if result, found, err := adapter.reconcileDeploymentCreate(ctx, effect); err != nil || found {
			return result, err
		}
	}
	requests, err := buildRequests(effect)
	if err != nil {
		return domain.EffectResult{}, err
	}
	var result domain.EffectResult
	for _, request := range requests {
		response, err := adapter.client.Do(ctx, request)
		if err != nil {
			return domain.EffectResult{}, err
		}
		if err := graphqlErrors(response); err != nil {
			return domain.EffectResult{}, err
		}
		result = resultFromResponse(response, false)
	}
	return result, nil
}

func (adapter *Adapter) checkPreconditions(ctx context.Context, effect domain.Effect) error {
	if effect.Preconditions == (domain.Preconditions{}) {
		return nil
	}
	projectPath := "/api/v4/projects/" + url.PathEscape(effect.Target.Repository)
	var path string
	switch effect.Kind {
	case domain.EffectTicketUpdate, domain.EffectTicketState:
		path = fmt.Sprintf("%s/issues/%d", projectPath, effect.Target.Number)
	case domain.EffectContextPublish:
		resource, err := noteResource(effect.Target.SubjectKind)
		if err != nil {
			return err
		}
		path = fmt.Sprintf("%s/%s/%d", projectPath, resource, effect.Target.Number)
	case domain.EffectReviewPublish:
		path = fmt.Sprintf("%s/merge_requests/%d", projectPath, effect.Target.Number)
	case domain.EffectDeploymentStatus, domain.EffectDeploymentApproval:
		path = fmt.Sprintf("%s/deployments/%d", projectPath, effect.Target.DeploymentID)
	default:
		return errors.New("preconditions are not supported for this GitLab effect")
	}
	response, err := adapter.client.Do(ctx, provider.Request{Method: http.MethodGet, Path: path})
	if err != nil {
		return err
	}
	var subject struct {
		State     string `json:"state"`
		Status    string `json:"status"`
		UpdatedAt string `json:"updated_at"`
		SHA       string `json:"sha"`
		DiffRefs  struct {
			HeadSHA string `json:"head_sha"`
		} `json:"diff_refs"`
	}
	if err := json.Unmarshal(response.Body, &subject); err != nil {
		return fmt.Errorf("decode GitLab precondition subject: %w", err)
	}
	observedState := subject.State
	if observedState == "" {
		observedState = subject.Status
	}
	if effect.Preconditions.State != "" && !strings.EqualFold(effect.Preconditions.State, observedState) {
		return fmt.Errorf("%w: expected state %q, observed %q", provider.ErrConflict, effect.Preconditions.State, observedState)
	}
	if effect.Preconditions.Revision != "" {
		observed := subject.UpdatedAt
		if subject.DiffRefs.HeadSHA != "" {
			observed = subject.DiffRefs.HeadSHA
		} else if subject.SHA != "" {
			observed = subject.SHA
		}
		if observed != effect.Preconditions.Revision {
			return fmt.Errorf("%w: expected revision %q, observed %q", provider.ErrConflict, effect.Preconditions.Revision, observed)
		}
	}
	if effect.Preconditions.Digest != "" && domain.Digest(response.Body) != effect.Preconditions.Digest {
		return fmt.Errorf("%w: provider subject digest changed", provider.ErrConflict)
	}
	return nil
}

func buildRequests(effect domain.Effect) ([]provider.Request, error) {
	project := url.PathEscape(effect.Target.Repository)
	projectPath := "/api/v4/projects/" + project
	switch effect.Kind {
	case domain.EffectTicketCreate:
		body := map[string]any{
			"title":       effect.Payload.Title,
			"description": appendEffectMarker(effect.Payload.Body, effect.IdempotencyKey),
		}
		if err := setOptionalTicketFields(body, effect.Payload); err != nil {
			return nil, err
		}
		return []provider.Request{{Method: http.MethodPost, Path: projectPath + "/issues", Body: body}}, nil
	case domain.EffectTicketUpdate:
		body := map[string]any{}
		if effect.Payload.Title != "" {
			body["title"] = effect.Payload.Title
		}
		if effect.Payload.Body != "" {
			body["description"] = effect.Payload.Body
		}
		if err := setOptionalTicketFields(body, effect.Payload); err != nil {
			return nil, err
		}
		if len(effect.Payload.AddLabels) > 0 {
			body["add_labels"] = strings.Join(effect.Payload.AddLabels, ",")
		}
		if len(effect.Payload.RemoveLabels) > 0 {
			body["remove_labels"] = strings.Join(effect.Payload.RemoveLabels, ",")
		}
		if len(body) == 0 {
			return nil, errors.New("ticket.update has no fields to update")
		}
		return []provider.Request{{
			Method: http.MethodPut,
			Path:   fmt.Sprintf("%s/issues/%d", projectPath, effect.Target.Number),
			Body:   body,
		}}, nil
	case domain.EffectTicketState:
		var event string
		switch strings.ToLower(effect.Payload.State) {
		case "open", "reopen":
			event = "reopen"
		case "closed", "close":
			event = "close"
		default:
			return nil, fmt.Errorf("unsupported GitLab issue state %q", effect.Payload.State)
		}
		return []provider.Request{{
			Method: http.MethodPut,
			Path:   fmt.Sprintf("%s/issues/%d", projectPath, effect.Target.Number),
			Body:   map[string]string{"state_event": event},
		}}, nil
	case domain.EffectKanbanStatus:
		if effect.Target.WorkItemID != "" {
			return []provider.Request{graphQLRequest(
				`mutation($input:WorkItemUpdateInput!){workItemUpdate(input:$input){workItem{id} errors}}`,
				map[string]any{"input": map[string]any{
					"id":               effect.Target.WorkItemID,
					"statusWidget":     map[string]string{"status": effect.Target.OptionID},
					"clientMutationId": effect.IdempotencyKey,
				}},
			)}, nil
		}
		body := map[string]any{}
		if len(effect.Payload.AddLabels) > 0 {
			body["add_labels"] = strings.Join(effect.Payload.AddLabels, ",")
		}
		if len(effect.Payload.RemoveLabels) > 0 {
			body["remove_labels"] = strings.Join(effect.Payload.RemoveLabels, ",")
		}
		return []provider.Request{{
			Method: http.MethodPut,
			Path:   fmt.Sprintf("%s/issues/%d", projectPath, effect.Target.Number),
			Body:   body,
		}}, nil
	case domain.EffectContextPublish:
		resource, err := noteResource(effect.Target.SubjectKind)
		if err != nil {
			return nil, err
		}
		return []provider.Request{{
			Method: http.MethodPost,
			Path:   fmt.Sprintf("%s/%s/%d/notes", projectPath, resource, effect.Target.Number),
			Body:   map[string]any{"body": effect.Payload.Context.Markdown()},
		}}, nil
	case domain.EffectReviewPublish:
		event := strings.ToUpper(effect.Payload.ReviewEvent)
		if event == "" {
			event = "COMMENT"
		}
		if event != "COMMENT" && event != "APPROVE" && event != "REQUEST_CHANGES" {
			return nil, fmt.Errorf("unsupported GitLab review event %q", event)
		}
		requests := []provider.Request{{
			Method: http.MethodPost,
			Path:   fmt.Sprintf("%s/merge_requests/%d/notes", projectPath, effect.Target.Number),
			Body:   map[string]string{"body": appendEffectMarker(effect.Payload.Body, effect.IdempotencyKey)},
		}}
		if event == "APPROVE" {
			approvalBody := map[string]string{}
			if effect.Payload.CommitID != "" {
				approvalBody["sha"] = effect.Payload.CommitID
			}
			requests = append(requests, provider.Request{
				Method: http.MethodPost,
				Path:   fmt.Sprintf("%s/merge_requests/%d/approve", projectPath, effect.Target.Number),
				Body:   approvalBody,
			})
		}
		return requests, nil
	case domain.EffectDeploymentCreate:
		status := strings.ToLower(effect.Payload.Status)
		if status == "" {
			status = "running"
		}
		if !contains([]string{"running", "success", "failed", "canceled"}, status) {
			return nil, fmt.Errorf("unsupported GitLab deployment status %q", status)
		}
		return []provider.Request{{
			Method: http.MethodPost,
			Path:   projectPath + "/deployments",
			Body: map[string]any{
				"environment": effect.Payload.Environment,
				"sha":         effect.Payload.SHA,
				"ref":         effect.Payload.Ref,
				"tag":         false,
				"status":      status,
			},
		}}, nil
	case domain.EffectDeploymentStatus:
		status := strings.ToLower(effect.Payload.Status)
		if !contains([]string{"running", "success", "failed", "canceled"}, status) {
			return nil, fmt.Errorf("unsupported GitLab deployment status %q", status)
		}
		return []provider.Request{{
			Method: http.MethodPut,
			Path:   fmt.Sprintf("%s/deployments/%d", projectPath, effect.Target.DeploymentID),
			Body:   map[string]string{"status": status},
		}}, nil
	case domain.EffectDeploymentApproval:
		approval := strings.ToLower(effect.Payload.Approval)
		if approval != "approved" && approval != "rejected" {
			return nil, fmt.Errorf("GitLab deployment approval must be approved or rejected")
		}
		body := map[string]string{"status": approval}
		if effect.Payload.Comment != "" {
			body["comment"] = effect.Payload.Comment
		}
		if effect.Payload.RepresentedAs != "" {
			body["represented_as"] = effect.Payload.RepresentedAs
		}
		return []provider.Request{{
			Method: http.MethodPost,
			Path:   fmt.Sprintf("%s/deployments/%d/approval", projectPath, effect.Target.DeploymentID),
			Body:   body,
		}}, nil
	default:
		return nil, fmt.Errorf("GitLab does not implement effect kind %q", effect.Kind)
	}
}

func (adapter *Adapter) publishContext(
	ctx context.Context,
	effect domain.Effect,
) (domain.EffectResult, error) {
	resource, err := noteResource(effect.Target.SubjectKind)
	if err != nil {
		return domain.EffectResult{}, err
	}
	path := fmt.Sprintf(
		"/api/v4/projects/%s/%s/%d/notes",
		url.PathEscape(effect.Target.Repository),
		resource,
		effect.Target.Number,
	)
	body := effect.Payload.Context.Markdown()
	identity := fmt.Sprintf(
		"<!-- xonovex-context id=%s version=%d ",
		effect.Payload.Context.ID,
		effect.Payload.Context.Version,
	)
	matches, err := adapter.listNotes(ctx, path, identity)
	if err != nil {
		return domain.EffectResult{}, err
	}
	if len(matches) > 1 {
		return domain.EffectResult{}, provider.ErrDuplicateContext
	}
	if len(matches) == 1 {
		if matches[0].Body != body {
			return domain.EffectResult{}, provider.ErrConflict
		}
		return domain.EffectResult{
			NativeReference: matches[0].URL,
			StatusCode:      http.StatusOK,
			Reconciled:      true,
		}, nil
	}
	requests, _ := buildRequests(effect)
	response, err := adapter.client.Do(ctx, requests[0])
	if err != nil {
		return domain.EffectResult{}, err
	}
	return resultFromResponse(response, false), nil
}

func (adapter *Adapter) publishReview(
	ctx context.Context,
	effect domain.Effect,
) (domain.EffectResult, error) {
	path := fmt.Sprintf(
		"/api/v4/projects/%s/merge_requests/%d/notes",
		url.PathEscape(effect.Target.Repository),
		effect.Target.Number,
	)
	marker := effectMarker(effect.IdempotencyKey)
	matches, err := adapter.listNotes(ctx, path, marker)
	if err != nil {
		return domain.EffectResult{}, err
	}
	expected := appendEffectMarker(effect.Payload.Body, effect.IdempotencyKey)
	if len(matches) > 1 {
		return domain.EffectResult{}, provider.ErrConflict
	}
	var result domain.EffectResult
	if len(matches) == 1 {
		if matches[0].Body != expected {
			return domain.EffectResult{}, provider.ErrConflict
		}
		result = domain.EffectResult{
			NativeReference: matches[0].URL,
			StatusCode:      http.StatusOK,
			Reconciled:      true,
		}
	} else {
		requests, _ := buildRequests(effect)
		response, err := adapter.client.Do(ctx, requests[0])
		if err != nil {
			return domain.EffectResult{}, err
		}
		result = resultFromResponse(response, false)
	}
	if strings.EqualFold(effect.Payload.ReviewEvent, "APPROVE") {
		requests, _ := buildRequests(effect)
		response, err := adapter.client.Do(ctx, requests[1])
		if err != nil {
			return domain.EffectResult{}, err
		}
		approval := resultFromResponse(response, result.Reconciled)
		if result.NativeReference != "" {
			approval.NativeReference = result.NativeReference
		}
		result = approval
	}
	return result, nil
}

func (adapter *Adapter) reconcileTicketCreate(
	ctx context.Context,
	effect domain.Effect,
) (domain.EffectResult, bool, error) {
	response, err := adapter.client.Do(ctx, provider.Request{
		Method: http.MethodGet,
		Path:   "/api/v4/projects/" + url.PathEscape(effect.Target.Repository) + "/issues",
		Query: url.Values{
			"search":   []string{effect.IdempotencyKey},
			"in":       []string{"description"},
			"scope":    []string{"all"},
			"per_page": []string{"100"},
		},
	})
	if err != nil {
		return domain.EffectResult{}, false, err
	}
	var issues []struct {
		Description string `json:"description"`
		WebURL      string `json:"web_url"`
	}
	if err := json.Unmarshal(response.Body, &issues); err != nil {
		return domain.EffectResult{}, false, fmt.Errorf("decode GitLab issue search: %w", err)
	}
	marker := effectMarker(effect.IdempotencyKey)
	var matches []struct {
		Body string
		URL  string
	}
	for _, issue := range issues {
		if strings.Contains(issue.Description, marker) {
			matches = append(matches, struct {
				Body string
				URL  string
			}{Body: issue.Description, URL: issue.WebURL})
		}
	}
	if len(matches) > 1 {
		return domain.EffectResult{}, false, provider.ErrConflict
	}
	if len(matches) == 1 {
		if matches[0].Body != appendEffectMarker(effect.Payload.Body, effect.IdempotencyKey) {
			return domain.EffectResult{}, false, provider.ErrConflict
		}
		return domain.EffectResult{
			NativeReference: matches[0].URL,
			StatusCode:      http.StatusOK,
			Reconciled:      true,
		}, true, nil
	}
	return domain.EffectResult{}, false, nil
}

func (adapter *Adapter) reconcileDeploymentCreate(
	ctx context.Context,
	effect domain.Effect,
) (domain.EffectResult, bool, error) {
	status := effect.Payload.Status
	if status == "" {
		status = "running"
	}
	response, err := adapter.client.Do(ctx, provider.Request{
		Method: http.MethodGet,
		Path:   "/api/v4/projects/" + url.PathEscape(effect.Target.Repository) + "/deployments",
		Query: url.Values{
			"sha":      []string{effect.Payload.SHA},
			"status":   []string{status},
			"order_by": []string{"created_at"},
			"sort":     []string{"desc"},
			"per_page": []string{"100"},
		},
	})
	if err != nil {
		return domain.EffectResult{}, false, err
	}
	var deployments []struct {
		ID          int64  `json:"id"`
		SHA         string `json:"sha"`
		Ref         string `json:"ref"`
		Status      string `json:"status"`
		Environment struct {
			Name string `json:"name"`
		} `json:"environment"`
	}
	if err := json.Unmarshal(response.Body, &deployments); err != nil {
		return domain.EffectResult{}, false, fmt.Errorf("decode GitLab deployments: %w", err)
	}
	var matches []int64
	for _, deployment := range deployments {
		if deployment.SHA == effect.Payload.SHA &&
			deployment.Ref == effect.Payload.Ref &&
			deployment.Environment.Name == effect.Payload.Environment {
			matches = append(matches, deployment.ID)
		}
	}
	if len(matches) > 1 {
		return domain.EffectResult{}, false, provider.ErrConflict
	}
	if len(matches) == 1 {
		return domain.EffectResult{
			NativeReference: strconv.FormatInt(matches[0], 10),
			StatusCode:      http.StatusOK,
			Reconciled:      true,
		}, true, nil
	}
	return domain.EffectResult{}, false, nil
}

type noteRecord struct {
	Body string
	URL  string
}

func (adapter *Adapter) listNotes(
	ctx context.Context,
	path string,
	containsValue string,
) ([]noteRecord, error) {
	var matches []noteRecord
	for page := 1; page <= maxReconciliationPages; page++ {
		response, err := adapter.client.Do(ctx, provider.Request{
			Method: http.MethodGet,
			Path:   path,
			Query: url.Values{
				"per_page": []string{"100"},
				"page":     []string{strconv.Itoa(page)},
				"sort":     []string{"asc"},
				"order_by": []string{"created_at"},
			},
		})
		if err != nil {
			return nil, err
		}
		var notes []struct {
			ID   int64  `json:"id"`
			Body string `json:"body"`
		}
		if err := json.Unmarshal(response.Body, &notes); err != nil {
			return nil, fmt.Errorf("decode GitLab note reconciliation page: %w", err)
		}
		for _, note := range notes {
			if strings.Contains(note.Body, containsValue) {
				matches = append(matches, noteRecord{
					Body: note.Body,
					URL:  path + "/" + strconv.FormatInt(note.ID, 10),
				})
			}
		}
		if len(notes) < 100 {
			return matches, nil
		}
	}
	return nil, errors.New("GitLab reconciliation exceeded the pagination safety limit")
}

func graphQLRequest(query string, variables map[string]any) provider.Request {
	return provider.Request{
		Method: http.MethodPost,
		Path:   "/api/graphql",
		Body: map[string]any{
			"query":     query,
			"variables": variables,
		},
	}
}

func graphqlErrors(response provider.Response) error {
	var result struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
		Data struct {
			WorkItemUpdate struct {
				Errors []string `json:"errors"`
			} `json:"workItemUpdate"`
		} `json:"data"`
	}
	if json.Unmarshal(response.Body, &result) != nil {
		return nil
	}
	if len(result.Errors) > 0 {
		return &provider.PermanentError{StatusCode: response.StatusCode, Message: result.Errors[0].Message, Body: response.Body}
	}
	if len(result.Data.WorkItemUpdate.Errors) > 0 {
		return &provider.PermanentError{
			StatusCode: response.StatusCode,
			Message:    strings.Join(result.Data.WorkItemUpdate.Errors, "; "),
			Body:       response.Body,
		}
	}
	return nil
}

func setOptionalTicketFields(body map[string]any, payload domain.EffectPayload) error {
	if payload.Labels != nil {
		body["labels"] = strings.Join(payload.Labels, ",")
	}
	if payload.Assignees != nil {
		ids := make([]int64, 0, len(payload.Assignees))
		for _, value := range payload.Assignees {
			id, err := strconv.ParseInt(value, 10, 64)
			if err != nil || id <= 0 {
				return fmt.Errorf("GitLab assignee %q must be a positive numeric ID", value)
			}
			ids = append(ids, id)
		}
		body["assignee_ids"] = ids
	}
	if payload.Milestone != nil {
		body["milestone_id"] = *payload.Milestone
	}
	if payload.IssueType != "" {
		body["issue_type"] = payload.IssueType
	}
	return nil
}

func noteResource(subjectKind string) (string, error) {
	switch subjectKind {
	case "issue":
		return "issues", nil
	case "merge_request":
		return "merge_requests", nil
	default:
		return "", fmt.Errorf("unsupported GitLab note subject kind %q", subjectKind)
	}
}

func appendEffectMarker(body string, idempotencyKey string) string {
	body = strings.TrimRight(body, "\n")
	if body == "" {
		return effectMarker(idempotencyKey)
	}
	return body + "\n\n" + effectMarker(idempotencyKey)
}

func effectMarker(idempotencyKey string) string {
	return "<!-- xonovex-effect idempotency=" + idempotencyKey + " -->"
}

func resultFromResponse(response provider.Response, reconciled bool) domain.EffectResult {
	var value struct {
		ID     any    `json:"id"`
		WebURL string `json:"web_url"`
	}
	_ = json.Unmarshal(response.Body, &value)
	nativeReference := value.WebURL
	if nativeReference == "" && value.ID != nil {
		nativeReference = fmt.Sprint(value.ID)
	}
	return domain.EffectResult{
		NativeReference: nativeReference,
		StatusCode:      response.StatusCode,
		Body:            response.Body,
		Reconciled:      reconciled,
	}
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
