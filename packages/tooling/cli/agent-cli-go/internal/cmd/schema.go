package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	contract "github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/record"
)

// schemaSource returns one JSON Schema.
type schemaSource func() ([]byte, error)

// schemas are the JSON Schemas the CLI prints, by name.
var schemas = map[string]schemaSource{
	"config":     infallible(config.Schema),
	"provider":   providerSchema,
	"run-record": infallible(record.Schema),
}

func infallible(build func() []byte) schemaSource {
	return func() ([]byte, error) { return build(), nil }
}

// providerSchema returns the provider schema, which shared-contract-go owns
// because the providers section of agentic.json references it.
func providerSchema() ([]byte, error) {
	documents, err := contract.Documents()
	return documents[contract.Provider], err
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
	return schemaCommand(out, schemas)
}

func schemaCommand(out io.Writer, sources map[string]schemaSource) *cobra.Command {
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
			build, ok := sources[args[0]]
			if !ok {
				return UsageError{Err: fmt.Errorf("unknown schema %q; available schemas: %s", args[0], strings.Join(schemaNames(), ", "))}
			}
			document, err := build()
			if err != nil {
				return fmt.Errorf("read the %s schema: %w", args[0], err)
			}
			// Every schema ends with a newline.
			_, err = out.Write(document)
			return err
		},
	}
}
