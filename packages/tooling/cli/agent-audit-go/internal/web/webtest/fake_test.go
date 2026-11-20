package webtest_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web/webtest"
)

func TestFakeReplaysTheScriptAndRepeatsTheLastOutcome(t *testing.T) {
	boom := errors.New("down")
	fake := webtest.New(map[string]webtest.Page{
		"https://e.test": {{Err: boom}, {Response: web.Response{Status: 200}}},
	})
	if _, err := fake.Get(context.Background(), "https://e.test"); !errors.Is(err, boom) {
		t.Fatalf("first: %v", err)
	}
	for i := 0; i < 2; i++ {
		if got, err := fake.Get(context.Background(), "https://e.test"); err != nil || got.Status != 200 {
			t.Fatalf("request %d: %+v, %v", i+2, got, err)
		}
	}
	if fake.Requests["https://e.test"] != 3 {
		t.Errorf("Requests = %v", fake.Requests)
	}
}

func TestFakeFailsAnUnscriptedURL(t *testing.T) {
	fake := webtest.New(map[string]webtest.Page{"https://empty.test": {}})
	for _, url := range []string{"https://none.test", "https://empty.test"} {
		if _, err := fake.Get(context.Background(), url); err == nil {
			t.Errorf("%s: an unscripted URL must be an error", url)
		}
	}
}
