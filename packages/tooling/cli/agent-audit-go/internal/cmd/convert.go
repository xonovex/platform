package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/legacy"
)

const legacyFile = "SOURCES.md"

type convertFlags struct {
	stdout bool
	force  bool
}

func newConvertCommand(deps Dependencies) *cobra.Command {
	var flags convertFlags
	command := &cobra.Command{
		Use:   "convert <skill-directory | SOURCES.md>",
		Short: "Convert the SOURCES.md of one skill into sources.json",
		Long: `Convert one SOURCES.md into a sources.json beside it. The conversion keeps
every word of the file. The command checks the result against the catalog
schema, refuses to replace an existing sources.json unless --force is given,
and leaves SOURCES.md in place.

Exit status: 0 when the catalog is written, 1 when the file cannot be converted
without loss, 2 when the run could not complete.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConvert(deps, flags, args[0])
		},
	}
	command.Flags().BoolVar(&flags.stdout, "stdout", false, "write the catalog to standard output instead of a file")
	command.Flags().BoolVar(&flags.force, "force", false, "replace an existing sources.json")
	return command
}

func runConvert(deps Dependencies, flags convertFlags, target string) error {
	legacyPath, dir, err := locate(deps.FS, target)
	if err != nil {
		return err
	}
	raw, err := deps.FS.ReadFile(legacyPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", legacyPath, err)
	}

	result := legacy.Convert(string(raw), legacyPath)
	data, err := sources.Encode(result.Document)
	if err != nil {
		return err
	}
	problems := result.Findings
	violations, err := deps.Validator.Validate(schemas.Sources, data)
	if err != nil {
		return err
	}
	for _, v := range violations {
		problems = append(problems, findings.Finding{
			Check: legacy.CheckName, Level: findings.Fail, Code: "convert.invalid", Path: legacyPath,
			Message: fmt.Sprintf("%s: %s", v.Pointer, v.Message),
		})
	}
	if len(problems) > 0 {
		for _, f := range problems {
			if _, err := fmt.Fprintln(deps.Err, f); err != nil {
				return err
			}
		}
		return ExitError{Code: findings.ExitFindings}
	}

	if flags.stdout {
		_, err = deps.Out.Write(data)
		return err
	}
	out := path.Join(dir, sources.CatalogFile)
	if exists, err := fs.Exists(deps.FS, out); err != nil {
		return err
	} else if exists && !flags.force {
		return fmt.Errorf("%s exists; use --force to replace it", out)
	}
	if err := deps.FS.WriteFile(out, data); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	_, err = fmt.Fprintf(deps.Err, "wrote %s (%d source(s))\n", out, len(result.Document.Sources))
	return err
}

// locate resolves the argument, a skill directory or the SOURCES.md itself, to
// the path of the legacy file and the directory it sits in.
func locate(fsys fs.FileSystem, target string) (legacyPath, dir string, err error) {
	info, err := fsys.Stat(target)
	if err != nil {
		return "", "", fmt.Errorf("stat %s: %w", target, err)
	}
	if info.IsDir {
		return path.Join(target, legacyFile), target, nil
	}
	return target, path.Dir(target), nil
}
