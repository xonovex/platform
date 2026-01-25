package nixenv

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/xonovex/platform/packages/tools/tool-lib-go/pkg/scriptlib"
)

// DefaultBuildTimeout is the default build timeout (30 minutes)
const DefaultBuildTimeout = 30 * time.Minute

// BuildOptions are options for building a Nix environment
type BuildOptions struct {
	Verbose bool
	Debug   bool
	Timeout time.Duration
}

// EnsureDirectories ensures required directories exist
func EnsureDirectories() error {
	if err := os.MkdirAll(GetSpecsDir(), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(GetEnvsDir(), 0755); err != nil {
		return err
	}
	return nil
}

// writeSpecAtomic writes a spec file atomically using temp file + rename
func writeSpecAtomic(specPath string, content string) error {
	dir := filepath.Dir(specPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpPath := filepath.Join(dir, ".tmp-spec")
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, specPath)
}

// runNixBuild runs nix-build and returns the result
func runNixBuild(specPath string, outLink string, timeout time.Duration, verbose bool) *BuildResult {
	startTime := time.Now()

	args := []string{specPath, "-o", outLink}

	if verbose {
		scriptlib.LogInfo("Running: nix-build " + specPath + " -o " + outLink)
	}

	cmd := exec.Command("nix-build", args...)
	cmd.Env = append(os.Environ(), "NIXPKGS_ALLOW_UNFREE=1")

	var stderr bytes.Buffer
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = &stderr
	}

	// Create a channel for the result
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Wait with timeout
	select {
	case err := <-done:
		duration := time.Since(startTime).Milliseconds()
		if err != nil {
			return &BuildResult{
				Success:  false,
				Error:    stderr.String(),
				Duration: duration,
			}
		}

		// Get the real path of the output
		storePath, err := filepath.EvalSymlinks(outLink)
		if err != nil {
			return &BuildResult{
				Success:  false,
				Error:    "Build completed but output link invalid: " + err.Error(),
				Duration: duration,
			}
		}

		return &BuildResult{
			Success:   true,
			StorePath: storePath,
			Duration:  duration,
		}

	case <-time.After(timeout):
		cmd.Process.Kill()
		return &BuildResult{
			Success:  false,
			Error:    "Build timed out",
			Duration: time.Since(startTime).Milliseconds(),
		}
	}
}

// BuildEnv builds a Nix environment from an EnvSpec
// If the environment is already built (cache hit), returns immediately.
// Otherwise, generates the .nix file and runs nix-build.
func BuildEnv(spec *EnvSpec, opts BuildOptions) (*ResolvedEnv, *BuildResult, error) {
	if opts.Timeout == 0 {
		opts.Timeout = DefaultBuildTimeout
	}

	if err := EnsureDirectories(); err != nil {
		return nil, nil, err
	}

	resolved, err := ResolveEnv(spec)
	if err != nil {
		return nil, nil, err
	}

	if opts.Debug {
		scriptlib.LogDebug(opts.Debug, "EnvID: "+resolved.EnvID)
		scriptlib.LogDebug(opts.Debug, "SpecPath: "+resolved.SpecPath)
		scriptlib.LogDebug(opts.Debug, "OutLink: "+resolved.OutLink)
	}

	// Cache hit - environment already built
	if resolved.Ready {
		if opts.Verbose {
			scriptlib.LogInfo("Using cached environment: " + resolved.EnvID)
		}
		storePath, _ := filepath.EvalSymlinks(resolved.OutLink)
		return resolved, &BuildResult{
			Success:   true,
			StorePath: storePath,
			Duration:  0,
		}, nil
	}

	// Generate and write the Nix expression
	nixExpr, err := RenderNixExpression(spec, resolved.EnvID)
	if err != nil {
		return nil, nil, err
	}

	if opts.Debug {
		scriptlib.LogDebug(opts.Debug, "Generated Nix expression:\n"+nixExpr)
	}

	if _, err := os.Stat(resolved.SpecPath); os.IsNotExist(err) {
		if err := writeSpecAtomic(resolved.SpecPath, nixExpr); err != nil {
			return nil, nil, err
		}
		if opts.Verbose {
			scriptlib.LogInfo("Wrote spec file: " + resolved.SpecPath)
		}
	}

	// Run nix-build
	if opts.Verbose {
		scriptlib.LogInfo("Building environment: " + resolved.EnvID)
	}

	result := runNixBuild(resolved.SpecPath, resolved.OutLink, opts.Timeout, opts.Verbose)

	if !result.Success {
		scriptlib.LogError("Build failed: " + result.Error)
	} else if opts.Verbose {
		scriptlib.LogInfo("Build completed")
	}

	return resolved, result, nil
}
