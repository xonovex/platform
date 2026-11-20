// Package audit runs the audit of one source catalog: it loads the catalog,
// asks the transports what the sources look like now, and hands the sources and
// the observations to the rules. It owns the order of the steps and nothing
// else: reading is the transports' work and judging is the rules' work.
package audit

import (
	"context"
	"errors"
	"path"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/rules"
)

// Options select what an audit looks at.
type Options struct {
	// Now is the instant that review windows are measured against.
	Now     time.Time
	Windows rules.Windows
	// Fetch lets the transports that use the network run.
	Fetch bool
	// Pull refreshes the tags of a checkout before it is compared.
	Pull bool
	// CoverDir is the directory below the catalog whose Markdown files each need
	// a source. An empty value turns the check off.
	CoverDir string
}

// Deps are the ports an audit reads through.
type Deps struct {
	FS         fs.FileSystem
	Loader     sources.Loader
	Transports *observe.Registry
}

// SourceReport is what the audit learned about one source.
type SourceReport struct {
	ID           string                `json:"id"`
	Title        string                `json:"title"`
	ReviewedAt   string                `json:"reviewedAt"`
	Observations []observe.Observation `json:"observations,omitempty"`
}

// CatalogReport is the result of auditing one catalog.
type CatalogReport struct {
	Dir      string             `json:"dir"`
	Sources  []SourceReport     `json:"sources"`
	Findings []findings.Finding `json:"findings"`
	// Catalog is the loaded catalog, kept for the steps that follow an audit.
	Catalog sources.Catalog `json:"-"`
}

// Catalog audits the catalog in dir. A problem in the content is a finding. The
// error is for a fault of the file system.
func Catalog(ctx context.Context, deps Deps, dir string, opts Options) (CatalogReport, error) {
	catalog, loaded, err := deps.Loader.Load(dir)
	if err != nil {
		return CatalogReport{}, err
	}
	report := CatalogReport{Dir: dir, Catalog: catalog, Findings: append([]findings.Finding(nil), loaded...)}
	if hasCode(loaded, sources.CodeMissing) {
		return report, nil
	}

	exists := func(feed string) bool {
		ok, err := fs.Exists(deps.FS, path.Join(dir, feed))
		return err == nil && ok
	}
	for _, source := range catalog.Sources {
		observations, faults := observeSource(ctx, deps.Transports, source, dir, opts)
		report.Sources = append(report.Sources, SourceReport{
			ID: source.ID, Title: source.Title, ReviewedAt: source.ReviewedAt, Observations: observations,
		})
		report.Findings = append(report.Findings, faults...)
		report.Findings = append(report.Findings, rules.Review(source, opts.Windows, opts.Now)...)
		report.Findings = append(report.Findings, rules.Mapping(source, exists)...)
		report.Findings = append(report.Findings, rules.Pin(source)...)
		report.Findings = append(report.Findings, rules.Observed(source, observations)...)
	}

	references, err := referenceFiles(deps.FS, dir, opts.CoverDir)
	if err != nil {
		return CatalogReport{}, err
	}
	report.Findings = append(report.Findings, rules.Coverage(catalog, references)...)

	if findings.Worst(report.Findings) == findings.Pass {
		report.Findings = append(report.Findings, findings.Finding{
			Check: rules.CheckName, Level: findings.Pass, Code: rules.CodeOK, Path: dir,
			Message: "every source is current",
		})
	}
	return report, nil
}

// observeSource asks each transport that can read the source. A transport that
// uses the network is skipped unless the options allow it. A transport that
// cannot work at all becomes a finding and the audit goes on.
func observeSource(ctx context.Context, registry *observe.Registry, source sources.Source, dir string, opts Options) ([]observe.Observation, []findings.Finding) {
	var observations []observe.Observation
	var faults []findings.Finding
	for _, transport := range registry.For(source) {
		if transport.Remote() && !opts.Fetch {
			continue
		}
		observation, err := transport.Observe(ctx, observe.Request{Source: source, Dir: dir, Pull: opts.Pull})
		if err != nil {
			faults = append(faults, rules.ProviderFault(source, transport.Name(), err))
			continue
		}
		observations = append(observations, observation)
	}
	return observations, faults
}

// referenceFiles lists the Markdown files directly below dir/coverDir as slash
// paths relative to dir. A missing directory has none.
func referenceFiles(fsys fs.FileSystem, dir, coverDir string) ([]string, error) {
	if coverDir == "" {
		return nil, nil
	}
	entries, err := fsys.ReadDir(path.Join(dir, coverDir))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir && strings.HasSuffix(entry.Name, ".md") {
			files = append(files, path.Join(coverDir, entry.Name))
		}
	}
	return files, nil
}

func hasCode(all []findings.Finding, code string) bool {
	for _, f := range all {
		if f.Code == code {
			return true
		}
	}
	return false
}
