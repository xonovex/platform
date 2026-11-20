package wiring_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
)

const (
	digest    = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	usageJSON = `{"inputTokens":10,"outputTokens":5,"costUsd":0.01}`
)

type invalid struct {
	name     string
	document string
	pointer  string
}

type contract struct {
	valid   []string
	invalid []invalid
}

func contracts() map[string]contract {
	return map[string]contract{
		schemas.Sources: {
			valid: []string{
				`{"version":1,"sources":[]}`,
				`{"$schema":"x","version":1,"sources":[
					{"id":"a-1","kind":"git","title":"T","url":"https://e.test/r.git","ref":"v1","reviewedAt":"2026-01-01","reviewAfter":"2026-07-01","feeds":["a/b.md"],"claims":[{"id":"c1","statement":"s","anchor":"h","evidence":"e"}]},
					{"id":"b","kind":"filesystem","title":"T","path":"docs/x.md","reviewedAt":"2026-01-01","supersededBy":"a-1"},
					{"id":"c","kind":"repository","title":"T","path":".","reviewedAt":"2026-01-01"}]}`,
			},
			invalid: []invalid{
				{"unknown kind", `{"version":1,"sources":[{"id":"a","kind":"ftp","title":"T","url":"https://e.test","reviewedAt":"2026-01-01"}]}`, "/sources/0/kind"},
				{"git needs url", `{"version":1,"sources":[{"id":"a","kind":"git","title":"T","reviewedAt":"2026-01-01"}]}`, "/sources/0"},
				{"filesystem needs path", `{"version":1,"sources":[{"id":"a","kind":"filesystem","title":"T","reviewedAt":"2026-01-01"}]}`, "/sources/0"},
				{"impossible date", `{"version":1,"sources":[{"id":"a","kind":"web","title":"T","url":"https://e.test","reviewedAt":"2026-02-30"}]}`, "/sources/0/reviewedAt"},
				{"bad id", `{"version":1,"sources":[{"id":"Bad_Id","kind":"web","title":"T","url":"https://e.test","reviewedAt":"2026-01-01"}]}`, "/sources/0/id"},
				{"absolute path", `{"version":1,"sources":[{"id":"a","kind":"filesystem","title":"T","path":"/etc/x","reviewedAt":"2026-01-01"}]}`, "/sources/0/path"},
				{"parent segment", `{"version":1,"sources":[{"id":"a","kind":"filesystem","title":"T","path":"a/../b","reviewedAt":"2026-01-01"}]}`, "/sources/0/path"},
				{"feed repeats", `{"version":1,"sources":[{"id":"a","kind":"web","title":"T","url":"https://e.test","reviewedAt":"2026-01-01","feeds":["x.md","x.md"]}]}`, "/sources/0/feeds"},
				{"wrong version", `{"version":2,"sources":[]}`, "/version"},
				{"extra field", `{"version":1,"sources":[],"extra":1}`, ""},
			},
		},
		schemas.EvalTrigger: {
			valid: []string{
				`{"version":1,"skill":"using-hono","queries":[
					{"id":"q1","prompt":"add a route","shouldTrigger":true,"split":"train","tags":["api"]},
					{"id":"q2","promptFile":"prompts/long.md","shouldTrigger":false,"split":"validation"}]}`,
			},
			invalid: []invalid{
				{"no queries", `{"version":1,"skill":"s","queries":[]}`, "/queries"},
				{"both prompt forms", `{"version":1,"skill":"s","queries":[{"id":"q","prompt":"a","promptFile":"b.md","shouldTrigger":true,"split":"train"}]}`, "/queries/0"},
				{"neither prompt form", `{"version":1,"skill":"s","queries":[{"id":"q","shouldTrigger":true,"split":"train"}]}`, "/queries/0"},
				{"unknown split", `{"version":1,"skill":"s","queries":[{"id":"q","prompt":"a","shouldTrigger":true,"split":"dev"}]}`, "/queries/0/split"},
				{"unknown field", `{"version":1,"skill":"s","queries":[{"id":"q","prompt":"a","shouldTrigger":true,"split":"train","other":1}]}`, "/queries/0/other"},
			},
		},
		schemas.EvalOutput: {
			valid: []string{
				`{"version":1,"skill":"s","judge":{"harness":"claude","model":"m","rubric":"r"},"cases":[
					{"id":"c1","prompt":"do it","gate":"required","fixtures":[{"path":"a.txt","content":"x"},{"path":"b.txt","contentFile":"fixtures/b.txt"}],
					 "expect":[{"kind":"filesystem","path":"out.md","check":"contains","value":"x"},{"kind":"llm","criterion":"is clear","threshold":0.8}]}]}`,
			},
			invalid: []invalid{
				{"no expectations", `{"version":1,"skill":"s","cases":[{"id":"c","prompt":"p","gate":"required","expect":[]}]}`, "/cases/0/expect"},
				{"unknown gate", `{"version":1,"skill":"s","cases":[{"id":"c","prompt":"p","gate":"soft","expect":[{"kind":"llm","criterion":"x"}]}]}`, "/cases/0/gate"},
				{"fixture with both contents", `{"version":1,"skill":"s","cases":[{"id":"c","prompt":"p","gate":"required","fixtures":[{"path":"a","content":"x","contentFile":"b"}],"expect":[{"kind":"llm","criterion":"x"}]}]}`, "/cases/0/fixtures/0"},
			},
		},
		schemas.EvalInstruction: {
			valid: []string{
				`{"version":1,"instruction":"AGENTS.md","cases":[
					{"id":"c1","prompt":"p","anchor":{"heading":"Commits"},"expect":{"present":[{"kind":"session","event":"command","match":"^git commit","min":1}],"absent":[{"kind":"session","event":"command","match":"^git commit","max":0}]}},
					{"id":"c2","promptFile":"p.md","anchor":{"text":"never push"},"ablate":"never push","expect":{"ablated":[{"kind":"llm","criterion":"does not push"}]}}]}`,
			},
			invalid: []invalid{
				{"ablated without ablate", `{"version":1,"instruction":"AGENTS.md","cases":[{"id":"c","prompt":"p","anchor":{"text":"t"},"expect":{"ablated":[{"kind":"llm","criterion":"x"}]}}]}`, "/cases/0"},
				{"unknown condition", `{"version":1,"instruction":"AGENTS.md","cases":[{"id":"c","prompt":"p","anchor":{"text":"t"},"expect":{"removed":[{"kind":"llm","criterion":"x"}]}}]}`, "/cases/0/expect"},
				{"empty expect", `{"version":1,"instruction":"AGENTS.md","cases":[{"id":"c","prompt":"p","anchor":{"text":"t"},"expect":{}}]}`, "/cases/0/expect"},
				{"anchor with both", `{"version":1,"instruction":"AGENTS.md","cases":[{"id":"c","prompt":"p","anchor":{"text":"t","heading":"h"},"expect":{"present":[{"kind":"llm","criterion":"x"}]}}]}`, "/cases/0/anchor"},
			},
		},
		schemas.EvalOrchestration: {
			valid: []string{
				`{"version":1,"id":"review-flow","agents":[{"id":"author","harness":"claude"},{"id":"reviewer","harness":"codex","model":"m","role":"reviewer"}],
				  "steps":[{"id":"write","agent":"author","prompt":"write"},{"id":"review","agent":"reviewer","promptFile":"r.md","after":["write"],"expect":[{"kind":"session","event":"message","match":"LGTM"}]}],
				  "budget":{"maxModelRuns":4,"maxCostUsd":1.5,"timeoutSeconds":600}}`,
			},
			invalid: []invalid{
				{"no agents", `{"version":1,"id":"f","agents":[],"steps":[{"id":"s","agent":"a","prompt":"p"}]}`, "/agents"},
				{"unknown harness", `{"version":1,"id":"f","agents":[{"id":"a","harness":"vim"}],"steps":[{"id":"s","agent":"a","prompt":"p"}]}`, "/agents/0/harness"},
				{"step without agent", `{"version":1,"id":"f","agents":[{"id":"a","harness":"pi"}],"steps":[{"id":"s","prompt":"p"}]}`, "/steps/0"},
				{"zero cost cap", `{"version":1,"id":"f","agents":[{"id":"a","harness":"pi"}],"steps":[{"id":"s","agent":"a","prompt":"p"}],"budget":{"maxCostUsd":0}}`, "/budget/maxCostUsd"},
			},
		},
		schemas.EvalProject: {
			valid: []string{
				`{"version":1,"pluginDirs":["packages/runtime/plugin"],"instructionRoots":["."],"defaultProfile":"cheap",
				  "profiles":{"cheap":{"harness":"claude","model":"haiku","effort":"low","timeoutSeconds":120}},
				  "protocols":{"trigger":{"runs":3,"batchSize":8,"threshold":0.8},"routing":{"runs":1},"output":{"minimumWinRate":0.6},"instruction":{"minimumLift":0.2},"orchestration":{"runs":2}},
				  "budgets":{"maxModelRuns":200,"maxCostUsd":25}}`,
			},
			invalid: []invalid{
				{"no profiles", `{"version":1,"profiles":{}}`, "/profiles"},
				{"batch over the cap", `{"version":1,"profiles":{"p":{"harness":"claude"}},"protocols":{"trigger":{"batchSize":25}}}`, "/protocols/trigger/batchSize"},
				{"profile without harness", `{"version":1,"profiles":{"p":{"model":"m"}}}`, "/profiles/p"},
				{"unknown effort", `{"version":1,"profiles":{"p":{"harness":"claude","effort":"extreme"}}}`, "/profiles/p/effort"},
				{"profile name not a slug", `{"version":1,"profiles":{"Not A Slug":{"harness":"claude"}}}`, "/profiles"},
			},
		},
		schemas.Results: {
			valid: []string{
				`{"version":1,"runId":"run-1","protocol":"instruction","profile":"cheap","inputsDigest":"` + digest + `","startedAt":"2026-01-01T10:00:00Z","finishedAt":"2026-01-01T10:05:00Z","status":"complete",
				  "cases":[{"id":"c1","condition":"present","outcome":"passed","score":1,"durationMs":1200,"usage":` + usageJSON + `,"predicates":[{"kind":"session","passed":true,"detail":"saw it"}]},
				           {"id":"c2","outcome":"errored","error":"timeout"}],"usage":` + usageJSON + `}`,
				`{"version":1,"runId":"run-2","protocol":"trigger","profile":"p","inputsDigest":"` + digest + `","startedAt":"2026-01-01T10:00:00Z","status":"incomplete","cases":[]}`,
			},
			invalid: []invalid{
				{"complete without finish time", `{"version":1,"runId":"r","protocol":"trigger","profile":"p","inputsDigest":"` + digest + `","startedAt":"2026-01-01T10:00:00Z","status":"complete","cases":[]}`, ""},
				{"errored without error", `{"version":1,"runId":"r","protocol":"trigger","profile":"p","inputsDigest":"` + digest + `","startedAt":"2026-01-01T10:00:00Z","status":"incomplete","cases":[{"id":"c","outcome":"errored"}]}`, "/cases/0"},
				{"short digest", `{"version":1,"runId":"r","protocol":"trigger","profile":"p","inputsDigest":"sha256:abc","startedAt":"2026-01-01T10:00:00Z","status":"incomplete","cases":[]}`, "/inputsDigest"},
				{"unknown protocol", `{"version":1,"runId":"r","protocol":"vibes","profile":"p","inputsDigest":"` + digest + `","startedAt":"2026-01-01T10:00:00Z","status":"incomplete","cases":[]}`, "/protocol"},
				{"bad timestamp", `{"version":1,"runId":"r","protocol":"trigger","profile":"p","inputsDigest":"` + digest + `","startedAt":"yesterday","status":"incomplete","cases":[]}`, "/startedAt"},
			},
		},
		schemas.Baseline: {
			valid: []string{
				`{"version":1,"protocol":"output","profile":"p","inputsDigest":"` + digest + `","recordedAt":"2026-01-01T10:00:00Z","tolerance":0.05,"cases":[{"id":"c1","condition":"absent","passRate":0.9,"runs":5}]}`,
			},
			invalid: []invalid{
				{"rate above one", `{"version":1,"protocol":"output","profile":"p","inputsDigest":"` + digest + `","recordedAt":"2026-01-01T10:00:00Z","tolerance":0.05,"cases":[{"id":"c","passRate":1.5,"runs":5}]}`, "/cases/0/passRate"},
				{"no cases", `{"version":1,"protocol":"output","profile":"p","inputsDigest":"` + digest + `","recordedAt":"2026-01-01T10:00:00Z","tolerance":0.05,"cases":[]}`, "/cases"},
				{"zero runs", `{"version":1,"protocol":"output","profile":"p","inputsDigest":"` + digest + `","recordedAt":"2026-01-01T10:00:00Z","tolerance":0.05,"cases":[{"id":"c","passRate":1,"runs":0}]}`, "/cases/0/runs"},
			},
		},
		schemas.CostHistory: {
			valid: []string{
				`{"version":1,"entries":[]}`,
				`{"version":1,"entries":[{"runId":"r","recordedAt":"2026-01-01T10:00:00Z","protocol":"trigger","profile":"p","model":"m","modelRuns":24,"usage":` + usageJSON + `}]}`,
			},
			invalid: []invalid{
				{"negative cost", `{"version":1,"entries":[{"runId":"r","recordedAt":"2026-01-01T10:00:00Z","protocol":"trigger","profile":"p","model":"m","modelRuns":1,"usage":{"inputTokens":1,"outputTokens":1,"costUsd":-1}}]}`, "/entries/0/usage/costUsd"},
				{"fractional tokens", `{"version":1,"entries":[{"runId":"r","recordedAt":"2026-01-01T10:00:00Z","protocol":"trigger","profile":"p","model":"m","modelRuns":1,"usage":{"inputTokens":1.5,"outputTokens":1,"costUsd":1}}]}`, "/entries/0/usage/inputTokens"},
			},
		},
		schemas.Predicate: {
			valid: []string{
				`{"kind":"filesystem","path":"a.md","check":"exists"}`,
				`{"kind":"filesystem","path":"a.md","check":"equals","value":"x"}`,
				`{"kind":"session","event":"tool-call","match":"^Read$"}`,
				`{"kind":"llm","criterion":"is correct"}`,
			},
			invalid: []invalid{
				{"contains needs value", `{"kind":"filesystem","path":"a.md","check":"contains"}`, ""},
				{"unknown kind", `{"kind":"telepathy"}`, ""},
				{"unbalanced regex", `{"kind":"session","event":"command","match":"(unclosed"}`, "/match"},
				{"traversing path", `{"kind":"filesystem","path":"../x","check":"exists"}`, "/path"},
			},
		},
	}
}

