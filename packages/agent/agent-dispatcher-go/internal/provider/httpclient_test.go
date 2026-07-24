package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestClientPreviewAndAuthentication(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authorization = request.Header.Get("Authorization")
		writeJSON(t, writer, http.StatusCreated, map[string]any{"id": 1})
	}))
	defer server.Close()
	baseURL, _ := url.Parse(server.URL)
	client := NewClientWithHTTP(baseURL, "token", AuthenticationGitHub, server.Client())
	request := Request{Method: http.MethodPost, Path: "/repos/a/b/issues", Query: url.Values{"x": []string{"1"}}, Body: map[string]string{"title": "test"}}
	preview, err := client.Preview(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Requests) != 1 || !strings.Contains(preview.Requests[0].URL, "x=1") {
		t.Fatalf("unexpected preview: %+v", preview)
	}
	response, err := client.Do(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusCreated || authorization != "Bearer token" {
		t.Fatalf("unexpected response or authentication: %+v %q", response, authorization)
	}
}

func TestClientClassifiesFailures(t *testing.T) {
	tests := []struct {
		status    int
		retryable bool
		auth      Authentication
	}{
		{http.StatusTooManyRequests, true, AuthenticationGitLab},
		{http.StatusBadGateway, true, AuthenticationGitLab},
		{http.StatusConflict, true, AuthenticationGitLab},
		{http.StatusBadRequest, false, AuthenticationGitLab},
	}
	for _, test := range tests {
		t.Run(http.StatusText(test.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Retry-After", "2")
				writeJSON(t, writer, test.status, map[string]string{"message": "failed"})
			}))
			defer server.Close()
			baseURL, _ := url.Parse(server.URL)
			client := NewClientWithHTTP(baseURL, "token", test.auth, server.Client())
			_, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/resource"})
			if err == nil {
				t.Fatal("expected request failure")
			}
			var retryable *RetryableError
			if errors.As(err, &retryable) != test.retryable {
				t.Fatalf("unexpected failure type: %T", err)
			}
			if test.retryable && retryable.RetryAfter != 2*time.Second {
				t.Fatalf("unexpected retry delay: %s", retryable.RetryAfter)
			}
		})
	}
}

func TestClientClassifiesGitHubRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("X-RateLimit-Remaining", "0")
		writeJSON(t, writer, http.StatusForbidden, map[string]string{"message": "rate limited"})
	}))
	defer server.Close()
	baseURL, _ := url.Parse(server.URL)
	client := NewClientWithHTTP(baseURL, "token", AuthenticationGitHub, server.Client())
	_, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/resource"})
	var retryable *RetryableError
	if !errors.As(err, &retryable) {
		t.Fatalf("expected retryable GitHub rate limit, got %v", err)
	}
}

func TestClientRejectsUnsafePathAndRedirect(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")
	client := NewClient(baseURL, "token", AuthenticationGitHub, time.Second)
	if _, err := client.Preview(Request{Method: http.MethodGet, Path: "../secret"}); err == nil {
		t.Fatal("expected unsafe path rejection")
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "https://example.org", http.StatusFound)
	}))
	defer server.Close()
	baseURL, _ = url.Parse(server.URL)
	client = NewClient(baseURL, "token", AuthenticationGitHub, time.Second)
	_, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/"})
	var permanent *PermanentError
	if !errors.As(err, &permanent) || permanent.StatusCode != http.StatusFound {
		t.Fatalf("expected redirect to fail permanently, got %v", err)
	}
}

func writeJSON(t *testing.T, writer http.ResponseWriter, status int, value any) {
	t.Helper()
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Fatal(err)
	}
}
