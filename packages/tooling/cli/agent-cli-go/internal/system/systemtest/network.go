package systemtest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
)

// ErrNoRoute reports a dial the test registered no route for.
var ErrNoRoute = errors.New("systemtest: no route")

// DialFunc connects to address on the named network.
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// MemListener is an in-memory net.Listener: Dial hands a connection to Accept.
type MemListener struct {
	network string
	address string

	accepted chan net.Conn
	closed   chan struct{}
	once     sync.Once
}

func newMemListener(network, address string) *MemListener {
	return &MemListener{network: network, address: address, accepted: make(chan net.Conn), closed: make(chan struct{})}
}

// Accept waits for the next Dial.
func (l *MemListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.accepted:
		return conn, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

// Close stops the listener.
func (l *MemListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

// Addr returns the address the listener was opened on.
func (l *MemListener) Addr() net.Addr { return memAddr{network: l.network, address: l.address} }

// Dial connects a client to the listener and returns the client end.
func (l *MemListener) Dial() (net.Conn, error) {
	client, server := net.Pipe()
	select {
	case l.accepted <- server:
		return client, nil
	case <-l.closed:
		_ = client.Close()
		_ = server.Close()
		return nil, net.ErrClosed
	}
}

// Closed reports whether Close was called.
func (l *MemListener) Closed() bool {
	select {
	case <-l.closed:
		return true
	default:
		return false
	}
}

type memAddr struct{ network, address string }

func (a memAddr) Network() string { return a.network }
func (a memAddr) String() string  { return a.address }

// Listen opens an in-memory listener. A tcp address that ends in ":0" gets the
// fixed port 40000 so a test can name it.
func (f *Fake) Listen(network, address string) (net.Listener, error) {
	f.netMu.Lock()
	defer f.netMu.Unlock()
	if f.ListenErr != nil {
		return nil, f.ListenErr
	}
	if network == "tcp" && len(address) > 2 && address[len(address)-2:] == ":0" {
		address = address[:len(address)-1] + "40000"
	}
	listener := newMemListener(network, address)
	if f.listeners == nil {
		f.listeners = map[string]*MemListener{}
	}
	f.listeners[network+" "+address] = listener
	return listener, nil
}

// Listener returns the listener opened on network and address.
func (f *Fake) Listener(network, address string) (*MemListener, bool) {
	f.netMu.Lock()
	defer f.netMu.Unlock()
	listener, ok := f.listeners[network+" "+address]
	return listener, ok
}

// DialContext calls Dial, or fails with ErrNoRoute when the test set none.
func (f *Fake) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if f.Dial == nil {
		return nil, fmt.Errorf("%w to %s %s", ErrNoRoute, network, address)
	}
	return f.Dial(ctx, network, address)
}

// DialPublic calls DialPublicFunc, or fails with ErrNoRoute when the test set none.
func (f *Fake) DialPublic(ctx context.Context, network, address string) (net.Conn, error) {
	if f.DialPublicFunc == nil {
		return nil, fmt.Errorf("%w to %s %s", ErrNoRoute, network, address)
	}
	return f.DialPublicFunc(ctx, network, address)
}

// Executable returns ExecutablePath, or an error when it is empty.
func (f *Fake) Executable() (string, error) {
	if f.ExecutablePath == "" {
		return "", errors.New("systemtest: no executable path")
	}
	return f.ExecutablePath, nil
}
