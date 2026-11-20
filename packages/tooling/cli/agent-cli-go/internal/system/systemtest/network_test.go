package systemtest

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
)

func TestListenAndDialConnectThroughMemory(t *testing.T) {
	fake := New()
	listener, err := fake.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if listener.Addr().String() != "127.0.0.1:40000" || listener.Addr().Network() != "tcp" {
		t.Errorf("Addr = %s %s, want tcp 127.0.0.1:40000", listener.Addr().Network(), listener.Addr())
	}
	mem, ok := fake.Listener("tcp", "127.0.0.1:40000")
	if !ok || mem.Closed() {
		t.Fatalf("Listener = %v, %v", mem, ok)
	}
	go func() {
		client, err := mem.Dial()
		if err == nil {
			_, _ = client.Write([]byte("hi"))
			_ = client.Close()
		}
	}()
	conn, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(conn)
	if string(data) != "hi" {
		t.Errorf("read %q", data)
	}
	if _, ok := fake.Listener("tcp", "elsewhere:1"); ok {
		t.Error("an unopened address must have no listener")
	}
}

func TestCloseStopsAcceptAndDial(t *testing.T) {
	fake := New()
	listener, _ := fake.Listen("unix", "/run/s.sock")
	mem, _ := fake.Listener("unix", "/run/s.sock")
	if err := listener.Close(); err != nil || !mem.Closed() {
		t.Fatalf("Close = %v, closed = %v", err, mem.Closed())
	}
	if err := listener.Close(); err != nil {
		t.Errorf("a second Close = %v", err)
	}
	if _, err := listener.Accept(); !errors.Is(err, net.ErrClosed) {
		t.Errorf("Accept after Close = %v", err)
	}
	if _, err := mem.Dial(); !errors.Is(err, net.ErrClosed) {
		t.Errorf("Dial after Close = %v", err)
	}
}

func TestListenFailureAndPlainPorts(t *testing.T) {
	fake := New()
	fake.ListenErr = errors.New("in use")
	if _, err := fake.Listen("tcp", "127.0.0.1:0"); err == nil {
		t.Error("ListenErr must fail Listen")
	}
	fake.ListenErr = nil
	if listener, _ := fake.Listen("tcp", "127.0.0.1:8080"); listener.Addr().String() != "127.0.0.1:8080" {
		t.Errorf("a named port must stay: %s", listener.Addr())
	}
}

func TestDialsFailWithoutARoute(t *testing.T) {
	fake := New()
	if _, err := fake.DialContext(context.Background(), "tcp", "a:1"); !errors.Is(err, ErrNoRoute) {
		t.Errorf("DialContext = %v", err)
	}
	if _, err := fake.DialPublic(context.Background(), "tcp", "a:1"); !errors.Is(err, ErrNoRoute) {
		t.Errorf("DialPublic = %v", err)
	}
	near, far := net.Pipe()
	defer far.Close()
	fake.Dial = func(context.Context, string, string) (net.Conn, error) { return near, nil }
	fake.DialPublicFunc = func(context.Context, string, string) (net.Conn, error) { return near, nil }
	if conn, err := fake.DialContext(context.Background(), "tcp", "a:1"); err != nil || conn != near {
		t.Errorf("DialContext = %v, %v", conn, err)
	}
	if conn, err := fake.DialPublic(context.Background(), "tcp", "a:1"); err != nil || conn != near {
		t.Errorf("DialPublic = %v, %v", conn, err)
	}
}

func TestExecutable(t *testing.T) {
	fake := New()
	if _, err := fake.Executable(); err == nil {
		t.Error("an unset executable path must fail")
	}
	fake.ExecutablePath = "/opt/agent-cli"
	if path, err := fake.Executable(); err != nil || path != "/opt/agent-cli" {
		t.Errorf("Executable = %q, %v", path, err)
	}
}
