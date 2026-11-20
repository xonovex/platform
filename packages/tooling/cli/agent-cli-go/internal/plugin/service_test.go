package plugin

import (
	"errors"
	"strings"
	"testing"

	libplugin "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/plugin"
	origin "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func install(t *testing.T, f fixture, spec string, allow ...libplugin.Capability) libplugin.Locked {
	t.Helper()
	fetched, err := f.service.Fetch(spec)
	if err != nil {
		t.Fatalf("Fetch(%s): %v", spec, err)
	}
	record, err := f.service.Commit(fetched, fetched.Digest, allow)
	if err != nil {
		t.Fatalf("Commit(%s): %v", spec, err)
	}
	return record
}

func TestInstallNeedsTheApprovalOfTheExactDigest(t *testing.T) {
	f := newFixture(t, nil)
	fetched, err := f.service.Fetch("/src/hello")
	if err != nil {
		t.Fatal(err)
	}
	if fetched.Manifest.Name != "hello" || fetched.FileCount != 3 || fetched.Source != "/src/hello" || fetched.Marketplace != "" || len(fetched.Capabilities) != 0 {
		t.Errorf("fetched = %+v", fetched)
	}

	if _, err := f.service.Commit(fetched, "", nil); err == nil || !strings.Contains(err.Error(), "--approve-digest "+fetched.Digest) {
		t.Errorf("preview: %v", err)
	}
	if _, err := f.service.Commit(fetched, "sha256:"+strings.Repeat("0", 64), nil); err == nil || !strings.Contains(err.Error(), "differs") {
		t.Errorf("wrong digest: %v", err)
	}
	if f.host.Exists(storeDir) {
		t.Error("a refused install wrote into the store")
	}

	record, err := f.service.Commit(fetched, fetched.Digest, nil)
	if err != nil || record.Name != "hello" || record.Digest != fetched.Digest || record.Source != "/src/hello" {
		t.Fatalf("Commit = %+v, %v", record, err)
	}
	list, err := f.service.List()
	if err != nil || len(list) != 1 || list[0].Name != "hello" {
		t.Errorf("List = %+v, %v", list, err)
	}
	set, err := f.service.Load([]string{"hello"})
	if err != nil || len(set.Skills) != 1 || set.Skills[0].Name != "greet" || set.Skills[0].Source != "hello" || len(set.Skills[0].Files) != 1 {
		t.Errorf("Load = %+v, %v", set, err)
	}
	if err := f.service.Remove("hello"); err != nil {
		t.Fatal(err)
	}
	if list, _ := f.service.List(); len(list) != 0 {
		t.Errorf("List after Remove = %+v", list)
	}
	if _, err := f.service.Load([]string{"hello"}); err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("Load after Remove: %v", err)
	}
}

func TestRemoveDeletesTheStoredFilesAndFailsClosed(t *testing.T) {
	f := newFixture(t, nil)
	record := install(t, f, "/src/hello")
	content := storeDir + "/content/" + strings.TrimPrefix(record.Digest, "sha256:")
	if !f.host.Exists(content) {
		t.Fatal("the plugin files are not stored")
	}
	if err := f.service.Remove("hello"); err != nil || f.host.Exists(content) {
		t.Errorf("Remove: %v, files kept = %v", err, f.host.Exists(content))
	}
	if err := f.service.Remove("hello"); err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("Remove of a missing plugin: %v", err)
	}
}

