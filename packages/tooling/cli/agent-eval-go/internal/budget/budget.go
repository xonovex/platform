// Package budget bounds the model work of one invocation. A run asks the guard
// before it starts a model run and tells it what the run cost, so a loop over
// many cases stops at the limit instead of after it.
package budget

import (
	"errors"
	"fmt"
	"sync"

	evalsection "github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/eval"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// ErrExceeded reports that the next model runs would pass a limit.
var ErrExceeded = errors.New("budget exceeded")

// Limits are the ceilings of one invocation. A zero value means the default.
type Limits struct {
	MaxModelRuns int
	MaxCostUSD   float64
}

// Resolve fills the zero members with the ceilings of the eval section.
func (l Limits) Resolve() Limits {
	if l.MaxModelRuns <= 0 {
		l.MaxModelRuns = evalsection.DefaultMaxModelRuns
	}
	if l.MaxCostUSD <= 0 {
		l.MaxCostUSD = evalsection.DefaultMaxCostUSD
	}
	return l
}

// Guard counts model runs and cost against the limits. It is safe for
// concurrent use.
type Guard struct {
	mu     sync.Mutex
	limits Limits
	runs   int
	spent  model.Usage
}

// New returns a guard for the limits, with zero members defaulted.
func New(limits Limits) *Guard { return &Guard{limits: limits.Resolve()} }

// Admit reserves model runs. It fails without reserving when they would pass the
// run limit or the cost limit is already spent.
func (g *Guard) Admit(runs int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.runs+runs > g.limits.MaxModelRuns {
		return fmt.Errorf("%w: %d model runs reserved of %d, %d more asked", ErrExceeded, g.runs, g.limits.MaxModelRuns, runs)
	}
	if g.spent.CostUSD >= g.limits.MaxCostUSD {
		return fmt.Errorf("%w: $%.4f spent of $%.2f", ErrExceeded, g.spent.CostUSD, g.limits.MaxCostUSD)
	}
	g.runs += runs
	return nil
}

// Record adds what finished runs cost.
func (g *Guard) Record(usage model.Usage) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.spent = g.spent.Add(usage)
}

// Spent returns the model runs reserved and the usage recorded.
func (g *Guard) Spent() (int, model.Usage) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.runs, g.spent
}
