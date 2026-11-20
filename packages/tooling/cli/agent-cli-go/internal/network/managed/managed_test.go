package managed

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

var testLog = logging.New(io.Discard, io.Discard, false, false)

func allowlist(t *testing.T, rules ...string) egress.Allowlist {
	t.Helper()
	list, err := egress.NewAllowlist(rules)
	if err != nil {
		t.Fatal(err)
	}
	return list
}

// dialer records the addresses the proxy asked for and answers with pipes.
type dialer struct {
	public  []string
	any     []string
	upstrm  chan net.Conn
	failAll bool
}

func newDialer() *dialer { return &dialer{upstrm: make(chan net.Conn, 4)} }

func (d *dialer) open() (net.Conn, error) {
	if d.failAll {
		return nil, errors.New("unreachable")
	}
	near, far := net.Pipe()
	d.upstrm <- far
	return near, nil
}

func (d *dialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.any = append(d.any, address)
	return d.open()
}

func (d *dialer) DialPublic(_ context.Context, _, address string) (net.Conn, error) {
	d.public = append(d.public, address)
	return d.open()
}

// hijackWriter is a ResponseWriter whose connection a test holds.
type hijackWriter struct {
	*httptest.ResponseRecorder
	server  net.Conn
	refuse  bool
	hijacks int
}

func (h *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijacks++
	if h.refuse {
		return nil, nil, errors.New("no hijack")
	}
	return h.server, bufio.NewReadWriter(bufio.NewReader(h.server), bufio.NewWriter(h.server)), nil
}

func connectRequest(target string) *http.Request {
	r := httptest.NewRequest(http.MethodConnect, "http://"+target, nil)
	r.Host = target
	return r
}

func TestTunnelRelaysAnAllowedHost(t *testing.T) {
	d := newDialer()
	proxy := Proxy{Allow: allowlist(t, "api.example.com"), Dialer: d, Log: testLog}
	client, server := net.Pipe()
	w := &hijackWriter{ResponseRecorder: httptest.NewRecorder(), server: server}
	done := make(chan struct{})
	go func() { proxy.ServeHTTP(w, connectRequest("api.example.com:443")); close(done) }()

	reader := bufio.NewReader(client)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "HTTP/1.1 200") {
		t.Fatalf("status line = %q, %v", line, err)
	}
	if _, err := reader.ReadString('\n'); err != nil {
		t.Fatal(err)
	}
	upstream := <-d.upstrm
	go func() { _, _ = client.Write([]byte("ping")) }()
	got := make([]byte, 4)
	if _, err := io.ReadFull(upstream, got); err != nil || string(got) != "ping" {
		t.Fatalf("upstream got %q, %v", got, err)
	}
	_ = upstream.Close()
	<-done
	_ = client.Close()
	if len(d.public) != 1 || d.public[0] != "api.example.com:443" || len(d.any) != 0 {
		t.Errorf("dials: public=%v any=%v, want one public dial", d.public, d.any)
	}
}

func TestTunnelUsesTheUnrestrictedDialerOnlyForAnAddressRule(t *testing.T) {
	d := newDialer()
	proxy := Proxy{Allow: allowlist(t, "192.0.2.10:9000"), Dialer: d, Log: testLog}
	client, server := net.Pipe()
	w := &hijackWriter{ResponseRecorder: httptest.NewRecorder(), server: server}
	go proxy.ServeHTTP(w, connectRequest("192.0.2.10:9000"))
	if _, err := bufio.NewReader(client).ReadString('\n'); err != nil {
		t.Fatal(err)
	}
	_ = client.Close()
	if len(d.any) != 1 || len(d.public) != 0 {
		t.Errorf("dials: public=%v any=%v, want one unrestricted dial", d.public, d.any)
	}
}

func TestTunnelRefusals(t *testing.T) {
	cases := []struct {
		name   string
		target string
		dial   func(*dialer)
		hijack func(*hijackWriter)
		status int
	}{
		{"host off the allowlist", "evil.example.com:443", nil, nil, http.StatusForbidden},
		{"port off the allowlist", "api.example.com:80", nil, nil, http.StatusForbidden},
		{"unparsable target", "api.example.com:port", nil, nil, http.StatusBadRequest},
		{"unreachable upstream", "api.example.com:443", func(d *dialer) { d.failAll = true }, nil, http.StatusBadGateway},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := newDialer()
			if c.dial != nil {
				c.dial(d)
			}
			proxy := Proxy{Allow: allowlist(t, "api.example.com"), Dialer: d, Log: testLog}
			w := &hijackWriter{ResponseRecorder: httptest.NewRecorder()}
			proxy.ServeHTTP(w, connectRequest(c.target))
			if w.Code != c.status || w.hijacks != 0 {
				t.Errorf("status = %d hijacks = %d, want %d and none", w.Code, w.hijacks, c.status)
			}
		})
	}
}

