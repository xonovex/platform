package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/record"
)

// schemas are the JSON Schemas the CLI prints, by name.
var schemas = map[string]func() []byte{
	"config":     config.Schema,
	"provider":   providers.Schema,
	"run-record": record.Schema,
}

func schemaNames() []string {
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func writeJSON(out io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	_, err = fmt.Fprintf(out, "%s\n", data)
	return err
}

// newSchemaCommand builds `agent-cli schema`: it prints a JSON Schema, so an
// editor or a pipeline can validate a configuration file or a run record without
// reading the source.
func newSchemaCommand(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:       "schema [" + strings.Join(schemaNames(), "|") + "]",
		Short:     "Print a JSON Schema of the configuration file, a provider file or the run record",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: schemaNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				_, err := fmt.Fprintln(out, strings.Join(schemaNames(), "\n"))
				return err
			}
			build, ok := schemas[args[0]]
			if !ok {
				return UsageError{Err: fmt.Errorf("unknown schema %q; available schemas: %s", args[0], strings.Join(schemaNames(), ", "))}
			}
			// Both schemas end with a newline.
			_, err := out.Write(build())
			return err
		},
	}
}
