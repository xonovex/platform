package legacy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// The legacy places of the Fake host, whose home is /home/user and whose
// working directory is /work.
const (
	managedYAML      = "/etc/agent-cli/managed.yaml"
	managedProviders = "/etc/agent-cli/providers"
	userConfig       = "/home/user/.config/sandboxed-claude/config"
	userProviders    = "/home/user/.config/sandboxed-claude/providers"
	projectTOML      = "/work/.agent-cli.toml"
	projectProviders = "/work/.agent-cli/providers"
	ledgerAtHome     = "/home/user/.local/share/agent-cli/approvals.json"
)

const route = `{"name":"%s","agent":"claude","models":{"big":{"id":"big-id"}},"environment":{"BASE_URL":"https://models.example.test/v1"}}`

func provider(name string) string { return strings.ReplaceAll(route, "%s", name) }

var validator = func() jsonschema.Validator {
	v, err := wiring.NewValidator()
	if err != nil {
		panic(err)
	}
	return v
}()

func env(host runconfig.Host) runconfig.Env {
	return runconfig.Env{Host: host, Validator: validator, Out: &bytes.Buffer{}}
}

// everyLayer is a host with a legacy file and a provider file in each layer.
func everyLayer() *oshosttest.Fake {
	return oshosttest.New().
		AddFile(managedYAML, "network: none\nrequireEgressRestricted: true\n", 0o644).
		AddFile(managedProviders+"/corp.json", provider("corp"), 0o644).
		AddFile(userConfig, "# sandbox\nprovider=mine\nhomeDir=/sandbox\n", 0o600).
		AddFile(userProviders+"/mine.json", provider("mine"), 0o600).
		AddFile(projectTOML, "agent = \"claude\"\nmodels = [\"big=other\"]\n", 0o644).
		AddFile(projectProviders+"/proj.json", provider("proj"), 0o644)
}

func TestReadReturnsTheContentOfEveryLayer(t *testing.T) {
	content, err := Adapter{}.Read(env(everyLayer()), "/work")
	if err != nil {
		t.Fatal(err)
	}

	var paths []string
	for _, document := range content.Documents {
		paths = append(paths, string(document.Layer)+" "+document.Path)
	}
	wantPaths := []string{"managed " + managedYAML + " (legacy)", "user " + userConfig + " (legacy)", "project " + projectTOML + " (legacy)"}
	if !reflect.DeepEqual(paths, wantPaths) {
		t.Errorf("documents = %q, want %q", paths, wantPaths)
	}
	var section bytes.Buffer
	if err := json.Compact(&section, content.Documents[1].Sections["run"]); err != nil || section.String() != `{"homeDir":"/sandbox","provider":"mine"}` {
		got := section.String()
		t.Errorf("user run section = %s", got)
	}
	wantFiles := []runconfig.LegacyFile{
		{Layer: spec.Managed, Path: managedYAML}, {Layer: spec.Managed, Path: managedProviders},
		{Layer: spec.User, Path: userConfig}, {Layer: spec.User, Path: userProviders},
		{Layer: spec.Project, Path: projectTOML}, {Layer: spec.Project, Path: projectProviders},
	}
	if !reflect.DeepEqual(content.Files, wantFiles) {
		t.Errorf("files = %v, want %v", content.Files, wantFiles)
	}
	var origins []string
	for _, document := range content.Providers {
		origins = append(origins, document.Layer+" "+document.Origin)
	}
	wantOrigins := []string{
		"managed " + managedProviders + "/corp.json (legacy)",
		"user " + userProviders + "/mine.json (legacy)",
		"project " + projectProviders + "/proj.json (legacy)",
	}
	if !reflect.DeepEqual(origins, wantOrigins) {
		t.Errorf("providers = %q, want %q", origins, wantOrigins)
	}
	if len(content.Pending) != 1 {
		t.Fatalf("pending = %v, want the project material", content.Pending)
	}
	pending := content.Pending[0]
	if pending.Subject.Path != "/work (legacy)" || !strings.HasPrefix(pending.Subject.Digest, "sha256:") {
		t.Errorf("subject = %+v", pending.Subject)
	}
	for _, want := range []string{"Provider proj for claude", "Route hosts: models.example.test", "Model slots: big=big-id", "Model overrides: big=other"} {
		if !strings.Contains(pending.Subject.Review, want) {
			t.Errorf("review %q lacks %q", pending.Subject.Review, want)
		}
	}
	if !reflect.DeepEqual(pending.Origins, []string{projectProviders + "/proj.json (legacy)"}) {
		t.Errorf("origins = %q", pending.Origins)
	}
}

func TestReadOfAHostWithoutLegacyFilesIsEmpty(t *testing.T) {
	content, err := Adapter{}.Read(env(oshosttest.New()), "/work")
	if err != nil || !reflect.DeepEqual(content, runconfig.LegacyContent{}) {
		t.Errorf("Read() = %+v, %v, want nothing", content, err)
	}
}

