package runconfig

import (
	"bytes"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

// The places of the Fake host, whose home is /home/user and whose working
// directory is /work.
const (
	userFile         = "/home/user/.config/agentic/agentic.json"
	userProviders    = "/home/user/.config/agentic/providers.d"
	managedFile      = "/etc/agentic/agentic.json"
	managedProviders = "/etc/agentic/providers.d"
	projectFile      = "/work/agentic.json"
	projectProviders = "/work/providers.d"
	ledgerFile       = "/home/user/.local/state/agentic/trust.json"
)

var schemas = func() jsonschema.Validator {
	validator, err := wiring.NewValidator()
	if err != nil {
		panic(err)
	}
	return validator
}()

// anything is a validator that accepts every document, so a test reaches the
// checks agent-cli makes after the schema.
type anything struct{}

func (anything) Validate(string, []byte) ([]jsonschema.Violation, error) { return nil, nil }

func env(host *oshosttest.Fake) (Env, *bytes.Buffer) {
	var out bytes.Buffer
	return Env{Host: host, Validator: schemas, Out: &out}, &out
}

func document(t *testing.T, l spec.Layer, path, data string) layer.Document {
	t.Helper()
	parsed, err := layer.Parse(l, path, []byte(data))
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestKeysAreASetAndBecomeCommandLineValues(t *testing.T) {
	keys := KeysOf(KeyAgent, KeyNetwork)
	if !keys.Has(KeyAgent) || keys.Has(KeyProvider) {
		t.Errorf("KeysOf() = %v", keys)
	}
	flags := []merge.Flag{Text(KeyAgent, "codex"), List(KeyBindPaths, []string{"/a"}), Demand(KeyRequireEgressRestricted)}
	want := []string{`/agent="codex"`, `/bindPaths=["/a"]`, `/requireEgressRestricted=true`}
	for i, flag := range flags {
		if got := flag.Pointer + "=" + string(flag.Value); flag.Section != "run" || got != want[i] {
			t.Errorf("flag %d = %s %s, want %s", i, flag.Section, got, want[i])
		}
	}
}

func TestMergeNamesTheSourceOfEachValue(t *testing.T) {
	documents := []layer.Document{
		document(t, spec.Managed, managedFile, `{"version":1,"run":{"network":"none","bindPaths":["/m"],"requireEgressRestricted":true}}`),
		document(t, spec.User, userFile, `{"version":1,"run":{"agent":"claude","bindPaths":["/u","/m"],"requireEgressRestricted":true}}`),
		document(t, spec.Project, projectFile, `{"version":1,"run":{"agent":"codex"}}`),
		document(t, spec.Project, "/work/other.json", `{"version":1}`),
	}

	resolved, err := Merge(documents, []merge.Flag{List(KeyBindPaths, []string{"/f"})})

	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if resolved.Text(KeyAgent) != "codex" || resolved.Text(KeyProvider) != "" || !resolved.On(KeyRequireEgressRestricted) || resolved.On(KeyRequireKernelIsolation) {
		t.Errorf("values: agent %q, provider %q", resolved.Text(KeyAgent), resolved.Text(KeyProvider))
	}
	if !resolved.Has(KeyAgent) || resolved.Has(KeyProvider) || resolved.Has(KeyRequireKernelIsolation) || !resolved.Has(KeyRequireEgressRestricted) {
		t.Error("Has() does not follow the layers")
	}
	if got := resolved.Items(KeyBindPaths); !reflect.DeepEqual(got, []string{"/u", "/m", "/f"}) {
		t.Errorf("Items(bindPaths) = %q, want each path once, the weakest source first", got)
	}
	if resolved.Items(KeyAllowHosts) != nil || resolved.Sourced(KeyAllowHosts) != nil {
		t.Error("an unset Append key has items")
	}
	if got := resolved.Sourced(KeyBindPaths)[1].Source.String(); got != "user "+userFile {
		t.Errorf("the source of /m = %q, want the user file that gave it first", got)
	}

	var sources []string
	for _, source := range resolved.Sources() {
		sources = append(sources, source.String())
	}
	if want := []string{"user " + userFile, "project " + projectFile, "flags command line", "managed " + managedFile}; !reflect.DeepEqual(sources, want) {
		t.Errorf("Sources() = %q, want %q", sources, want)
	}

	want := []string{
		"agent: codex (project " + projectFile + ")",
		"network: none (managed " + managedFile + ")",
		"bindPaths: /u (user " + userFile + ")",
		"bindPaths: /m (user " + userFile + ")",
		"bindPaths: /f (flags command line)",
		"requireEgressRestricted: required (managed " + managedFile + ")",
	}
	if got := resolved.Provenance(); !reflect.DeepEqual(got, want) {
		t.Errorf("Provenance() =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	layers := map[Key][]spec.Layer{
		KeyAgent: {spec.Project}, KeyNetwork: {spec.Managed}, KeyBindPaths: {spec.User, spec.Flags},
		KeyRequireEgressRestricted: {spec.Managed},
	}
	if got := resolved.Layers(); !reflect.DeepEqual(got, layers) {
		t.Errorf("Layers() = %v, want %v", got, layers)
	}
	if resolved.Merged().Section("run") == nil {
		t.Error("Merged() holds no run section")
	}
}

func TestARequireKeyNobodyDemandsIsNotReported(t *testing.T) {
	resolved, err := Merge([]layer.Document{document(t, spec.User, userFile, `{"version":1,"run":{"requirePinnedProvision":false}}`)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Has(KeyRequirePinnedProvision) || len(resolved.Provenance()) != 0 || len(resolved.Layers()) != 0 {
		t.Errorf("Provenance() = %q, Layers() = %v, want nothing", resolved.Provenance(), resolved.Layers())
	}
}

func TestMergeRefusesACommandLineThatContradictsTheManagedFile(t *testing.T) {
	documents := []layer.Document{document(t, spec.Managed, managedFile, `{"version":1,"run":{"network":"none"}}`)}
	if _, err := Merge(documents, []merge.Flag{Text(KeyNetwork, "host")}); err == nil || !strings.Contains(err.Error(), "but the managed file") {
		t.Errorf("Merge() error = %v", err)
	}
}

func TestReadFindsTheProvidersOfEveryLayer(t *testing.T) {
	host := oshosttest.New()
	host.AddFile(userFile, `{"version":1,"providers":{"inline":{"agent":"claude"}}}`, 0o600)
	host.AddFile(userProviders+"/alpha.json", `{"name":"alpha","agent":"claude"}`, 0o644)
	host.AddFile(managedFile, `{"version":1,"run":{"network":"none"}}`, 0o644)
	host.AddFile(managedProviders+"/alpha.json", `{"name":"alpha","agent":"claude","portable":true}`, 0o644)
	project := `{"version":1,"run":{"agent":"claude"}}`
	host.AddFile(projectFile, project, 0o644)
	host.AddFile(projectProviders+"/beta.json", `{"name":"beta","agent":"codex"}`, 0o644)
	e, _ := env(host)

	loaded, err := Read(e, Request{Flags: []merge.Flag{Text(KeyProvider, "alpha")}})

	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	var names []string
	for _, provider := range loaded.Providers {
		names = append(names, provider.Preset.Name+" "+provider.Layer+" "+provider.Origin)
	}
	want := []string{
		"inline user " + userFile, "alpha user " + userProviders + "/alpha.json",
		"beta project " + projectProviders + "/beta.json", "alpha managed " + managedProviders + "/alpha.json",
	}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("providers = %q, want %q", names, want)
	}
	dirs := []ProviderDir{{spec.Managed, managedProviders}, {spec.User, userProviders}, {spec.Project, projectProviders}}
	if !reflect.DeepEqual(loaded.ProviderDirs, dirs) || loaded.UserProviderDir() != userProviders {
		t.Errorf("ProviderDirs = %v", loaded.ProviderDirs)
	}
	if len(loaded.Pending) != 2 || loaded.Pending[0].Subject.Digest != layer.Digest([]byte(project)) || loaded.Pending[1].Origins[0] != projectProviders+"/beta.json" {
		t.Errorf("Pending = %+v, want the project file and its shard", loaded.Pending)
	}
	if loaded.Resolved.Text(KeyProvider) != "alpha" || loaded.Resolved.Text(KeyNetwork) != "none" {
		t.Error("the run section was not merged")
	}
	if (Loaded{}).UserProviderDir() != "" {
		t.Error("an empty configuration names a user provider directory")
	}
}

func TestReadAppliesTheHarnessOverridesOfEveryLayer(t *testing.T) {
	host := oshosttest.New()
	host.AddFile("/home/user/.config/agentic/harnesses.d/claude.json", `{"binary":"claude-user"}`, 0o644)
	host.AddFile("/home/user/.config/agentic/harnesses.d/notes.json", `{"binary":"ignored"}`, 0o644)
	host.AddFile("/etc/agentic/harnesses.d/codex.json", `{"binary":"codex-managed"}`, 0o644)
	host.AddFile(projectFile, `{"version":1}`, 0o644)
	project := `{"binary":"claude-project"}`
	host.AddFile("/work/harnesses.d/claude.json", project, 0o644)
	e, _ := env(host)

	loaded, err := Read(e, Request{})

	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	claude, _ := loaded.Harnesses.Get(types.AgentClaude)
	codex, _ := loaded.Harnesses.Get(types.AgentCodex)
	if claude.Binary != "claude-project" || codex.Binary != "codex-managed" {
		t.Errorf("binaries = %q, %q; want the project and the managed override", claude.Binary, codex.Binary)
	}
	if len(loaded.Pending) != 1 || loaded.Pending[0].Subject.Digest != layer.Digest([]byte(project)) || !strings.Contains(loaded.Pending[0].Subject.Notice, "harness") {
		t.Errorf("Pending = %+v, want the project override", loaded.Pending)
	}
}

func TestAProjectFileThatSetsNoGatedSectionWaitsForNothing(t *testing.T) {
	host := oshosttest.New()
	host.AddFile(projectFile, `{"version":1}`, 0o644)
	e, _ := env(host)
	loaded, err := Load(e, Request{})
	if err != nil || len(loaded.Pending) != 0 || len(loaded.ProviderDirs) != 3 {
		t.Errorf("Load() = %+v, %v", loaded, err)
	}
}

func TestReadFailsClosed(t *testing.T) {
	cases := map[string]struct {
		files     map[string]string
		validator jsonschema.Validator
		homeless  bool
		want      string
	}{
		"no home":           {homeless: true, want: "find the user configuration directory"},
		"invalid file":      {files: map[string]string{userFile: `{"version":1,"run":{"agnet":"x"}}`}, want: userFile},
		"widening project":  {files: map[string]string{projectFile: `{"version":1,"run":{"bindPaths":["/"]}}`}, want: "only the managed and user files may set"},
		"shard directory":   {files: map[string]string{userProviders: "a file"}, want: userProviders},
		"broken shard":      {files: map[string]string{userProviders + "/a.json": "{"}, want: "provider file " + userProviders + "/a.json"},
		"misnamed shard":    {files: map[string]string{userProviders + "/a.json": `{"name":"b","agent":"claude"}`}, want: "name the file b.json"},
		"defined twice":     {files: map[string]string{userFile: `{"version":1,"providers":{"a":{"agent":"claude"}}}`, userProviders + "/a.json": `{"name":"a","agent":"claude"}`}, want: "is defined in " + userFile + " and in " + userProviders + "/a.json"},
		"renamed inline":    {files: map[string]string{userFile: `{"version":1,"providers":{"a":{"name":"b","agent":"claude"}}}`}, want: `the provider a names itself "b"`},
		"broken inline":     {files: map[string]string{userFile: `{"version":1,"providers":{"a":{"agent":"Claude"}}}`}, validator: anything{}, want: "the provider a:"},
		"not a map":         {files: map[string]string{userFile: `{"version":1,"providers":{"a":1}}`}, validator: anything{}, want: userFile + ": the providers section"},
		"missing project":   {want: "project configuration /missing.json"},
		"harness directory": {files: map[string]string{"/home/user/.config/agentic/harnesses.d": "a file"}, want: "harnesses.d"},
		"broken override":   {files: map[string]string{"/etc/agentic/harnesses.d/claude.json": `{"binary":1}`}, want: "harness override /etc/agentic/harnesses.d/claude.json"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			host := oshosttest.New()
			for path, data := range tc.files {
				host.AddFile(path, data, 0o644)
			}
			if tc.homeless {
				host.HomeErr = errors.New("no home")
			}
			e, _ := env(host)
			if tc.validator != nil {
				e.Validator = tc.validator
			}
			request := Request{}
			if name == "missing project" {
				request.Project = "/missing.json"
			}
			if _, err := Load(e, request); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("Load() error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestGateApprovesEachPendingFileByItsDigest(t *testing.T) {
	host := oshosttest.New()
	project := `{"version":1,"run":{"agent":"claude"}}`
	shard := `{"name":"beta","agent":"codex"}`
	host.AddFile(projectFile, project, 0o644)
	host.AddFile(projectProviders+"/beta.json", shard, 0o644)
	e, _ := env(host)

	if _, err := Load(e, Request{ApproveDigest: layer.Digest([]byte(project))}); !errors.Is(err, trust.ErrNotApproved) || !strings.Contains(err.Error(), layer.Digest([]byte(shard))) {
		t.Fatalf("Load() error = %v, want the shard to wait", err)
	}
	if _, err := Load(e, Request{ApproveDigest: layer.Digest([]byte(shard))}); err != nil {
		t.Fatalf("Load() error = %v, want both approved", err)
	}
	if _, err := Load(e, Request{ApproveDigest: "sha256:other"}); err == nil || !strings.Contains(err.Error(), "--approve-config-digest sha256:other is the digest of no project file") {
		t.Errorf("Load() error = %v, want the stale digest refused", err)
	}

	loaded, err := Read(e, Request{})
	if err != nil {
		t.Fatal(err)
	}
	if yes, err := Approved(e, loaded.Pending[0]); !yes || err != nil {
		t.Errorf("Approved() = %v, %v", yes, err)
	}
	revoked, err := Revoke(e, loaded.Pending)
	if err != nil || len(revoked) != 2 {
		t.Fatalf("Revoke() = %v, %v", revoked, err)
	}
	if revoked, err := Revoke(e, loaded.Pending); err != nil || len(revoked) != 0 {
		t.Errorf("a second Revoke() = %v, %v", revoked, err)
	}
	host.AddFile(ledgerFile, "{", 0o600)
	if _, err := Revoke(e, loaded.Pending); err == nil {
		t.Error("Revoke() read a corrupt ledger")
	}
}

// failing is a host whose file writes fail at one step.
type failing struct {
	*oshosttest.Fake
	step string
}

func (f failing) MkdirAll(p string, perm fs.FileMode) error {
	if f.step == "mkdir" {
		return errors.New("mkdir failed")
	}
	return f.Fake.MkdirAll(p, perm)
}

func (f failing) WriteFile(p string, data []byte, perm fs.FileMode) error {
	if f.step == "write" {
		return errors.New("write failed")
	}
	return f.Fake.WriteFile(p, data, perm)
}

func (f failing) Rename(oldPath, newPath string) error {
	if f.step == "rename" {
		return errors.New("rename failed")
	}
	return f.Fake.Rename(oldPath, newPath)
}

func TestFilesWriteThroughARenameAndReportEachFailure(t *testing.T) {
	host := oshosttest.New()
	if err := (Files{Host: host}).WriteFile("/state/trust.json", []byte("{}")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if data, _ := host.File("/state/trust.json"); data != "{}" || host.Exists("/state/trust.json.tmp") {
		t.Errorf("the file = %q, want the content and no staged file", data)
	}
	for _, step := range []string{"mkdir", "write", "rename"} {
		if err := (Files{Host: failing{Fake: oshosttest.New(), step: step}}).WriteFile("/state/trust.json", nil); err == nil || !strings.Contains(err.Error(), step) {
			t.Errorf("WriteFile() with a failing %s = %v", step, err)
		}
	}
	if _, err := (Files{Host: host}).Stat("/missing"); err == nil {
		t.Error("Stat() found a missing file")
	}
	if _, err := (Files{Host: host}).ReadDir("/missing"); err == nil {
		t.Error("ReadDir() listed a missing directory")
	}
}

func TestPlacesFollowThePlatform(t *testing.T) {
	host := oshosttest.New()
	if got := (Source{Layer: spec.Default}).String(); got != "default" {
		t.Errorf("String() = %q, want the layer alone", got)
	}
	host.OS = "windows"
	if dir := (Places{Host: host}).ManagedDir(); dir != "" {
		t.Errorf("ManagedDir() = %q, want none without ProgramData", dir)
	}
	host.OS = "linux"
	if dir, err := (Places{Host: host}).StateDir(); err != nil || dir != "/home/user/.local/state/agentic" {
		t.Errorf("StateDir() = %q, %v", dir, err)
	}
}

func TestReadSearchesTheProjectFromAnAbsoluteStart(t *testing.T) {
	host := oshosttest.New()
	host.AddFile("/repo/agentic.json", `{"version":1,"run":{"agent":"codex"}}`, 0o644)
	e, _ := env(host)

	loaded, err := Read(e, Request{Start: "/repo"})

	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got := loaded.ProviderDirs[len(loaded.ProviderDirs)-1]; got != (ProviderDir{Layer: spec.Project, Path: "/repo/providers.d"}) {
		t.Errorf("project provider directory = %+v, want the one beside /repo/agentic.json", got)
	}
}

func TestReadFailsWithoutAWorkingDirectory(t *testing.T) {
	e := Env{Host: wdFails{oshosttest.New()}, Validator: schemas, Out: &bytes.Buffer{}}
	if _, err := Read(e, Request{}); err == nil || !strings.Contains(err.Error(), "find the working directory") {
		t.Errorf("Read() error = %v, want the working directory fault", err)
	}
}

func TestPlacesNameTheWorkingDirectory(t *testing.T) {
	if wd, err := (Places{Host: oshosttest.New()}).Getwd(); err != nil || wd != "/work" {
		t.Errorf("Getwd() = %q, %v", wd, err)
	}
}

// wdFails is a host without a working directory.
type wdFails struct{ *oshosttest.Fake }

func (wdFails) Getwd() (string, error) { return "", errors.New("no working directory") }

func TestPathNamesTheKeyInsideTheRunSection(t *testing.T) {
	if got := KeyImage.Path(); got != "run.image" {
		t.Errorf("Path() = %q, want run.image", got)
	}
}