func TestACapabilityNeedsAGrantAtInstallAndAtLoad(t *testing.T) {
	f := newFixture(t, nil)
	f.host.AddFile("/src/hello/hooks/hooks.json", "{}", 0o644)
	fetched, err := f.service.Fetch("/src/hello")
	if err != nil || len(fetched.Capabilities) != 1 || fetched.Capabilities[0] != libplugin.CapabilityHooks {
		t.Fatalf("Fetch = %+v, %v", fetched, err)
	}
	if _, err := f.service.Commit(fetched, fetched.Digest, nil); err == nil || !strings.Contains(err.Error(), "hooks") {
		t.Errorf("an ungranted capability passed: %v", err)
	}
	record, err := f.service.Commit(fetched, fetched.Digest, hooks())
	if err != nil || len(record.Grants) != 1 {
		t.Fatalf("Commit = %+v, %v", record, err)
	}
	if _, err := f.service.Load([]string{"hello"}); err != nil {
		t.Errorf("Load of a granted plugin: %v", err)
	}

	record.Grants = nil
	lock, _ := f.service.Store.Lock()
	if err := f.service.Store.SaveLock(lock.With(record)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Load([]string{"hello"}); err == nil || !strings.Contains(err.Error(), "carries hooks") {
		t.Errorf("Load without the grant: %v", err)
	}
}

func TestLoadFailsClosedOnTamperedOrMismatchedFiles(t *testing.T) {
	f := newFixture(t, nil)
	record := install(t, f, "/src/hello")
	hex := strings.TrimPrefix(record.Digest, "sha256:")
	if err := f.host.WriteFile(storeDir+"/content/"+hex+"/skills/greet/SKILL.md", []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Load([]string{"hello"}); err == nil || !strings.Contains(err.Error(), "changed on disk") {
		t.Errorf("tampered: %v", err)
	}

	// The lock names a plugin that the stored files do not name.
	g := newFixture(t, nil)
	other := install(t, g, "/src/hello")
	other.Name = "renamed"
	lock, _ := g.service.Store.Lock()
	if err := g.service.Store.SaveLock(lock.With(other)); err != nil {
		t.Fatal(err)
	}
	if _, err := g.service.Load([]string{"renamed"}); err == nil || !strings.Contains(err.Error(), `name the plugin "hello"`) {
		t.Errorf("renamed: %v", err)
	}
}

func TestLoadFailsClosedOnAnUnreadableManifestAndAnUnknownGrant(t *testing.T) {
	f := newFixture(t, nil)
	record := install(t, f, "/src/hello")
	hex := strings.TrimPrefix(record.Digest, "sha256:")
	// A manifest that is not valid JSON changes the digest, so replace the record too.
	host := f.host
	host.AddFile(storeDir+"/content/"+hex+"/.claude-plugin/plugin.json", "{", 0o644)
	if _, err := f.service.Load([]string{"hello"}); err == nil {
		t.Error("a corrupt stored plugin passed")
	}

	// A plugin whose only manifest is broken is refused before it is stored.
	g := newFixture(t, nil)
	g.host.AddFile("/src/broken/plugin.json", "{", 0o644)
	if _, err := g.service.Fetch("/src/broken"); err == nil {
		t.Error("a broken manifest passed")
	}
}

func TestADependencyMustBeInstalledFirstAndCannotBeRemovedWhileUsed(t *testing.T) {
	f := newFixture(t, nil)
	f.host.AddFile("/src/app/plugin.json", manifest("app", "hello"), 0o644).
		AddFile("/src/app/skills/run/SKILL.md", "---\nname: run\ndescription: Use when running.\n---\n", 0o644)
	app, err := f.service.Fetch("/src/app")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Commit(app, app.Digest, nil); err == nil || !strings.Contains(err.Error(), "install it first") {
		t.Errorf("missing dependency: %v", err)
	}
	install(t, f, "/src/hello")
	if _, err := f.service.Commit(app, app.Digest, nil); err != nil {
		t.Fatal(err)
	}
	set, err := f.service.Load([]string{"app"})
	if err != nil || len(set.Skills) != 2 {
		t.Errorf("Load = %+v, %v", set, err)
	}
	if err := f.service.Remove("hello"); err == nil || !strings.Contains(err.Error(), "dependency of app") {
		t.Errorf("Remove of a used dependency: %v", err)
	}
	if err := f.service.Remove("app"); err != nil {
		t.Fatal(err)
	}
	if err := f.service.Remove("hello"); err != nil {
		t.Error(err)
	}
}

func TestLoadFailsWhenTwoPluginsDefineOneSkill(t *testing.T) {
	f := newFixture(t, nil)
	install(t, f, "/src/hello")
	f.host.AddFile("/src/twin/plugin.json", manifest("twin"), 0o644).
		AddFile("/src/twin/skills/greet/SKILL.md", "---\nname: greet\ndescription: Use when twinning.\n---\n", 0o644)
	install(t, f, "/src/twin")
	if _, err := f.service.Load([]string{"hello", "twin"}); err == nil || !strings.Contains(err.Error(), "greet") {
		t.Errorf("a name conflict passed: %v", err)
	}
}

func TestReplacingAPluginKeepsFilesAnotherPluginStillUses(t *testing.T) {
	f := newFixture(t, nil)
	old := install(t, f, "/src/hello")
	// Another record shares the old digest.
	lock, _ := f.service.Store.Lock()
	shared := old
	shared.Name = "sibling"
	if err := f.service.Store.SaveLock(lock.With(shared)); err != nil {
		t.Fatal(err)
	}
	f.host.AddFile("/src/hello/skills/greet/extra.md", "new", 0o644)
	updated := install(t, f, "/src/hello")
	if updated.Digest == old.Digest {
		t.Fatal("the digest did not change")
	}
	oldDir := storeDir + "/content/" + strings.TrimPrefix(old.Digest, "sha256:")
	if !f.host.Exists(oldDir) {
		t.Error("files another plugin uses were removed")
	}
	// Reinstalling identical content changes nothing.
	again := install(t, f, "/src/hello")
	if again.Digest != updated.Digest {
		t.Error("an identical reinstall changed the digest")
	}
	// Once nothing uses the old files, a replacement removes them.
	if err := f.service.Remove("sibling"); err != nil {
		t.Fatal(err)
	}
	if f.host.Exists(oldDir) {
		t.Error("the unused files stayed")
	}

	g := newFixture(t, nil)
	first := install(t, g, "/src/hello")
	g.host.AddFile("/src/hello/skills/greet/extra.md", "new", 0o644)
	install(t, g, "/src/hello")
	if g.host.Exists(storeDir + "/content/" + strings.TrimPrefix(first.Digest, "sha256:")) {
		t.Error("the replaced files stayed")
	}
}

func TestMarketplaceLifecycle(t *testing.T) {
	f := newFixture(t, nil)
	source, listing, err := f.service.AddMarketplace("/market")
	if err != nil || source.Name != "main" || source.Source != "/market" || len(listing.Entries) != 3 {
		t.Fatalf("AddMarketplace = %+v, %+v, %v", source, listing, err)
	}
	registered, err := f.service.Marketplaces()
	if err != nil || len(registered) != 1 {
		t.Errorf("Marketplaces = %+v, %v", registered, err)
	}
	browsed, err := f.service.Browse("main")
	if err != nil || browsed.Name != "main" {
		t.Errorf("Browse = %+v, %v", browsed, err)
	}

	fetched, err := f.service.Fetch("hello@main")
	if err != nil || fetched.Marketplace != "main" || fetched.Source != "/market" || fetched.FileCount != 2 {
		t.Fatalf("Fetch = %+v, %v", fetched, err)
	}
	record, err := f.service.Commit(fetched, fetched.Digest, nil)
	if err != nil || record.Marketplace != "main" {
		t.Fatalf("Commit = %+v, %v", record, err)
	}
	set, err := f.service.Load([]string{"hello"})
	if err != nil || len(set.Agents) != 1 {
		t.Errorf("Load = %+v, %v", set, err)
	}

	if err := f.service.RemoveMarketplace("main"); err == nil || !strings.Contains(err.Error(), "source of hello") {
		t.Errorf("RemoveMarketplace while used: %v", err)
	}
	if err := f.service.Remove("hello"); err != nil {
		t.Fatal(err)
	}
	if err := f.service.RemoveMarketplace("main"); err != nil {
		t.Fatal(err)
	}
	if err := f.service.RemoveMarketplace("main"); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Errorf("RemoveMarketplace twice: %v", err)
	}
	if _, err := f.service.Browse("main"); err == nil {
		t.Error("Browse of a removed marketplace passed")
	}
}

func TestListedInstallFailsClosed(t *testing.T) {
	f := newFixture(t, nil)
	if _, err := f.service.Fetch("hello@main"); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Errorf("unregistered: %v", err)
	}
	if _, _, err := f.service.AddMarketplace("/market"); err != nil {
		t.Fatal(err)
	}
	for spec, want := range map[string]string{
		"missing@main": "lists no plugin",
		"remote@main":  "source that is not a path",
		"escape@main":  "not clean and confined",
	} {
		if _, err := f.service.Fetch(spec); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Fetch(%s) = %v, want %q", spec, err, want)
		}
	}

	f.host.AddFile("/market/plugins/other/plugin.json", manifest("different"), 0o644)
	f.host.AddFile("/market/.claude-plugin/marketplace.json", `{"name":"main","plugins":[{"name":"other","source":"./plugins/other"},{"name":"linked","source":"./plugins/linked"}]}`, 0o644)
	f.host.AddLink("/market/plugins/linked", "/src/hello")
	if _, err := f.service.Fetch("other@main"); err == nil || !strings.Contains(err.Error(), "names itself") {
		t.Errorf("a renamed plugin passed: %v", err)
	}
	if _, err := f.service.Fetch("linked@main"); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Errorf("a symlinked plugin passed: %v", err)
	}
}

