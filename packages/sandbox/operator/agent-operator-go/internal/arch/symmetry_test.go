//go:build arch

package arch

import (
	"os"
	"path/filepath"
	"testing"
)

// sharedAxes are the orthogonal axes that the operator and the CLI both
// realize, each as a top-level internal directory.
var sharedAxes = []string{"isolation", "network", "provision", "workspace"}

// consumerOnlyAxes are the deliberate asymmetries and the consumer that owns
// each. The CLI takes harness commands and provider environments from the
// shared agent library, and the operator runs pods, not terminals.
var consumerOnlyAxes = map[string]string{
	"harness":  "operator",
	"provider": "operator",
	"terminal": "cli",
}

func hasDir(base, name string) bool {
	info, err := os.Stat(filepath.Join(base, name))
	return err == nil && info.IsDir()
}

func TestAxisSymmetry(t *testing.T) {
	operatorRoot := moduleRoot(t)
	packagesRoot := filepath.Dir(filepath.Dir(filepath.Dir(operatorRoot)))
	internal := map[string]string{
		"operator": filepath.Join(operatorRoot, "internal"),
		"cli":      filepath.Join(packagesRoot, "tooling", "cli", "agent-cli-go", "internal"),
	}
	if !hasDir(internal["cli"], "") {
		t.Fatalf("CLI module not found at %s", internal["cli"])
	}

	for _, axis := range sharedAxes {
		for consumer, dir := range internal {
			if !hasDir(dir, axis) {
				t.Errorf("axis %q is missing from the %s", axis, consumer)
			}
		}
	}
	for axis, owner := range consumerOnlyAxes {
		for consumer, dir := range internal {
			if present := hasDir(dir, axis); present != (consumer == owner) {
				t.Errorf("axis %q in the %s: present %v, want it only in the %s", axis, consumer, present, owner)
			}
		}
	}
}
