//go:build integration

package managed

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system"
)

var integrationLog = logging.New(io.Discard, io.Discard, false, false)

// upstream serves one fixed answer on a loopback port that an allowlist rule can
// name by address, because the dialer refuses a loopback host name.
func upstream(t *testing.T) (host, port string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "upstream-answer")
	}))
	t.Cleanup(server.Close)
	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	return parsed.Hostname(), parsed.Port()
}

func openProxy(t *testing.T, rules ...string) isoRoute {
	t.Helper()
	allow, err := egress.NewAllowlist(rules)
	if err != nil {
		t.Fatal(err)
	}
	host := system.New(os.Stdin, io.Discard, io.Discard)
	route, stop, err := Service{Host: host, Log: integrationLog}.Open(netshared.TransportUnixSocket, allow)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stop)
	return isoRoute{socket: route.Socket}
}

type isoRoute struct{ socket string }

func TestProxyForwardsAnAllowedRequestOverTheUnixSocket(t *testing.T) {
	host, port := upstream(t)
	route := openProxy(t, host+":"+port)
	conn, err := net.Dial("unix", route.socket)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	request := "GET http://" + host + ":" + port + "/ HTTP/1.1\r\nHost: " + host + ":" + port + "\r\nConnection: close\r\n\r\n"
	if _, err := io.WriteString(conn, request); err != nil {
		t.Fatal(err)
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK || string(body) != "upstream-answer" {
		t.Errorf("response = %d %q", response.StatusCode, body)
	}
}

func TestProxyTunnelsAnAllowedConnectRequest(t *testing.T) {
	host, port := upstream(t)
	route := openProxy(t, host+":"+port)
	conn, err := net.Dial("unix", route.socket)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	target := host + ":" + port
	_, _ = io.WriteString(conn, "CONNECT "+target+" HTTP/1.1\r\nHost: "+target+"\r\n\r\n")
	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(status, "200") {
		t.Fatalf("status = %q, %v", status, err)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
	}
	_, _ = io.WriteString(conn, "GET / HTTP/1.1\r\nHost: "+target+"\r\nConnection: close\r\n\r\n")
	response, err := http.ReadResponse(reader, nil)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	if string(body) != "upstream-answer" {
		t.Errorf("tunnelled body = %q", body)
	}
}

func TestProxyRefusesWhatTheAllowlistDoesNotName(t *testing.T) {
	host, port := upstream(t)
	route := openProxy(t, host+":"+port)
	for _, target := range []string{"evil.example.com:443", host + ":1"} {
		conn, err := net.Dial("unix", route.socket)
		if err != nil {
			t.Fatal(err)
		}
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
		_, _ = io.WriteString(conn, "CONNECT "+target+" HTTP/1.1\r\nHost: "+target+"\r\n\r\n")
		status, _ := bufio.NewReader(conn).ReadString('\n')
		_ = conn.Close()
		if !strings.Contains(status, "403") {
			t.Errorf("CONNECT %s status = %q, want 403", target, status)
		}
	}
}

func TestProxyRefusesALoopbackHostNameEvenWhenAllowed(t *testing.T) {
	_, port := upstream(t)
	route := openProxy(t, "localhost:"+port)
	conn, err := net.Dial("unix", route.socket)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	_, _ = io.WriteString(conn, "CONNECT localhost:"+port+" HTTP/1.1\r\nHost: localhost:"+port+"\r\n\r\n")
	status, _ := bufio.NewReader(conn).ReadString('\n')
	if !strings.Contains(status, "502") {
		t.Errorf("status = %q, want 502: an allowed name must not reach a service of this machine", status)
	}
}

func TestStopRemovesTheSocket(t *testing.T) {
	host := system.New(os.Stdin, io.Discard, io.Discard)
	allow, _ := egress.NewAllowlist([]string{"api.example.com"})
	route, stop, err := Service{Host: host, Log: integrationLog}.Open(netshared.TransportUnixSocket, allow)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(route.Socket); err != nil {
		t.Fatalf("the socket must exist while the proxy runs: %v", err)
	}
	stop()
	if _, err := os.Stat(route.Socket); !os.IsNotExist(err) {
		t.Errorf("the socket must be gone after stop: %v", err)
	}
}

func TestBridgeRelaysALoopbackConnectionToTheProxy(t *testing.T) {
	var stdout bytes.Buffer
	machine := system.New(os.Stdin, &stdout, io.Discard)
	bash, err := machine.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	host, port := upstream(t)
	route := openProxy(t, host+":"+port)
	free, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	bridgeAddress := free.Addr().String()
	_, bridgePort, _ := net.SplitHostPort(bridgeAddress)
	_ = free.Close()
	script := `exec 3<>/dev/tcp/127.0.0.1/` + bridgePort + `
printf 'GET http://` + host + `:` + port + `/ HTTP/1.1\r\nHost: ` + host + `:` + port + `\r\nConnection: close\r\n\r\n' >&3
cat <&3`
	code, err := Bridge{Host: machine}.Run(route.socket, bridgeAddress, []string{bash, "-c", script})
	if err != nil || code != 0 {
		t.Fatalf("Run = %d, %v", code, err)
	}
	if !strings.Contains(stdout.String(), "upstream-answer") {
		t.Errorf("output = %q", stdout.String())
	}
}