func TestTunnelFailsWhenTheConnectionCannotBeHijacked(t *testing.T) {
	d := newDialer()
	proxy := Proxy{Allow: allowlist(t, "api.example.com"), Dialer: d, Log: testLog}
	plain := httptest.NewRecorder() // not a Hijacker
	proxy.ServeHTTP(plain, connectRequest("api.example.com:443"))
	if plain.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", plain.Code)
	}
	w := &hijackWriter{ResponseRecorder: httptest.NewRecorder(), refuse: true}
	proxy.ServeHTTP(w, connectRequest("api.example.com:443"))
	upstream := <-d.upstrm
	if _, err := upstream.Write([]byte("x")); err == nil {
		t.Error("the upstream connection must be closed after a failed hijack")
	}
}

type brokenConn struct{ net.Conn }

func (brokenConn) Write([]byte) (int, error) { return 0, errors.New("broken") }

type brokenHijacker struct {
	*httptest.ResponseRecorder
	conn net.Conn
}

func (b brokenHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return brokenConn{b.conn}, bufio.NewReadWriter(bufio.NewReader(b.conn), bufio.NewWriter(b.conn)), nil
}

func TestTunnelClosesBothSidesWhenTheClientCannotBeAnswered(t *testing.T) {
	d := newDialer()
	proxy := Proxy{Allow: allowlist(t, "api.example.com"), Dialer: d, Log: testLog}
	near, far := net.Pipe()
	defer far.Close()
	proxy.ServeHTTP(brokenHijacker{httptest.NewRecorder(), near}, connectRequest("api.example.com:443"))
	upstream := <-d.upstrm
	if _, err := upstream.Write([]byte("x")); err == nil {
		t.Error("the upstream connection must be closed")
	}
}

func TestForwardRelaysAnAllowedPlainRequest(t *testing.T) {
	d := newDialer()
	proxy := Proxy{Allow: allowlist(t, "files.example.com:80"), Dialer: d, Log: testLog}
	go func() {
		upstream := <-d.upstrm
		request, err := http.ReadRequest(bufio.NewReader(upstream))
		if err != nil {
			return
		}
		if request.Header.Get("Proxy-Connection") != "" {
			_, _ = io.WriteString(upstream, "HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n")
			return
		}
		_, _ = io.WriteString(upstream, "HTTP/1.1 200 OK\r\nX-Reply: yes\r\nConnection: close\r\nContent-Length: 5\r\n\r\nhello")
		_ = upstream.Close()
	}()
	request := httptest.NewRequest(http.MethodGet, "http://files.example.com/a", nil)
	request.Header.Set("Proxy-Connection", "keep-alive")
	recorder := httptest.NewRecorder()
	proxy.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || recorder.Body.String() != "hello" || recorder.Header().Get("X-Reply") != "yes" {
		t.Errorf("response = %d %q %v", recorder.Code, recorder.Body.String(), recorder.Header())
	}
	if recorder.Header().Get("Connection") != "" {
		t.Error("hop-by-hop headers must not reach the client")
	}
}

func TestForwardRefusals(t *testing.T) {
	d := newDialer()
	proxy := Proxy{Allow: allowlist(t, "files.example.com:80"), Dialer: d, Log: testLog}
	cases := []struct {
		name   string
		target string
		status int
	}{
		{"a relative request", "/a", http.StatusBadRequest},
		{"a host off the allowlist", "http://other.example.com/a", http.StatusForbidden},
		{"a port off the allowlist", "https://files.example.com/a", http.StatusForbidden},
		{"a bad port", "http://files.example.com/a", http.StatusBadRequest},
	}
	for _, c := range cases {
		request := httptest.NewRequest(http.MethodGet, c.target, nil)
		if c.name == "a bad port" {
			request.URL.Host = "files.example.com:xx"
		}
		recorder := httptest.NewRecorder()
		proxy.ServeHTTP(recorder, request)
		if recorder.Code != c.status {
			t.Errorf("%s: status = %d, want %d", c.name, recorder.Code, c.status)
		}
	}
	d.failAll = true
	recorder := httptest.NewRecorder()
	proxy.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://files.example.com/a", nil))
	if recorder.Code != http.StatusBadGateway {
		t.Errorf("unreachable upstream: status = %d, want 502", recorder.Code)
	}
}

