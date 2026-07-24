package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

type Provider string

const (
	ProviderGitHub Provider = "github"
	ProviderGitLab Provider = "gitlab"
)

type DeliveryState string

const (
	DeliveryPending    DeliveryState = "pending"
	DeliveryNormalized DeliveryState = "normalized"
	DeliveryIgnored    DeliveryState = "ignored"
	DeliveryFailed     DeliveryState = "failed"
)

type Delivery struct {
	Provider      Provider          `json:"provider"`
	Tenant        string            `json:"tenant"`
	DeliveryID    string            `json:"delivery_id"`
	EventUUID     string            `json:"event_uuid,omitempty"`
	Event         string            `json:"event"`
	Action        string            `json:"action,omitempty"`
	PayloadDigest string            `json:"payload_digest"`
	Payload       json.RawMessage   `json:"payload"`
	Headers       map[string]string `json:"headers"`
	State         DeliveryState     `json:"state"`
	ReceivedAt    time.Time         `json:"received_at"`
}

type WorkflowEvent struct {
	Provider      Provider        `json:"provider"`
	Tenant        string          `json:"tenant"`
	DeliveryID    string          `json:"delivery_id"`
	EventUUID     string          `json:"event_uuid,omitempty"`
	Kind          string          `json:"kind"`
	Action        string          `json:"action,omitempty"`
	Repository    string          `json:"repository,omitempty"`
	SubjectKind   string          `json:"subject_kind,omitempty"`
	SubjectID     string          `json:"subject_id,omitempty"`
	SubjectNumber int64           `json:"subject_number,omitempty"`
	Revision      string          `json:"revision,omitempty"`
	Actor         string          `json:"actor,omitempty"`
	Suppressed    bool            `json:"suppressed"`
	Payload       json.RawMessage `json:"payload"`
}

type EffectMode string

const (
	EffectModePreview EffectMode = "preview"
	EffectModeApply   EffectMode = "apply"
)

type EffectState string

const (
	EffectQueued     EffectState = "queued"
	EffectProcessing EffectState = "processing"
	EffectSucceeded  EffectState = "succeeded"
	EffectRetry      EffectState = "retry"
	EffectDeadLetter EffectState = "dead_letter"
)

type EffectKind string

const (
	EffectTicketCreate       EffectKind = "ticket.create"
	EffectTicketUpdate       EffectKind = "ticket.update"
	EffectTicketState        EffectKind = "ticket.state"
	EffectKanbanAdd          EffectKind = "kanban.add"
	EffectKanbanStatus       EffectKind = "kanban.status"
	EffectKanbanArchive      EffectKind = "kanban.archive"
	EffectContextPublish     EffectKind = "context.publish"
	EffectReviewPublish      EffectKind = "review.publish"
	EffectDeploymentCreate   EffectKind = "deployment.create"
	EffectDeploymentStatus   EffectKind = "deployment.status"
	EffectDeploymentApproval EffectKind = "deployment.approval"
)

var EffectKinds = []EffectKind{
	EffectTicketCreate,
	EffectTicketUpdate,
	EffectTicketState,
	EffectKanbanAdd,
	EffectKanbanStatus,
	EffectKanbanArchive,
	EffectContextPublish,
	EffectReviewPublish,
	EffectDeploymentCreate,
	EffectDeploymentStatus,
	EffectDeploymentApproval,
}

type EffectTarget struct {
	Repository   string `json:"repository"`
	SubjectKind  string `json:"subject_kind,omitempty"`
	Number       int64  `json:"number,omitempty"`
	ProjectID    string `json:"project_id,omitempty"`
	ItemID       string `json:"item_id,omitempty"`
	ContentID    string `json:"content_id,omitempty"`
	FieldID      string `json:"field_id,omitempty"`
	OptionID     string `json:"option_id,omitempty"`
	WorkItemID   string `json:"work_item_id,omitempty"`
	DeploymentID int64  `json:"deployment_id,omitempty"`
}

