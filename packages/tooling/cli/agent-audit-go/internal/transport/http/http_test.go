package http_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
	httptransport "github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/http"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web/webtest"
)

func reply(status int, body string) webtest.Outcome {
	return webtest.Outcome{Response: web.Response{Status: status, Body: []byte(body)}}
}

func observeURLs(t *testing.T, fake *webtest.Fake, urls ...string) []observe.Document {
	t.Helper()
	got, err := httptransport.Transport{Client: fake}.Observe(context.Background(), observe.Request{Source: sources.Source{URLs: urls}})
	if err != nil || got.Transport != httptransport.Name {
		t.Fatalf("Observe = %+v, %v", got, err)
	}
	return got.Documents
}

func TestObserveClassifiesEveryStatus(t *testing.T) {
	fake := webtest.New(map[string]webtest.Page{
		"https://ok.test":    {{Response: web.Response{Status: 200, LastModified: "Mon", Body: []byte("body")}}},
		"https://gone.test":  {reply(404, "")},
		"https://auth.test":  {reply(401, "")},
		"https://deny.test":  {reply(403, "")},
		"https://slow.test":  {reply(429, "")},
		"https://boom.test":  {reply(500, "")},
		"https://zero.test":  {reply(0, "")},
		"https://redir.test": {reply(302, "moved")},
	})
	got := observeURLs(t, fake, "https://ok.test", "https://gone.test", "https://auth.test", "https://deny.test", "https://slow.test", "https://boom.test", "https://zero.test", "https://redir.test")
	want := []observe.Status{observe.OK, observe.Missing, observe.Restricted, observe.Restricted, observe.Restricted, observe.Failed, observe.Failed, observe.OK}
	for i, d := range got {
		if d.Status != want[i] {
			t.Errorf("%s: status = %s; want %s", d.Locator, d.Status, want[i])
		}
	}
	if got[0].SHA256 != observe.Hash([]byte("body")) || string(got[0].Body) != "body" || got[0].Detail != "HTTP 200, Last-Modified: Mon" {
		t.Errorf("ok document = %+v", got[0])
	}
	if got[1].Detail != "HTTP 404" || got[5].Detail != "HTTP 500" {
		t.Errorf("details = %q, %q", got[1].Detail, got[5].Detail)
	}
	if got[3].SHA256 != "" {
		t.Error("a restricted document has no digest")
	}
}

func TestObserveRetriesAFailedRequestOnce(t *testing.T) {
	boom := errors.New("connection refused")
	fake := webtest.New(map[string]webtest.Page{
		"https://flaky.test": {{Err: boom}, reply(200, "ok")},
		"https://down.test":  {{Err: boom}},
		"https://error.test": {reply(503, ""), reply(200, "back")},
	})
	got := observeURLs(t, fake, "https://flaky.test", "https://down.test", "https://error.test")
	if got[0].Status != observe.OK || got[2].Status != observe.OK {
		t.Errorf("a failure that clears on the retry must pass: %+v", got)
	}
	if got[1].Status != observe.Failed || got[1].Detail != "unreachable: connection refused" {
		t.Errorf("down = %+v", got[1])
	}
	if fake.Requests["https://flaky.test"] != 2 || fake.Requests["https://down.test"] != 2 {
		t.Errorf("requests = %v", fake.Requests)
	}
	if fake.Requests["https://error.test"] != 2 {
		t.Errorf("a 5xx counts as a failure and is retried: %v", fake.Requests)
	}
}

func TestTransportFacts(t *testing.T) {
	tr := httptransport.Transport{}
	if tr.Name() != "http" || !tr.Remote() {
		t.Error("http is a remote transport")
	}
	if !tr.Applies(sources.Source{URLs: []string{"https://e.test"}}) || tr.Applies(sources.Source{Path: "x"}) {
		t.Error("Applies must follow the URLs")
	}
}
