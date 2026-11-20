package hook_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/hook"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, ".")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, hook.Check{}, ctx)
}

func hooks(body string) map[string]string {
	return map[string]string{"plugin/hooks/hooks.json": body}
}

func TestIdentity(t *testing.T) {
	c := hook.Check{}
	if c.Name() != "hook" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	if _, err := c.Run(checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})); err == nil {
		t.Fatal("unknown option")
	}
}

const valid = `{
  "description": "Guards",
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [
        {"type": "command", "command": "guard", "args": ["a"], "shell": "bash", "timeout": 5, "once": true, "async": false},
        {"type": "http", "url": "https://x", "headers": {"a": "b"}, "allowedEnvVars": ["A", "B"]},
        {"type": "mcp_tool", "server": "s", "tool": "t", "input": {}},
        {"type": "prompt", "prompt": "p", "model": "m"},
        {"type": "agent", "prompt": "p"}
      ]}
    ]
  }
}`

func TestValid(t *testing.T) {
	if got := run(t, hooks(valid)); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
	files := map[string]string{"other/hooks.json": "{", "hooks/notes.json": "{", "hooks.json": "{"}
	if got := run(t, files); len(got) != 0 {
		t.Fatalf("only hooks/hooks.json is read: %v", got)
	}
	custom := map[string]string{"cfg/settings.json": "{"}
	if got := run(t, custom, map[string]any{"file": "settings.json", "directory": "cfg"}); !checktest.Has(got, "hook.json") {
		t.Fatalf("custom names: %v", got)
	}
}

func TestInvalid(t *testing.T) {
	handler := func(h string) string {
		return `{"hooks": {"Stop": [{"hooks": [` + h + `]}]}}`
	}
	cases := map[string]struct {
		body    string
		message string
	}{
		"json":           {`{`, "invalid JSON"},
		"not object":     {`[]`, "/ must be an object"},
		"top field":      {`{"hooks": {"Stop": [{"hooks": [{"type": "agent", "prompt": "p"}]}]}, "x": 1}`, "/x is not a known top-level field"},
		"description":    {`{"description": "", "hooks": {}}`, "description must be a non-empty string"},
		"no hooks":       {`{}`, "required property 'hooks'"},
		"hooks type":     {`{"hooks": []}`, "must be an object keyed by event name"},
		"no events":      {`{"hooks": {}}`, "must name at least one event"},
		"unknown event":  {`{"hooks": {"Nope": []}}`, "is not a known hook event"},
		"groups type":    {`{"hooks": {"Stop": {}}}`, "must be a list of matcher groups"},
		"groups empty":   {`{"hooks": {"Stop": []}}`, "at least one matcher group"},
		"group type":     {`{"hooks": {"Stop": [1]}}`, "/hooks/Stop/0 must be an object"},
		"group field":    {`{"hooks": {"Stop": [{"hooks": [], "x": 1}]}}`, "is not a matcher-group field"},
		"matcher":        {`{"hooks": {"Stop": [{"matcher": 1, "hooks": []}]}}`, "must be a string"},
		"group no hooks": {`{"hooks": {"Stop": [{}]}}`, "required property 'hooks'"},
		"handlers type":  {`{"hooks": {"Stop": [{"hooks": {}}]}}`, "must be a list of handlers"},
		"handlers empty": {`{"hooks": {"Stop": [{"hooks": []}]}}`, "at least one handler"},
		"handler type":   {handler(`1`), "/hooks/Stop/0/hooks/0 must be an object"},
		"no type":        {handler(`{}`), "required property 'type'"},
		"type text":      {handler(`{"type": 1}`), "must be a string"},
		"type name":      {handler(`{"type": "x"}`), "must be one of command, http"},
		"required":       {handler(`{"type": "mcp_tool", "server": "s"}`), "required property 'tool'"},
		"foreign field":  {handler(`{"type": "prompt", "prompt": "p", "url": "x"}`), "is not a field of a prompt handler"},
		"string field":   {handler(`{"type": "command", "command": ""}`), "command must be a non-empty string"},
		"bool field":     {handler(`{"type": "command", "command": "c", "async": "yes"}`), "async must be true or false"},
		"number field":   {handler(`{"type": "command", "command": "c", "timeout": 0}`), "timeout must be a number greater than 0"},
		"number type":    {handler(`{"type": "command", "command": "c", "timeout": "1"}`), "timeout must be a number"},
		"list field":     {handler(`{"type": "command", "command": "c", "args": "x"}`), "args must be a list of strings"},
		"list entry":     {handler(`{"type": "command", "command": "c", "args": [1]}`), "args must be a list of strings"},
		"unique":         {handler(`{"type": "http", "url": "u", "allowedEnvVars": ["A", "A"]}`), "allowedEnvVars must be a list of unique"},
		"unique empty":   {handler(`{"type": "http", "url": "u", "allowedEnvVars": [""]}`), "allowedEnvVars must be a list of unique"},
		"unique type":    {handler(`{"type": "http", "url": "u", "allowedEnvVars": [1]}`), "allowedEnvVars must be a list of unique"},
		"mapping":        {handler(`{"type": "http", "url": "u", "headers": {"a": 1}}`), "headers must be a mapping of strings"},
		"mapping type":   {handler(`{"type": "http", "url": "u", "headers": []}`), "headers must be a mapping of strings"},
		"any mapping":    {handler(`{"type": "mcp_tool", "server": "s", "tool": "t", "input": 1}`), "input must be a mapping"},
		"shell":          {handler(`{"type": "command", "command": "c", "shell": "fish"}`), "shell must be either"},
	}
	for name, tc := range cases {
		got := run(t, hooks(tc.body))
		found := false
		for _, f := range got {
			found = found || strings.Contains(f.Message, tc.message)
		}
		if !found {
			t.Errorf("%s: findings = %v", name, got)
		}
	}
}

func TestLines(t *testing.T) {
	body := "{\n  \"hooks\": {\n    \"Nope\": []\n  }\n}"
	f, ok := checktest.Find(run(t, hooks(body)), "hook.schema")
	if !ok || f.Line != 3 {
		t.Fatalf("finding = %v", f)
	}
	f, _ = checktest.Find(run(t, hooks(`{"hooks": {"Stop": [1]}}`)), "hook.schema")
	if f.Line != 1 {
		t.Fatalf("an index maps to line 1: %v", f)
	}
	f, _ = checktest.Find(run(t, hooks(`{"hooks": {"Stop": []}}`)), "hook.schema")
	if f.Line != 1 {
		t.Fatalf("a pointer with no key in the text maps to line 1: %v", f)
	}
}
