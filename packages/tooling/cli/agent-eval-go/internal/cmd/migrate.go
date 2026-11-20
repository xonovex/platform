package cmd

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/migrate"
)

// maxWalkDepth bounds how far below a given directory migrate looks for skills:
// a repository root, a plugin group, a plugin, its skills directory, a skill.
const maxWalkDepth = 6

func newMigrateCommand(deps Dependencies) *cobra.Command {
	var removeLegacy bool
	c := &cobra.Command{
		Use:   "migrate <path>...",
		Short: "Convert the legacy evals.json and eval-queries.json of skills to the eval documents",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			converter := migrate.Converter{FS: deps.FS, Reader: deps.reader(), RemoveLegacy: removeLegacy}
			var lines strings.Builder
			total := 0
			for _, arg := range args {
				dirs, err := skillDirs(deps.FS, resolve(deps.Dir, arg), 0)
				if err != nil {
					return err
				}
				for _, dir := range dirs {
					res, err := converter.Skill(dir)
					if err != nil {
						return fmt.Errorf("%s: %w", dir, err)
					}
					for _, w := range res.Written {
						fmt.Fprintf(&lines, "wrote %s\n", w)
					}
					for _, r := range res.Removed {
						fmt.Fprintf(&lines, "removed %s\n", r)
					}
					total += len(res.Written)
				}
			}
			fmt.Fprintf(&lines, "%d documents written\n", total)
			_, err := io.WriteString(c.OutOrStdout(), lines.String())
			return err
		},
	}
	c.Flags().BoolVar(&removeLegacy, "remove-legacy", false, "delete the legacy files after their replacements are written")
	return c
}

// skillDirs finds the skill directories at or below dir: a directory that holds
// a SKILL.md is a skill and is not searched further.
func skillDirs(fsys fs.FileSystem, dir string, depth int) ([]string, error) {
	if ok, err := fs.Exists(fsys, path.Join(dir, "SKILL.md")); err != nil || ok {
		if ok {
			return []string{dir}, nil
		}
		return nil, err
	}
	if depth >= maxWalkDepth {
		return nil, nil
	}
	entries, err := fsys.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir {
			continue
		}
		found, err := skillDirs(fsys, path.Join(dir, e.Name), depth+1)
		if err != nil {
			return nil, err
		}
		out = append(out, found...)
	}
	return out, nil
}
