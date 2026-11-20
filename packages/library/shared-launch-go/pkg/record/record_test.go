package record

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema/tekuri"
)

var started = time.Date(2026, time.November, 20, 12, 30, 5, 0, time.UTC)

func exited(code int) Record {
	return Record{
		StartedAt: started,
		EndedAt:   started.Add(2 * time.Second),
		Status:    StatusExited,
		ExitCode:  &code,
		Agent:     Agent{Name: "claude", VersionProbe: "claude --version", Provider: "glm"},
		Axes: Axes{
			Isolation: "bwrap", Provision: "nix", Network: "managed",
			Workspace: Workspace{Mode: ModeCheckout, VCS: "git", Branch: "feature"}, Terminal: "tmux",
		},
		Guarantees:  []Guarantee{{Name: "pinned-provision", Required: true, Provided: true}},
		Overlay:     Overlay{Plugins: []Digest{{Name: "plugin-core", Digest: "sha256:" + strings.Repeat("a", 64)}}},
		Credentials: []string{"TOKEN"},
		WorkDir:     "/work/repo",
		Config:      Config{Sources: []Source{{Layer: "user", Path: "/home/user/.config/sandboxed-claude/config"}, {Layer: "flags"}}, Decided: map[string][]string{"isolation": {"flags"}}},
	}
}

// validateRecord checks data against the run record schema and returns the
// violations as one error, or nil.
func validateRecord(t *testing.T, data []byte) error {
	t.Helper()
	validator, err := tekuri.New(map[string][]byte{SchemaID: Schema()})
	if err != nil {
		t.Fatalf("compile the run record schema: %v", err)
	}
	violations, err := validator.Validate(SchemaID, data)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(violations) > 0 {
		return fmt.Errorf("%+v", violations)
	}
	return nil
}

func TestMarshalProducesADocumentThatSatisfiesTheSchema(t *testing.T) {
	failed := Record{StartedAt: started, EndedAt: started, Status: StatusFailed, Error: "isolation bwrap is not available",
		Agent: Agent{Name: "claude", VersionProbe: "claude --version"}, Axes: Axes{Isolation: "bwrap", Provision: "none", Network: "host", Workspace: Workspace{Mode: ModeInPlace}}, WorkDir: "/work"}
	for name, r := range map[string]Record{"exited": exited(0), "failed to start": failed} {
		r.ID = "20261120T123005Z"
		data, err := Marshal(r)
		if err != nil {
			t.Fatalf("%s: Marshal() error = %v", name, err)
		}
		if err := validateRecord(t, data); err != nil {
			t.Errorf("%s: the record violates the schema: %v\n%s", name, err, data)
		}
		if !strings.HasSuffix(string(data), "}\n") {
			t.Errorf("%s: the record must end with a newline", name)
		}
	}
}

func TestSchemaRejectsABrokenRecord(t *testing.T) {
	good := exited(0)
	good.ID = "20261120T123005Z"
	data, err := Marshal(good)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("the record is not JSON: %v", err)
	}
	broken := map[string]func(map[string]any){
		"unknown member":          func(d map[string]any) { d["secret"] = "x" },
		"bad id":                  func(d map[string]any) { d["id"] = "../x" },
		"bad status":              func(d map[string]any) { d["status"] = "crashed" },
		"exited without a code":   func(d map[string]any) { delete(d, "exitCode") },
		"failed with a code":      func(d map[string]any) { d["status"] = "failed-to-start"; d["error"] = "x" },
		"failed without an error": func(d map[string]any) { d["status"] = "failed-to-start"; delete(d, "exitCode") },
		"a digest that is no digest": func(d map[string]any) {
			d["overlay"].(map[string]any)["plugins"] = []any{map[string]any{"name": "p", "digest": "abc"}}
		},
		"credential as an object": func(d map[string]any) { d["credentials"] = []any{map[string]any{"TOKEN": "value"}} },
		"bad layer": func(d map[string]any) {
			d["config"].(map[string]any)["decided"] = map[string]any{"agent": []any{"cloud"}}
		},
		"missing agent": func(d map[string]any) { delete(d, "agent") },
	}
	for name, mutate := range broken {
		var copyOf map[string]any
		if err := json.Unmarshal(data, &copyOf); err != nil {
			t.Fatalf("copy: %v", err)
		}
		mutate(copyOf)
		mutated, _ := json.Marshal(copyOf)
		if err := validateRecord(t, mutated); err == nil {
			t.Errorf("%s: the schema accepted the document", name)
		}
	}
}

