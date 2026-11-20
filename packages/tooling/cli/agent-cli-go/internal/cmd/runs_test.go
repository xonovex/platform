package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/record"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/plugin"
)

type fakeRecords struct {
	records []record.Record
	err     error
}

func (f fakeRecords) List() ([]record.Record, error) { return f.records, f.err }
func (f fakeRecords) Read(id string) (record.Record, error) {
	for _, r := range f.records {
		if r.ID == id {
			return r, f.err
		}
	}
	return record.Record{}, errors.New("no such record")
}

func sampleRecords() []record.Record {
	code, other := 0, 2
	started := time.Date(2026, time.November, 20, 12, 0, 0, 0, time.UTC)
	return []record.Record{
		{ID: "20261120T120100Z", StartedAt: started, EndedAt: started, Status: record.StatusExited, ExitCode: &other,
			Agent: record.Agent{Name: "claude", VersionProbe: "claude --version"}, Axes: record.Axes{Isolation: "bwrap", Provision: "nix", Network: "none", Workspace: record.Workspace{Mode: record.ModeInPlace}},
			WorkDir: "/work/a", Credentials: []string{"TOKEN"},
			Guarantees: []record.Guarantee{{Name: "pinned-provision", Required: true, Provided: true}},
			Overlay:    record.Overlay{Plugins: []record.Digest{{Name: "p", Digest: "sha256:aa"}}, Directories: []record.Digest{{Name: "/d", Digest: "sha256:bb"}}}},
		{ID: "20261120T120000Z", StartedAt: started, EndedAt: started, Status: record.StatusFailed, Error: "boom",
			Agent: record.Agent{Name: "codex", VersionProbe: "codex --version"}, Axes: record.Axes{Isolation: "none"}, WorkDir: "/work/b"},
		{ID: "20261119T120000Z", Status: record.StatusExited, ExitCode: &code, Agent: record.Agent{Name: "pi"}},
	}
}

func runs(t *testing.T, records launch.RecordReader, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	root := NewRootCommand(newTestRoot().deps, plugin.Manager{}, RecordPorts{Records: records, Out: &out}, &recordingBridge{})
	root.SetArgs(append([]string{"runs"}, args...))
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	err := root.Execute()
	return out.String(), err
}

func TestRunsListShowsEachRecordNewestFirst(t *testing.T) {
	out, err := runs(t, fakeRecords{records: sampleRecords()}, "list")

	if err != nil {
		t.Fatalf("runs list error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 4 || !strings.HasPrefix(lines[0], "ID") || !strings.Contains(lines[1], "20261120T120100Z") || !strings.Contains(lines[1], "exited") || !strings.Contains(lines[2], "failed-to-start") {
		t.Errorf("runs list =\n%s", out)
	}
}

func TestRunsListSaysWhenThereAreNoRuns(t *testing.T) {
	out, err := runs(t, fakeRecords{}, "list")
	if err != nil || out != "No runs recorded.\n" {
		t.Errorf("runs list = %q, %v", out, err)
	}
	out, err = runs(t, fakeRecords{}, "list", "--json")
	if err != nil || strings.TrimSpace(out) != "[]" {
		t.Errorf("runs list --json = %q, %v; want an empty array", out, err)
	}
}

func TestRunsListPrintsJSON(t *testing.T) {
	out, err := runs(t, fakeRecords{records: sampleRecords()}, "list", "--json")

	var records []record.Record
	if err != nil || json.Unmarshal([]byte(out), &records) != nil || len(records) != 3 {
		t.Errorf("runs list --json = %q, %v", out, err)
	}
}

func TestRunsShowPrintsARecordByIDOrAsTheLatest(t *testing.T) {
	records := fakeRecords{records: sampleRecords()}
	for _, id := range []string{"20261120T120100Z", "latest"} {
		out, err := runs(t, records, "show", id)
		if err != nil {
			t.Fatalf("runs show %s error = %v", id, err)
		}
		for _, want := range []string{"Run 20261120T120100Z: exited (exit 2)", "Isolation: bwrap, provision: nix, network: none", "Credential: TOKEN", "Plugin p: sha256:aa", "Overlay directory /d: sha256:bb", "Guarantee pinned-provision: required true, provided true"} {
			if !strings.Contains(out, want) {
				t.Errorf("runs show %s lacks %q:\n%s", id, want, out)
			}
		}
	}
	out, err := runs(t, records, "show", "20261120T120000Z")
	if err != nil || !strings.Contains(out, "Error: boom") || !strings.Contains(out, "(exit -)") {
		t.Errorf("runs show of a failed run = %q, %v", out, err)
	}
}

func TestRunsShowPrintsJSON(t *testing.T) {
	out, err := runs(t, fakeRecords{records: sampleRecords()}, "show", "latest", "--json")

	got, parseErr := record.Unmarshal([]byte(out))
	if err != nil || parseErr != nil || got.ID != "20261120T120100Z" {
		t.Errorf("runs show --json = %q, %v, %v", out, err, parseErr)
	}
}

func TestRunsShowFailsForAMissingRecordAndAnEmptyStore(t *testing.T) {
	if _, err := runs(t, fakeRecords{records: sampleRecords()}, "show", "20260101T000000Z"); err == nil {
		t.Error("runs show of an unknown id succeeded")
	}
	if _, err := runs(t, fakeRecords{}, "show", "latest"); err == nil || !strings.Contains(err.Error(), "no run is recorded") {
		t.Errorf("runs show latest of an empty store error = %v", err)
	}
	if _, err := runs(t, fakeRecords{err: errors.New("denied")}, "show", "latest"); err == nil {
		t.Error("runs show latest hid a store failure")
	}
	if _, err := runs(t, fakeRecords{err: errors.New("denied")}, "list"); err == nil {
		t.Error("runs list hid a store failure")
	}
}

func TestRunsReportsABuildWithoutRecords(t *testing.T) {
	for _, args := range [][]string{{"list"}, {"show", "latest"}} {
		if _, err := runs(t, nil, args...); err == nil || !strings.Contains(err.Error(), "keeps no run records") {
			t.Errorf("runs %v error = %v", args, err)
		}
	}
}

func TestRunsReportsAFailureToWrite(t *testing.T) {
	records := fakeRecords{records: sampleRecords()}
	for _, args := range [][]string{{"list"}, {"list", "--json"}, {"show", "latest"}, {"show", "latest", "--json"}} {
		command := newRunsCommand(RecordPorts{Records: records, Out: brokenWriter{}})
		command.SetArgs(args)
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)
		if err := command.Execute(); err == nil {
			t.Errorf("runs %v error = nil, want the write failure", args)
		}
	}
}

func TestCompletionOffersTheRunIDs(t *testing.T) {
	var out bytes.Buffer
	root := NewRootCommand(newTestRoot().deps, plugin.Manager{}, RecordPorts{Records: fakeRecords{records: sampleRecords()}, Out: io.Discard}, &recordingBridge{})
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"__complete", "runs", "show", ""})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"latest", "20261120T120100Z"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("completion lacks %q: %q", want, out.String())
		}
	}
	for _, args := range [][]string{{"__complete", "runs", "show", "latest", ""}} {
		out.Reset()
		root.SetArgs(args)
		_ = root.Execute()
		if strings.Contains(out.String(), "20261120T120100Z") {
			t.Errorf("completion %v offered an id after the id was given", args)
		}
	}
}
