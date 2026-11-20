//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web/nethttp"
)

func TestNetHTTPReadsABodyOnlyBelowStatus400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("User-Agent"), "agent-audit") {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		switch r.URL.Path {
		case "/ok":
			w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
			_, _ = w.Write([]byte("hello"))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("error page"))
		}
	}))
	defer server.Close()

	client := nethttp.New()
	ok, err := client.Get(context.Background(), server.URL+"/ok")
	if err != nil || ok.Status != 200 || string(ok.Body) != "hello" || ok.LastModified == "" {
		t.Fatalf("ok = %+v, %v", ok, err)
	}
	gone, err := client.Get(context.Background(), server.URL+"/gone")
	if err != nil || gone.Status != 404 || len(gone.Body) != 0 {
		t.Fatalf("gone = %+v, %v", gone, err)
	}
}

func TestNetHTTPReportsFaults(t *testing.T) {
	client := nethttp.New()
	if _, err := client.Get(context.Background(), "http://127.0.0.1:1/"); err == nil {
		t.Error("a refused connection must be an error")
	}
	if _, err := client.Get(context.Background(), "://bad"); err == nil {
		t.Error("a malformed URL must be an error")
	}
}