func TestMarketplaceSourcesFailClosed(t *testing.T) {
	f := newFixture(t, nil)
	f.host.AddDir("/empty")
	if _, _, err := f.service.AddMarketplace("/empty"); err == nil || !strings.Contains(err.Error(), "no marketplace file") {
		t.Errorf("an empty directory: %v", err)
	}
	if _, _, err := f.service.AddMarketplace("nowhere"); err == nil {
		t.Error("a missing source passed")
	}
	if _, err := f.service.Fetch("nowhere"); err == nil {
		t.Error("a missing plugin source passed")
	}
	if _, err := f.service.Browse("none"); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Errorf("Browse: %v", err)
	}
}

func TestSplitListed(t *testing.T) {
	for spec, ok := range map[string]bool{"a@b": true, "git@github.com:o/r": false, "a": false, "A@b": false, "a@": false, "@b": false, "https://x/y@z": false} {
		if _, _, got := splitListed(spec); got != ok {
			t.Errorf("splitListed(%q) = %v", spec, got)
		}
	}
}

func TestAPluginFromAnotherOriginRecordsTheResolvedRevision(t *testing.T) {
	f := newFixture(t, nil)
	f.service.Origins = origin.Registry{"stub": stubOrigin{dir: "/src/hello"}}
	fetched, err := f.service.Fetch("stub:x")
	if err != nil || fetched.Resolved != "abc123" || fetched.Source != "stub:x#main" {
		t.Fatalf("Fetch = %+v, %v", fetched, err)
	}
	record, err := f.service.Commit(fetched, fetched.Digest, nil)
	if err != nil || record.Resolved != "abc123" {
		t.Errorf("Commit = %+v, %v", record, err)
	}
}

