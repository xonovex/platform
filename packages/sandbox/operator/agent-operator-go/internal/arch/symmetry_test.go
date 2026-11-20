//go:build arch

package arch

import (
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
)

// sharedAxes are the orthogonal axes that the operator realizes, each as a
// top-level internal directory, and the shared port or type that declares the
// axis for every consumer.
var sharedAxes = map[string]reflect.Type{
	"isolation": reflect.TypeFor[isolation.Confinement](),
	"provision": reflect.TypeFor[provision.Provisioner](),
	"network":   reflect.TypeFor[network.NetworkMethod](),
	"workspace": reflect.TypeFor[workspace.VCSType](),
}

// operatorOnlyAxes are the deliberate asymmetries: true for an axis only the
// operator realizes, false for one it must not have. The operator renders the
// harness command and the provider environment into a Job itself, and it runs
// pods, not terminals.
var operatorOnlyAxes = map[string]bool{
	"harness":  true,
	"provider": true,
	"terminal": false,
}

func hasDir(base, name string) bool {
	info, err := os.Stat(filepath.Join(base, name))
	return err == nil && info.IsDir()
}

func TestAxisSymmetry(t *testing.T) {
	internal := filepath.Join(moduleRoot(t), "internal")
	for axis, declaration := range sharedAxes {
		if declared := path.Base(declaration.PkgPath()); declared != axis {
			t.Errorf("axis %q is declared by the shared package %q; the operator directory and the shared package share one name", axis, declared)
		}
		if !hasDir(internal, axis) {
			t.Errorf("axis %q is missing from the operator", axis)
		}
	}
	for axis, want := range operatorOnlyAxes {
		if present := hasDir(internal, axis); present != want {
			t.Errorf("axis %q in the operator: present %v, want %v", axis, present, want)
		}
	}
}
