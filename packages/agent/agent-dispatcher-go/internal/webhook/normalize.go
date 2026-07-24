package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
)

func NormalizeGitHub(
	tenant string,
	deliveryID string,
	eventName string,
	body []byte,
	isBot func(string) bool,
) (domain.WorkflowEvent, error) {
	var payload githubPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return domain.WorkflowEvent{}, fmt.Errorf("decode GitHub webhook: %w", err)
	}
	if payload.Repository.FullName == "" {
		return domain.WorkflowEvent{}, errors.New("GitHub webhook repository.full_name is required")
	}
	subjectKind, subjectID, subjectNumber, revision := githubSubject(eventName, payload)
	actor := payload.Sender.Login
	action := payload.Action
	return domain.WorkflowEvent{
		Provider:      domain.ProviderGitHub,
		Tenant:        tenant,
		DeliveryID:    deliveryID,
		Kind:          eventName,
		Action:        action,
		Repository:    payload.Repository.FullName,
		SubjectKind:   subjectKind,
		SubjectID:     subjectID,
		SubjectNumber: subjectNumber,
		Revision:      revision,
		Actor:         actor,
		Suppressed:    isBot(actor),
		Payload:       append(json.RawMessage(nil), body...),
	}, nil
}

func NormalizeGitLab(
	tenant string,
	deliveryID string,
	eventUUID string,
	eventName string,
	body []byte,
	isBot func(string) bool,
) (domain.WorkflowEvent, error) {
	var payload gitlabPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return domain.WorkflowEvent{}, fmt.Errorf("decode GitLab webhook: %w", err)
	}
	repository := payload.Project.PathWithNamespace
	if repository == "" {
		repository = payload.Project.WebURL
	}
	if repository == "" {
		return domain.WorkflowEvent{}, errors.New("GitLab webhook project identity is required")
	}
	actor := payload.User.Username
	if actor == "" {
		actor = payload.UserUsername
	}
	action := payload.ObjectAttributes.Action
	if action == "" {
		action = payload.EventType
	}
	subjectKind := payload.ObjectKind
	if subjectKind == "" {
		subjectKind = normalizeGitLabEventName(eventName)
	}
	subjectNumber := payload.ObjectAttributes.IID
	subjectID := strconv.FormatInt(payload.ObjectAttributes.ID, 10)
	if payload.ObjectAttributes.ID == 0 {
		subjectID = ""
	}
	revision := payload.ObjectAttributes.LastCommit.ID
	if revision == "" {
		revision = payload.ObjectAttributes.SHA
	}
	return domain.WorkflowEvent{
		Provider:      domain.ProviderGitLab,
		Tenant:        tenant,
		DeliveryID:    deliveryID,
		EventUUID:     eventUUID,
		Kind:          normalizeGitLabEventName(eventName),
		Action:        action,
		Repository:    repository,
		SubjectKind:   subjectKind,
		SubjectID:     subjectID,
		SubjectNumber: subjectNumber,
		Revision:      revision,
		Actor:         actor,
		Suppressed:    isBot(actor),
		Payload:       append(json.RawMessage(nil), body...),
	}, nil
}

type githubPayload struct {
	Action     string `json:"action"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
	Issue struct {
		ID          int64 `json:"id"`
		Number      int64 `json:"number"`
		PullRequest any   `json:"pull_request"`
	} `json:"issue"`
	PullRequest struct {
		ID     int64 `json:"id"`
		Number int64 `json:"number"`
		Head   struct {
			SHA string `json:"sha"`
		} `json:"head"`
	} `json:"pull_request"`
	ProjectV2Item struct {
		ID          int64  `json:"id"`
		NodeID      string `json:"node_id"`
		ContentNode string `json:"content_node_id"`
	} `json:"projects_v2_item"`
	Deployment struct {
		ID  int64  `json:"id"`
		SHA string `json:"sha"`
	} `json:"deployment"`
	DeploymentStatus struct {
		ID    int64  `json:"id"`
		State string `json:"state"`
	} `json:"deployment_status"`
}

func githubSubject(eventName string, payload githubPayload) (string, string, int64, string) {
	switch eventName {
	case "issues", "issue_comment":
		kind := "issue"
		if payload.Issue.PullRequest != nil {
			kind = "pull_request"
		}
		return kind, strconv.FormatInt(payload.Issue.ID, 10), payload.Issue.Number, ""
	case "pull_request", "pull_request_review", "pull_request_review_comment":
		return "pull_request", strconv.FormatInt(payload.PullRequest.ID, 10), payload.PullRequest.Number, payload.PullRequest.Head.SHA
	case "projects_v2_item":
		id := payload.ProjectV2Item.NodeID
		if id == "" {
			id = strconv.FormatInt(payload.ProjectV2Item.ID, 10)
		}
		return "project_item", id, 0, payload.ProjectV2Item.ContentNode
	case "deployment", "deployment_status":
		return "deployment", strconv.FormatInt(payload.Deployment.ID, 10), 0, payload.Deployment.SHA
	default:
		return eventName, "", 0, ""
	}
}

type gitlabPayload struct {
	ObjectKind string `json:"object_kind"`
	EventType  string `json:"event_type"`
	Project    struct {
		PathWithNamespace string `json:"path_with_namespace"`
		WebURL            string `json:"web_url"`
	} `json:"project"`
	User struct {
		Username string `json:"username"`
	} `json:"user"`
	UserUsername     string `json:"user_username"`
	ObjectAttributes struct {
		ID         int64  `json:"id"`
		IID        int64  `json:"iid"`
		Action     string `json:"action"`
		State      string `json:"state"`
		SHA        string `json:"sha"`
		LastCommit struct {
			ID string `json:"id"`
		} `json:"last_commit"`
	} `json:"object_attributes"`
}

func normalizeGitLabEventName(value string) string {
	normalized := strings.TrimSpace(strings.TrimSuffix(strings.ToLower(value), " hook"))
	return strings.ReplaceAll(normalized, " ", "_")
}
