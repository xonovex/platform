package config

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

func TestLoadLayersReadsManagedUserAndProject(t *testing.T) {
	files := oshosttest.New().
		AddFile("/etc/agent-cli/managed.yaml", "isolation: docker\nrequireKernelIsolation: true\n", 0o644).
		AddFile(DefaultPath("/home/user"), "provider: glm\nbindPaths:\n  - /data\n", 0o600).
		AddFile("/work/repo/.agent-cli.toml", "agent = \"codex\"\n", 0o644)

	inputs, err := LoadLayers(files, "", "/work/repo")
	if err != nil {
		t.Fatalf("LoadLayers() error = %v", err)
	}
	if len(inputs) != 3 {
		t.Fatalf("LoadLayers() = %d layers, want managed, user and project", len(inputs))
	}
	got := map[Layer]string{}
	for _, in := range inputs {
		got[in.Layer] = in.Origin
	}
	want := map[Layer]string{
		LayerManaged: "/etc/agent-cli/managed.yaml",
		LayerUser:    DefaultPath("/home/user"),
		LayerProject: "/work/repo/.agent-cli.toml",
	}
	for layer, origin := range want {
		if got[layer] != origin {
			t.Errorf("%s origin = %q, want %q", layer, got[layer], origin)
		}
	}
}

func TestLoadLayersSkipsAbsentFiles(t *testing.T) {
	inputs, err := LoadLayers(oshosttest.New(), "", "/work")
	if err != nil || len(inputs) != 0 {
		t.Fatalf("LoadLayers() = (%v, %v), want no layer and no error", inputs, err)
	}
	files := oshosttest.New()
	files.ManagedDir = ""
	if inputs, err := LoadLayers(files, "", ""); err != nil || len(inputs) != 0 {
		t.Fatalf("LoadLayers(no dirs) = (%v, %v), want no layer", inputs, err)
	}
}

func TestLoadLayersTakesTheExplicitFileAsTheUserLayer(t *testing.T) {
	files := oshosttest.New().
		AddFile(DefaultPath("/home/user"), "provider: glm\n", 0o600).
		AddFile("/etc/mine.yaml", "provider: gemini\n", 0o600)

	inputs, err := LoadLayers(files, "/etc/mine.yaml", "/work")
	if err != nil || len(inputs) != 1 || inputs[0].Origin != "/etc/mine.yaml" || inputs[0].Entries[KeyProvider].Text != "gemini" {
		t.Fatalf("LoadLayers(explicit) = (%+v, %v), want only the named file as the user layer", inputs, err)
	}
	if _, err := LoadLayers(files, "/etc/missing.yaml", "/work"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("LoadLayers(missing explicit) error = %v, want a missing-file error", err)
	}
}

func TestLoadLayersRefusesAManagedFileOthersCanWrite(t *testing.T) {
	for _, mode := range []fs.FileMode{0o666, 0o664, 0o646} {
		files := oshosttest.New().AddFile("/etc/agent-cli/managed.yaml", "isolation: docker\n", mode)
		if _, err := LoadLayers(files, "", "/work"); err == nil || !strings.Contains(err.Error(), "writable") {
			t.Errorf("mode %o: LoadLayers() error = %v, want a refusal of a writable policy", mode, err)
		}
	}
	files := oshosttest.New().AddFile("/work/.agent-cli.yaml", "agent: codex\n", 0o666)
	if _, err := LoadLayers(files, "", "/work"); err != nil {
		t.Errorf("a writable project file is not guarded, got %v", err)
	}
}

func TestLoadLayersNamesTheFileOfAParseError(t *testing.T) {
	cases := map[string]*oshosttest.Fake{
		"managed": oshosttest.New().AddFile("/etc/agent-cli/managed.yaml", "colour: red\n", 0o644),
		"user":    oshosttest.New().AddFile(DefaultPath("/home/user"), "colour: red\n", 0o600),
		"project": oshosttest.New().AddFile("/work/.agent-cli.yml", "colour: red\n", 0o644),
	}
	for layer, files := range cases {
		if _, err := LoadLayers(files, "", "/work"); err == nil || !strings.Contains(err.Error(), layer) {
			t.Errorf("%s: LoadLayers() error = %v, want the layer named", layer, err)
		}
	}
}

func TestLoadLayersFailsWithoutAHome(t *testing.T) {
	files := oshosttest.New()
	files.HomeErr = errors.New("no home")
	if _, err := LoadLayers(files, "", "/work"); err == nil || !strings.Contains(err.Error(), "home directory") {
		t.Errorf("LoadLayers() error = %v, want a home lookup error", err)
	}
}

type failingFiles struct {
	*oshosttest.Fake
	statErr error
	readErr error
}

func (f failingFiles) Stat(path string) (fs.FileInfo, error) {
	if f.statErr != nil && strings.HasSuffix(path, ".agent-cli.yaml") {
		return nil, f.statErr
	}
	return f.Fake.Stat(path)
}

func (f failingFiles) ReadFile(path string) ([]byte, error) {
	if f.readErr != nil && strings.HasSuffix(path, ".agent-cli.yaml") {
		return nil, f.readErr
	}
	return f.Fake.ReadFile(path)
}

func TestLoadLayersReportsAnUnreadableFile(t *testing.T) {
	base := oshosttest.New().AddFile("/work/.agent-cli.yaml", "agent: codex\n", 0o644)
	if _, err := LoadLayers(failingFiles{Fake: base, statErr: errors.New("denied")}, "", "/work"); err == nil || !strings.Contains(err.Error(), "denied") {
		t.Errorf("stat failure: LoadLayers() error = %v, want it reported", err)
	}
	if _, err := LoadLayers(failingFiles{Fake: base, readErr: errors.New("io error")}, "", "/work"); err == nil || !strings.Contains(err.Error(), "io error") {
		t.Errorf("read failure: LoadLayers() error = %v, want it reported", err)
	}
}
