// Package managed is the network=managed leaf: an allowlist forward proxy that
// the consumer runs on the host for the length of a run, and the bridge that relays a
// sandbox's private loopback port to it. The allowlist denies by default.
package managed

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
)

// Dialer opens the outbound connections of the proxy.
type Dialer interface {
	// DialContext connects to any address; the proxy uses it for a host that an
	// allowlist rule names by address.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	// DialPublic connects and refuses an address that is not public, so an allowed
	// host name cannot lead to a service of this machine or its network.
	DialPublic(ctx context.Context, network, address string) (net.Conn, error)
}

// dialTimeout bounds one outbound connection attempt.
const dialTimeout = 30 * time.Second

// Proxy is an http.Handler that tunnels CONNECT requests and forwards plain HTTP
// requests to the hosts of Allow, and answers 403 to everything else.
type Proxy struct {
	Allow  egress.Allowlist
	Dialer Dialer
	Log    logging.Logger
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.tunnel(w, r)
		return
	}
	p.forward(w, r)
}

// target splits an authority into host and port.
func target(authority string, defaultPort int) (string, int, error) {
	host, portText, err := net.SplitHostPort(authority)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			return strings.Trim(authority, "[]"), defaultPort, nil
		}
		return "", 0, err
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

// dial connects to host and port, through the unrestricted dialer only for a
// host that is an address, which a rule can name only on purpose.
func (p Proxy) dial(ctx context.Context, host string, port int) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()
	address := net.JoinHostPort(host, strconv.Itoa(port))
	if net.ParseIP(host) != nil {
		return p.Dialer.DialContext(ctx, "tcp", address)
	}
	return p.Dialer.DialPublic(ctx, "tcp", address)
}

func (p Proxy) deny(w http.ResponseWriter, host string, port int) {
	p.Log.Debug(fmt.Sprintf("egress denied: %s:%d", host, port))
	http.Error(w, fmt.Sprintf("egress denied: %s:%d is not on the allowlist", host, port), http.StatusForbidden)
}

func (p Proxy) tunnel(w http.ResponseWriter, r *http.Request) {
	host, port, err := target(r.Host, 443)
	if err != nil {
		http.Error(w, "bad CONNECT target", http.StatusBadRequest)
		return
	}
	if !p.Allow.Allows(host, port) {
		p.deny(w, host, port)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "connection cannot be hijacked", http.StatusInternalServerError)
		return
	}
	upstream, err := p.dial(r.Context(), host, port)
	if err != nil {
		p.Log.Debug(fmt.Sprintf("egress dial %s:%d failed: %v", host, port, err))
		http.Error(w, "upstream unreachable", http.StatusBadGateway)
		return
	}
	client, buffered, err := hijacker.Hijack()
	if err != nil {
		_ = upstream.Close()
		return
	}
	if _, err := io.WriteString(client, "HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		_ = client.Close()
		_ = upstream.Close()
		return
	}
	relay(client, buffered, upstream)
}

// relay copies both directions until either side ends, then closes both.
func relay(client net.Conn, clientReader io.Reader, upstream net.Conn) {
	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(upstream, clientReader); done <- struct{}{} }()
	go func() { _, _ = io.Copy(client, upstream); done <- struct{}{} }()
	<-done
	_ = client.Close()
	_ = upstream.Close()
	<-done
}

// hopHeaders are the connection-scoped headers a proxy must not forward.
var hopHeaders = []string{"Connection", "Proxy-Connection", "Proxy-Authenticate", "Proxy-Authorization", "Keep-Alive", "Te", "Trailer", "Transfer-Encoding", "Upgrade"}

func (p Proxy) forward(w http.ResponseWriter, r *http.Request) {
	if !r.URL.IsAbs() || r.URL.Host == "" {
		http.Error(w, "this is a forward proxy: send an absolute URL or CONNECT", http.StatusBadRequest)
		return
	}
	defaultPort := 80
	if r.URL.Scheme == "https" {
		defaultPort = 443
	}
	host, port, err := target(r.URL.Host, defaultPort)
	if err != nil {
		http.Error(w, "bad request target", http.StatusBadRequest)
		return
	}
	if !p.Allow.Allows(host, port) {
		p.deny(w, host, port)
		return
	}
	out := r.Clone(r.Context())
	out.RequestURI = ""
	for _, name := range hopHeaders {
		out.Header.Del(name)
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) { return p.dial(ctx, host, port) },
	}
	defer transport.CloseIdleConnections()
	response, err := transport.RoundTrip(out)
	if err != nil {
		p.Log.Debug(fmt.Sprintf("egress forward to %s:%d failed: %v", host, port, err))
		http.Error(w, "upstream unreachable", http.StatusBadGateway)
		return
	}
	defer func() { _ = response.Body.Close() }()
	for _, name := range hopHeaders {
		response.Header.Del(name)
	}
	for name, values := range response.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}