func TestSchemaReturnsACopy(t *testing.T) {
	first := Schema()
	first[0] = 'X'
	if Schema()[0] == 'X' {
		t.Error("Schema() must return a copy of the embedded document")
	}
}

func TestUnmarshalRoundTripsAndRefusesAnotherVersion(t *testing.T) {
	r := exited(3)
	r.ID = "20261120T123005Z"
	data, err := Marshal(r)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	want := r.normalize()
	want.Overlay.Directories = []Digest{}
	if !reflect.DeepEqual(got.Agent, want.Agent) || got.ID != want.ID || *got.ExitCode != 3 || got.Status != StatusExited {
		t.Errorf("round trip = %+v, want %+v", got, want)
	}
	if _, err := Unmarshal([]byte(`{"schemaVersion": 2}`)); err == nil || !strings.Contains(err.Error(), "version 2") {
		t.Errorf("Unmarshal(version 2) error = %v, want a version refusal", err)
	}
	if _, err := Unmarshal([]byte(`not json`)); err == nil {
		t.Error("Unmarshal(garbage) error = nil, want a decode error")
	}
}

func TestMarshalNeverEmitsNullCollections(t *testing.T) {
	data, err := Marshal(Record{StartedAt: started, EndedAt: started, Status: StatusFailed, Error: "x"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(data), "null") {
		t.Errorf("the record has a null member:\n%s", data)
	}
}

func TestDefaultDirFollowsTheXDGStateDirectory(t *testing.T) {
	env := map[string]string{"XDG_STATE_HOME": "/state"}
	getenv := func(key string) string { return env[key] }
	if got := DefaultDir(getenv, "/home/user", "agent-cli"); got != "/state/agent-cli/runs" {
		t.Errorf("DefaultDir = %q, want the XDG state directory", got)
	}
	env["XDG_STATE_HOME"] = "relative"
	if got := DefaultDir(getenv, "/home/user", "agent-cli"); got != "/home/user/.local/state/agent-cli/runs" {
		t.Errorf("DefaultDir = %q, want the default below the home directory", got)
	}
}

func newStore(t *testing.T) (Store, *memFiles) {
	t.Helper()
	files := newMemFiles()
	store, err := NewStore(files, "/state/agent-cli/runs")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	return store, files
}

func TestStoreWritesPrivateRecordsWithDistinctIDs(t *testing.T) {
	store, files := newStore(t)
	first, err := store.Write(exited(0))
	if err != nil || first != "20261120T123005Z" {
		t.Fatalf("Write() = (%q, %v), want the id of the start second", first, err)
	}
	second, err := store.Write(exited(1))
	if err != nil || second != "20261120T123005Z-2" {
		t.Fatalf("Write(same second) = (%q, %v), want a suffixed id", second, err)
	}
	info, err := files.Stat("/state/agent-cli/runs/" + first + ".json")
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("record file = (%v, %v), want a private file", info, err)
	}
	if store.Dir() != "/state/agent-cli/runs" {
		t.Errorf("Dir() = %q", store.Dir())
	}
	read, err := store.Read(second)
	if err != nil || read.ID != second || *read.ExitCode != 1 {
		t.Errorf("Read() = (%+v, %v), want the second record", read, err)
	}
}

func TestStoreListsNewestFirst(t *testing.T) {
	store, _ := newStore(t)
	older := exited(0)
	older.StartedAt = started.Add(-time.Hour)
	for _, r := range []Record{older, exited(1), exited(2)} {
		if _, err := store.Write(r); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}
	records, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	var ids []string
	for _, r := range records {
		ids = append(ids, r.ID)
	}
	want := []string{"20261120T123005Z-2", "20261120T123005Z", "20261120T113005Z"}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("List() ids = %v, want %v", ids, want)
	}
}

func TestStoreListIgnoresForeignFilesAndAnAbsentDirectory(t *testing.T) {
	store, files := newStore(t)
	if records, err := store.List(); err != nil || len(records) != 0 {
		t.Fatalf("List(absent) = (%v, %v), want none", records, err)
	}
	files.AddFile("/state/agent-cli/runs/notes.txt", "x", 0o600).AddDir("/state/agent-cli/runs/20261120T123005Z.json")
	if records, err := store.List(); err != nil || len(records) != 0 {
		t.Fatalf("List(foreign files) = (%v, %v), want none", records, err)
	}
}

