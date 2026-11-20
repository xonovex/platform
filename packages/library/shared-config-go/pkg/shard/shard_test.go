package shard_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/configtest"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/shard"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
)

var places = configtest.Places{Managed: "/etc/agentic", User: "/home/u/.config/agentic", State: "/home/u/.local/state/agentic", Wd: "/repo"}

const projectShard = `{"binary":"./tools/claude"}`

func env(files discover.Files, store trust.Store, prompter trust.Prompter, out *strings.Builder) shard.Env {
	return shard.Env{
		Files: files, Places: places,
		Approvals: trust.Ledger{Store: store, StateDir: places.StateDir},
		Prompter:  prompter, Out: out,
	}
}

func request(approve string) shard.Request {
	return shard.Request{Dir: "harnesses.d", Names: []string{"claude", "codex"}, Kind: "harness", Start: ".", ApproveDigest: approve}
}

func layered() *configtest.Files {
	return configtest.NewFiles(map[string]string{
		"/etc/agentic/harnesses.d/claude.json":            `{"binary":"/opt/claude"}`,
		"/home/u/.config/agentic/harnesses.d/claude.json": `{"binary":"claude-user"}`,
		"/home/u/.config/agentic/harnesses.d/codex.json":  `{"binary":"codex-user"}`,
		"/home/u/.config/agentic/harnesses.d/notes.txt":   `x`,
		"/repo/.git/HEAD":                    "",
		"/repo/agentic.json":                 `{"version":1}`,
		"/repo/harnesses.d/claude.json":      projectShard,
		"/elsewhere/harnesses.d/claude.json": `{}`,
	})
}

func names(shards []shard.Shard) []string {
	var out []string
	for _, s := range shards {
		out = append(out, string(s.Layer)+":"+s.Path)
	}
	return out
}

func TestReadOrdersTheLayersSoTheManagedShardWins(t *testing.T) {
	files := layered()
	shards, err := shard.Read(env(files, files, &configtest.Prompter{}, &strings.Builder{}), request(layer.Digest([]byte(projectShard))))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"user:/home/u/.config/agentic/harnesses.d/claude.json",
		"user:/home/u/.config/agentic/harnesses.d/codex.json",
		"project:/repo/harnesses.d/claude.json",
		"managed:/etc/agentic/harnesses.d/claude.json",
	}
	if got := names(shards); !reflect.DeepEqual(got, want) {
		t.Fatalf("Read() = %v, want %v", got, want)
	}
	if shards[2].Name != "claude" || string(shards[2].Data) != projectShard {
		t.Errorf("project shard = %+v", shards[2])
	}
}

func TestAnUnapprovedProjectShardFailsClosed(t *testing.T) {
	files := layered()
	_, err := shard.Read(env(files, files, &configtest.Prompter{}, &strings.Builder{}), request(""))
	if !errors.Is(err, trust.ErrNotApproved) || !strings.Contains(err.Error(), "/repo/harnesses.d/claude.json sets the harness configuration") {
		t.Fatalf("Read() error = %v, want the gate to name the shard", err)
	}
}

func TestAnApprovalAtTheTerminalLetsTheProjectShardIn(t *testing.T) {
	files := layered()
	out := &strings.Builder{}
	shards, err := shard.Read(env(files, files, &configtest.Prompter{Terminal: true, Answers: []string{"y"}}, out), request(""))
	if err != nil || len(shards) != 4 {
		t.Fatalf("Read() = %v, %v", names(shards), err)
	}
	if !strings.Contains(out.String(), projectShard) {
		t.Errorf("the prompt does not show the shard: %q", out.String())
	}
}

func TestWithoutAProjectFileThereIsNoProjectShard(t *testing.T) {
	files := configtest.NewFiles(map[string]string{
		"/repo/.git/HEAD":               "",
		"/repo/harnesses.d/claude.json": projectShard,
	})
	shards, err := shard.Read(env(files, files, &configtest.Prompter{}, &strings.Builder{}), request(""))
	if err != nil || len(shards) != 0 {
		t.Fatalf("Read() = %v, %v, want no shard", names(shards), err)
	}
}

type faultyFiles struct {
	*configtest.Files
	statErr, readErr string
}

func (f faultyFiles) Stat(path string) (discover.Info, error) {
	if path == f.statErr {
		return discover.Info{}, errors.New("stat fault")
	}
	return f.Files.Stat(path)
}

func (f faultyFiles) ReadFile(path string) ([]byte, error) {
	if path == f.readErr {
		return nil, errors.New("read fault")
	}
	return f.Files.ReadFile(path)
}

func TestReadReportsEveryFault(t *testing.T) {
	userShard := "/home/u/.config/agentic/harnesses.d/claude.json"
	dirShard := configtest.NewFiles(map[string]string{"/home/u/.config/agentic/harnesses.d/claude.json/x": ""})
	cases := map[string]struct {
		files  discover.Files
		places configtest.Places
		want   string
	}{
		"stat":        {faultyFiles{Files: layered(), statErr: userShard}, places, "stat fault"},
		"read":        {faultyFiles{Files: layered(), readErr: userShard}, places, "read fault"},
		"not a file":  {dirShard, places, "not a regular file"},
		"no user dir": {layered(), configtest.Places{Err: errors.New("no home")}, "no home"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			e := env(c.files, configtest.NewFiles(nil), &configtest.Prompter{}, &strings.Builder{})
			e.Places = c.places
			if _, err := shard.Read(e, request("")); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Read() error = %v, want %q", err, c.want)
			}
		})
	}
}

type faultyLister struct{ err error }

func (l faultyLister) ReadDir(string) ([]string, error) { return nil, l.err }

func TestARequestWithoutNamesReadsEveryShard(t *testing.T) {
	files := layered()
	e := env(files, configtest.NewFiles(nil), &configtest.Prompter{}, &strings.Builder{})
	e.Lister = files
	all := shard.Request{Dir: "harnesses.d", Kind: "harness", Start: "."}
	shards, err := shard.Find(e, all)
	want := []string{
		"user:/home/u/.config/agentic/harnesses.d/claude.json",
		"user:/home/u/.config/agentic/harnesses.d/codex.json",
		"project:/repo/harnesses.d/claude.json",
		"managed:/etc/agentic/harnesses.d/claude.json",
	}
	if err != nil || !reflect.DeepEqual(names(shards), want) {
		t.Fatalf("Find() = %v, %v, want %v", names(shards), err, want)
	}
	if shards[2].Digest != layer.Digest([]byte(projectShard)) {
		t.Fatalf("Digest = %s", shards[2].Digest)
	}
	if _, err := shard.Read(e, all); !errors.Is(err, trust.ErrNotApproved) {
		t.Fatalf("Read() of an unapproved project shard = %v", err)
	}
	e.Lister = faultyLister{err: errors.New("list fault")}
	if _, err := shard.Find(e, all); err == nil || !strings.Contains(err.Error(), "list fault") {
		t.Fatalf("Find() with a list fault = %v", err)
	}
	e.Lister = nil
	if _, err := shard.Find(e, all); err == nil || !strings.Contains(err.Error(), "needs a Lister") {
		t.Fatalf("Find() without a Lister = %v", err)
	}
	empty := configtest.NewFiles(map[string]string{"/repo/.git/HEAD": "", "/repo/harnesses.d/.json": "{}"})
	e = env(empty, empty, &configtest.Prompter{}, &strings.Builder{})
	e.Lister = empty
	if shards, err := shard.Find(e, all); err != nil || len(shards) != 0 {
		t.Fatalf("Find() without shard files = %v, %v", names(shards), err)
	}
}
