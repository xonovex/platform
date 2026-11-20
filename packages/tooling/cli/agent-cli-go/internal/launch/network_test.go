package launch

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	credshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
)

// fakeOpener records the managed egress a run starts and stops.
type fakeOpener struct {
	transport netshared.Transport
	allow     egress.Allowlist
	route     isoshared.EgressRoute
	err       error
	opened    int
	stopped   int
}

func (o *fakeOpener) Open(transport netshared.Transport, allow egress.Allowlist) (isoshared.EgressRoute, func(), error) {
	o.opened++
	o.transport, o.allow = transport, allow
	if o.err != nil {
		return isoshared.EgressRoute{}, nil, o.err
	}
	return o.route, func() { o.stopped++ }, nil
}

func managedOptions() Options {
	options := bwrapOptions()
	options.Network = "managed"
	options.AllowHosts = []string{"api.example.com"}
	return options
}

func TestExecuteStartsManagedEgressAndStopsItAfterTheRun(t *testing.T) {
	f := newExecuteFixture()
	opener := &fakeOpener{route: isoshared.EgressRoute{Transport: netshared.TransportUnixSocket, Socket: "/run/e/p.sock", Bridge: "/bin/agent-cli"}}
	deps := f.deps()
	deps.Egress = opener
	if err := Execute(deps, managedOptions()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	iso := f.sandbox.isolator
	if opener.transport != netshared.TransportUnixSocket || !opener.allow.Allows("api.example.com", 443) {
		t.Errorf("opened for %q %v", opener.transport, opener.allow.Rules())
	}
	if iso.commandRun.Egress.Socket != "/run/e/p.sock" || !iso.ran {
		t.Errorf("run egress = %+v, ran = %v", iso.commandRun.Egress, iso.ran)
	}
	if opener.opened != 1 || opener.stopped != 1 {
		t.Errorf("opened %d, stopped %d; want one each", opener.opened, opener.stopped)
	}
}

func TestExecuteStopsManagedEgressAfterAFailedRun(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.runErr = errors.New("boom")
	opener := &fakeOpener{}
	deps := f.deps()
	deps.Egress = opener
	if err := Execute(deps, managedOptions()); err == nil {
		t.Fatal("Execute() succeeded")
	}
	if opener.stopped != 1 {
		t.Errorf("stopped %d times, want 1", opener.stopped)
	}
}

func TestExecuteFailsWithoutManagedEgress(t *testing.T) {
	f := newExecuteFixture()
	if err := Execute(f.deps(), managedOptions()); err == nil || !strings.Contains(err.Error(), "no managed egress") {
		t.Errorf("a build without egress error = %v", err)
	}
	deps := f.deps()
	deps.Egress = &fakeOpener{err: errors.New("address in use")}
	if err := Execute(deps, managedOptions()); err == nil || !strings.Contains(err.Error(), "address in use") {
		t.Errorf("an opener failure error = %v", err)
	}
	if f.sandbox.isolator.ran {
		t.Error("the agent must not run when egress cannot start")
	}
}

func TestExecuteDryRunReportsTheNetworkAndTheCredentials(t *testing.T) {
	f := newExecuteFixture()
	opener := &fakeOpener{}
	deps := f.deps()
	deps.Egress = opener
	options := managedOptions()
	options.DryRun = true
	options.Credentials = []string{"API_KEY=keychain:service/account"}
	if err := Execute(deps, options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, want := range []string{"Network: managed", "Allowed hosts: api.example.com:443", "Egress transport: unix-socket", "Credential: API_KEY from keychain:service/account ([redacted])"} {
		if !strings.Contains(f.log.String(), want) {
			t.Errorf("log lacks %q:\n%s", want, f.log.String())
		}
	}
	if opener.stopped != opener.opened {
		t.Error("a dry run must stop what it started")
	}
	if got := f.sandbox.isolator.commandRun.Secrets["API_KEY"]; got != credshared.Redacted {
		t.Errorf("a dry run must carry the placeholder, got %q", got)
	}
}

type fakeStore struct {
	value string
	err   error
	seen  *string
}

func (s fakeStore) Available() bool { return true }
func (s fakeStore) Lookup(locator string) (credshared.Secret, error) {
	if s.seen != nil {
		*s.seen = locator
	}
	if s.err != nil {
		return credshared.Secret{}, s.err
	}
	return credshared.NewSecret(s.value), nil
}

func TestExecuteResolvesCredentialsForTheRunOnly(t *testing.T) {
	f := newExecuteFixture()
	var locator string
	deps := f.deps()
	deps.Credentials = credshared.NewRegistry().Register("test", func() credshared.Store { return fakeStore{value: "s3cret", seen: &locator} })
	options := bwrapOptions()
	options.Credentials = []string{"API_KEY=test:vault/key"}
	if err := Execute(deps, options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if locator != "vault/key" || f.sandbox.isolator.commandRun.Secrets["API_KEY"] != "s3cret" {
		t.Errorf("locator = %q, secrets = %v", locator, f.sandbox.isolator.commandRun.Secrets)
	}
	if strings.Contains(f.log.String(), "s3cret") || strings.Contains(f.out.String(), "s3cret") {
		t.Error("a credential value reached the log or the output")
	}
}

func TestExecuteDryRunResolvesNoCredential(t *testing.T) {
	f := newExecuteFixture()
	var locator string
	deps := f.deps()
	deps.Credentials = credshared.NewRegistry().Register("test", func() credshared.Store { return fakeStore{value: "s3cret", seen: &locator} })
	options := bwrapOptions()
	options.DryRun = true
	options.Credentials = []string{"API_KEY=test:vault/key"}
	if err := Execute(deps, options); err != nil {
		t.Fatal(err)
	}
	if locator != "" {
		t.Error("a dry run must not read the store")
	}
}

func TestExecuteFailsClosedOnACredential(t *testing.T) {
	options := bwrapOptions()
	options.Credentials = []string{"API_KEY=test:vault/key"}

	f := newExecuteFixture()
	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), "no credential stores") {
		t.Errorf("a build without stores error = %v", err)
	}
	deps := f.deps()
	deps.Credentials = credshared.NewRegistry().Register("test", func() credshared.Store { return fakeStore{err: credshared.ErrNotFound} })
	if err := Execute(deps, options); !errors.Is(err, credshared.ErrNotFound) {
		t.Errorf("a missing credential error = %v", err)
	}
	if f.sandbox.isolator.ran {
		t.Error("the agent must not start without a credential the user asked for")
	}
}