func TestStoreReadRefusesABadIDAndAMissingOrBrokenRecord(t *testing.T) {
	store, files := newStore(t)
	for _, id := range []string{"../../etc/passwd", "", "latest", "20261120T123005Z/x"} {
		if _, err := store.Read(id); err == nil || !strings.Contains(err.Error(), "not a run record id") {
			t.Errorf("Read(%q) error = %v, want an id refusal", id, err)
		}
	}
	if _, err := store.Read("20261120T123005Z"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Read(missing) error = %v, want a missing-file error", err)
	}
	files.AddFile("/state/agent-cli/runs/20261120T123005Z.json", "not json", 0o600)
	if _, err := store.Read("20261120T123005Z"); err == nil || !strings.Contains(err.Error(), "20261120T123005Z") {
		t.Errorf("Read(broken) error = %v, want the record named", err)
	}
	if _, err := store.List(); err == nil {
		t.Error("List() with a broken record error = nil, want it reported")
	}
}

func TestNewStoreRefusesARelativeDirectory(t *testing.T) {
	for _, dir := range []string{"runs", "/a/../b", "/a/"} {
		if _, err := NewStore(newMemFiles(), dir); err == nil {
			t.Errorf("NewStore(%q) error = nil, want a refusal", dir)
		}
	}
}

type brokenFiles struct {
	*memFiles
	mkdirErr, writeErr, statErr, readDirErr error
}

func (b brokenFiles) MkdirAll(p string, perm fs.FileMode) error {
	if b.mkdirErr != nil {
		return b.mkdirErr
	}
	return b.memFiles.MkdirAll(p, perm)
}

func (b brokenFiles) WriteFile(p string, data []byte, perm fs.FileMode) error {
	if b.writeErr != nil {
		return b.writeErr
	}
	return b.memFiles.WriteFile(p, data, perm)
}

func (b brokenFiles) Stat(p string) (fs.FileInfo, error) {
	if b.statErr != nil {
		return nil, b.statErr
	}
	return b.memFiles.Stat(p)
}

func (b brokenFiles) ReadDir(p string) ([]fs.DirEntry, error) {
	if b.readDirErr != nil {
		return nil, b.readDirErr
	}
	return b.memFiles.ReadDir(p)
}

func TestStoreReportsDiskFailures(t *testing.T) {
	cases := map[string]brokenFiles{
		"mkdir": {memFiles: newMemFiles(), mkdirErr: errors.New("read-only")},
		"write": {memFiles: newMemFiles(), writeErr: errors.New("disk full")},
		"stat":  {memFiles: newMemFiles(), statErr: errors.New("denied")},
	}
	for name, files := range cases {
		store, err := NewStore(files, "/state/runs")
		if err != nil {
			t.Fatalf("NewStore() error = %v", err)
		}
		if _, err := store.Write(exited(0)); err == nil {
			t.Errorf("%s: Write() error = nil, want the failure reported", name)
		}
	}
	store, _ := NewStore(brokenFiles{memFiles: newMemFiles(), readDirErr: errors.New("denied")}, "/state/runs")
	if _, err := store.List(); err == nil || !strings.Contains(err.Error(), "denied") {
		t.Errorf("List() error = %v, want the failure reported", err)
	}
}

func TestStoreGivesUpWhenEveryIDIsTaken(t *testing.T) {
	store, files := newStore(t)
	files.AddFile("/state/agent-cli/runs/20261120T123005Z.json", "{}", 0o600)
	for suffix := 2; suffix <= maxSuffix; suffix++ {
		files.AddFile("/state/agent-cli/runs/20261120T123005Z-"+strconv.Itoa(suffix)+".json", "{}", 0o600)
	}
	if _, err := store.Write(exited(0)); err == nil || !strings.Contains(err.Error(), "no free run record id") {
		t.Errorf("Write() error = %v, want the exhaustion reported", err)
	}
}

func TestMarshalReportsATimeJSONCannotEncode(t *testing.T) {
	r := exited(0)
	r.EndedAt = time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC)
	if _, err := Marshal(r); err == nil {
		t.Error("Marshal() error = nil, want the encoding failure")
	}
	store, _ := newStore(t)
	if _, err := store.Write(r); err == nil {
		t.Error("Write() error = nil, want the encoding failure")
	}
}
