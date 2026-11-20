package plugin

import (
	"errors"
	"io/fs"

	libplugin "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/plugin"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/plugins"
	origin "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
)

const storeDir = "/store"

// faulty fails the operations that fail names, and passes the rest to the fake.
type faulty struct {
	*oshosttest.Fake
	fail func(op, path string) error
}

func (f faulty) check(op, path string) error {
	if f.fail == nil {
		return nil
	}
	return f.fail(op, path)
}

func (f faulty) Lstat(p string) (fs.FileInfo, error) {
	if err := f.check("lstat", p); err != nil {
		return nil, err
	}
	return f.Fake.Lstat(p)
}

func (f faulty) ReadDir(p string) ([]fs.DirEntry, error) {
	if err := f.check("readdir", p); err != nil {
		return nil, err
	}
	return f.Fake.ReadDir(p)
}

func (f faulty) ReadFile(p string) ([]byte, error) {
	if err := f.check("read", p); err != nil {
		return nil, err
	}
	return f.Fake.ReadFile(p)
}

func (f faulty) WriteFile(p string, data []byte, perm fs.FileMode) error {
	if err := f.check("write", p); err != nil {
		return err
	}
	return f.Fake.WriteFile(p, data, perm)
}

func (f faulty) MkdirAll(p string, perm fs.FileMode) error {
	if err := f.check("mkdir", p); err != nil {
		return err
	}
	return f.Fake.MkdirAll(p, perm)
}

func (f faulty) Rename(oldPath, newPath string) error {
	if err := f.check("rename", newPath); err != nil {
		return err
	}
	return f.Fake.Rename(oldPath, newPath)
}

func (f faulty) RemoveAll(p string) error {
	if err := f.check("removeall", p); err != nil {
		return err
	}
	return f.Fake.RemoveAll(p)
}

var errInjected = errors.New("injected")

func failOn(op string) func(string, string) error {
	return func(got, _ string) error {
		if got == op {
			return errInjected
		}
		return nil
	}
}

// fixture is a service over the in-memory machine with a plugin directory and a
// marketplace directory in place.
type fixture struct {
	host    *oshosttest.Fake
	service Service
}

func manifest(name string, dependencies ...string) string {
	text := `{"name":"` + name + `","version":"1.0.0","description":"d"`
	if len(dependencies) > 0 {
		text += `,"dependencies":["` + dependencies[0] + `"]`
	}
	return text + "}"
}

func newFixture(t interface {
	Helper()
	Fatal(...any)
}, wrap func(*oshosttest.Fake) Files) fixture {
	t.Helper()
	host := oshosttest.New()
	host.WorkingDir = "/work"
	host.AddFile("/src/hello/.claude-plugin/plugin.json", manifest("hello"), 0o644).
		AddFile("/src/hello/skills/greet/SKILL.md", "---\nname: greet\ndescription: Use when greeting.\n---\n# Greet\n", 0o644).
		AddFile("/src/hello/skills/greet/run.sh", "#!/bin/sh\n", 0o755).
		AddFile("/src/hello/.git/config", "ignored", 0o644).
		AddFile("/market/.claude-plugin/marketplace.json", `{"name":"main","plugins":[{"name":"hello","description":"d","source":"./plugins/hello"},{"name":"remote","source":{"source":"github"}},{"name":"escape","source":"./../out"}]}`, 0o644).
		AddFile("/market/plugins/hello/.claude-plugin/plugin.json", manifest("hello"), 0o644).
		AddFile("/market/plugins/hello/agents/reviewer.md", "---\ndescription: Reviews.\n---\nBe strict.\n", 0o644)
	var files Files = host
	if wrap != nil {
		files = wrap(host)
	}
	store, err := NewStore(files, storeDir)
	if err != nil {
		t.Fatal(err)
	}
	return fixture{host: host, service: Service{Store: store, Files: files, Origins: plugins.DefaultRegistry(host)}}
}

func hooks() []libplugin.Capability { return []libplugin.Capability{libplugin.CapabilityHooks} }

var _ origin.Files = faulty{}
