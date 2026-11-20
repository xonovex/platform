package resolve_test

import (
	"encoding/json"
	"errors"
	iofs "io/fs"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/configtest"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/resolve"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
)

var (
	open   = spec.Section{Name: "audit", SchemaID: "audit", Keys: []spec.Key{{Section: "audit", Pointer: "/maxAge", Rule: spec.Replace, Layers: spec.EveryFile}}}
	gate   = spec.Section{Name: "eval", SchemaID: "eval", Gated: true, Keys: []spec.Key{{Section: "eval", Pointer: "/runs", Rule: spec.Cap, Layers: spec.EveryFile, Ceiling: 24}}}
	places = configtest.Places{Managed: "/etc/agentic", User: "/home/u/.config/agentic", State: "/home/u/.local/state/agentic", Wd: "/repo"}
)

type validator struct{ violations []jsonschema.Violation }

func (v validator) Validate(string, []byte) ([]jsonschema.Violation, error) { return v.violations, nil }

func env(files *configtest.Files, prompter trust.Prompter) resolve.Env {
	return resolve.Env{
		Files: files, Places: places, Validator: validator{},
		Approvals: trust.Ledger{Store: files, StateDir: places.StateDir},
		Prompter:  prompter, Out: &strings.Builder{},
	}
}

func TestResolveMergesEveryLayer(t *testing.T) {
	files := configtest.NewFiles(map[string]string{
		"/etc/agentic/agentic.json":            `{"version":1,"audit":{"maxAge":1}}`,
		"/home/u/.config/agentic/agentic.json": `{"version":1,"audit":{"maxAge":2},"eval":{"runs":50}}`,
		"/repo/.git/HEAD":                      "",
		"/repo/agentic.json":                   `{"version":1,"audit":{"maxAge":3},"eval":{"bad":true}}`,
	})
	r, err := resolve.Resolve(env(files, &configtest.Prompter{}), resolve.Request{Sections: []spec.Section{open}, Start: "."})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(r.Section("audit")); got != `{"maxAge":1}` {
		t.Fatalf("audit = %s", got)
	}
}

func TestAnUngatedSectionNeedsNoApproval(t *testing.T) {
	files := configtest.NewFiles(map[string]string{"/repo/.git/HEAD": "", "/repo/agentic.json": `{"version":1,"audit":{"maxAge":3},"eval":{"runs":5}}`})
	r, err := resolve.Resolve(env(files, &configtest.Prompter{}), resolve.Request{Sections: []spec.Section{open}, Start: "/repo"})
	if err != nil || string(r.Section("audit")) != `{"maxAge":3}` {
		t.Fatalf("audit = %s, %v", r.Section("audit"), err)
	}
}

func TestAGatedProjectSectionFailsClosed(t *testing.T) {
	files := configtest.NewFiles(map[string]string{"/repo/.git/HEAD": "", "/repo/agentic.json": `{"version":1,"eval":{"runs":5}}`})
	_, err := resolve.Resolve(env(files, &configtest.Prompter{}), resolve.Request{Sections: []spec.Section{gate}, Start: "/repo"})
	if !errors.Is(err, trust.ErrNotApproved) {
		t.Fatalf("err = %v", err)
	}
}

func TestAnApprovedGatedSectionTakesEffect(t *testing.T) {
	const content = `{"version":1,"eval":{"runs":5}}`
	files := configtest.NewFiles(map[string]string{"/repo/.git/HEAD": "", "/repo/agentic.json": content})
	request := resolve.Request{Sections: []spec.Section{gate}, Start: "/repo", ApproveDigest: layer.Digest([]byte(content))}
	r, err := resolve.Resolve(env(files, &configtest.Prompter{}), request)
	if err != nil || string(r.Section("eval")) != `{"runs":5}` {
		t.Fatalf("eval = %s, %v", r.Section("eval"), err)
	}
	if _, err := files.ReadFile("/home/u/.local/state/agentic/trust.json"); err != nil {
		t.Fatalf("the approval must be kept: %v", err)
	}
}

func TestTheFlagsJoinTheMerge(t *testing.T) {
	r, err := resolve.Resolve(env(configtest.NewFiles(nil), &configtest.Prompter{}), resolve.Request{
		Sections: []spec.Section{gate}, Start: "/repo",
		Flags: []merge.Flag{{Section: "eval", Pointer: "/runs", Value: json.RawMessage("100")}},
	})
	if err != nil || string(r.Section("eval")) != `{"runs":100}` {
		t.Fatalf("eval = %s, %v", r.Section("eval"), err)
	}
}

func TestANamedProjectFileMustExist(t *testing.T) {
	_, err := resolve.Resolve(env(configtest.NewFiles(nil), &configtest.Prompter{}), resolve.Request{Sections: []spec.Section{open}, Start: "/repo", Project: "other.json"})
	if !errors.Is(err, iofs.ErrNotExist) || !strings.Contains(err.Error(), "project configuration /repo/other.json") {
		t.Fatalf("err = %v", err)
	}
}

type unreadable struct{ *configtest.Files }

func (unreadable) ReadFile(path string) ([]byte, error) {
	return nil, &iofs.PathError{Op: "read", Path: path, Err: iofs.ErrPermission}
}

func TestResolveReportsEveryFault(t *testing.T) {
	user := "/home/u/.config/agentic/agentic.json"
	cases := map[string]struct {
		env  func() resolve.Env
		want string
	}{
		"no places": {func() resolve.Env {
			e := env(configtest.NewFiles(nil), &configtest.Prompter{})
			e.Places = configtest.Places{Err: errors.New("no home")}
			return e
		}, "no home"},
		"directory": {func() resolve.Env {
			return env(configtest.NewFiles(map[string]string{user + "/x": ""}), &configtest.Prompter{})
		}, "is not a regular file"},
		"unreadable": {func() resolve.Env {
			e := env(configtest.NewFiles(map[string]string{user: "{}"}), &configtest.Prompter{})
			e.Files = unreadable{configtest.NewFiles(map[string]string{user: "{}"})}
			return e
		}, "permission denied"},
		"broken envelope": {func() resolve.Env {
			return env(configtest.NewFiles(map[string]string{user: `{"version":2}`}), &configtest.Prompter{})
		}, "version is 2"},
		"broken section": {func() resolve.Env {
			e := env(configtest.NewFiles(map[string]string{user: `{"version":1,"audit":{}}`}), &configtest.Prompter{})
			e.Validator = validator{violations: []jsonschema.Violation{{Pointer: "/maxAge", Message: "bad"}}}
			return e
		}, "/audit/maxAge: bad"},
	}
	for name, c := range cases {
		if _, err := resolve.Resolve(c.env(), resolve.Request{Sections: []spec.Section{open}, Start: "/repo"}); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