type ContextRecord struct {
	ID              string `json:"id"`
	Version         int    `json:"version"`
	Digest          string `json:"digest"`
	Supersedes      string `json:"supersedes,omitempty"`
	Type            string `json:"type"`
	Summary         string `json:"summary"`
	Rationale       string `json:"rationale"`
	Alternatives    string `json:"alternatives,omitempty"`
	Tradeoffs       string `json:"tradeoffs,omitempty"`
	AppliesTo       string `json:"applies_to"`
	Source          string `json:"source"`
	Status          string `json:"status"`
	Audience        string `json:"audience"`
	Visibility      string `json:"visibility"`
	NativeReference string `json:"native_reference,omitempty"`
}

type EffectPayload struct {
	Title                 string         `json:"title,omitempty"`
	Body                  string         `json:"body,omitempty"`
	State                 string         `json:"state,omitempty"`
	Labels                []string       `json:"labels,omitempty"`
	AddLabels             []string       `json:"add_labels,omitempty"`
	RemoveLabels          []string       `json:"remove_labels,omitempty"`
	Assignees             []string       `json:"assignees,omitempty"`
	Milestone             *int64         `json:"milestone,omitempty"`
	IssueType             string         `json:"issue_type,omitempty"`
	Status                string         `json:"status,omitempty"`
	Context               *ContextRecord `json:"context,omitempty"`
	ReviewEvent           string         `json:"review_event,omitempty"`
	CommitID              string         `json:"commit_id,omitempty"`
	Ref                   string         `json:"ref,omitempty"`
	SHA                   string         `json:"sha,omitempty"`
	Environment           string         `json:"environment,omitempty"`
	Description           string         `json:"description,omitempty"`
	LogURL                string         `json:"log_url,omitempty"`
	EnvironmentURL        string         `json:"environment_url,omitempty"`
	TransientEnvironment  bool           `json:"transient_environment,omitempty"`
	ProductionEnvironment bool           `json:"production_environment,omitempty"`
	Approval              string         `json:"approval,omitempty"`
	Comment               string         `json:"comment,omitempty"`
	RepresentedAs         string         `json:"represented_as,omitempty"`
}

type Preconditions struct {
	Revision string `json:"revision,omitempty"`
	State    string `json:"state,omitempty"`
	Digest   string `json:"digest,omitempty"`
}

type Effect struct {
	ID             string        `json:"id"`
	CorrelationID  string        `json:"correlation_id"`
	IdempotencyKey string        `json:"idempotency_key"`
	Provider       Provider      `json:"provider"`
	Tenant         string        `json:"tenant"`
	Kind           EffectKind    `json:"kind"`
	Mode           EffectMode    `json:"mode"`
	Target         EffectTarget  `json:"target"`
	Payload        EffectPayload `json:"payload"`
	Preconditions  Preconditions `json:"preconditions,omitempty"`
	State          EffectState   `json:"state,omitempty"`
	Attempts       int           `json:"attempts,omitempty"`
	NextAttemptAt  time.Time     `json:"next_attempt_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at,omitempty"`
}

type RequestPreview struct {
	Requests []ProviderRequestPreview `json:"requests"`
}

type ProviderRequestPreview struct {
	Method string          `json:"method"`
	URL    string          `json:"url"`
	Body   json.RawMessage `json:"body,omitempty"`
}

type EffectResult struct {
	NativeReference string          `json:"native_reference,omitempty"`
	StatusCode      int             `json:"status_code"`
	Body            json.RawMessage `json:"body,omitempty"`
	Reconciled      bool            `json:"reconciled"`
}

var contextIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._:-]{0,127}$`)
var operationIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,191}$`)
var digestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

