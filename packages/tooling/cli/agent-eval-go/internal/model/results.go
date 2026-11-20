package model

// Status says how far a run got. Complete holds a verdict, incomplete awaits a
// resume and invalid means an infrastructure fault voids the run.
type Status string

// The statuses.
const (
	Complete   Status = "complete"
	Incomplete Status = "incomplete"
	Invalid    Status = "invalid"
)

// Stop says why a run ended early.
type Stop string

// The stop reasons.
const (
	StopBudget         Stop = "budget"
	StopInfrastructure Stop = "infrastructure"
)

// Outcome is the verdict on one case.
type Outcome string

// The outcomes.
const (
	Passed  Outcome = "passed"
	Failed  Outcome = "failed"
	Errored Outcome = "errored"
)

// PredicateResult is the verdict on one predicate of a case.
type PredicateResult struct {
	Kind   string `json:"kind"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

// CaseResult is the record of one finished case.
type CaseResult struct {
	ID             string            `json:"id"`
	Condition      Condition         `json:"condition,omitempty"`
	Outcome        Outcome           `json:"outcome"`
	Score          *float64          `json:"score,omitempty"`
	ControlScore   *float64          `json:"controlScore,omitempty"`
	Runs           int               `json:"runs,omitempty"`
	Selected       []string          `json:"selected,omitempty"`
	SkillTriggered *bool             `json:"skillTriggered,omitempty"`
	DurationMs     int64             `json:"durationMs,omitempty"`
	Usage          *Usage            `json:"usage,omitempty"`
	Predicates     []PredicateResult `json:"predicates,omitempty"`
	Error          string            `json:"error,omitempty"`
}

// Key identifies a case result inside one run: the case id and its condition.
func (c CaseResult) Key() string { return CaseKey(c.ID, c.Condition) }

// CaseKey joins a case id and a condition.
func CaseKey(id string, condition Condition) string {
	if condition == "" {
		return id
	}
	return id + "@" + string(condition)
}

// Results is the record of one eval run. An incomplete run lists the cases it
// finished, so a resumer restarts at the first case it lacks.
type Results struct {
	Schema       string       `json:"$schema,omitempty"`
	Version      int          `json:"version"`
	RunID        string       `json:"runId"`
	Protocol     Protocol     `json:"protocol"`
	Profile      string       `json:"profile"`
	InputsDigest string       `json:"inputsDigest"`
	StartedAt    string       `json:"startedAt"`
	FinishedAt   string       `json:"finishedAt,omitempty"`
	Status       Status       `json:"status"`
	Stop         Stop         `json:"stop,omitempty"`
	Cases        []CaseResult `json:"cases"`
	Usage        *Usage       `json:"usage,omitempty"`
}

// Done returns the case results that are finished, keyed by CaseResult.Key.
func (r Results) Done() map[string]CaseResult {
	done := make(map[string]CaseResult, len(r.Cases))
	for _, c := range r.Cases {
		done[c.Key()] = c
	}
	return done
}

// BaselineCase is the pass rate recorded for one case.
type BaselineCase struct {
	ID        string    `json:"id"`
	Condition Condition `json:"condition,omitempty"`
	PassRate  float64   `json:"passRate"`
	Runs      int       `json:"runs"`
}

// Baseline holds the pass rates that a later run is compared with.
type Baseline struct {
	Schema       string         `json:"$schema,omitempty"`
	Version      int            `json:"version"`
	Protocol     Protocol       `json:"protocol"`
	Profile      string         `json:"profile"`
	InputsDigest string         `json:"inputsDigest"`
	RecordedAt   string         `json:"recordedAt"`
	Tolerance    float64        `json:"tolerance"`
	Cases        []BaselineCase `json:"cases"`
}

// CostEntry is the model cost of one run.
type CostEntry struct {
	RunID      string   `json:"runId"`
	RecordedAt string   `json:"recordedAt"`
	Protocol   Protocol `json:"protocol"`
	Profile    string   `json:"profile"`
	Model      string   `json:"model"`
	ModelRuns  int      `json:"modelRuns"`
	Usage      Usage    `json:"usage"`
}

// CostHistory lists the cost of each run, appended once per run.
type CostHistory struct {
	Schema  string      `json:"$schema,omitempty"`
	Version int         `json:"version"`
	Entries []CostEntry `json:"entries"`
}