func newValidator(t *testing.T) jsonschema.Validator {
	t.Helper()
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	return validator
}

func TestEverySchemaHasFixtures(t *testing.T) {
	var covered []string
	for id := range contracts() {
		covered = append(covered, id)
	}
	sort.Strings(covered)
	want := []string{}
	for _, id := range schemas.IDs() {
		if id != schemas.Common {
			want = append(want, id)
		}
	}
	if !reflect.DeepEqual(covered, want) {
		t.Fatalf("fixtures cover %v; schemas are %v", covered, want)
	}
}

func TestValidDocuments(t *testing.T) {
	validator := newValidator(t)
	for id, c := range contracts() {
		for i, document := range c.valid {
			violations, err := validator.Validate(id, []byte(document))
			if err != nil || len(violations) != 0 {
				t.Errorf("%s valid[%d]: %+v, %v", id, i, violations, err)
			}
		}
	}
}

func TestInvalidDocuments(t *testing.T) {
	validator := newValidator(t)
	for id, c := range contracts() {
		for _, tc := range c.invalid {
			t.Run(id+"/"+tc.name, func(t *testing.T) {
				violations, err := validator.Validate(id, []byte(tc.document))
				if err != nil {
					t.Fatal(err)
				}
				if len(violations) == 0 {
					t.Fatal("the document must break the schema")
				}
				for _, v := range violations {
					if v.Pointer == tc.pointer {
						return
					}
				}
				t.Fatalf("no violation at %q: %+v", tc.pointer, violations)
			})
		}
	}
}