func TestReadHonoursThePreviousLedger(t *testing.T) {
	host := oshosttest.New().AddFile(projectTOML, "models = [\"big=x\"]\n", 0o644)
	digest, err := material{dir: "/work", models: []string{"big=x"}}.digest()
	if err != nil {
		t.Fatal(err)
	}
	ledger := `{"version":1,"approved":[{"kind":"plugin","subject":"p","digest":"` + digest + `"},{"kind":"project-providers","subject":"/work","digest":"` + digest + `"}]}`

	for name, setup := range map[string]func(*oshosttest.Fake){
		"default data directory": func(h *oshosttest.Fake) { h.AddFile(ledgerAtHome, ledger, 0o600) },
		"XDG_DATA_HOME": func(h *oshosttest.Fake) {
			h.Env["XDG_DATA_HOME"] = "/data"
			h.AddFile("/data/agent-cli/approvals.json", ledger, 0o600)
		},
	} {
		t.Run(name, func(t *testing.T) {
			h := oshosttest.New().AddFile(projectTOML, "models = [\"big=x\"]\n", 0o644)
			setup(h)
			content, err := Adapter{}.Read(env(h), "/work")
			if err != nil || len(content.Pending) != 0 {
				t.Errorf("Read() pending = %v, %v, want the material approved", content.Pending, err)
			}
		})
	}

	other := `{"version":1,"approved":[{"kind":"project-providers","subject":"/work","digest":"sha256:00"}]}`
	host.AddFile(ledgerAtHome, other, 0o600)
	content, err := Adapter{}.Read(env(host), "/work")
	if err != nil || len(content.Pending) != 1 || content.Pending[0].Subject.Digest != digest {
		t.Errorf("Read() pending = %v, %v, want the material pending", content.Pending, err)
	}
}

func TestReadReportsEveryFault(t *testing.T) {
	homeless := oshosttest.New()
	homeless.HomeErr = errors.New("no home")
	noData := oshosttest.New().AddFile(projectTOML, "models = [\"big=x\"]\n", 0o644)
	noData.HomeErr = errors.New("no home")
	for name, c := range map[string]struct {
		host runconfig.Host
		want string
	}{
		"no home": {homeless, "find the home directory"},
		"writable managed file": {
			oshosttest.New().AddFile(managedYAML, "network: none\n", 0o664),
			"writable by its group or by others",
		},
		"unstattable file": {
			statFails{Fake: oshosttest.New(), path: projectTOML},
			"project configuration /work/.agent-cli.toml: stat failed",
		},
		"unreadable file": {
			readFails{Fake: oshosttest.New().AddFile(projectTOML, "", 0o644), path: projectTOML},
			"project configuration /work/.agent-cli.toml: read failed",
		},
		"unknown key": {
			oshosttest.New().AddFile(projectTOML, "colour = \"red\"\n", 0o644),
			"project configuration /work/.agent-cli.toml",
		},
		"value the schema refuses": {
			oshosttest.New().AddFile(userConfig, "network: everywhere\n", 0o600),
			userConfig + " (legacy)",
		},
		"broken provider file": {
			oshosttest.New().AddFile(userProviders+"/mine.json", "{", 0o600),
			"user providers",
		},
		"unreadable previous ledger": {
			readFails{Fake: oshosttest.New().AddFile(projectTOML, "models = [\"big=x\"]\n", 0o644).AddFile(ledgerAtHome, "", 0o600), path: ledgerAtHome},
			"read the legacy approval ledger",
		},
		"broken previous ledger": {
			oshosttest.New().AddFile(projectTOML, "models = [\"big=x\"]\n", 0o644).AddFile(ledgerAtHome, `{"version":2}`, 0o600),
			"is not a version 1 ledger",
		},
		"no data directory": {noData, "find the home directory"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Adapter{}.Read(env(c.host), "/work")
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("Read() error = %v, want %q", err, c.want)
			}
		})
	}
}

func TestReadChecksTheEnvelope(t *testing.T) {
	_, err := Adapter{}.Read(runconfig.Env{Host: oshosttest.New().AddFile(userConfig, "provider=mine\n", 0o600), Validator: refusing{}}, "/work")
	if err == nil || !strings.Contains(err.Error(), userConfig+" (legacy)") {
		t.Errorf("Read() error = %v, want the refusal to name the legacy file", err)
	}
}

