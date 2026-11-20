package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/baseline"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// DefaultBaselineFile is where a baseline is kept, below the working directory.
const DefaultBaselineFile = ".agent-eval/baseline.json"

func newBaselineCommand(deps Dependencies) *cobra.Command {
	root := &cobra.Command{Use: "baseline", Short: "Record the pass rate of each case and compare a later run with it"}
	var file string
	var tolerance float64
	record := &cobra.Command{
		Use:   "record <results.json>",
		Short: "Record the pass rate of each case of a finished run as the baseline",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			results, err := deps.reader().Results(resolvePath(deps.Dir, args[0]))
			if err != nil {
				return err
			}
			if results.Status != model.Complete {
				return fmt.Errorf("run %s is %s: only a complete run makes a baseline", results.RunID, results.Status)
			}
			b := baseline.Record(results, tolerance, deps.Now().UTC().Format(time.RFC3339))
			target := resolvePath(deps.Dir, file)
			if err := (baseline.Store{FS: deps.FS, Reader: deps.reader()}).Save(target, b); err != nil {
				return err
			}
			_, err = fmt.Fprintf(c.OutOrStdout(), "recorded %d cases in %s\n", len(b.Cases), target)
			return err
		},
	}
	record.Flags().Float64Var(&tolerance, "tolerance", baseline.DefaultTolerance, "pass-rate drop that counts as a regression")
	compare := &cobra.Command{
		Use:   "compare <results.json>",
		Short: "Compare a finished run with the baseline; exit 1 on a regression",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			results, err := deps.reader().Results(resolvePath(deps.Dir, args[0]))
			if err != nil {
				return err
			}
			if results.Status != model.Complete {
				return fmt.Errorf("run %s is %s: only a complete run can be compared", results.RunID, results.Status)
			}
			base, found, err := (baseline.Store{FS: deps.FS, Reader: deps.reader()}).Load(resolvePath(deps.Dir, file))
			if err != nil {
				return err
			}
			if !found {
				return fmt.Errorf("no baseline in %s: run baseline record first", file)
			}
			cmp := baseline.Compare(base, results)
			var lines strings.Builder
			for _, d := range cmp.Regressions {
				fmt.Fprintf(&lines, "regressed %s: %.2f -> %.2f\n", d.Key, d.Baseline, d.Current)
			}
			for _, d := range cmp.Improvements {
				fmt.Fprintf(&lines, "improved  %s: %.2f -> %.2f\n", d.Key, d.Baseline, d.Current)
			}
			for _, key := range cmp.Missing {
				fmt.Fprintf(&lines, "missing   %s\n", key)
			}
			for _, key := range cmp.Added {
				fmt.Fprintf(&lines, "new       %s\n", key)
			}
			if !cmp.Regressed() {
				lines.WriteString("no regression against the baseline\n")
			}
			if _, err := io.WriteString(c.OutOrStdout(), lines.String()); err != nil {
				return err
			}
			if cmp.Regressed() {
				return ExitError{Code: findings.ExitFindings}
			}
			return nil
		},
	}
	root.PersistentFlags().StringVar(&file, "file", DefaultBaselineFile, "baseline file")
	root.AddCommand(record, compare)
	return root
}
