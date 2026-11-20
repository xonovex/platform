package none_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine/none"
)

func TestTheRequestRunsUnchanged(t *testing.T) {
	request := process.Request{Dir: "/tmp/w", Name: "claude", Args: []string{"-p", "--", "hi"}, Env: map[string]string{"A": "1"}, Timeout: time.Minute}
	confined, err := none.New().Confine(confine.Request{Harness: "claude", Process: request})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(confined.Process, request) || confined.Stop != nil {
		t.Errorf("confined = %+v, want the request unchanged and nothing to stop", confined)
	}
	confined.Release()
}
