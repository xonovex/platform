package cmd

import (
	"bytes"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/explain"
	evalsection "github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/eval"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
)

// bindConfig binds the flags that choose and approve the project agentic.json.
func bindConfig(c *cobra.Command, project, approveDigest *string) {
	f := c.Flags()
	f.StringVar(project, "project", "", "project agentic.json (default the nearest agentic.json from the working directory up to the repository root)")
	f.StringVar(approveDigest, trust.ApproveFlag, "", "approve the eval section of the project agentic.json with this digest")
}

func newConfigCommand(deps Dependencies) *cobra.Command {
	config := &cobra.Command{Use: "config", Short: "Inspect the agentic.json configuration", Args: cobra.NoArgs}
	var r configRequest
	var sources bool
	show := &cobra.Command{
		Use:   "show",
		Short: "Print the merged eval section",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			resolved, err := deps.resolveConfig(r)
			if err != nil {
				return err
			}
			if sources {
				return explain.Render(c.OutOrStdout(), resolved)
			}
			var out bytes.Buffer
			// The merged section is valid JSON, which always indents.
			_ = json.Indent(&out, resolved.Section(evalsection.Name), "", "  ")
			out.WriteByte('\n')
			_, err = out.WriteTo(c.OutOrStdout())
			return err
		},
	}
	bindConfig(show, &r.project, &r.approveDigest)
	show.Flags().BoolVar(&sources, "sources", false, "name the layer and the file of each value")
	config.AddCommand(show)
	return config
}
