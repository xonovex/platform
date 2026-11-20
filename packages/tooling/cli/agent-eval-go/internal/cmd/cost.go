package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/cost"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

func newCostCommand(deps Dependencies) *cobra.Command {
	var file, protocolName string
	c := &cobra.Command{
		Use:   "cost",
		Short: "Show what the eval runs cost",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			history, err := (cost.Store{FS: deps.FS, Reader: deps.reader()}).Load(resolvePath(deps.Dir, file))
			if err != nil {
				return err
			}
			summary := cost.Summarize(history, model.Protocol(protocolName))
			var lines strings.Builder
			for _, p := range summary.Protocols() {
				t := summary.ByProtocol[p]
				fmt.Fprintf(&lines, "%-14s %3d runs %4d model runs $%.4f\n", p, t.Runs, t.ModelRuns, t.Usage.CostUSD)
			}
			t := summary.All
			fmt.Fprintf(&lines, "%-14s %3d runs %4d model runs $%.4f (%d input tokens, %d output tokens)\n",
				"total", t.Runs, t.ModelRuns, t.Usage.CostUSD, t.Usage.InputTokens, t.Usage.OutputTokens)
			_, err = io.WriteString(c.OutOrStdout(), lines.String())
			return err
		},
	}
	c.Flags().StringVar(&file, "file", DefaultResultsDir+"/"+CostFile, "cost history file")
	c.Flags().StringVar(&protocolName, "protocol", "", "only this protocol")
	return c
}