func TestManagedDirFollowsThePlatform(t *testing.T) {
	for _, c := range []struct {
		goos, programData, want string
	}{
		{"linux", "", "/etc/agent-cli"},
		{"darwin", "", "/Library/Application Support/agent-cli"},
		{"windows", "/ProgramData", "/ProgramData/agent-cli"},
		{"windows", "", ""},
	} {
		host := oshosttest.New()
		host.OS = c.goos
		host.Env["ProgramData"] = c.programData
		if got := managedDir(host); got != c.want {
			t.Errorf("managedDir(%s, %q) = %q, want %q", c.goos, c.programData, got, c.want)
		}
	}
	host := oshosttest.New().AddFile(userConfig, "provider=mine\n", 0o600)
	host.OS = "windows"
	content, err := Adapter{}.Read(env(host), "/work")
	if err != nil || len(content.Documents) != 1 {
		t.Errorf("Read() on Windows without ProgramData = %+v, %v, want the user file only", content, err)
	}
}

func TestMaterialDigestCoversTheFilesAndTheModels(t *testing.T) {
	document := func(raw string) providers.Document {
		preset, err := providers.ParsePreset([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		return providers.Document{Preset: preset, Raw: []byte(raw)}
	}
	base := material{dir: "/work", documents: []providers.Document{document(provider("proj"))}, models: []string{"big=x"}}
	digest, err := base.digest()
	if err != nil || digest == "" {
		t.Fatalf("digest() = %q, %v", digest, err)
	}
	for name, changed := range map[string]material{
		"file":       {documents: []providers.Document{document(provider("proj") + "\n")}, models: base.models},
		"models":     {documents: base.documents, models: []string{"big=y"}},
		"no models":  {documents: base.documents},
		"no file":    {models: base.models},
		"other name": {documents: []providers.Document{document(provider("next"))}, models: base.models},
	} {
		if got, _ := changed.digest(); got == digest {
			t.Errorf("a changed %s kept the digest", name)
		}
	}
	escaping := material{dir: "/work", documents: []providers.Document{{Preset: providers.Preset{Name: "../x"}}}}
	if _, err := escaping.digest(); err == nil {
		t.Error("digest() of a name that leaves its directory = nil error")
	}
	if _, err := escaping.pending(oshosttest.New(), nil); err == nil {
		t.Error("pending() of a name that leaves its directory = nil error")
	}
	if review := (material{documents: []providers.Document{document(`{"name":"bare","agent":"claude"}`)}}).review(); !strings.Contains(review, "Route hosts: (none)") {
		t.Errorf("review() = %q", review)
	}
}

// statFails is a host whose Stat of one path fails.
type statFails struct {
	*oshosttest.Fake
	path string
}

func (s statFails) Stat(p string) (fs.FileInfo, error) {
	if p == s.path {
		return nil, errors.New("stat failed")
	}
	return s.Fake.Stat(p)
}

// readFails is a host whose ReadFile of one path fails.
type readFails struct {
	*oshosttest.Fake
	path string
}

func (r readFails) ReadFile(p string) ([]byte, error) {
	if p == r.path {
		return nil, errors.New("read failed")
	}
	return r.Fake.ReadFile(p)
}

// refusing is a validator that refuses every document.
type refusing struct{}

func (refusing) Validate(string, []byte) ([]jsonschema.Violation, error) {
	return []jsonschema.Violation{{Pointer: "/x", Message: "refused"}}, nil
}

func TestReadLeavesAProjectWithoutProvidersOrModelsUngated(t *testing.T) {
	host := oshosttest.New().AddFile(projectTOML, "agent = \"claude\"\n", 0o644)
	content, err := Adapter{}.Read(env(host), "/work")
	if err != nil || len(content.Documents) != 1 || len(content.Pending) != 0 {
		t.Errorf("Read() = %+v, %v, want one document and nothing pending", content, err)
	}
}

func TestPreviousLedgerNeedsAHomeWithoutXDGDataHome(t *testing.T) {
	host := oshosttest.New()
	host.HomeErr = errors.New("no home")
	if _, err := previouslyApproved(host, "sha256:00"); err == nil || !strings.Contains(err.Error(), "find the home directory") {
		t.Errorf("previouslyApproved() error = %v", err)
	}
}

func TestMigrateWritesEveryLayerItIsAskedFor(t *testing.T) {
	host := everyLayer()

	migrated, err := Adapter{}.Migrate(env(host), runconfig.Migration{ProjectDir: "/work", Managed: true})
	if err != nil {
		t.Fatal(err)
	}

	want := []runconfig.Migrated{
		{Path: "/etc/agentic/agentic.json", From: []string{managedYAML}},
		{Path: "/etc/agentic/providers.d/corp.json", From: []string{managedProviders + "/corp.json"}},
		{Path: "/home/user/.config/agentic/agentic.json", From: []string{userConfig}},
		{Path: "/home/user/.config/agentic/providers.d/mine.json", From: []string{userProviders + "/mine.json"}},
		{Path: "/work/agentic.json", From: []string{projectTOML}},
		{Path: "/work/providers.d/proj.json", From: []string{projectProviders + "/proj.json"}},
	}
	if !reflect.DeepEqual(migrated, want) {
		t.Errorf("Migrate() = %v, want %v", migrated, want)
	}
	if data, _ := host.File("/home/user/.config/agentic/agentic.json"); !strings.Contains(data, `"provider": "mine"`) {
		t.Errorf("user agentic.json = %s", data)
	}
	if data, _ := host.File("/work/providers.d/proj.json"); data != provider("proj") {
		t.Errorf("project provider shard = %s, want the legacy file unchanged", data)
	}
	if info, err := host.Stat("/etc/agentic/agentic.json"); err != nil || info.Mode().Perm() != 0o644 {
		t.Errorf("managed agentic.json mode = %v, %v, want 0644", info, err)
	}
}

func TestMigrateLeavesOutTheLayersItIsNotAskedFor(t *testing.T) {
	host := everyLayer()

	migrated, err := Adapter{}.Migrate(env(host), runconfig.Migration{})
	if err != nil {
		t.Fatal(err)
	}

	if len(migrated) != 2 || migrated[0].Path != "/home/user/.config/agentic/agentic.json" {
		t.Errorf("Migrate() = %v, want the user layer only", migrated)
	}
	if host.Exists("/etc/agentic/agentic.json") || host.Exists("/work/agentic.json") {
		t.Error("Migrate() wrote a layer it was not asked for")
	}
}

func TestMigrateNeverReplacesAFile(t *testing.T) {
	host := everyLayer().AddFile("/work/providers.d/proj.json", "{}", 0o644)

	_, err := Adapter{}.Migrate(env(host), runconfig.Migration{ProjectDir: "/work"})

	if err == nil || !strings.Contains(err.Error(), "/work/providers.d/proj.json exists") {
		t.Fatalf("Migrate() error = %v, want the existing file named", err)
	}
	if host.Exists("/home/user/.config/agentic/agentic.json") {
		t.Error("Migrate() wrote a file before it refused")
	}
}

func TestMigrateReportsEveryFault(t *testing.T) {
	noUserDir := everyLayer()
	noUserDir.HomeErr = errors.New("no home")
	for name, c := range map[string]struct {
		env  runconfig.Env
		want string
	}{
		"unreadable legacy file": {env(readFails{Fake: everyLayer(), path: userConfig}), "read failed"},
		"file the schema refuses": {
			runconfig.Env{Host: everyLayer(), Validator: refusing{}},
			"refused",
		},
		"provider the schema refuses": {
			runconfig.Env{Host: oshosttest.New().AddFile(userProviders+"/mine.json", provider("mine"), 0o600), Validator: refusing{}},
			"does not match the provider schema: /x refused",
		},
		"provider the validator cannot check": {
			runconfig.Env{Host: oshosttest.New().AddFile(userProviders+"/mine.json", provider("mine"), 0o600), Validator: broken{}},
			"check " + userProviders + "/mine.json: validator broken",
		},
		"no user directory": {
			env(&homeOnce{Fake: everyLayer()}),
			"find the user configuration directory",
		},
		"unstattable target": {env(statFails{Fake: everyLayer(), path: "/work/agentic.json"}), "check /work/agentic.json: stat failed"},
		"directory fails":    {env(mkdirFails{everyLayer()}), "create the directory of"},
		"write fails":        {env(writeFails{everyLayer()}), "write /home/user/.config/agentic/agentic.json"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Adapter{}.Migrate(c.env, runconfig.Migration{ProjectDir: "/work"})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("Migrate() error = %v, want %q", err, c.want)
			}
		})
	}
	if _, err := (Adapter{}).Migrate(env(noUserDir), runconfig.Migration{}); err == nil {
		t.Error("Migrate() without a home = nil error")
	}
}

// homeOnce is a host that finds the home directory once: the legacy places
// read it, and the agentic places that follow find none.
type homeOnce struct {
	*oshosttest.Fake
	calls int
}

func (h *homeOnce) UserHomeDir() (string, error) {
	h.calls++
	if h.calls > 1 {
		return "", errors.New("no home")
	}
	return h.Fake.UserHomeDir()
}

// mkdirFails is a host that creates no directory.
type mkdirFails struct{ *oshosttest.Fake }

func (mkdirFails) MkdirAll(string, fs.FileMode) error { return errors.New("mkdir failed") }

// writeFails is a host that writes no file.
type writeFails struct{ *oshosttest.Fake }

func (writeFails) WriteFile(string, []byte, fs.FileMode) error { return errors.New("write failed") }

// broken is a validator that cannot check a document.
type broken struct{}

func (broken) Validate(string, []byte) ([]jsonschema.Violation, error) {
	return nil, errors.New("validator broken")
}
