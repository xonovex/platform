package github

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
	request, err := buildRequest(effect)
	if err != nil {
		return domain.RequestPreview{}, err
	}
	return adapter.client.Preview(request)
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
	request, err := buildRequest(effect)
	if err != nil {
		return domain.EffectResult{}, err
	}
	response, err := adapter.client.Do(ctx, request)
	if err != nil {
		return domain.EffectResult{}, err
	}
	if err := graphqlErrors(response); err != nil {
		return domain.EffectResult{}, err
	}
	return resultFromResponse(response, false), nil
}

func (adapter *Adapter) checkPreconditions(ctx context.Context, effect domain.Effect) error {
	if effect.Preconditions == (domain.Preconditions{}) {
		return nil
	}
	owner, repository, err := repositoryParts(effect.Target.Repository)
	if err != nil {
		return err
	}
	var path string
	switch effect.Kind {
	case domain.EffectTicketUpdate, domain.EffectTicketState, domain.EffectContextPublish:
		path = fmt.Sprintf("/repos/%s/%s/issues/%d", url.PathEscape(owner), url.PathEscape(repository), effect.Target.Number)
	case domain.EffectReviewPublish:
		path = fmt.Sprintf("/repos/%s/%s/pulls/%d", url.PathEscape(owner), url.PathEscape(repository), effect.Target.Number)
	case domain.EffectDeploymentStatus:
		path = fmt.Sprintf("/repos/%s/%s/deployments/%d", url.PathEscape(owner), url.PathEscape(repository), effect.Target.DeploymentID)
	default:
		return errors.New("preconditions are not supported for this GitHub effect")
	}
	response, err := adapter.client.Do(ctx, provider.Request{Method: http.MethodGet, Path: path})
	if err != nil {
		return err
	}
	var subject struct {
		State     string `json:"state"`
		UpdatedAt string `json:"updated_at"`
		SHA       string `json:"sha"`
		Head      struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := json.Unmarshal(response.Body, &subject); err != nil {
		return fmt.Errorf("decode GitHub precondition subject: %w", err)
	}
	if effect.Preconditions.State != "" && !strings.EqualFold(effect.Preconditions.State, subject.State) {
		return fmt.Errorf("%w: expected state %q, observed %q", provider.ErrConflict, effect.Preconditions.State, subject.State)
	}
	if effect.Preconditions.Revision != "" {
		observed := subject.UpdatedAt
		if subject.Head.SHA != "" {
			observed = subject.Head.SHA
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

func buildRequest(effect domain.Effect) (provider.Request, error) {
	owner, repository, err := repositoryParts(effect.Target.Repository)
	if err != nil {
		return provider.Request{}, err
	}
	repositoryPath := "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repository)
	switch effect.Kind {
	case domain.EffectTicketCreate:
		body := map[string]any{
			"title": effect.Payload.Title,
			"body":  appendEffectMarker(effect.Payload.Body, effect.IdempotencyKey),
		}
		setOptionalTicketFields(body, effect.Payload)
		return provider.Request{Method: http.MethodPost, Path: repositoryPath + "/issues", Body: body}, nil
	case domain.EffectTicketUpdate:
		body := map[string]any{}
		if effect.Payload.Title != "" {
			body["title"] = effect.Payload.Title
		}
		if effect.Payload.Body != "" {
			body["body"] = effect.Payload.Body
		}
		setOptionalTicketFields(body, effect.Payload)
		if len(body) == 0 {
			return provider.Request{}, errors.New("ticket.update has no fields to update")
		}
		return provider.Request{
			Method: http.MethodPatch,
			Path:   fmt.Sprintf("%s/issues/%d", repositoryPath, effect.Target.Number),
			Body:   body,
		}, nil
	case domain.EffectTicketState:
		state := strings.ToLower(effect.Payload.State)
		if state != "open" && state != "closed" {
			return provider.Request{}, errors.New("GitHub ticket state must be open or closed")
		}
		body := map[string]any{"state": state}
		if state == "closed" && (effect.Payload.Status == "completed" || effect.Payload.Status == "not_planned") {
			body["state_reason"] = effect.Payload.Status
		}
		return provider.Request{
			Method: http.MethodPatch,
			Path:   fmt.Sprintf("%s/issues/%d", repositoryPath, effect.Target.Number),
			Body:   body,
		}, nil
	case domain.EffectKanbanAdd:
		return graphQLRequest(
			`mutation($input:AddProjectV2ItemByIdInput!){addProjectV2ItemById(input:$input){item{id}}}`,
			map[string]any{"input": map[string]any{
				"projectId":        effect.Target.ProjectID,
				"contentId":        effect.Target.ContentID,
				"clientMutationId": effect.IdempotencyKey,
			}},
		), nil
	case domain.EffectKanbanStatus:
		return graphQLRequest(
			`mutation($input:UpdateProjectV2ItemFieldValueInput!){updateProjectV2ItemFieldValue(input:$input){projectV2Item{id}}}`,
			map[string]any{"input": map[string]any{
				"projectId":        effect.Target.ProjectID,
				"itemId":           effect.Target.ItemID,
				"fieldId":          effect.Target.FieldID,
				"value":            map[string]string{"singleSelectOptionId": effect.Target.OptionID},
				"clientMutationId": effect.IdempotencyKey,
			}},
		), nil
	case domain.EffectKanbanArchive:
		return graphQLRequest(
			`mutation($input:ArchiveProjectV2ItemInput!){archiveProjectV2Item(input:$input){item{id isArchived}}}`,
			map[string]any{"input": map[string]any{
				"projectId":        effect.Target.ProjectID,
				"itemId":           effect.Target.ItemID,
				"clientMutationId": effect.IdempotencyKey,
			}},
		), nil
	case domain.EffectContextPublish:
		return provider.Request{
			Method: http.MethodPost,
			Path:   fmt.Sprintf("%s/issues/%d/comments", repositoryPath, effect.Target.Number),
			Body:   map[string]string{"body": effect.Payload.Context.Markdown()},
		}, nil
	case domain.EffectReviewPublish:
		event := strings.ToUpper(effect.Payload.ReviewEvent)
		if event == "" {
			event = "COMMENT"
		}
		if event != "APPROVE" && event != "REQUEST_CHANGES" && event != "COMMENT" {
			return provider.Request{}, fmt.Errorf("unsupported GitHub review event %q", event)
		}
		body := map[string]any{
			"body":  appendEffectMarker(effect.Payload.Body, effect.IdempotencyKey),
			"event": event,
		}
		if effect.Payload.CommitID != "" {
			body["commit_id"] = effect.Payload.CommitID
		}
		return provider.Request{
			Method: http.MethodPost,
			Path:   fmt.Sprintf("%s/pulls/%d/reviews", repositoryPath, effect.Target.Number),
			Body:   body,
		}, nil
	case domain.EffectDeploymentCreate:
		body := map[string]any{
			"ref":                    effect.Payload.SHA,
			"environment":            effect.Payload.Environment,
			"description":            effect.Payload.Description,
			"transient_environment":  effect.Payload.TransientEnvironment,
			"production_environment": effect.Payload.ProductionEnvironment,
			"payload": map[string]string{
				"xonovex_idempotency_key": effect.IdempotencyKey,
				"requested_ref":           effect.Payload.Ref,
			},
		}
		return provider.Request{Method: http.MethodPost, Path: repositoryPath + "/deployments", Body: body}, nil
	case domain.EffectDeploymentStatus:
		state := strings.ToLower(effect.Payload.Status)
		if !contains([]string{"error", "failure", "inactive", "in_progress", "queued", "pending", "success"}, state) {
			return provider.Request{}, fmt.Errorf("unsupported GitHub deployment status %q", state)
		}
		body := map[string]any{"state": state}
		if effect.Payload.Description != "" {
			body["description"] = effect.Payload.Description
		}
		if effect.Payload.LogURL != "" {
			body["log_url"] = effect.Payload.LogURL
		}
		if effect.Payload.EnvironmentURL != "" {
			body["environment_url"] = effect.Payload.EnvironmentURL
		}
		return provider.Request{
			Method: http.MethodPost,
			Path:   fmt.Sprintf("%s/deployments/%d/statuses", repositoryPath, effect.Target.DeploymentID),
			Body:   body,
		}, nil
	default:
		return provider.Request{}, fmt.Errorf("GitHub does not implement effect kind %q", effect.Kind)
	}
}

func (adapter *Adapter) publishContext(
	ctx context.Context,
	effect domain.Effect,
) (domain.EffectResult, error) {
	body := effect.Payload.Context.Markdown()
	identity := fmt.Sprintf(
		"<!-- xonovex-context id=%s version=%d ",
		effect.Payload.Context.ID,
		effect.Payload.Context.Version,
	)
	repositoryPath, err := issuePath(effect.Target.Repository, effect.Target.Number)
	if err != nil {
		return domain.EffectResult{}, err
	}
	matches, err := adapter.listBodies(ctx, repositoryPath+"/comments", identity)
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
	request, _ := buildRequest(effect)
	response, err := adapter.client.Do(ctx, request)
	if err != nil {
		return domain.EffectResult{}, err
	}
	return resultFromResponse(response, false), nil
}

func (adapter *Adapter) publishReview(
	ctx context.Context,
	effect domain.Effect,
) (domain.EffectResult, error) {
	marker := effectMarker(effect.IdempotencyKey)
	owner, repository, err := repositoryParts(effect.Target.Repository)
	if err != nil {
		return domain.EffectResult{}, err
	}
	path := fmt.Sprintf(
		"/repos/%s/%s/pulls/%d/reviews",
		url.PathEscape(owner),
		url.PathEscape(repository),
		effect.Target.Number,
	)
	matches, err := adapter.listBodies(ctx, path, marker)
	if err != nil {
		return domain.EffectResult{}, err
	}
	expected := appendEffectMarker(effect.Payload.Body, effect.IdempotencyKey)
	if len(matches) > 1 {
		return domain.EffectResult{}, provider.ErrConflict
	}
	if len(matches) == 1 {
		if matches[0].Body != expected {
			return domain.EffectResult{}, provider.ErrConflict
		}
		return domain.EffectResult{
			NativeReference: matches[0].URL,
			StatusCode:      http.StatusOK,
			Reconciled:      true,
		}, nil
	}
	request, _ := buildRequest(effect)
	response, err := adapter.client.Do(ctx, request)
	if err != nil {
		return domain.EffectResult{}, err
	}
	return resultFromResponse(response, false), nil
}

func (adapter *Adapter) reconcileTicketCreate(
	ctx context.Context,
	effect domain.Effect,
) (domain.EffectResult, bool, error) {
	owner, repository, err := repositoryParts(effect.Target.Repository)
	if err != nil {
		return domain.EffectResult{}, false, err
	}
	query := fmt.Sprintf(`repo:%s/%s in:body "%s"`, owner, repository, effectMarker(effect.IdempotencyKey))
	response, err := adapter.client.Do(ctx, provider.Request{
		Method: http.MethodGet,
		Path:   "/search/issues",
		Query:  url.Values{"q": []string{query}, "per_page": []string{"100"}},
	})
	if err != nil {
		return domain.EffectResult{}, false, err
	}
	var result struct {
		Items []struct {
			Body    string `json:"body"`
			HTMLURL string `json:"html_url"`
		} `json:"items"`
	}
	if err := json.Unmarshal(response.Body, &result); err != nil {
		return domain.EffectResult{}, false, fmt.Errorf("decode GitHub issue search: %w", err)
	}
	marker := effectMarker(effect.IdempotencyKey)
	var found []struct {
		Body string
		URL  string
	}
	for _, item := range result.Items {
		if strings.Contains(item.Body, marker) {
			found = append(found, struct {
				Body string
				URL  string
			}{Body: item.Body, URL: item.HTMLURL})
		}
	}
	if len(found) > 1 {
		return domain.EffectResult{}, false, provider.ErrConflict
	}
	if len(found) == 1 {
		expected := appendEffectMarker(effect.Payload.Body, effect.IdempotencyKey)
		if found[0].Body != expected {
			return domain.EffectResult{}, false, provider.ErrConflict
		}
		return domain.EffectResult{
			NativeReference: found[0].URL,
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
	owner, repository, err := repositoryParts(effect.Target.Repository)
	if err != nil {
		return domain.EffectResult{}, false, err
	}
	response, err := adapter.client.Do(ctx, provider.Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/repos/%s/%s/deployments", url.PathEscape(owner), url.PathEscape(repository)),
		Query: url.Values{
			"sha":         []string{effect.Payload.SHA},
			"environment": []string{effect.Payload.Environment},
			"per_page":    []string{"100"},
		},
	})
	if err != nil {
		return domain.EffectResult{}, false, err
	}
	var deployments []struct {
		ID      int64          `json:"id"`
		URL     string         `json:"url"`
		Payload map[string]any `json:"payload"`
	}
	if err := json.Unmarshal(response.Body, &deployments); err != nil {
		return domain.EffectResult{}, false, fmt.Errorf("decode GitHub deployments: %w", err)
	}
	var matches []struct {
		ID  int64
		URL string
	}
	for _, deployment := range deployments {
		if deployment.Payload["xonovex_idempotency_key"] == effect.IdempotencyKey {
			matches = append(matches, struct {
				ID  int64
				URL string
			}{ID: deployment.ID, URL: deployment.URL})
		}
	}
	if len(matches) > 1 {
		return domain.EffectResult{}, false, provider.ErrConflict
	}
	if len(matches) == 1 {
		return domain.EffectResult{
			NativeReference: matches[0].URL,
			StatusCode:      http.StatusOK,
			Body:            json.RawMessage(strconv.FormatInt(matches[0].ID, 10)),
			Reconciled:      true,
		}, true, nil
	}
	return domain.EffectResult{}, false, nil
}

type bodyRecord struct {
	Body string
	URL  string
}

func (adapter *Adapter) listBodies(
	ctx context.Context,
	path string,
	containsValue string,
) ([]bodyRecord, error) {
	var matches []bodyRecord
	for page := 1; page <= maxReconciliationPages; page++ {
		response, err := adapter.client.Do(ctx, provider.Request{
			Method: http.MethodGet,
			Path:   path,
			Query: url.Values{
				"per_page": []string{"100"},
				"page":     []string{strconv.Itoa(page)},
			},
		})
		if err != nil {
			return nil, err
		}
		var records []struct {
			Body    string `json:"body"`
			HTMLURL string `json:"html_url"`
			URL     string `json:"url"`
		}
		if err := json.Unmarshal(response.Body, &records); err != nil {
			return nil, fmt.Errorf("decode GitHub reconciliation page: %w", err)
		}
		for _, record := range records {
			if strings.Contains(record.Body, containsValue) {
				nativeURL := record.HTMLURL
				if nativeURL == "" {
					nativeURL = record.URL
				}
				matches = append(matches, bodyRecord{Body: record.Body, URL: nativeURL})
			}
		}
		if len(records) < 100 {
			return matches, nil
		}
	}
	return nil, errors.New("GitHub reconciliation exceeded the pagination safety limit")
}

func graphQLRequest(query string, variables map[string]any) provider.Request {
	return provider.Request{
		Method: http.MethodPost,
		Path:   "/graphql",
		Body: map[string]any{
			"query":     query,
			"variables": variables,
		},
	}
}

func setOptionalTicketFields(body map[string]any, payload domain.EffectPayload) {
	if payload.Labels != nil {
		body["labels"] = payload.Labels
	}
	if payload.Assignees != nil {
		body["assignees"] = payload.Assignees
	}
	if payload.Milestone != nil {
		body["milestone"] = *payload.Milestone
	}
	if payload.IssueType != "" {
		body["type"] = payload.IssueType
	}
}

func repositoryParts(value string) (string, string, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("GitHub repository must be owner/name, got %q", value)
	}
	return parts[0], parts[1], nil
}

func issuePath(repository string, number int64) (string, error) {
	owner, name, err := repositoryParts(repository)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/repos/%s/%s/issues/%d", url.PathEscape(owner), url.PathEscape(name), number), nil
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
		ID      any    `json:"id"`
		HTMLURL string `json:"html_url"`
		URL     string `json:"url"`
	}
	_ = json.Unmarshal(response.Body, &value)
	nativeReference := value.HTMLURL
	if nativeReference == "" {
		nativeReference = value.URL
	}
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

func graphqlErrors(response provider.Response) error {
	var result struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if json.Unmarshal(response.Body, &result) == nil && len(result.Errors) > 0 {
		return &provider.PermanentError{StatusCode: response.StatusCode, Message: result.Errors[0].Message, Body: response.Body}
	}
	return nil
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
