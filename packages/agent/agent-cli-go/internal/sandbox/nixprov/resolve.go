package nixprov

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/nix"
)

// ResolveClosure resolves a NixSource to a content-pinned closure using the HOST
// nix daemon. Resolution happens on the host; the isolator later binds the
// resulting requisites read-only — the sandbox never runs nix, so the host nix
// daemon socket is never exposed through a container/pod boundary.
//
// The committed lock is enforced at resolve time (--no-write-lock-file): a
// dirty or missing lock makes nix error rather than silently re-pin, so the pin
// is a real guarantee, not a doc note (fail closed).
func ResolveClosure(src nix.NixSource) (nix.ClosureDescriptor, error) {
	switch src.Kind {
	case nix.NixSourcePackages:
		return resolvePackages(src)
	case nix.NixSourceProjectFlake:
		return resolveFlake(src)
	default:
		return nix.ClosureDescriptor{}, fmt.Errorf("unknown nix source kind %q", src.Kind)
	}
}

// resolvePackages realizes a rev-pinned package set to its store paths. The rev
// in the flake URL is the pin; --no-write-lock-file keeps resolution frozen.
func resolvePackages(src nix.NixSource) (nix.ClosureDescriptor, error) {
	if src.Rev == "" {
		return nix.ClosureDescriptor{}, fmt.Errorf("nix packages source requires a pinned --nix-rev")
	}
	pkgs := nix.ExpandPackageSets(src.Packages)
	if len(pkgs) == 0 {
		return nix.ClosureDescriptor{}, fmt.Errorf("nix packages source requires at least one package")
	}

	args := []string{"build", "--no-link", "--print-out-paths", "--no-write-lock-file"}
	for _, p := range pkgs {
		args = append(args, fmt.Sprintf("github:NixOS/nixpkgs/%s#%s", src.Rev, p))
	}
	out, err := runNix(args...)
	if err != nil {
		return nix.ClosureDescriptor{}, err
	}
	storePaths := nonEmptyLines(out)
	if len(storePaths) == 0 {
		return nix.ClosureDescriptor{}, fmt.Errorf("nix build produced no store paths")
	}

	pathEntries := make([]string, len(storePaths))
	for i, p := range storePaths {
		pathEntries[i] = p + "/bin"
	}
	requisites, err := pathInfoRequisites(storePaths)
	if err != nil {
		return nix.ClosureDescriptor{}, err
	}
	return nix.ClosureDescriptor{StorePaths: storePaths, Requisites: requisites, PathEntries: pathEntries}, nil
}

// resolveFlake resolves the project's devShell to its tool closure. It reads the
// resolved PATH from `nix print-dev-env --json` and derives the bound store paths
// from it (the devShell's buildInputs land on PATH).
func resolveFlake(src nix.NixSource) (nix.ClosureDescriptor, error) {
	if src.FlakeRef == "" {
		return nix.ClosureDescriptor{}, fmt.Errorf("nix flake source requires a flake reference")
	}
	installable := fmt.Sprintf("%s#%s", src.FlakeRef, src.Shell)
	out, err := runNix("print-dev-env", "--json", "--no-write-lock-file", installable)
	if err != nil {
		return nix.ClosureDescriptor{}, err
	}

	var devEnv struct {
		Variables map[string]struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"variables"`
	}
	if err := json.Unmarshal([]byte(out), &devEnv); err != nil {
		return nix.ClosureDescriptor{}, fmt.Errorf("parse print-dev-env JSON: %w", err)
	}

	pathVar, ok := devEnv.Variables["PATH"]
	if !ok || pathVar.Value == "" {
		return nix.ClosureDescriptor{}, fmt.Errorf("devShell %q exposes no PATH", installable)
	}

	pathEntries := storePathEntries(pathVar.Value)
	if len(pathEntries) == 0 {
		return nix.ClosureDescriptor{}, fmt.Errorf("devShell %q PATH has no /nix/store entries", installable)
	}
	storePaths := storeRoots(pathEntries)
	requisites, err := pathInfoRequisites(storePaths)
	if err != nil {
		return nix.ClosureDescriptor{}, err
	}
	return nix.ClosureDescriptor{StorePaths: storePaths, Requisites: requisites, PathEntries: pathEntries}, nil
}

// pathInfoRequisites enumerates the transitive closure of the given store paths
// (`nix path-info -r`). This is the exact set the isolator binds read-only —
// never the whole world-readable /nix/store.
func pathInfoRequisites(storePaths []string) ([]string, error) {
	if len(storePaths) == 0 {
		return nil, nil
	}
	out, err := runNix(append([]string{"path-info", "-r", "--no-write-lock-file"}, storePaths...)...)
	if err != nil {
		return nil, err
	}
	requisites := nonEmptyLines(out)
	sort.Strings(requisites)
	return dedupe(requisites), nil
}

// runNix runs the host nix CLI and returns stdout, wrapping stderr on failure.
func runNix(args ...string) (string, error) {
	return runCmd("nix", args...)
}

// runCmd runs bin with args and returns stdout, wrapping stderr on failure.
func runCmd(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w: %s", bin, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// storePathEntries returns the /nix/store/.../bin (and friends) entries of a PATH.
func storePathEntries(path string) []string {
	var entries []string
	for _, p := range strings.Split(path, ":") {
		if strings.HasPrefix(p, "/nix/store/") {
			entries = append(entries, p)
		}
	}
	return entries
}

// storeRoots maps store sub-paths to their top-level store path (/nix/store/<hash>-<name>).
func storeRoots(paths []string) []string {
	seen := make(map[string]bool)
	var roots []string
	for _, p := range paths {
		if root := storeRoot(p); root != "" && !seen[root] {
			seen[root] = true
			roots = append(roots, root)
		}
	}
	return roots
}

// storeRoot returns the top-level store path for a path under /nix/store, or "".
func storeRoot(p string) string {
	if !strings.HasPrefix(p, "/nix/store/") {
		return ""
	}
	parts := strings.SplitN(p, "/", 5) // "", "nix", "store", "<hash>-<name>", rest
	if len(parts) < 4 || parts[3] == "" {
		return ""
	}
	return "/nix/store/" + parts[3]
}

func nonEmptyLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			lines = append(lines, t)
		}
	}
	return lines
}

func dedupe(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
