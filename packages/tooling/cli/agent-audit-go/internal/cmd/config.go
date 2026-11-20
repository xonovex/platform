package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/explain"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/resolve"
	auditsection "github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/audit"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// settings are the merged audit section, with the defaults of the unset keys.
type settings struct {
	MaxAge        int    `json:"maxAge"`
	VersionMaxAge int    `json:"versionMaxAge"`
	CoverDir      string `json:"coverDir"`
}

// resolveConfig merges the audit section of every layer with the flags the
// user set. The section is not gated, so the resolution needs no approval
// port.
func resolveConfig(deps Dependencies, project string, flags []merge.Flag) (merge.Resolved, error) {
	env := resolve.Env{Files: deps.ConfigFiles, Places: deps.Places, Validator: deps.Validator}
	return resolve.Resolve(env, resolve.Request{Sections: []spec.Section{auditsection.Section()}, Start: ".", Project: project, Flags: flags})
}

// loadSettings resolves the audit section and fills the keys no layer sets
// with the defaults of the sources command.
func loadSettings(deps Dependencies, project string, flags []merge.Flag) (settings, error) {
	resolved, err := resolveConfig(deps, project, flags)
	if err != nil {
		return settings{}, err
	}
	merged := settings{MaxAge: defaultMaxAge, VersionMaxAge: defaultVersionMaxAge, CoverDir: defaultCoverDir}
	if err := json.Unmarshal(resolved.Section(auditsection.Name), &merged); err != nil {
		return settings{}, fmt.Errorf("decode the audit section: %w", err)
	}
	return merged, nil
}

// configFlags turns the flags the user set into the flags layer of the merge.
func (f sourcesFlags) configFlags(c *cobra.Command) []merge.Flag {
	var flags []merge.Flag
	set := c.Flags().Changed
	if set("max-age") {
		flags = append(flags, merge.Flag{Section: auditsection.Name, Pointer: "/maxAge", Value: json.RawMessage(strconv.Itoa(f.maxAge))})
	}
	if set("version-max-age") {
		flags = append(flags, merge.Flag{Section: auditsection.Name, Pointer: "/versionMaxAge", Value: json.RawMessage(strconv.Itoa(f.versionMaxAge))})
	}
	if set("cover-dir") {
		// A string always encodes.
		value, _ := json.Marshal(f.coverDir)
		flags = append(flags, merge.Flag{Section: auditsection.Name, Pointer: "/coverDir", Value: value})
	}
	return flags
}

func bindProject(c *cobra.Command, project *string) {
	c.Flags().StringVar(project, "project", "", "project agentic.json (default the nearest agentic.json from the working directory up to the repository root)")
}

func newConfigCommand(deps Dependencies) *cobra.Command {
	config := &cobra.Command{Use: "config", Short: "Inspect the agentic.json configuration", Args: cobra.NoArgs}
	var project string
	var sources bool
	show := &cobra.Command{
		Use:   "show",
		Short: "Print the merged audit section",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			resolved, err := resolveConfig(deps, project, nil)
			if err != nil {
				return err
			}
			if sources {
				return explain.Render(c.OutOrStdout(), resolved)
			}
			var out bytes.Buffer
			// The merged section is valid JSON, which always indents.
			_ = json.Indent(&out, resolved.Section(auditsection.Name), "", "  ")
			out.WriteByte('\n')
			_, err = out.WriteTo(c.OutOrStdout())
			return err
		},
	}
	bindProject(show, &project)
	show.Flags().BoolVar(&sources, "sources", false, "name the layer and the file of each value")
	config.AddCommand(show)
	return config
}
