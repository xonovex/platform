package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/report"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/audit"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/catalogs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/compare"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/compare/gitbase"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/render"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/rules"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/stamp"
)

const (
	defaultMaxAge        = 180
	defaultVersionMaxAge = 90
	defaultCoverDir      = "references"
)

// sourcesFlags are the options of the sources command.
type sourcesFlags struct {
	maxAge, versionMaxAge int
	fetch, pull           bool
	markReviewed          bool
	match                 string
	base                  string
	coverDir              string
	format                string
	project               string
	json                  bool
	strict                bool
}

func newSourcesCommand(deps Dependencies) *cobra.Command {
	var flags sourcesFlags
	command := &cobra.Command{
		Use:   "sources [path...]",
		Short: "Audit the source catalogs below each path",
		Long: `Audit every source catalog (sources.json and sources.d) below each path,
the current directory by default.

The audit reports sources whose review is overdue, claims whose anchor is
gone from the source, pins that no longer match the upstream, and reference
files that no source backs. The network and git checkouts are read only when
--fetch asks for the network or a source names a checkout.

The maxAge, versionMaxAge and coverDir keys of the audit section of
agentic.json set the defaults of --max-age, --version-max-age and --cover-dir.

With --base REF the audit also compares each catalog with the catalog at REF
and reports pins or claims that changed without a new review.

Exit status: 0 when nothing reaches the failing level, 1 when a finding does,
2 when the run could not complete.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			merged, err := loadSettings(deps, flags.project, flags.configFlags(cmd))
			if err != nil {
				return err
			}
			flags.maxAge, flags.versionMaxAge, flags.coverDir = merged.MaxAge, merged.VersionMaxAge, merged.CoverDir
			return runSources(cmd.Context(), deps, flags, args)
		},
	}
	f := command.Flags()
	bindProject(command, &flags.project)
	f.IntVar(&flags.maxAge, "max-age", defaultMaxAge, "days after which a review is overdue")
	f.IntVar(&flags.versionMaxAge, "version-max-age", defaultVersionMaxAge, "days after which the review of a versioned source is overdue")
	f.BoolVar(&flags.fetch, "fetch", false, "read the documents of web and http sources")
	f.BoolVar(&flags.pull, "pull", false, "refresh the tags of each checkout before it is compared")
	f.BoolVar(&flags.markReviewed, "mark-reviewed", false, "set the review date of the sources to today before the audit")
	f.StringVar(&flags.match, "match", "", "limit --mark-reviewed to the sources whose title contains this text")
	f.StringVar(&flags.base, "base", "", "compare each catalog with the catalog at this git ref")
	f.StringVar(&flags.coverDir, "cover-dir", defaultCoverDir, "directory whose Markdown files each need a source, empty to turn the check off")
	f.StringVar(&flags.format, "format", string(report.Text), "output format: "+report.Names(render.Formats()...))
	f.BoolVar(&flags.json, "json", false, "write JSON, the same as --format json")
	f.BoolVar(&flags.strict, "strict", false, "fail on a warning as well")
	return command
}

func runSources(ctx context.Context, deps Dependencies, flags sourcesFlags, roots []string) error {
	format, err := outputFormat(flags)
	if err != nil {
		return err
	}
	if len(roots) == 0 {
		roots = []string{"."}
	}
	outcome := collect(ctx, deps, flags, roots)

	if err := render.Write(deps.Out, format, outcome, Name, version); err != nil {
		return err
	}
	policy := findings.DefaultPolicy()
	if flags.strict {
		policy = findings.StrictPolicy()
	}
	if code := findings.ExitCode(outcome.Findings(), outcome.Err, policy); code != findings.ExitClean {
		return ExitError{Code: code}
	}
	return nil
}

func outputFormat(flags sourcesFlags) (report.Format, error) {
	if flags.json {
		return report.JSON, nil
	}
	return report.Parse(flags.format, render.Formats()...)
}

// collect audits every catalog below the roots. A fault stops the run and goes
// into the outcome next to what was found until then.
func collect(ctx context.Context, deps Dependencies, flags sourcesFlags, roots []string) render.Outcome {
	var outcome render.Outcome
	loader, err := deps.LoaderFor(deps.FS)
	if err != nil {
		outcome.Err = err
		return outcome
	}
	run := audit.Deps{FS: deps.FS, Loader: loader, Transports: deps.Transports}
	now := deps.Now()
	opts := audit.Options{
		Now:      now,
		Windows:  rules.Windows{Default: flags.maxAge, Versioned: flags.versionMaxAge},
		Fetch:    flags.fetch,
		Pull:     flags.pull,
		CoverDir: flags.coverDir,
	}
	base := gitbase.Base{Runner: deps.Runner, Ref: flags.base, LoaderFor: deps.LoaderFor}

	for _, root := range roots {
		dirs, err := catalogs.Discover(deps.FS, root)
		if err != nil {
			outcome.Err = err
			return outcome
		}
		if len(dirs) == 0 {
			outcome.Catalogs = append(outcome.Catalogs, audit.CatalogReport{Dir: root, Findings: []findings.Finding{{
				Check: rules.CheckName, Level: findings.Warn, Code: rules.CodeNoSources, Path: root,
				Message: "no source catalog found",
			}}})
			continue
		}
		for _, dir := range dirs {
			if flags.markReviewed {
				if err := markReviewed(deps, loader, dir, flags.match, now); err != nil {
					outcome.Err = err
					return outcome
				}
			}
			report, err := audit.Catalog(ctx, run, dir, opts)
			if err != nil {
				outcome.Err = err
				return outcome
			}
			outcome.Catalogs = append(outcome.Catalogs, report)
			if flags.base == "" {
				continue
			}
			changes, err := compare.Against(base, dir, report.Catalog)
			if err != nil {
				outcome.Err = err
				return outcome
			}
			outcome.Compare = append(outcome.Compare, changes...)
		}
	}
	return outcome
}

// markReviewed stamps the sources of a catalog and tells what changed.
func markReviewed(deps Dependencies, loader sources.Loader, dir, match string, now time.Time) error {
	catalog, _, err := loader.Load(dir)
	if err != nil {
		return err
	}
	result, err := stamp.Reviewed(deps.FS, catalog, match, now)
	if err != nil {
		return err
	}
	if len(result.Edited) > 0 {
		_, err = fmt.Fprintf(deps.Err, "%s: marked %d source(s) as reviewed on %s\n", dir, len(result.Edited), now.Format("2006-01-02"))
	}
	return err
}