type stubOrigin struct{ dir string }

func (stubOrigin) Accepts(spec string) bool { return strings.HasPrefix(spec, "stub:") }
func (s stubOrigin) Fetch(spec string) (origin.Checkout, error) {
	return origin.Checkout{Dir: s.dir, Source: spec + "#main", Resolved: "abc123", Release: func() {}}, nil
}

func TestEveryFailureOfTheMachineStopsTheCommand(t *testing.T) {
	steps := map[string]func(f fixture) error{
		"install": func(f fixture) error {
			fetched, err := f.service.Fetch("/src/hello")
			if err != nil {
				return err
			}
			_, err = f.service.Commit(fetched, fetched.Digest, nil)
			return err
		},
		"add marketplace": func(f fixture) error { _, _, err := f.service.AddMarketplace("/market"); return err },
	}
	for name, step := range steps {
		for _, op := range []string{"lstat", "readdir", "read", "write", "mkdir", "rename", "removeall"} {
			f := newFixture(t, func(host *systemtest.Fake) Files { return faulty{host, failOn(op)} })
			err := step(f)
			if err == nil {
				continue // the step never used this operation
			}
			if !errors.Is(err, errInjected) {
				t.Errorf("%s with %s failing: %v", name, op, err)
			}
		}
	}
}

func TestCorruptStoreFilesStopEveryCommand(t *testing.T) {
	f := newFixture(t, nil)
	f.host.AddFile(storeDir+"/lock.json", "{", 0o600).AddFile(storeDir+"/marketplaces.json", "{", 0o600)
	fetched, _ := f.service.Fetch("/src/hello")
	checks := map[string]error{
		"Commit":            func() error { _, err := f.service.Commit(fetched, fetched.Digest, nil); return err }(),
		"List":              func() error { _, err := f.service.List(); return err }(),
		"Remove":            f.service.Remove("hello"),
		"Load":              func() error { _, err := f.service.Load([]string{"hello"}); return err }(),
		"AddMarketplace":    func() error { _, _, err := f.service.AddMarketplace("/market"); return err }(),
		"Marketplaces":      func() error { _, err := f.service.Marketplaces(); return err }(),
		"Browse":            func() error { _, err := f.service.Browse("main"); return err }(),
		"RemoveMarketplace": f.service.RemoveMarketplace("main"),
		"Fetch listed":      func() error { _, err := f.service.Fetch("hello@main"); return err }(),
	}
	for name, err := range checks {
		if err == nil {
			t.Errorf("%s passed over a corrupt store", name)
		}
	}
}

