package archtest_test

import (
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/archtest"
)

func file(source string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(source)} }

func syntax(t *testing.T, source string) archtest.File {
	t.Helper()
	files, err := archtest.Load(fstest.MapFS{"pkg/a/a.go": file(source)}, "pkg")
	if err != nil {
		t.Fatal(err)
	}
	return files[0]
}

func TestLoadReadsEveryGoFileUnderTheRoots(t *testing.T) {
	tree := fstest.MapFS{
		"pkg/a/a.go":               file("package a\n\nimport (\n\t\"os\"\n\t\"example.com/b\"\n)\n"),
		"pkg/a/a_test.go":          file("package a\n"),
		"pkg/a/README.md":          file("text"),
		"pkg/a/testdata/broken.go": file("not go"),
		"pkg/.cache/broken.go":     file("not go"),
		"pkg/_old/broken.go":       file("not go"),
		"cmd/main/main.go":         file("package main\n"),
		"other/skipped.go":         file("package other\n"),
	}
	files, err := archtest.Load(tree, "pkg", "cmd")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	if !slices.Equal(paths, []string{"pkg/a/a.go", "pkg/a/a_test.go", "cmd/main/main.go"}) {
		t.Fatal(paths)
	}
	first := files[0]
	if first.Dir != "pkg/a" || first.Test || !slices.Equal(first.Imports, []string{"os", "example.com/b"}) {
		t.Fatal(first.Dir, first.Test, first.Imports)
	}
	if !files[1].Test {
		t.Fatal("a _test.go file is a test")
	}
	if production := archtest.Production(files); len(production) != 2 || production[1].Path != "cmd/main/main.go" {
		t.Fatal(production)
	}
	whole, err := archtest.Load(fstest.MapFS{"main.go": file("package main\n"), ".git/hook.go": file("not go")}, ".")
	if err != nil || len(whole) != 1 || whole[0].Dir != "." {
		t.Fatal(whole, err)
	}
}

type unreadable struct{ fstest.MapFS }

func (unreadable) ReadFile(string) ([]byte, error) { return nil, errors.New("unreadable") }

func TestLoadFailsForAMissingRootAnUnreadableFileAndBadSource(t *testing.T) {
	if _, err := archtest.Load(fstest.MapFS{}, "pkg"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	if _, err := archtest.Load(unreadable{fstest.MapFS{"pkg/a.go": file("package a\n")}}, "pkg"); err == nil {
		t.Fatal("an unreadable file must fail")
	}
	if _, err := archtest.Load(fstest.MapFS{"pkg/a.go": file("not go")}, "pkg"); err == nil {
		t.Fatal("a file that does not parse must fail")
	}
}

func TestBuildsOnlyWithNamesTheTierOfAFile(t *testing.T) {
	cases := []struct {
		source string
		want   bool
	}{
		{"//go:build integration\n\npackage a\n", true},
		{"//go:build e2e || e2e_kata\n\npackage a\n", true},
		{"//go:build integration && linux\n\npackage a\n", false},
		{"//go:build !integration\n\npackage a\n", false},
		{"//go:build arch\n\npackage a\n", false},
		{"// Package a.\npackage a\n", false},
		{"//go:build integration (\n\npackage a\n", false},
		{"package a\n\n//go:build integration\n", false},
	}
	for _, c := range cases {
		if got := syntax(t, c.source).BuildsOnlyWith(tierOf(c.source)); got != c.want {
			t.Errorf("%q: BuildsOnlyWith = %v, want %v", c.source, got, c.want)
		}
	}
}

// tierOf names the tag a case asks about: e2e for the end-to-end case.
func tierOf(source string) string {
	if strings.Contains(source, "e2e") {
		return "e2e"
	}
	return "integration"
}

func TestIsStandardAndIsHostClassifyImportPaths(t *testing.T) {
	for _, path := range []string{"fmt", "go/ast", "net/netip"} {
		if !archtest.IsStandard(path) {
			t.Error(path, "is standard")
		}
	}
	if archtest.IsStandard("github.com/spf13/cobra") {
		t.Error("a module path is not standard")
	}
	for _, path := range []string{"os", "os/exec", "os/user", "os/signal", "io/ioutil", "net", "net/http", "syscall"} {
		if !archtest.IsHost(path) {
			t.Error(path, "reaches the host")
		}
	}
	for _, path := range []string{"io/fs", "net/netip", "net/url", "path/filepath"} {
		if archtest.IsHost(path) {
			t.Error(path, "does not reach the host")
		}
	}
}

func parsed(t *testing.T, source string) *archtest.File {
	t.Helper()
	f := syntax(t, source)
	return &f
}

func TestReadsClockFindsNowSinceAndUntil(t *testing.T) {
	for _, call := range []string{"time.Now()", "time.Since(t0)", "time.Until(t0)"} {
		if !archtest.ReadsClock(parsed(t, "package a\n\nimport \"time\"\n\nvar t0 time.Time\n\nfunc f() { _ = "+call+" }\n").Syntax) {
			t.Error(call, "reads the clock")
		}
	}
	if archtest.ReadsClock(parsed(t, "package a\n\nimport \"time\"\n\nfunc f(d time.Duration) { _ = d.Round(d).Seconds() }\n").Syntax) {
		t.Error("a duration reads no clock")
	}
}

func TestPrintsFindsTheStandardStreams(t *testing.T) {
	source := `package a

import (
	"fmt"
	"io"
	"os"
)

func f(w io.Writer) {
	fmt.Println("x")
	fmt.Printf("%d", 1)
	fmt.Print("x")
	fmt.Fprintln(w, "kept")
	println("x")
	print("x")
	_ = os.Stdout
	fmt.Fprintln(os.Stderr, "x")
	_ = fmt.Sprint("kept")
	_ = os.Getenv("kept")
}
`
	got := archtest.Prints(parsed(t, source).Syntax)
	want := []string{"fmt.Println", "fmt.Printf", "fmt.Print", "println", "print", "os.Stdout", "os.Stderr"}
	if !slices.Equal(got, want) {
		t.Fatal(got)
	}
}

func TestPackageStateFindsInitAndReassignableVariables(t *testing.T) {
	source := `package a

import (
	"embed"
	"errors"
	"io/fs"
	"regexp"
	"strings"
	"sync"
)

//go:embed *.json
var files embed.FS

var (
	//go:embed schema.json
	schema []byte

	table   = map[string]int{"a": 1}
	limit   = 3
	pointer = &struct{}{}
	pattern = regexp.MustCompile("a")
	errBad  = errors.New("bad")
	errGone = fs.ErrNotExist
	catalog = sync.OnceValue(func() int { return 1 })
	_       = strings.ToUpper
)

var counter int
var now = strings.ToLower
var lower = (strings.ToLower)
var joined = strings.Join(nil, "")
var a, b = 1, counter
var named = limit

type thing struct{}

func (thing) init() {}

func init() {}
`
	got := archtest.PackageState(parsed(t, source).Syntax)
	want := []string{"var counter", "var now", "var lower", "var joined", "var a", "var b", "var named", "func init"}
	if !slices.Equal(got, want) {
		t.Fatal(got)
	}
}

func TestPackageStateAcceptsConstantsTypesAndFunctions(t *testing.T) {
	source := "package a\n\nimport \"fmt\"\n\nconst limit = 3\n\ntype name string\n\nfunc f() string { return fmt.Sprint(limit) }\n"
	if got := archtest.PackageState(parsed(t, source).Syntax); len(got) != 0 {
		t.Fatal(got)
	}
}