func TestTarget(t *testing.T) {
	cases := []struct {
		in   string
		host string
		port int
		bad  bool
	}{
		{"a.example.com:8443", "a.example.com", 8443, false},
		{"a.example.com", "a.example.com", 443, false},
		{"[2001:db8::1]", "2001:db8::1", 443, false},
		{"a.example.com:x", "", 0, true},
		{"a:b:c", "", 0, true},
	}
	for _, c := range cases {
		host, port, err := target(c.in, 443)
		if (err != nil) != c.bad || host != c.host || port != c.port {
			t.Errorf("target(%q) = %q, %d, %v", c.in, host, port, err)
		}
	}
}

func readStatus(t *testing.T, conn net.Conn) string {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	return line
}

func TestServiceRunsAUnixSocketProxyAndCleansUp(t *testing.T) {
	host := systemtest.New()
	host.ExecutablePath = "/opt/agent-cli"
	service := Service{Host: host, Log: testLog}
	route, stop, err := service.Open(netshared.TransportUnixSocket, allowlist(t, "api.example.com"))
	if err != nil {
		t.Fatal(err)
	}
	if route.Bridge != "/opt/agent-cli" || !strings.HasSuffix(route.Socket, "/proxy.sock") || route.Transport != netshared.TransportUnixSocket {
		t.Fatalf("route = %+v", route)
	}
	listener, ok := host.Listener("unix", route.Socket)
	if !ok {
		t.Fatal("the proxy did not listen on its socket")
	}
	conn, err := listener.Dial()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, _ = io.WriteString(conn, "CONNECT denied.example.com:443 HTTP/1.1\r\nHost: denied.example.com:443\r\n\r\n")
	}()
	if status := readStatus(t, conn); !strings.Contains(status, "403") {
		t.Errorf("status = %q, want 403", status)
	}
	_ = conn.Close()
	dir := route.Socket[:strings.LastIndex(route.Socket, "/")]
	if !host.Exists(dir) {
		t.Fatal("the socket directory must exist while the proxy runs")
	}
	stop()
	if host.Exists(dir) || !listener.Closed() {
		t.Error("stop must remove the socket directory and close the listener")
	}
}

func TestServiceRunsALoopbackProxy(t *testing.T) {
	host := systemtest.New()
	route, stop, err := Service{Host: host, Log: testLog}.Open(netshared.TransportLoopback, allowlist(t, "api.example.com"))
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if route.Proxy != "127.0.0.1:40000" || route.Socket != "" {
		t.Fatalf("route = %+v", route)
	}
}

func TestServicePolicyTransportRunsNoProxy(t *testing.T) {
	host := systemtest.New()
	route, stop, err := Service{Host: host, Log: testLog}.Open(netshared.TransportPolicy, allowlist(t, "api.example.com"))
	if err != nil {
		t.Fatal(err)
	}
	stop()
	if route.Transport != netshared.TransportPolicy || len(route.Allow.Rules()) != 1 {
		t.Errorf("route = %+v", route)
	}
	if _, _, err := (Service{Host: host, Log: testLog}).Open(netshared.TransportPolicy, egress.Allowlist{}); err == nil {
		t.Error("a policy route without an allowlist must fail")
	}
}

func TestServiceFailures(t *testing.T) {
	list := allowlist(t, "api.example.com")
	noExecutable := systemtest.New()
	if _, _, err := (Service{Host: noExecutable, Log: testLog}).Open(netshared.TransportUnixSocket, list); err == nil {
		t.Error("a missing executable must fail")
	}
	noListen := systemtest.New()
	noListen.ExecutablePath = "/opt/agent-cli"
	noListen.ListenErr = errors.New("address in use")
	if _, _, err := (Service{Host: noListen, Log: testLog}).Open(netshared.TransportUnixSocket, list); err == nil {
		t.Error("a failed unix listen must fail")
	}
	if _, _, err := (Service{Host: noListen, Log: testLog}).Open(netshared.TransportLoopback, list); err == nil {
		t.Error("a failed loopback listen must fail")
	}
	if len(noListen.Calls) != 0 {
		t.Errorf("calls = %v", noListen.Calls)
	}
}

type failingTempHost struct {
	*systemtest.Fake
}

func (failingTempHost) MkdirTemp(string) (string, error) { return "", errors.New("disk full") }

func TestServiceFailsWhenTheSocketDirectoryCannotBeMade(t *testing.T) {
	fake := systemtest.New()
	fake.ExecutablePath = "/opt/agent-cli"
	if _, _, err := (Service{Host: failingTempHost{fake}, Log: testLog}).Open(netshared.TransportUnixSocket, allowlist(t, "api.example.com")); err == nil {
		t.Error("a failed temporary directory must fail")
	}
}