func TestACorruptLockStopsRemoveMarketplaceAfterTheRegistryLoads(t *testing.T) {
	f := newFixture(t, nil)
	if _, _, err := f.service.AddMarketplace("/market"); err != nil {
		t.Fatal(err)
	}
	f.host.AddFile(storeDir+"/lock.json", "{", 0o600)
	if err := f.service.RemoveMarketplace("main"); err == nil {
		t.Error("a corrupt lock passed")
	}
}

func TestFailuresOfASingleStepStopTheCommand(t *testing.T) {
	// A registered marketplace whose source vanished.
	f := newFixture(t, nil)
	if _, _, err := f.service.AddMarketplace("/market"); err != nil {
		t.Fatal(err)
	}
	f.host.RemoveAll("/market")
	if _, err := f.service.Fetch("hello@main"); err == nil {
		t.Error("Fetch from a vanished marketplace passed")
	}
	if _, err := f.service.Browse("main"); err == nil {
		t.Error("Browse of a vanished marketplace passed")
	}

	// A listed plugin without a directory, and one with a broken manifest.
	g := newFixture(t, nil)
	g.host.AddFile("/market/.claude-plugin/marketplace.json", `{"name":"main","plugins":[{"name":"gone","source":"./plugins/gone"},{"name":"broken","source":"./plugins/broken"}]}`, 0o644).
		AddFile("/market/plugins/broken/plugin.json", "{", 0o644)
	if _, _, err := g.service.AddMarketplace("/market"); err != nil {
		t.Fatal(err)
	}
	for _, spec := range []string{"gone@main", "broken@main"} {
		if _, err := g.service.Fetch(spec); err == nil {
			t.Errorf("Fetch(%s) passed", spec)
		}
	}

	// A lock file that cannot be written after the files were stored.
	h := newFixture(t, func(host *systemtest.Fake) Files {
		return faulty{host, func(op, p string) error {
			if op == "rename" && strings.HasSuffix(p, "lock.json") {
				return errInjected
			}
			return nil
		}}
	})
	fetched, _ := h.service.Fetch("/src/hello")
	if _, err := h.service.Commit(fetched, fetched.Digest, nil); !errors.Is(err, errInjected) {
		t.Errorf("Commit: %v", err)
	}
	// A stored plugin whose lock record outlives the files of its manifest.
	i := newFixture(t, nil)
	record := install(t, i, "/src/hello")
	if err := i.service.Remove("hello"); err != nil {
		t.Fatal(err)
	}
	i.host.AddFile("/src/bare/skills/x/SKILL.md", "---\nname: x\ndescription: d\n---\n", 0o644)
	tree, err := origin.ReadTree(i.host, "/src/bare")
	if err != nil {
		t.Fatal(err)
	}
	if err := i.service.Store.Put(tree); err != nil {
		t.Fatal(err)
	}
	record.Name, record.Digest = "bare", tree.Digest()
	lock, _ := i.service.Store.Lock()
	if err := i.service.Store.SaveLock(lock.With(record)); err != nil {
		t.Fatal(err)
	}
	if _, err := i.service.Load([]string{"bare"}); err == nil {
		t.Error("a stored plugin without a manifest passed")
	}
}

func TestRemoveFailsClosedWhenTheLockCannotBeWritten(t *testing.T) {
	f := newFixture(t, nil)
	install(t, f, "/src/hello")
	failing := Service{Store: f.service.Store, Files: f.service.Files, Origins: f.service.Origins}
	failing.Store.files = faulty{f.host, failOn("rename")}
	if err := failing.Remove("hello"); !errors.Is(err, errInjected) {
		t.Errorf("Remove: %v", err)
	}
}

func TestAListedPluginThatCannotBeReadStopsTheInstall(t *testing.T) {
	f := newFixture(t, func(host *systemtest.Fake) Files {
		return faulty{host, func(op, p string) error {
			if op == "readdir" && p == "/market/plugins/hello" {
				return errInjected
			}
			return nil
		}}
	})
	if _, _, err := f.service.AddMarketplace("/market"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Fetch("hello@main"); !errors.Is(err, errInjected) {
		t.Errorf("Fetch: %v", err)
	}
}