func NewID(prefix string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("create random id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(random), nil
}

func Digest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (effect Effect) Validate() error {
	if effect.ID == "" || effect.CorrelationID == "" || effect.IdempotencyKey == "" {
		return errors.New("id, correlation_id, and idempotency_key are required")
	}
	if !operationIDPattern.MatchString(effect.ID) ||
		!operationIDPattern.MatchString(effect.CorrelationID) ||
		!operationIDPattern.MatchString(effect.IdempotencyKey) {
		return errors.New("operation identifiers must use 1-192 letters, digits, dots, underscores, colons, or hyphens")
	}
	if effect.Provider != ProviderGitHub && effect.Provider != ProviderGitLab {
		return fmt.Errorf("unsupported provider %q", effect.Provider)
	}
	if strings.TrimSpace(effect.Tenant) == "" || strings.TrimSpace(effect.Target.Repository) == "" {
		return errors.New("tenant and target.repository are required")
	}
	if effect.Mode != EffectModePreview && effect.Mode != EffectModeApply {
		return fmt.Errorf("unsupported effect mode %q", effect.Mode)
	}
	if !slices.Contains(EffectKinds, effect.Kind) {
		return fmt.Errorf("unsupported effect kind %q", effect.Kind)
	}

	switch effect.Kind {
	case EffectTicketCreate:
		if strings.TrimSpace(effect.Payload.Title) == "" {
			return errors.New("ticket.create requires payload.title")
		}
	case EffectTicketUpdate, EffectTicketState:
		if effect.Target.Number <= 0 {
			return fmt.Errorf("%s requires target.number", effect.Kind)
		}
	case EffectKanbanAdd:
		if effect.Provider != ProviderGitHub || effect.Target.ProjectID == "" || effect.Target.ContentID == "" {
			return errors.New("kanban.add requires GitHub target.project_id and target.content_id")
		}
	case EffectKanbanStatus:
		if effect.Provider == ProviderGitHub &&
			(effect.Target.ProjectID == "" || effect.Target.ItemID == "" || effect.Target.FieldID == "" || effect.Target.OptionID == "") {
			return errors.New("GitHub kanban.status requires project_id, item_id, field_id, and option_id")
		}
		if effect.Provider == ProviderGitLab && effect.Target.Number <= 0 && effect.Target.WorkItemID == "" {
			return errors.New("GitLab kanban.status requires target.number or target.work_item_id")
		}
		if effect.Provider == ProviderGitLab && effect.Target.WorkItemID != "" && effect.Target.OptionID == "" {
			return errors.New("GitLab native kanban.status requires target.option_id")
		}
		if effect.Provider == ProviderGitLab && effect.Target.WorkItemID == "" &&
			len(effect.Payload.AddLabels) == 0 && len(effect.Payload.RemoveLabels) == 0 {
			return errors.New("GitLab label kanban.status requires add_labels or remove_labels")
		}
	case EffectKanbanArchive:
		if effect.Provider != ProviderGitHub || effect.Target.ProjectID == "" || effect.Target.ItemID == "" {
			return errors.New("kanban.archive requires GitHub target.project_id and target.item_id")
		}
	case EffectContextPublish:
		if effect.Target.Number <= 0 || effect.Payload.Context == nil {
			return errors.New("context.publish requires target.number and payload.context")
		}
		if effect.Provider == ProviderGitLab &&
			effect.Target.SubjectKind != "issue" && effect.Target.SubjectKind != "merge_request" {
			return errors.New("GitLab context.publish requires subject_kind issue or merge_request")
		}
		if err := effect.Payload.Context.Validate(); err != nil {
			return err
		}
		if effect.Payload.Context.Visibility == "internal" {
			return errors.New("internal context cannot be published to a provider")
		}
	case EffectReviewPublish:
		if effect.Target.Number <= 0 || strings.TrimSpace(effect.Payload.Body) == "" {
			return errors.New("review.publish requires target.number and payload.body")
		}
	case EffectDeploymentCreate:
		if effect.Payload.Ref == "" || effect.Payload.SHA == "" || effect.Payload.Environment == "" {
			return errors.New("deployment.create requires ref, sha, and environment")
		}
	case EffectDeploymentStatus:
		if effect.Target.DeploymentID <= 0 || effect.Payload.Status == "" {
			return errors.New("deployment.status requires target.deployment_id and payload.status")
		}
	case EffectDeploymentApproval:
		if effect.Provider != ProviderGitLab || effect.Target.DeploymentID <= 0 || effect.Payload.Approval == "" {
			return errors.New("deployment.approval requires GitLab target.deployment_id and payload.approval")
		}
	}
	return nil
}

func (record ContextRecord) Validate() error {
	if !contextIDPattern.MatchString(record.ID) {
		return errors.New("context id must match [a-z0-9][a-z0-9._:-]{0,127}")
	}
	if record.Version < 1 {
		return errors.New("context version must be positive")
	}
	if record.Version > 1 && record.Supersedes == "" {
		return errors.New("context versions after the first must identify the superseded record")
	}
	if !digestPattern.MatchString(record.Digest) {
		return errors.New("context digest must be a sha256 digest")
	}
	if record.Type == "" || record.Summary == "" || record.Rationale == "" || record.AppliesTo == "" ||
		record.Source == "" || record.Status == "" || record.Audience == "" || record.Visibility == "" {
		return errors.New("context semantic and routing fields are required")
	}
	for name, value := range map[string]string{
		"type": record.Type, "summary": record.Summary, "rationale": record.Rationale,
		"alternatives": record.Alternatives, "tradeoffs": record.Tradeoffs,
		"applies_to": record.AppliesTo, "source": record.Source, "status": record.Status,
		"audience": record.Audience, "visibility": record.Visibility, "supersedes": record.Supersedes,
	} {
		if strings.ContainsAny(value, "\r\n") {
			return fmt.Errorf("context field %s must be a single canonical line", name)
		}
	}
	if Digest(record.SemanticPayload()) != record.Digest {
		return errors.New("context digest does not match the canonical semantic payload")
	}
	return nil
}

func (record ContextRecord) SemanticPayload() []byte {
	lines := []string{
		"Type: " + record.Type,
		"Summary: " + record.Summary,
		"Rationale: " + record.Rationale,
		"Alternatives: " + record.Alternatives,
		"Tradeoffs: " + record.Tradeoffs,
		"Applies to: " + record.AppliesTo,
		"Source: " + record.Source,
		"Status: " + record.Status,
		"Audience: " + record.Audience,
		"Visibility: " + record.Visibility,
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func (record ContextRecord) Markdown() string {
	lines := []string{
		fmt.Sprintf("<!-- xonovex-context id=%s version=%d digest=%s -->", record.ID, record.Version, record.Digest),
		"",
		"Context ID: " + record.ID,
		fmt.Sprintf("Context version: %d", record.Version),
		"Context digest: " + record.Digest,
	}
	if record.Supersedes != "" {
		lines = append(lines, "Supersedes: "+record.Supersedes)
	}
	lines = append(lines,
		"Type: "+record.Type,
		"Summary: "+record.Summary,
		"Rationale: "+record.Rationale,
		"Alternatives: "+record.Alternatives,
		"Tradeoffs: "+record.Tradeoffs,
		"Applies to: "+record.AppliesTo,
		"Source: "+record.Source,
		"Status: "+record.Status,
		"Audience: "+record.Audience,
		"Visibility: "+record.Visibility,
	)
	return strings.Join(lines, "\n") + "\n"
}

func (effect Effect) ResourceKey() string {
	base := strings.Join([]string{string(effect.Provider), effect.Tenant, effect.Target.Repository}, ":")
	switch effect.Kind {
	case EffectTicketCreate:
		return base + ":ticket-create:" + effect.IdempotencyKey
	case EffectTicketUpdate, EffectTicketState:
		return fmt.Sprintf("%s:ticket:%d", base, effect.Target.Number)
	case EffectKanbanAdd, EffectKanbanStatus, EffectKanbanArchive:
		item := effect.Target.ItemID
		if item == "" {
			item = effect.Target.ContentID
		}
		return base + ":kanban:" + effect.Target.ProjectID + ":" + item
	case EffectContextPublish:
		return fmt.Sprintf("%s:context:%d:%s", base, effect.Target.Number, effect.Payload.Context.ID)
	case EffectReviewPublish:
		return fmt.Sprintf("%s:review:%d", base, effect.Target.Number)
	case EffectDeploymentCreate:
		return base + ":deployment-create:" + effect.IdempotencyKey
	case EffectDeploymentStatus, EffectDeploymentApproval:
		return fmt.Sprintf("%s:deployment:%d", base, effect.Target.DeploymentID)
	default:
		return base + ":" + string(effect.Kind) + ":" + effect.IdempotencyKey
	}
}