// bridgeHost adds the fake's process outcomes to the bridge port.
func TestBridgeRelaysAndRunsTheCommand(t *testing.T) {
	host := systemtest.New()
	host.Env["HTTP_PROXY"] = "http://elsewhere:1"
	host.Env["KEEP"] = "yes"
	host.Outcomes["agent --flag"] = systemtest.Outcome{ExitCode: 7}
	proxyEnd := make(chan net.Conn, 1)
	host.Dial = func(_ context.Context, network, address string) (net.Conn, error) {
		if network != "unix" || address != "/run/egress/proxy.sock" {
			return nil, errors.New("wrong route")
		}
		near, far := net.Pipe()
		proxyEnd <- far
		return near, nil
	}
	code, err := (Bridge{Host: host}).Run("/run/egress/proxy.sock", "127.0.0.1:3128", []string{"agent", "--flag"})
	if err != nil || code != 7 {
		t.Fatalf("Run = %d, %v; want 7, nil", code, err)
	}
	call := host.Calls[0]
	env := strings.Join(call.Env, "\n")
	for _, want := range []string{"HTTP_PROXY=http://127.0.0.1:3128", "https_proxy=http://127.0.0.1:3128", "NO_PROXY=", "KEEP=yes"} {
		if !strings.Contains(env, want) {
			t.Errorf("env lacks %q:\n%s", want, env)
		}
	}
	if strings.Contains(env, "elsewhere") {
		t.Error("an inherited proxy variable must be replaced")
	}
	listener, ok := host.Listener("tcp", "127.0.0.1:3128")
	if !ok || !listener.Closed() {
		t.Error("the bridge must close its listener when the command ends")
	}
}

func TestBridgeRelaysConnectionsWhileTheCommandRuns(t *testing.T) {
	host := systemtest.New()
	proxyEnd := make(chan net.Conn, 1)
	host.Dial = func(context.Context, string, string) (net.Conn, error) {
		near, far := net.Pipe()
		proxyEnd <- far
		return near, nil
	}
	started := make(chan struct{})
	release := make(chan struct{})
	runHost := blockingHost{Fake: host, started: started, release: release}
	finished := make(chan struct{})
	go func() {
		_, _ = (Bridge{Host: runHost}).Run("/s", "127.0.0.1:3128", []string{"agent"})
		close(finished)
	}()
	<-started
	listener, _ := host.Listener("tcp", "127.0.0.1:3128")
	client, err := listener.Dial()
	if err != nil {
		t.Fatal(err)
	}
	go func() { _, _ = client.Write([]byte("hi")) }()
	far := <-proxyEnd
	got := make([]byte, 2)
	if _, err := io.ReadFull(far, got); err != nil || string(got) != "hi" {
		t.Errorf("proxy end got %q, %v", got, err)
	}
	_ = far.Close()
	_ = client.Close()
	close(release)
	<-finished
}

type blockingHost struct {
	*systemtest.Fake
	started chan struct{}
	release chan struct{}
}

func (b blockingHost) Run(string, []string, []string, string) (int, error) {
	close(b.started)
	<-b.release
	return 0, nil
}

func TestBridgeDropsAConnectionWhenTheProxyIsUnreachable(t *testing.T) {
	host := systemtest.New()
	started := make(chan struct{})
	release := make(chan struct{})
	runHost := blockingHost{Fake: host, started: started, release: release}
	finished := make(chan struct{})
	go func() {
		_, _ = (Bridge{Host: runHost}).Run("/s", "127.0.0.1:3128", []string{"agent"})
		close(finished)
	}()
	<-started
	listener, _ := host.Listener("tcp", "127.0.0.1:3128")
	client, err := listener.Dial()
	if err != nil {
		t.Fatal(err)
	}
	_ = client.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, err := client.Read(make([]byte, 1)); err == nil {
		t.Error("the connection must be closed when the proxy socket cannot be reached")
	}
	close(release)
	<-finished
}

func TestBridgeFailures(t *testing.T) {
	host := systemtest.New()
	if _, err := (Bridge{Host: host}).Run("/s", "127.0.0.1:3128", nil); !errors.Is(err, ErrNoCommand) {
		t.Errorf("no command error = %v", err)
	}
	host.ListenErr = errors.New("in use")
	if _, err := (Bridge{Host: host}).Run("/s", "127.0.0.1:3128", []string{"agent"}); err == nil {
		t.Error("a failed listen must fail")
	}
	host.ListenErr = nil
	if _, err := (Bridge{Host: host}).Run("/s", "127.0.0.1:3128", []string{"missing"}); err == nil {
		t.Error("a command that cannot start must fail")
	}
}
