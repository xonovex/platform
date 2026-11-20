// Command agent-validate is the composition root of the validator: it binds the
// file system, the configuration layers, the schema validator and git to the
// ports of the command tree.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/osfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/osfiles"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/cmd"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/gitrepo"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/registry"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// version is set at build time.
var version = "dev"

func main() {
	validator, err := wiring.NewValidator()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(findings.ExitUsage)
	}
	deps := cmd.Dependencies{
		Open:        open,
		ConfigFiles: osfiles.Files{},
		Places:      osfiles.Places{},
		Validator:   validator,
		Checks:      registry.All(),
		Out:         os.Stdout,
		Version:     version,
	}
	err = cmd.NewRootCommand(deps).Execute()
	if err == nil {
		return
	}
	// A bare exit code means the report already said why.
	var exit cmd.ExitError
	if !errors.As(err, &exit) || exit.Err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(cmd.ExitCode(err))
}

// open reads the repository at root from the disk and from git.
func open(root string) (cmd.Environment, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return cmd.Environment{}, err
	}
	return cmd.Environment{
		Tree:       tree.Tree{FS: osfs.New(), Root: abs},
		Repository: gitrepo.Repository{Dir: abs, Run: git},
	}, nil
}

func git(dir string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = dir
	out, err := command.Output()
	return string(out), err
}
