// Package cost keeps the history of what each eval run cost, so a team sees the
// spend of its evals and notices a change that makes them dearer.
package cost

import (
	"sort"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// Append returns the history with the entry added. An entry of a run that is
// already in the history replaces it, so a resumed run counts once.
func Append(history model.CostHistory, entry model.CostEntry) model.CostHistory {
	history.Version = model.Version
	entries := make([]model.CostEntry, 0, len(history.Entries)+1)
	for _, e := range history.Entries {
		if e.RunID != entry.RunID {
			entries = append(entries, e)
		}
	}
	history.Entries = append(entries, entry)
	return history
}

// Total is the cost summed over a set of runs.
type Total struct {
	Runs      int
	ModelRuns int
	Usage     model.Usage
}

// Summary is the cost of the history, overall and by protocol.
type Summary struct {
	All        Total
	ByProtocol map[model.Protocol]Total
}

// Protocols lists the protocols that appear in the summary, sorted.
func (s Summary) Protocols() []model.Protocol {
	out := make([]model.Protocol, 0, len(s.ByProtocol))
	for p := range s.ByProtocol {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Summarize sums the history. A protocol filter of "" keeps every protocol.
func Summarize(history model.CostHistory, protocol model.Protocol) Summary {
	summary := Summary{ByProtocol: map[model.Protocol]Total{}}
	for _, e := range history.Entries {
		if protocol != "" && e.Protocol != protocol {
			continue
		}
		summary.All = add(summary.All, e)
		summary.ByProtocol[e.Protocol] = add(summary.ByProtocol[e.Protocol], e)
	}
	return summary
}

func add(t Total, e model.CostEntry) Total {
	return Total{Runs: t.Runs + 1, ModelRuns: t.ModelRuns + e.ModelRuns, Usage: t.Usage.Add(e.Usage)}
}

// Store reads and writes the cost history file.
type Store struct {
	FS     fs.FileSystem
	Reader document.Reader
}

// Load returns the history in the file, or an empty history when there is none.
func (s Store) Load(file string) (model.CostHistory, error) {
	if ok, err := fs.Exists(s.FS, file); err != nil || !ok {
		return model.CostHistory{Version: model.Version}, err
	}
	return s.Reader.CostHistory(file)
}

// Save writes the history to the file.
func (s Store) Save(file string, history model.CostHistory) error {
	data, err := model.Encode(history)
	if err != nil {
		return err
	}
	return s.FS.WriteFile(file, data)
}
