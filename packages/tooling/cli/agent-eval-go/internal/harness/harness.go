// Package harness is the port between an eval protocol and an agent harness.
// A protocol describes what it wants to learn as a Request; a harness leaf
// turns the request into a program invocation and its staged files, and reads
// the stream the program writes into an Observation. The protocols never see a
// command line, and the leaves never see a protocol, so a new harness is one
// leaf and a new protocol is one package.
package harness

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process"
)

// Role is what a run is for. It decides which tools the harness may use.
type Role string

// The roles.
const (
	// RoleTrigger asks which of the offered skills the model would load, and
	// stops at the first answer.
	RoleTrigger Role = "trigger"
	// RoleGenerate answers a prompt with only the Skill and Read tools, with
	// or without the skill on offer.
	RoleGenerate Role = "generate"
	// RoleJudge grades a response and returns strict JSON with no tools.
	RoleJudge Role = "judge"
	// RoleAgent works in the workspace with file tools; a tool that needs
	// approval is denied and observed.
	RoleAgent Role = "agent"
)

// Offer is a plugin whose skills a run may load. Only the listed skill
// directories are staged, so the run sees no other skill.
type Offer struct {
	// Name is the plugin name that namespaces its skills.
	Name string
	// Dir is the plugin directory.
	Dir string
	// SkillDirs are the skill directories of the plugin to stage.
	SkillDirs []string
}

// File is a file placed in the workspace before a run.
type File struct {
	Path    string
	Content string
}

// Caps are the hard limits of one run. A zero value means the harness default.
type Caps struct {
	// Timeout ends the program.
	Timeout time.Duration
	// BudgetUSD caps the spend of the run where the harness can enforce it.
	BudgetUSD float64
	// MaxTurns caps the model turns of the run.
	MaxTurns int
	// MaxOutputChars ends the run once the streamed answer is this long.
	MaxOutputChars int
}

// Request is one model run.
type Request struct {
	Role    Role
	Harness string
	Model   string
	Effort  model.Effort
	// Prompt is the text the model receives.
	Prompt string
	// Target is the short name of the skill a trigger or generate run expects.
	Target string
	// Offer lists the plugins whose skills the run may load.
	Offer []Offer
	// WithSkill selects the arm of a generate run that has the skill on offer.
	WithSkill bool
	// Instructions is the text of the instruction file the run starts with. An
	// empty value leaves the workspace without one.
	Instructions string
	// Fixtures are written into the workspace before the run.
	Fixtures []File
	// Assertions is the number of assertions a judge run must grade.
	Assertions int
	Caps       Caps
}

// StageKind is the kind of a staging step.
type StageKind string

// The staging kinds.
const (
	// StageCopy copies a directory tree.
	StageCopy StageKind = "copy"
	// StageWrite writes a file.
	StageWrite StageKind = "write"
	// StageAppend appends text to a file that an earlier step created.
	StageAppend StageKind = "append"
	// StageDir creates an empty directory.
	StageDir StageKind = "dir"
)

// Stage is one step that prepares the workspace of a run.
type Stage struct {
	Kind    StageKind
	Src     string
	Dst     string
	Content string
}

// Plan is the program invocation a harness leaf derives from a request.
type Plan struct {
	// Args are the arguments before the prompt; the prompt is the last argument.
	Args []string
	// Prompt is the final argument, after any harness-specific framing.
	Prompt string
	// Env replaces or extends the inherited environment.
	Env map[string]string
	// Dir is the working directory of the program.
	Dir string
	// Stage lists the files and directories to place before the program starts.
	Stage []Stage
}

// Event is one thing the agent did, as the session predicates read it.
type Event struct {
	// Kind is tool-call, file-read, file-write, command or message.
	Kind string
	// Name is the tool name for a tool-call.
	Name string
	// Text is the path, command or message text.
	Text string
}

// Event kinds.
const (
	EventToolCall  = "tool-call"
	EventFileRead  = "file-read"
	EventFileWrite = "file-write"
	EventCommand   = "command"
	EventMessage   = "message"
)

// Selected values of a trigger observation.
const (
	SelectedTarget      = "target"
	SelectedNone        = "none"
	SelectedOutputLimit = "output-limit"
	SelectedCompetitor  = "competitor:"
)

// Observation is what a run showed.
type Observation struct {
	// Text is the final answer.
	Text string
	// Structured is the raw JSON verdict of a judge run.
	Structured string
	Usage      model.Usage
	DurationMs int64
	// Events lists what the agent did, in order.
	Events []Event
	// Files is the workspace after the run, by path relative to it. The
	// executor fills it before it removes the workspace, so the filesystem
	// predicates read what the run left behind.
	Files map[string]string
	// TargetAvailable reports that the harness listed the target skill.
	TargetAvailable bool
	// TargetInvoked reports that the model loaded the target skill.
	TargetInvoked bool
	// Selected says how a trigger run ended: the target, a competitor, the
	// output limit or nothing.
	Selected string
	// OutputLimit reports that the answer reached the output cap.
	OutputLimit bool
	// TimedOut reports that the timeout ended the run.
	TimedOut bool
	// Failure is the infrastructure fault of the run, empty when it ran to a
	// verdict. A transient fault is retried; a repeated one voids the run.
	Failure string
}

// Scanner reads the standard output of one run line by line.
type Scanner interface {
	// Feed reads one line and reports whether the run is settled, so the
	// program can end there.
	Feed(line string) (settled bool)
	// Finish combines what the stream showed with how the program ended.
	Finish(exit process.Result, elapsedMs int64) Observation
}

// Harness is one agent harness.
type Harness interface {
	// Name is the harness name a profile and a document use.
	Name() string
	// Binary is the program to start.
	Binary() string
	// Plan derives the invocation of a request in a workspace directory.
	Plan(request Request, workspace string) (Plan, error)
	// Scan returns a scanner for the standard output of a request.
	Scan(request Request) Scanner
}

// Executor runs a request and returns its observation. A run that failed for an
// infrastructure reason on every attempt is an error, never a failed verdict.
type Executor interface {
	Execute(ctx context.Context, request Request) (Observation, error)
}

// Registry holds the harnesses by name.
type Registry struct {
	byName map[string]Harness
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{byName: map[string]Harness{}} }

// Register adds a harness. A second harness with the same name panics, because
// two leaves claiming one name is a programming error.
func (r *Registry) Register(h Harness) {
	if _, dup := r.byName[h.Name()]; dup {
		panic(fmt.Sprintf("harness %q registered twice", h.Name()))
	}
	r.byName[h.Name()] = h
}

// Get returns the harness with the name.
func (r *Registry) Get(name string) (Harness, error) {
	h, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("no eval support for harness %q (supported: %v)", name, r.Names())
	}
	return h, nil
}

// Names lists the registered harness names, sorted.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.byName))
	for name := range r.byName {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
