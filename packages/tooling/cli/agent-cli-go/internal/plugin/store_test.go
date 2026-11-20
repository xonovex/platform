package plugin

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	libplugin "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/plugin"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func TestDefaultDirPrefersAnAbsoluteXDGDataHome(t *testing.T) {
	env := map[string]string{"XDG_DATA_HOME": "/data"}
	if got := DefaultDir(func(k string) string { return env[k] }, "/home/u"); got != "/data/agent-cli/plugins" {
		t.Errorf("DefaultDir = %s", got)
	}
	for _, value := range []string{"", "relative"} {
		env["XDG_DATA_HOME"] = value
		if got := DefaultDir(func(k string) string { return env[k] }, "/home/u"); got != "/home/u/.local/share/agent-cli/plugins" {
			t.Errorf("DefaultDir(%q) = %s", value, got)
		}
	}
}

func TestNewStoreNeedsACleanAbsoluteDirectory(t *testing.T) {
	for _, dir := range []string{"", "rel", "/a/../b", "/a//b", "/a/"} {
		if _, err := NewStore(systemtest.New(), dir); err == nil {
			t.Errorf("NewStore(%q) passed", dir)
		}
	}
	store, err := NewStore(systemtest.New(), "/s")
	if err != nil || store.Dir() != "/s" {
		t.Errorf("NewStore = %v, %v", store, err)
	}
}

func TestAnEmptyStoreHasAnEmptyLockAndRegistry(t *testing.T) {
	store, _ := NewStore(systemtest.New(), "/s")
	if lock, err := store.Lock(); err != nil || len(lock.Plugins) != 0 || lock.Version != libplugin.LockVersion {
		t.Errorf("Lock = %+v, %v", lock, err)
	}
	if registry, err := store.Registry(); err != nil || len(registry.Marketplaces) != 0 {
		t.Errorf("Registry = %+v, %v", registry, err)
	}
}

func TestLockAndRegistryRoundTripThroughTheDisk(t *testing.T) {
	host := systemtest.New()
	store, _ := NewStore(host, "/s")
	digest := artifact.Tree{{Path: "a"}}.Digest()
	lock := libplugin.EmptyLock().With(libplugin.Locked{Name: "a", Version: "1", Source: "/x", Digest: digest})
	if err := store.SaveLock(lock); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRegistry(libplugin.EmptyRegistry().With(libplugin.Source{Name: "m", Source: "/m"})); err != nil {
		t.Fatal(err)
	}
	got, err := store.Lock()
	if err != nil || len(got.Plugins) != 1 || got.Plugins[0].Name != "a" {
		t.Errorf("Lock = %+v, %v", got, err)
	}
	registry, err := store.Registry()
	if err != nil || len(registry.Marketplaces) != 1 {
		t.Errorf("Registry = %+v, %v", registry, err)
	}
	if host.Exists("/s/lock.json.tmp") {
		t.Error("the staging file stayed behind")
	}
}

func TestACorruptLockOrRegistryFailsClosed(t *testing.T) {
	host := systemtest.New().AddFile("/s/lock.json", "{", 0o600).AddFile("/s/marketplaces.json", "{", 0o600)
	store, _ := NewStore(host, "/s")
	if _, err := store.Lock(); err == nil {
		t.Error("a corrupt lock passed")
	}
	if _, err := store.Registry(); err == nil {
		t.Error("a corrupt registry passed")
	}
	bad := libplugin.Lock{Version: 9}
	if err := store.SaveLock(bad); err == nil {
		t.Error("an invalid lock was saved")
	}
	if err := store.SaveRegistry(libplugin.Registry{Version: 9}); err == nil {
		t.Error("an invalid registry was saved")
	}
}

func TestUnreadableLockAndRegistryFilesFailClosed(t *testing.T) {
	host := systemtest.New().AddFile("/s/lock.json", "x", 0o600).AddFile("/s/marketplaces.json", "x", 0o600)
	store, _ := NewStore(faulty{host, failOn("read")}, "/s")
	if _, err := store.Lock(); !errors.Is(err, errInjected) {
		t.Errorf("Lock: %v", err)
	}
	if _, err := store.Registry(); !errors.Is(err, errInjected) {
		t.Errorf("Registry: %v", err)
	}
}

func TestWritesFailClosedOnEveryOperation(t *testing.T) {
	tree := artifact.Tree{{Path: "a/run.sh", Executable: true, Content: []byte("x")}}
	for _, op := range []string{"mkdir", "write", "rename"} {
		store, _ := NewStore(faulty{systemtest.New(), failOn(op)}, "/s")
		if err := store.SaveLock(libplugin.EmptyLock()); !errors.Is(err, errInjected) {
			t.Errorf("SaveLock with %s failing: %v", op, err)
		}
		if err := store.Put(tree); !errors.Is(err, errInjected) {
			t.Errorf("Put with %s failing: %v", op, err)
		}
	}
	store, _ := NewStore(faulty{systemtest.New(), failOn("removeall")}, "/s")
	if err := store.Put(tree); !errors.Is(err, errInjected) {
		t.Errorf("Put with removeall failing: %v", err)
	}
	if err := store.Delete(tree.Digest()); !errors.Is(err, errInjected) {
		t.Errorf("Delete: %v", err)
	}
	// Only the nested directory of a file fails: the staging directory itself exists.
	nested := faulty{systemtest.New(), func(op, p string) error {
		if op == "mkdir" && strings.HasSuffix(p, "/a") {
			return errInjected
		}
		return nil
	}}
	store, _ = NewStore(nested, "/s")
	if err := store.Put(tree); !errors.Is(err, errInjected) {
		t.Errorf("Put with the nested mkdir failing: %v", err)
	}
}

func TestPutKeepsAnExistingTreeAndGetVerifiesTheDigest(t *testing.T) {
	host := systemtest.New()
	store, _ := NewStore(host, "/s")
	tree, err := artifact.NewTree([]artifact.File{{Path: "skills/a/SKILL.md", Content: []byte("x")}, {Path: "run.sh", Executable: true, Content: []byte("y")}})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(tree); err != nil {
		t.Fatal(err)
	}
	if err := store.Put(tree); err != nil {
		t.Fatalf("second Put: %v", err)
	}
	got, err := store.Get(tree.Digest())
	if err != nil || got.Digest() != tree.Digest() {
		t.Fatalf("Get = %v, %v", got, err)
	}
	if file, _ := got.Get("run.sh"); !file.Executable {
		t.Error("the executable bit was lost")
	}

	hex := strings.TrimPrefix(tree.Digest(), artifact.DigestPrefix)
	if err := host.WriteFile("/s/content/"+hex+"/skills/a/SKILL.md", []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(tree.Digest()); err == nil || !strings.Contains(err.Error(), "changed on disk") {
		t.Errorf("tampered Get: %v", err)
	}
	if err := store.Delete(tree.Digest()); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(tree.Digest()); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("deleted Get: %v", err)
	}
}
