package discover_test

import (
	"errors"
	iofs "io/fs"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/configtest"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

var places = configtest.Places{Managed: "/etc/agentic", User: "/home/u/.config/agentic", Wd: "/repo/sub"}

func TestFindListsTheLayersInOrder(t *testing.T) {
	files := configtest.NewFiles(map[string]string{"/repo/.git/HEAD": "", "/repo/agentic.json": "{}"})
	got, err := discover.Find(files, places, "deep", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []discover.Candidate{
		{Layer: spec.Managed, Path: "/etc/agentic/agentic.json"},
		{Layer: spec.User, Path: "/home/u/.config/agentic/agentic.json"},
		{Layer: spec.Project, Path: "/repo/agentic.json"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Find = %+v", got)
	}
}

func TestFindTakesTheNamedProjectFile(t *testing.T) {
	got, err := discover.Find(configtest.NewFiles(nil), configtest.Places{User: "/u"}, "/repo", "conf/agentic.json")
	if err != nil {
		t.Fatal(err)
	}
	want := []discover.Candidate{
		{Layer: spec.User, Path: "/u/agentic.json"},
		{Layer: spec.Project, Path: "/repo/conf/agentic.json", Required: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Find = %+v", got)
	}
	got, _ = discover.Find(configtest.NewFiles(nil), configtest.Places{User: "/u"}, "/repo", "/abs/x.json")
	if got[1].Path != "/abs/x.json" {
		t.Fatalf("absolute project = %+v", got[1])
	}
}

func TestFindWithoutAProjectFile(t *testing.T) {
	got, err := discover.Find(configtest.NewFiles(map[string]string{"/repo/.jj/x": ""}), configtest.Places{User: "/u"}, "/repo", "")
	if err != nil || len(got) != 1 || got[0].Layer != spec.User {
		t.Fatalf("Find = %+v, %v", got, err)
	}
}

func TestFindReportsTheHostFaults(t *testing.T) {
	if _, err := discover.Find(configtest.NewFiles(nil), configtest.Places{Err: errBoom}, "rel", ""); !errors.Is(err, errBoom) || !strings.Contains(err.Error(), "working directory") {
		t.Fatalf("Getwd fault = %v", err)
	}
	if _, err := discover.Find(configtest.NewFiles(nil), configtest.Places{Err: errBoom}, "/abs", ""); !errors.Is(err, errBoom) || !strings.Contains(err.Error(), "user configuration") {
		t.Fatalf("UserDir fault = %v", err)
	}
	if _, err := discover.Find(failing{}, configtest.Places{User: "/u"}, "/abs", ""); !errors.Is(err, errBoom) {
		t.Fatalf("Stat fault = %v", err)
	}
}

func TestProjectStopsAtTheRepositoryRoot(t *testing.T) {
	files := configtest.NewFiles(map[string]string{
		"/agentic.json":           "{}",
		"/repo/.git/HEAD":         "",
		"/repo/a/agentic.json":    "{}",
		"/repo/a/b/c/placeholder": "",
	})
	if path, found, err := discover.Project(files, "/repo/a/b/c"); err != nil || !found || path != "/repo/a/agentic.json" {
		t.Fatalf("nearest = %s, %v, %v", path, found, err)
	}
	if path, found, err := discover.Project(files, "/repo/x"); err != nil || found {
		t.Fatalf("above the root = %s, %v, %v", path, found, err)
	}
}

func TestProjectOutsideARepositorySearchesOnlyTheStart(t *testing.T) {
	files := configtest.NewFiles(map[string]string{"/tmp/agentic.json": "{}", "/tmp/w/x": ""})
	if _, found, err := discover.Project(files, "/tmp/w"); err != nil || found {
		t.Fatalf("found a parent file outside a repository: %v, %v", found, err)
	}
	if path, found, _ := discover.Project(files, "/tmp"); !found || path != "/tmp/agentic.json" {
		t.Fatalf("start = %s, %v", path, found)
	}
}

func TestProjectReportsAFaultOnTheFile(t *testing.T) {
	files := faultOn{Files: configtest.NewFiles(map[string]string{"/r/.git/HEAD": ""}), path: "/r/agentic.json"}
	if _, _, err := discover.Project(files, "/r"); err == nil {
		t.Fatal("a fault must surface")
	}
}

var errBoom = errors.New("boom")

type failing struct{}

func (failing) ReadFile(string) ([]byte, error)    { return nil, errBoom }
func (failing) Stat(string) (discover.Info, error) { return discover.Info{}, errBoom }

type faultOn struct {
	*configtest.Files
	path string
}

func (f faultOn) Stat(path string) (discover.Info, error) {
	if path == f.path {
		return discover.Info{}, &iofs.PathError{Op: "stat", Path: path, Err: errBoom}
	}
	return f.Files.Stat(path)
}
