//go:build integration
// +build integration

package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/egress/managed"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost"
)

func TestRunHelpListsTheManagedNetworkFlags(t *testing.T) {
	output, err := testCommand(t, "run", "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("run --help: %v\n%s", err, output)
	}
	for _, want := range []string{"--allow-host", "--credential", "managed", "seatbelt", "docker-sandboxes", "openshell"} {
		if !strings.Contains(string(output), want) {
			t.Errorf("help lacks %q", want)
		}
	}
}

func TestTheRetiredProxyNetworkIsRejected(t *testing.T) {
	output, err := testCommand(t, "run", "--dry-run", "--network", "proxy").CombinedOutput()
	if err == nil || !strings.Contains(string(output), "network") {
		t.Errorf("--network proxy must fail with a network error: %v\n%s", err, output)
	}
}

func TestManagedNetworkFailsClosedWithoutAnAllowlistOrATransport(t *testing.T) {
	cases := map[string][]string{
		"no allowed host":        {"run", "--dry-run", "--isolation", "bwrap", "--network", "managed"},
		"no transport on none":   {"run", "--dry-run", "--network", "managed", "--allow-host", "api.example.com"},
		"hosts without managed":  {"run", "--dry-run", "--allow-host", "api.example.com"},
		"unavailable credential": {"run", "--credential", "API_KEY=env:AGENT_CLI_SURELY_UNSET_VARIABLE"},
	}
	for name, args := range cases {
		output, err := testCommand(t, args...).CombinedOutput()
		if err == nil {
			t.Errorf("%s: the command succeeded:\n%s", name, output)
		}
	}
}

func TestManagedDryRunNamesTheTransportWhenBwrapIsAvailable(t *testing.T) {
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bwrap is not installed")
	}
	cmd := testCommand(t, "run", "--dry-run", "--isolation", "bwrap", "--isolation-bwrap-passthrough", "--network", "managed",
		"--allow-host", "api.example.com", "--credential", "API_KEY=env:HOME")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dry run: %v\n%s", err, output)
	}
	for _, want := range []string{"Network: managed", "api.example.com:443", "unix-socket", "API_KEY from env:HOME ([redacted])", "egress-bridge"} {
		if !strings.Contains(string(output), want) {
			t.Errorf("dry run lacks %q:\n%s", want, output)
		}
	}
}

func TestEgressBridgeRunsAndExitsWithTheCodeOfItsCommand(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "absent.sock")
	cmd := testCommand(t, "egress-bridge", "--socket", socket, "--listen", "127.0.0.1:0", "--", "sh", "-c", "exit 4")
	err := cmd.Run()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 4 {
		t.Errorf("exit = %v, want code 4", err)
	}
}

// TestBwrapSandboxReachesOnlyTheAllowedHost proves the guarantee the managed mode
// makes: inside an isolated network namespace a direct socket to the host fails,
// and the bridge reaches the allowed host through the proxy.
func TestBwrapSandboxReachesOnlyTheAllowedHost(t *testing.T) {
	machine := oshost.New(os.Stdin, io.Discard, io.Discard)
	bwrap, err := machine.LookPath("bwrap")
	if err != nil {
		t.Skip("bwrap is not installed")
	}
	bash, err := machine.LookPath("bash")
	if err != nil {
		t.Skip("bash is not installed")
	}
	if code, err := machine.Run(bwrap, []string{"--unshare-net", "--ro-bind", "/", "/", "--", "true"}, nil, ""); err != nil || code != 0 {
		t.Skip("bwrap cannot create a network namespace here")
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "reached") }))
	defer upstream.Close()
	parsed, _ := url.Parse(upstream.URL)
	allow, err := egress.NewAllowlist([]string{parsed.Host})
	if err != nil {
		t.Fatal(err)
	}
	route, stop, err := managed.Service{Host: machine, Log: logging.New(io.Discard, io.Discard, false, false)}.Open(network.TransportUnixSocket, allow)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	socketDir := filepath.Dir(route.Socket)

	// The sandbox script makes two attempts: a direct connection to the upstream
	// on the host loopback, which has no route inside the namespace, and a request
	// through the bridge.
	script := `
if (exec 3<>/dev/tcp/` + parsed.Hostname() + `/` + parsed.Port() + `) 2>/dev/null; then echo DIRECT-REACHED; fi
exec 4<>/dev/tcp/127.0.0.1/3128
printf 'GET ` + upstream.URL + `/ HTTP/1.1\r\nHost: ` + parsed.Host + `\r\nConnection: close\r\n\r\n' >&4
cat <&4`
	args := []string{"--unshare-net", "--ro-bind", "/", "/", "--dev", "/dev", "--proc", "/proc", "--bind", socketDir, socketDir,
		"--", testBinary(t), "egress-bridge", "--socket", route.Socket, "--listen", "127.0.0.1:3128", "--", bash, "-c", script}
	output, err := exec.Command(bwrap, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("bwrap: %v\n%s", err, output)
	}
	text := string(output)
	if strings.Contains(text, "DIRECT-REACHED") {
		t.Error("a direct socket reached the host: the namespace must have no route")
	}
	if !strings.Contains(text, "reached") {
		t.Errorf("the bridge did not reach the allowed host:\n%s", text)
	}
}
