package gitbase_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process/processtest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/compare/gitbase"
)

const catalog = `{"version":1,"sources":[{"id":"a","kind":"original","title":"A","provenance":"p","reviewedAt":"2026-01-01"}]}`

func newBase(fake *processtest.Fake) gitbase.Base {
	return gitbase.Base{Runner: fake, Ref: "main", LoaderFor: wiring.NewSourcesLoader}
}

func known() *processtest.Fake {
	return processtest.Script().On("git rev-parse --verify --quiet main^{commit}", processtest.Reply{})
}

func TestCatalogReadsTheTreeOfTheRef(t *testing.T) {
	fake := known().
		On("git cat-file -t main:./sources.json", processtest.Reply{Stdout: "blob\n"}).
		On("git cat-file -s main:./sources.json", processtest.Reply{Stdout: "9\n"}).
		On("git show main:./sources.json", processtest.Reply{Stdout: catalog}).
		On("git cat-file -t main:./sources.d", processtest.Reply{ExitCode: 128})
	got, err := newBase(fake).Catalog("cat")
	if err != nil || len(got.Sources) != 1 || got.Sources[0].ID != "a" {
		t.Fatalf("Catalog = %+v, %v", got, err)
	}
}

func TestACatalogThatDidNotExistIsEmpty(t *testing.T) {
	fake := known().
		On("git cat-file -t main:./sources.json", processtest.Reply{ExitCode: 128}).
		On("git cat-file -t main:./sources.d", processtest.Reply{ExitCode: 128})
	got, err := newBase(fake).Catalog("cat")
	if err != nil || len(got.Sources) != 0 {
		t.Fatalf("Catalog = %+v, %v", got, err)
	}
}

func TestAnUnknownRefIsAnError(t *testing.T) {
	fake := processtest.Script().On("git rev-parse --verify --quiet main^{commit}", processtest.Reply{ExitCode: 1})
	if _, err := newBase(fake).Catalog("cat"); !errors.Is(err, gitbase.ErrUnknownRef) {
		t.Errorf("Catalog: %v", err)
	}
	if _, err := newBase(fake).ChangedFiles("cat"); !errors.Is(err, gitbase.ErrUnknownRef) {
		t.Errorf("ChangedFiles: %v", err)
	}
}

func TestFaultsPropagate(t *testing.T) {
	boom := errors.New("boom")
	failing := processtest.Script().On("git rev-parse --verify --quiet main^{commit}", processtest.Reply{Err: boom})
	if _, err := newBase(failing).Catalog("cat"); !errors.Is(err, boom) {
		t.Errorf("verify: %v", err)
	}

	unreadable := known().
		On("git cat-file -t main:./sources.json", processtest.Reply{Err: boom})
	if _, err := newBase(unreadable).Catalog("cat"); !errors.Is(err, boom) {
		t.Errorf("load: %v", err)
	}

	noLoader := newBase(known())
	noLoader.LoaderFor = func(fs.FileSystem) (sources.Loader, error) { return sources.Loader{}, boom }
	if _, err := noLoader.Catalog("cat"); !errors.Is(err, boom) {
		t.Errorf("loader: %v", err)
	}

	diffFails := known().On("git diff --name-only --relative main -- .", processtest.Reply{Err: boom})
	if _, err := newBase(diffFails).ChangedFiles("cat"); !errors.Is(err, boom) {
		t.Errorf("diff: %v", err)
	}
}

func TestChangedFiles(t *testing.T) {
	fake := known().On("git diff --name-only --relative main -- .", processtest.Reply{Stdout: "references/a.md\n\n sources.json \n"})
	got, err := newBase(fake).ChangedFiles("cat")
	if err != nil || !reflect.DeepEqual(got, []string{"references/a.md", "sources.json"}) {
		t.Fatalf("ChangedFiles = %v, %v", got, err)
	}
	failed := known().On("git diff --name-only --relative main -- .", processtest.Reply{ExitCode: 128})
	if _, err := newBase(failed).ChangedFiles("cat"); err == nil {
		t.Error("a failing diff must be an error")
	}
}
