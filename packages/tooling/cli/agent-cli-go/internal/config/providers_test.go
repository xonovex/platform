package config

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

type providerFiles struct {
	files      fstest.MapFS
	home       string
	homeErr    error
	managedDir string
}

func (p providerFiles) ReadDir(dir string) ([]fs.DirEntry, error) {
	return fs.ReadDir(p.files, strings.TrimPrefix(dir, "/"))
}

func (p providerFiles) ReadFile(file string) ([]byte, error) {
	return fs.ReadFile(p.files, strings.TrimPrefix(file, "/"))
}
func (p providerFiles) UserHomeDir() (string, error) { return p.home, p.homeErr }
func (p providerFiles) ManagedConfigDir() string     { return p.managedDir }

func document(name string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(`{"name":"` + name + `","agent":"claude"}`)}
}

func TestProviderDirsSitBesideEachConfiguration(t *testing.T) {
	files := providerFiles{home: "/home/u", managedDir: "/etc/agent-cli"}

	dirs, err := ProviderDirs(files, "", "/work/repo")

	if err != nil {
		t.Fatalf("ProviderDirs() error = %v", err)
	}
	want := []ProviderDir{
		{LayerManaged, "/etc/agent-cli/providers"},
		{LayerUser, "/home/u/.config/sandboxed-claude/providers"},
		{LayerProject, "/work/repo/.agent-cli/providers"},
	}
	if len(dirs) != len(want) {
		t.Fatalf("dirs = %v, want %v", dirs, want)
	}
	for index := range want {
		if dirs[index] != want[index] {
			t.Errorf("dirs[%d] = %v, want %v", index, dirs[index], want[index])
		}
	}
}

func TestProviderDirsFollowAnExplicitUserFileAndSkipMissingLayers(t *testing.T) {
	dirs, err := ProviderDirs(providerFiles{homeErr: errors.New("no home")}, "/srv/cfg/agent.yaml", "")

	if err != nil || len(dirs) != 1 || dirs[0] != (ProviderDir{LayerUser, "/srv/cfg/providers"}) {
		t.Errorf("ProviderDirs() = %v, %v", dirs, err)
	}
	if _, err := ProviderDirs(providerFiles{homeErr: errors.New("no home")}, "", ""); err == nil || !strings.Contains(err.Error(), "no home") {
		t.Errorf("ProviderDirs() error = %v, want the home failure", err)
	}
}

func TestLoadProvidersOrdersLayersSoManagedWinsAndSplitsTheProjectPart(t *testing.T) {
	files := providerFiles{
		home: "/home/u", managedDir: "/etc/agent-cli",
		files: fstest.MapFS{
			"etc/agent-cli/providers/m.json":                   document("m"),
			"home/u/.config/sandboxed-claude/providers/u.json": document("u"),
			"work/repo/.agent-cli/providers/p.json":            document("p"),
			"work/repo/.agent-cli/providers/readme.txt":        {Data: []byte("x")},
		},
	}

	set, err := LoadProviders(files, "", "/work/repo")

	if err != nil {
		t.Fatalf("LoadProviders() error = %v", err)
	}
	var order []string
	for _, document := range set.Documents {
		order = append(order, document.Layer+":"+document.Preset.Name)
	}
	if strings.Join(order, " ") != "user:u project:p managed:m" {
		t.Errorf("order = %v, want user, project, managed", order)
	}
	if len(set.Project) != 1 || set.Project[0].Preset.Name != "p" || len(set.Dirs) != 3 {
		t.Errorf("Project = %v, Dirs = %v", set.Project, set.Dirs)
	}
}

func TestLoadProvidersNamesTheLayerThatFails(t *testing.T) {
	files := providerFiles{home: "/home/u", files: fstest.MapFS{
		"home/u/.config/sandboxed-claude/providers/bad.json": {Data: []byte("{")},
	}}
	if _, err := LoadProviders(files, "", ""); err == nil || !strings.Contains(err.Error(), "user providers") || !strings.Contains(err.Error(), "bad.json") {
		t.Errorf("LoadProviders() error = %v", err)
	}
	if _, err := LoadProviders(providerFiles{homeErr: errors.New("no home")}, "", ""); err == nil {
		t.Error("LoadProviders() ignored the home failure")
	}
}

func TestProjectModelsReadTheProjectLayerOnly(t *testing.T) {
	inputs := []Input{
		{Layer: LayerUser, Entries: Entries{KeyModels: Value{Items: []string{"a=1"}}}},
		{Layer: LayerProject, Entries: Entries{KeyModels: Value{Items: []string{"b=2"}}}},
	}
	if got := ProjectModels(inputs); len(got) != 1 || got[0] != "b=2" {
		t.Errorf("ProjectModels() = %v", got)
	}
	if got := ProjectModels(inputs[:1]); got != nil {
		t.Errorf("ProjectModels() without a project layer = %v", got)
	}
}

func TestModelsMergeInLayerOrderWithTheirSources(t *testing.T) {
	resolved, err := Merge([]Input{
		{Layer: LayerManaged, Origin: "m", Entries: Entries{KeyModels: Value{Items: []string{"a=managed"}}}},
		{Layer: LayerUser, Origin: "u", Entries: Entries{KeyModels: Value{Items: []string{"a=user"}}}},
		{Layer: LayerProject, Origin: "p", Entries: Entries{KeyModels: Value{Items: []string{"a=project"}}}},
		{Layer: LayerFlags, Origin: "f", Entries: Entries{KeyModels: Value{Items: []string{"a=flag"}}}},
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	var order []string
	for _, item := range resolved.Sourced(KeyModels) {
		order = append(order, item.Source.Origin+"="+item.Text)
	}
	if strings.Join(order, " ") != "u=a=user p=a=project f=a=flag m=a=managed" {
		t.Errorf("Sourced() = %v, want the managed value last", order)
	}
}
