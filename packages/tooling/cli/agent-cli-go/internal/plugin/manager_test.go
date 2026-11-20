package plugin

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"

	libplugin "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/plugin"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/plugins"
)

// unplacedHost fails to name the working or the home directory.
type unplacedHost struct {
	*oshosttest.Fake
	noCwd, noHome bool
}

func (h unplacedHost) Getwd() (string, error) {
	if h.noCwd {
		return "", errors.New("no cwd")
	}
	return h.Fake.Getwd()
}

func (h unplacedHost) UserHomeDir() (string, error) {
	if h.noHome {
		return "", errors.New("no home")
	}
	return h.Fake.UserHomeDir()
}

func newManager(t *testing.T) (Manager, *oshosttest.Fake, *bytes.Buffer) {
	t.Helper()
	f := newFixture(t, nil)
	var out bytes.Buffer
	return Manager{Host: f.host, Origins: plugins.DefaultRegistry(f.host), Out: &out}, f.host, &out
}

func TestOpenChoosesTheStore(t *testing.T) {
	manager, host, _ := newManager(t)
	for dir, want := range map[string]string{
		"":              "/home/user/.local/share/agentic/plugins",
		"relative/s":    "/work/relative/s",
		"/abs/../store": "/store",
	} {
		service, err := manager.Open(dir)
		if err != nil || service.Store.Dir() != want {
			t.Errorf("Open(%q) = %q, %v, want %q", dir, service.Store.Dir(), err, want)
		}
	}
	host.Env["XDG_DATA_HOME"] = "/data"
	if service, err := manager.Open(""); err != nil || service.Store.Dir() != "/data/agentic/plugins" {
		t.Errorf("Open() under XDG_DATA_HOME = %q, %v", service.Store.Dir(), err)
	}
}

func TestOpenFailsWithoutTheDirectoriesItNeeds(t *testing.T) {
	for name, host := range map[string]unplacedHost{
		"no working directory": {Fake: oshosttest.New(), noCwd: true},
		"no home directory":    {Fake: oshosttest.New(), noHome: true},
	} {
		if _, err := (Manager{Host: host}).Open(""); err == nil {
			t.Errorf("%s: Open() error = nil", name)
		}
	}
	if _, err := (Manager{Host: unplacedHost{Fake: oshosttest.New(), noHome: true}}).Open("s"); err != nil {
		t.Errorf("a relative store needs no home directory: %v", err)
	}
	relative := oshosttest.New()
	relative.WorkingDir = "relative"
	if _, err := (Manager{Host: relative}).Open("s"); err == nil {
		t.Error("Open() under a relative working directory error = nil")
	}
}

func TestInstallPreviewsThenInstallsOnlyTheApprovedDigest(t *testing.T) {
	const notice = "The plugin hello from /src/hello is installed only after you approve its digest"
	manager, _, _ := newManager(t)
	var previewed []string
	preview := func(fetched Fetched) { previewed = append(previewed, fetched.Digest) }

	_, err := manager.Install(InstallRequest{Store: storeDir, Source: "/src/hello"}, preview)

	if err == nil || len(previewed) != 1 || !strings.Contains(err.Error(), notice) || !strings.Contains(err.Error(), "--approve-config-digest "+previewed[0]) {
		t.Fatalf("Install() without an approval = %v, previewed %v", err, previewed)
	}
	record, err := manager.Install(InstallRequest{Store: storeDir, Source: "/src/hello", ApproveDigest: previewed[0]}, preview)
	if err != nil || record.Name != "hello" || record.Digest != previewed[0] {
		t.Fatalf("Install() with the digest = %+v, %v", record, err)
	}

	asked, host, out := newManager(t)
	host.Terminal = true
	host.Input = []string{"y"}
	if _, err := asked.Install(InstallRequest{Store: storeDir, Source: "/src/hello"}, preview); err != nil {
		t.Fatalf("Install() approved at the terminal error = %v", err)
	}
	if !strings.Contains(out.String(), notice) {
		t.Errorf("the trust gate asked without its notice: %q", out.String())
	}
}

func TestInstallFailsClosedBeforeItFetches(t *testing.T) {
	manager, _, _ := newManager(t)
	fetched := false
	preview := func(Fetched) { fetched = true }
	for name, request := range map[string]InstallRequest{
		"unknown grant":  {Store: storeDir, Source: "/src/hello", Allow: []string{"everything"}},
		"missing source": {Store: storeDir, Source: "/src/none"},
	} {
		if _, err := manager.Install(request, preview); err == nil {
			t.Errorf("%s: Install() error = nil", name)
		}
	}
	unplaced := Manager{Host: unplacedHost{Fake: oshosttest.New(), noCwd: true}}
	if _, err := unplaced.Install(InstallRequest{Source: "/src/hello"}, preview); err == nil {
		t.Error("Install() without a working directory error = nil")
	}
	if fetched {
		t.Error("a failed install previewed a plugin")
	}
	if _, err := manager.Install(InstallRequest{Store: storeDir, Source: "/src/hello", Allow: []string{string(libplugin.CapabilityHooks)}}, preview); err == nil || !fetched {
		t.Errorf("an unapproved install with a grant = %v, previewed %t", err, fetched)
	}
}

func TestCapabilityNamesListEveryGrant(t *testing.T) {
	names := CapabilityNames()
	if len(names) != len(libplugin.Capabilities()) || !slices.Contains(names, string(libplugin.CapabilityHooks)) {
		t.Errorf("CapabilityNames() = %v", names)
	}
}
