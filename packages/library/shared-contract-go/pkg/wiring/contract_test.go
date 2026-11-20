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
					{"id":"a-1","kind":"git","title":"T","checkout":"../up","urls":["https://e.test/r.git"],"ref":"v1","version":"1.2.3","commit":"abcdef1","watch":[{"path":"src","feeds":["a/b.md"]}],"contentSha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","citation":"C","usage":["u"],"coversAllReferences":true,"remarks":"r","reviewedAt":"2026-01-01","reviewAfter":"2026-07-01","feeds":["a/b.md"],"claims":[{"id":"c1","statement":"s","anchor":"h","evidence":"e"}]},
					{"id":"b","kind":"filesystem","title":"T","path":"docs/x.md","reviewedAt":"2026-01-01","supersededBy":"a-1"},
					{"id":"c","kind":"repository","title":"T","path":".","reviewedAt":"2026-01-01"},
					{"id":"d","kind":"original","title":"T","provenance":"written here","reviewedAt":"2026-01-01"},
					{"id":"e","kind":"original","title":"T","urlNotes":["A book"],"reviewedAt":"2026-01-01"}],"title":"Sources","preamble":"p","notes":[{"title":"Refresh","text":"t","after":"d"}]}`,
			},
			invalid: []invalid{
				{"unknown kind", `{"version":1,"sources":[{"id":"a","kind":"ftp","title":"T","urls":["https://e.test"],"reviewedAt":"2026-01-01"}]}`, "/sources/0/kind"},
				{"git needs checkout", `{"version":1,"sources":[{"id":"a","kind":"git","title":"T","reviewedAt":"2026-01-01"}]}`, "/sources/0"},
				{"web needs urls", `{"version":1,"sources":[{"id":"a","kind":"web","title":"T","reviewedAt":"2026-01-01"}]}`, "/sources/0"},
				{"original needs provenance or notes", `{"version":1,"sources":[{"id":"a","kind":"original","title":"T","reviewedAt":"2026-01-01"}]}`, "/sources/0"},
				{"no urls", `{"version":1,"sources":[{"id":"a","kind":"web","title":"T","urls":[],"reviewedAt":"2026-01-01"}]}`, "/sources/0/urls"},
				{"bare digest", `{"version":1,"sources":[{"id":"a","kind":"original","title":"T","provenance":"p","contentSha256":"abc","reviewedAt":"2026-01-01"}]}`, "/sources/0/contentSha256"},
				{"short commit", `{"version":1,"sources":[{"id":"a","kind":"original","title":"T","provenance":"p","commit":"abc","reviewedAt":"2026-01-01"}]}`, "/sources/0/commit"},
				{"filesystem needs path", `{"version":1,"sources":[{"id":"a","kind":"filesystem","title":"T","reviewedAt":"2026-01-01"}]}`, "/sources/0"},
				{"impossible date", `{"version":1,"sources":[{"id":"a","kind":"web","title":"T","urls":["https://e.test"],"reviewedAt":"2026-02-30"}]}`, "/sources/0/reviewedAt"},
				{"bad id", `{"version":1,"sources":[{"id":"Bad_Id","kind":"web","title":"T","urls":["https://e.test"],"reviewedAt":"2026-01-01"}]}`, "/sources/0/id"},
				{"absolute path", `{"version":1,"sources":[{"id":"a","kind":"filesystem","title":"T","path":"/etc/x","reviewedAt":"2026-01-01"}]}`, "/sources/0/path"},
				{"parent segment", `{"version":1,"sources":[{"id":"a","kind":"filesystem","title":"T","path":"a/../b","reviewedAt":"2026-01-01"}]}`, "/sources/0/path"},
				{"feed repeats", `{"version":1,"sources":[{"id":"a","kind":"web","title":"T","urls":["https://e.test"],"reviewedAt":"2026-01-01","feeds":["x.md","x.md"]}]}`, "/sources/0/feeds"},
				{"wrong version", `{"version":2,"sources":[]}`, "/version"},
				{"extra field", `{"version":1,"sources":[],"extra":1}`, ""},
			},
		},
		schemas.AgenticValidate: {
			valid: []string{
				`{}`,
				`{"skip":["vendor"],"checks":{"skill":{"enabled":true,"paths":["packages/plugin"],"level":"warn","options":{"a":1}},"text":{"enabled":false}}}`,
			},
			invalid: []invalid{
				{"version inside the section", `{"version":1}`, ""},
				{"unknown level", `{"checks":{"skill":{"level":"pass"}}}`, "/checks/skill/level"},
				{"empty paths", `{"checks":{"skill":{"paths":[]}}}`, "/checks/skill/paths"},
				{"parent path", `{"checks":{"skill":{"paths":["../x"]}}}`, "/checks/skill/paths/0"},
				{"unknown field", `{"checks":{"skill":{"other":1}}}`, "/checks/skill"},
				{"repeated skip", `{"skip":["a","a"]}`, "/skip"},
			},
		},
		schemas.Agentic: {
			valid: []string{
				`{"version":1}`,
				`{"$schema":"x","version":1,"run":{"agent":"claude","isolation":"bwrap","requireEgressRestricted":true},"providers":{"local":{"agent":"claude"}},
				  "validate":{"checks":{"skill":{"level":"warn"}}},"audit":{"maxAge":200},"eval":{"budgets":{"maxModelRuns":10}}}`,
			},
			invalid: []invalid{
				{"no version", `{"validate":{}}`, ""},
				{"wrong version", `{"version":2}`, "/version"},
				{"unknown section", `{"version":1,"lint":{}}`, ""},
				{"section breaks its schema", `{"version":1,"audit":{"maxAge":0}}`, "/audit/maxAge"},
			},
		},
		schemas.AgenticRun: {
			valid: []string{
				`{}`,
				`{"agent":"claude","provider":"local","models":["main=m-1"],"isolation":"docker","provision":"nix","network":"managed","homeDir":"/home/agent",
				  "bindPaths":["/a"],"roBindPaths":["/b"],"customEnv":["A=1"],"allowHosts":["*.example.com:443"],"credentials":["TOKEN=env:TOKEN"],
				  "requirePinnedProvision":true,"requireHostToolsUnreachable":false,"requireEgressRestricted":true,"requireKernelIsolation":false}`,
			},
			invalid: []invalid{
				{"unknown isolation", `{"isolation":"chroot"}`, "/isolation"},
				{"unknown network", `{"network":"proxy"}`, "/network"},
				{"model without slot", `{"models":["m-1"]}`, "/models/0"},
				{"variable without name", `{"customEnv":["=1"]}`, "/customEnv/0"},
				{"unknown key", `{"image":"x"}`, ""},
			},
		},
		schemas.AgenticProviders: {
			valid: []string{
				`{}`,
				`{"local":{"agent":"claude"},"cloud-2":{"agent":"codex","extends":"cloud"}}`,
			},
			invalid: []invalid{
				{"name not a slug", `{"Local":{"agent":"claude"}}`, ""},
				{"route not an object", `{"local":"claude"}`, "/local"},
				{"route without agent", `{"local":{"extends":"cloud"}}`, "/local"},
				{"unknown route member", `{"local":{"agent":"claude","model":"m"}}`, "/local"},
				{"credential without target", `{"local":{"agent":"claude","credential":{"sourceEnv":"KEY"}}}`, "/local/credential"},
				{"model slot without id", `{"local":{"agent":"claude","models":{"main":{}}}}`, "/local/models/main"},
				{"bad variable name", `{"local":{"agent":"claude","environment":{"1X":"v"}}}`, "/local/environment"},
			},
		},
		schemas.Provider: {
			valid: []string{
				`{"name":"local","agent":"claude"}`,
				`{"$schema":"x","name":"cloud-2","displayName":"Cloud","agent":"codex","extends":"cloud","portable":true,
					"credential":{"sourceEnv":"HOST_KEY","targetEnv":"API_KEY"},
					"models":{"main":{"id":"m-1","effort":"high"}},"environment":{"BASE":"{model:main}"},"args":["-c","x"]}`,
			},
			invalid: []invalid{
				{"no name", `{"agent":"claude"}`, ""},
				{"no agent", `{"name":"local"}`, ""},
				{"name not a slug", `{"name":"Local","agent":"claude"}`, "/name"},
				{"model id with space", `{"name":"a","agent":"claude","models":{"main":{"id":"a b"}}}`, "/models/main/id"},
				{"model slot name", `{"name":"a","agent":"claude","models":{"Main":{"id":"m"}}}`, "/models"},
				{"args not strings", `{"name":"a","agent":"claude","args":[1]}`, "/args/0"},
			},
		},
		schemas.Harness: {
			valid: []string{
				`{"name":"h","displayName":"H","binary":"h","nixPackage":"h","access":"native-login","versionArgs":["--version"]}`,
				`{"$schema":"x","name":"h","displayName":"H","description":"d","binary":"/opt/h","nixPackage":"h","protocol":"responses","access":"provider",
					"providerEnv":true,"providerArgs":true,"launchArgs":["chat"],"sandboxArgs":["--open"],"versionArgs":["-V"],
					"userConfigPaths":[".h",".config/h"],"configHome":{"env":"H_HOME","dir":".h"},"instructionFile":"H.md","skillsDir":".h/skills",
					"plugin":{"manifest":".h-plugin/plugin.json","marketplace":".h-plugin/marketplace.json","skillsDir":"skills"},
					"headless":{"args":["-p"],"hermeticArgs":["--clean"],"output":{"stream":["--json"]},"options":{"model":["--model","{value}"]},"skillPrompt":"/{skill} {prompt}"}}`,
			},
			invalid: []invalid{
				{"no binary", `{"name":"h","displayName":"H","nixPackage":"h","access":"provider","versionArgs":["-v"]}`, ""},
				{"empty version probe", `{"name":"h","displayName":"H","binary":"h","nixPackage":"h","access":"provider","versionArgs":[]}`, "/versionArgs"},
				{"unknown access", `{"name":"h","displayName":"H","binary":"h","nixPackage":"h","access":"key","versionArgs":["-v"]}`, "/access"},
				{"absolute config path", `{"name":"h","displayName":"H","binary":"h","nixPackage":"h","access":"provider","versionArgs":["-v"],"userConfigPaths":["/etc/h"]}`, "/userConfigPaths/0"},
				{"unknown member", `{"name":"h","displayName":"H","binary":"h","nixPackage":"h","access":"provider","versionArgs":["-v"],"flags":[]}`, ""},
			},
		},
		schemas.RunRecord: {
			valid: []string{
				`{"schemaVersion":1,"id":"20261120T123005Z","startedAt":"2026-11-20T12:30:05Z","endedAt":"2026-11-20T12:31:00Z","status":"exited","exitCode":0,
					"agent":{"name":"claude","versionProbe":"claude --version","provider":"route"},
					"axes":{"isolation":"bwrap","provision":"none","network":"none","workspace":{"mode":"in-place"}},
					"guarantees":[{"name":"egress-restricted","required":true,"provided":true}],
					"overlay":{"plugins":[{"name":"hello","digest":"` + digest + `"}],"directories":[]},
					"credentials":["API_KEY"],"workDir":"/work",
					"config":{"sources":[{"layer":"user","path":"/home/u/.config/agentic/agentic.json"},{"layer":"flags"}],"decided":{"network":["flags"]}}}`,
			},
			invalid: []invalid{
				{"other version", `{"schemaVersion":2}`, "/schemaVersion"},
				{"id not a timestamp", `{"schemaVersion":1,"id":"latest","startedAt":"2026-11-20T12:30:05Z","endedAt":"2026-11-20T12:31:00Z","status":"exited","agent":{"name":"claude","versionProbe":"v"},"axes":{"isolation":"none","provision":"none","network":"host","workspace":{"mode":"in-place"}},"guarantees":[],"overlay":{"plugins":[],"directories":[]},"credentials":[],"workDir":"/w","config":{"sources":[],"decided":{}}}`, "/id"},
			},
		},
		schemas.HarnessOverride: {
			valid: []string{
				`{}`,
				`{"$schema":"x","binary":"/opt/claude/bin/claude"}`,
				`{"headless":{"options":{"maxTurns":["--turns","{value}"]}}}`,
			},
			invalid: []invalid{
				{"empty binary", `{"binary":""}`, "/binary"},
				{"option not a list", `{"headless":{"options":{"model":"--model"}}}`, "/headless/options/model"},
				{"unknown member", `{"argv":["x"]}`, ""},
				{"plugin manifest absolute", `{"plugin":{"manifest":"/p.json"}}`, "/plugin/manifest"},
			},
		},
		schemas.AgenticAudit: {
			valid: []string{
				`{}`,
				`{"maxAge":200,"versionMaxAge":120,"coverDir":"references"}`,
				`{"coverDir":""}`,
			},
			invalid: []invalid{
				{"zero age", `{"maxAge":0}`, "/maxAge"},
				{"absolute cover directory", `{"coverDir":"/references"}`, "/coverDir"},
				{"unknown key", `{"fetch":true}`, ""},
			},
		},
		schemas.Kit: {
			valid: []string{
				`{"schemaVersion":"2","kind":"sandbox","name":"a","sandbox":{}}`,
				`{"schemaVersion":"2","kind":"mixin","name":"a"}`,
			},
			invalid: []invalid{
				{"no name", `{"schemaVersion":"2","kind":"mixin"}`, ""},
				{"wrong version", `{"schemaVersion":"1","kind":"mixin","name":"a"}`, "/schemaVersion"},
				{"unknown kind", `{"schemaVersion":"2","kind":"other","name":"a"}`, "/kind"},
				{"sandbox needs sandbox", `{"schemaVersion":"2","kind":"sandbox","name":"a"}`, ""},
				{"mixin extends", `{"schemaVersion":"2","kind":"mixin","name":"a","extends":"b"}`, ""},
			},
		},
		schemas.EvalTrigger: {
			valid: []string{
				`{"version":1,"skill":"using-hono","queries":[
					{"id":"q1","prompt":"add a route","shouldTrigger":true,"split":"train","tags":["api"]},
					{"id":"q2","promptFile":"prompts/long.md","shouldTrigger":false,"split":"validation","rationale":"names another skill"}]}`,
			},
			invalid: []invalid{
				{"no queries", `{"version":1,"skill":"s","queries":[]}`, "/queries"},
				{"both prompt forms", `{"version":1,"skill":"s","queries":[{"id":"q","prompt":"a","promptFile":"b.md","shouldTrigger":true,"split":"train"}]}`, "/queries/0"},
				{"neither prompt form", `{"version":1,"skill":"s","queries":[{"id":"q","shouldTrigger":true,"split":"train"}]}`, "/queries/0"},
				{"unknown split", `{"version":1,"skill":"s","queries":[{"id":"q","prompt":"a","shouldTrigger":true,"split":"dev"}]}`, "/queries/0/split"},
				{"unknown field", `{"version":1,"skill":"s","queries":[{"id":"q","prompt":"a","shouldTrigger":true,"split":"train","other":1}]}`, "/queries/0/other"},
				{"empty rationale", `{"version":1,"skill":"s","queries":[{"id":"q","prompt":"a","shouldTrigger":true,"split":"train","rationale":""}]}`, "/queries/0/rationale"},
			},
		},
		schemas.EvalOutput: {
			valid: []string{
				`{"version":1,"skill":"s","tier":"moderate","judge":{"harness":"claude","model":"m","rubric":"r"},"cases":[
					{"id":"c1","prompt":"do it","gate":"required","fixtures":[{"path":"a.txt","content":"x"},{"path":"b.txt","contentFile":"fixtures/b.txt"}],
					 "expect":[{"kind":"filesystem","path":"out.md","check":"contains","value":"x"},{"kind":"llm","criterion":"is clear","threshold":0.8}]}]}`,
			},
			invalid: []invalid{
				{"no expectations", `{"version":1,"skill":"s","cases":[{"id":"c","prompt":"p","gate":"required","expect":[]}]}`, "/cases/0/expect"},
				{"unknown tier", `{"version":1,"skill":"s","tier":"lenient","cases":[{"id":"c","prompt":"p","gate":"required","expect":[{"kind":"llm","criterion":"x"}]}]}`, "/tier"},
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
		schemas.AgenticEval: {
			valid: []string{
				`{}`,
				`{"pluginDirs":["packages/runtime/plugin"],"instructionRoots":["."],"defaultProfile":"cheap",
				  "profiles":{"cheap":{"harness":"claude","model":"haiku","effort":"low","timeoutSeconds":120}},
				  "protocols":{"trigger":{"runs":3,"batchSize":8,"threshold":0.8},"routing":{"runs":1},"output":{"minimumWinRate":0.6},"instruction":{"minimumLift":0.2},"orchestration":{"runs":2}},
				  "budgets":{"maxModelRuns":200,"maxCostUsd":25}}`,
			},
			invalid: []invalid{
				{"version inside the section", `{"version":1}`, ""},
				{"batch over the cap", `{"protocols":{"trigger":{"batchSize":25}}}`, "/protocols/trigger/batchSize"},
				{"profile without harness", `{"profiles":{"p":{"model":"m"}}}`, "/profiles/p"},
				{"unknown effort", `{"profiles":{"p":{"harness":"claude","effort":"extreme"}}}`, "/profiles/p/effort"},
				{"profile name not a slug", `{"profiles":{"Not A Slug":{"harness":"claude"}}}`, "/profiles"},
				{"zero spend", `{"budgets":{"maxCostUsd":0}}`, "/budgets/maxCostUsd"},
			},
		},
		schemas.Results: {
			valid: []string{
				`{"version":1,"runId":"run-1","protocol":"instruction","profile":"cheap","inputsDigest":"` + digest + `","startedAt":"2026-01-01T10:00:00Z","finishedAt":"2026-01-01T10:05:00Z","status":"complete",
				  "cases":[{"id":"c1","condition":"present","outcome":"passed","score":1,"controlScore":0.25,"runs":3,"selected":["target","none","target"],"skillTriggered":true,"durationMs":1200,"usage":` + usageJSON + `,"predicates":[{"kind":"session","passed":true,"detail":"saw it"}]},
				           {"id":"c2","outcome":"errored","error":"timeout"}],"usage":` + usageJSON + `}`,
				`{"version":1,"runId":"run-2","protocol":"trigger","profile":"p","inputsDigest":"` + digest + `","startedAt":"2026-01-01T10:00:00Z","status":"incomplete","stop":"budget","cases":[]}`,
			},
			invalid: []invalid{
				{"unknown stop reason", `{"version":1,"runId":"r","protocol":"trigger","profile":"p","inputsDigest":"` + digest + `","startedAt":"2026-01-01T10:00:00Z","status":"incomplete","stop":"boredom","cases":[]}`, "/stop"},
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
