// Package protocol is the port between the run engine and an eval protocol. A
// protocol reads its documents and turns them into a plan: independent units of
// model work and a verdict over their results. The engine runs the units with
// resume, budget and checkpoints and never learns what a protocol measures.
package protocol

import (
	"context"
	"fmt"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

// Env holds the ports a protocol plans and runs with.
type Env struct {
	FS         fs.FileSystem
	Reader     document.Reader
	Catalog    catalog.Catalog
	Exec       harness.Executor
	Predicates *predicate.Registry
}

// Input is what one invocation asks of a protocol.
type Input struct {
	// Target names what to evaluate: a skill, a plugin, an instruction file or an
	// orchestration file, as the protocol defines.
	Target   string
	Settings settings.Settings
	Env      Env
}

// Unit is one independent piece of model work. Its id and condition name the
// case result it produces.
type Unit struct {
	ID        string
	Condition model.Condition
	// ModelRuns is how many model runs the unit may start, for the budget.
	ModelRuns int
	Run       func(ctx context.Context) (model.CaseResult, error)
}

// Key is the case key of the result the unit produces.
func (u Unit) Key() string { return model.CaseKey(u.ID, u.Condition) }

// Verdict is the outcome of a finished run.
type Verdict struct {
	Pass bool
	// Lines explain the outcome: gate figures and the cases that failed.
	Lines []string
}

// Plan is the work of one invocation.
type Plan struct {
	// Subject names what was evaluated, for the run id.
	Subject string
	// Inputs are the texts the results depend on. The engine digests them, so a
	// changed document or setting starts a new run instead of resuming an old one.
	Inputs []string
	Units  []Unit
	// Decide turns the results of every unit into the verdict.
	Decide func(cases []model.CaseResult) Verdict
}

// Protocol is one eval protocol.
type Protocol interface {
	Name() model.Protocol
	Plan(input Input) (Plan, error)
}

// Registry holds the protocols by name.
type Registry struct {
	byName map[model.Protocol]Protocol
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{byName: map[model.Protocol]Protocol{}} }

// Register adds a protocol. A second protocol with one name panics.
func (r *Registry) Register(p Protocol) {
	if _, dup := r.byName[p.Name()]; dup {
		panic(fmt.Sprintf("protocol %q registered twice", p.Name()))
	}
	r.byName[p.Name()] = p
}

// Get returns the protocol with the name.
func (r *Registry) Get(name model.Protocol) (Protocol, error) {
	p, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("unknown protocol %q", name)
	}
	return p, nil
}

// Names lists the registered protocols, sorted.
func (r *Registry) Names() []model.Protocol {
	names := make([]model.Protocol, 0, len(r.byName))
	for name := range r.byName {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names
}

// Offers groups skills by plugin into the harness offers of a run.
func Offers(skills ...catalog.Skill) []harness.Offer {
	var offers []harness.Offer
	index := map[string]int{}
	for _, skill := range skills {
		i, ok := index[skill.Plugin]
		if !ok {
			i = len(offers)
			index[skill.Plugin] = i
			offers = append(offers, harness.Offer{Name: skill.Plugin})
		}
		offers[i].SkillDirs = append(offers[i].SkillDirs, skill.Dir)
	}
	return offers
}

// Sum adds the usage of every observation.
func Sum(usages ...model.Usage) model.Usage {
	var total model.Usage
	for _, u := range usages {
		total = total.Add(u)
	}
	return total
}

// UsagePtr returns a pointer to a usage, or nil when it is empty.
func UsagePtr(u model.Usage) *model.Usage {
	if u == (model.Usage{}) {
		return nil
	}
	return &u
}

// RunSeconds returns the timeout in seconds: the profile timeout when set,
// else the protocol default.
func RunSeconds(profile, fallback int) int {
	if profile > 0 {
		return profile
	}
	return fallback
}

// Failures lists the keys of the cases that did not pass, as verdict lines.
func Failures(cases []model.CaseResult) []string {
	var lines []string
	for _, c := range cases {
		if c.Outcome == model.Passed {
			continue
		}
		line := "failed: " + c.Key()
		if c.Error != "" {
			line += " (" + c.Error + ")"
		}
		lines = append(lines, line)
	}
	return lines
}
