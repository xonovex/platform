package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/budget"
)

// defaultCatalog is the catalog that the budgets cover unless a path says
// otherwise.
const defaultCatalog = "packages/runtime/plugin"

// seedCommand records the current word count of every skill file as its budget,
// so the ratchet of the drift check starts from the catalog as it is.
func seedCommand(d Dependencies, root *string) *cobra.Command {
	var paths []string
	var file string
	c := &cobra.Command{
		Use:   "seed-budgets",
		Short: "Record the current word count of each skill file as its budget",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			env, err := d.Open(*root)
			if err != nil {
				return usage(err)
			}
			if len(paths) == 0 {
				paths = []string{defaultCatalog}
			}
			files, err := budget.Collect(env.Tree, repositoryPaths(env.Tree, paths))
			if err != nil {
				return usage(err)
			}
			seeded := budget.Seed(files)
			data, err := budget.Marshal(seeded)
			if err != nil {
				return usage(err)
			}
			if err := env.Tree.FS.WriteFile(env.Tree.Abs(file), data); err != nil {
				return usage(err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "recorded %d budgets in %s\n", len(seeded), file)
			if err != nil {
				return usage(err)
			}
			return nil
		},
	}
	c.Flags().StringArrayVar(&paths, "path", nil, "catalog path to measure, absolute or relative to the root, repeatable (default "+defaultCatalog+")")
	c.Flags().StringVar(&file, "file", "budgets.json", "budget manifest to write, absolute or relative to the root")
	return c
}
