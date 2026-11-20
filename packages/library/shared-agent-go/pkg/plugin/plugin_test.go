package plugin

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
)

func tree(t *testing.T, files map[string]string) artifact.Tree {
	t.Helper()
	var list []artifact.File
	for path, content := range files {
		list = append(list, artifact.File{Path: path, Content: []byte(content)})
	}
	built, err := artifact.NewTree(list)
	if err != nil {
		t.Fatal(err)
	}
	return built
}

const manifestJSON = `{"name":"xonovex-core","version":"6.0.0","description":"d","dependencies":["b-dep","a-dep"],"interface":{"x":1}}`

func TestReadManifestAgreesAcrossFormats(t *testing.T) {
	manifest, err := ReadManifest(tree(t, map[string]string{
		".claude-plugin/plugin.json": manifestJSON,
		".codex-plugin/plugin.json":  manifestJSON,
		"plugin.json":                manifestJSON,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "xonovex-core" || manifest.Version != "6.0.0" || manifest.Description != "d" {
		t.Errorf("manifest = %+v", manifest)
	}
	if strings.Join(manifest.Dependencies, ",") != "a-dep,b-dep" {
		t.Errorf("dependencies = %v", manifest.Dependencies)
	}
	if strings.Join(manifest.Keys, ",") != "dependencies,description,interface,name,version" {
		t.Errorf("keys = %v", manifest.Keys)
	}
}

func TestReadManifestFailsClosed(t *testing.T) {
	other := strings.Replace(manifestJSON, "6.0.0", "6.0.1", 1)
	for name, files := range map[string]map[string]string{
		"none":              {"README.md": "x"},
		"not json":          {"plugin.json": "{"},
		"wrong type":        {"plugin.json": `{"name":1}`},
		"bad name":          {"plugin.json": `{"name":"Bad","version":"1"}`},
		"no version":        {"plugin.json": `{"name":"a"}`},
		"bad dependency":    {"plugin.json": `{"name":"a","version":"1","dependencies":["Bad"]}`},
		"self dependency":   {"plugin.json": `{"name":"a","version":"1","dependencies":["a"]}`},
		"repeated":          {"plugin.json": `{"name":"a","version":"1","dependencies":["b","b"]}`},
		"disagree":          {".claude-plugin/plugin.json": manifestJSON, "plugin.json": other},
		"not an object":     {"plugin.json": `[]`},
		"second not an obj": {".claude-plugin/plugin.json": manifestJSON, ".codex-plugin/plugin.json": `{"name":"xonovex-core","version":"6.0.0","dependencies":["x"]}`},
	} {
		if _, err := ReadManifest(tree(t, files)); err == nil {
			t.Errorf("%s: ReadManifest passed", name)
		}
	}
}

func TestDetectCapabilities(t *testing.T) {
	cases := map[string]struct {
		files map[string]string
		keys  []string
		want  string
	}{
		"none":        {map[string]string{"skills/a/SKILL.md": "x"}, nil, ""},
		"hooks dir":   {map[string]string{"hooks/run.sh": "x"}, nil, "hooks"},
		"hooks json":  {map[string]string{"hooks.json": "x"}, nil, "hooks"},
		"nested json": {map[string]string{".claude-plugin/hooks.json": "x"}, nil, "hooks"},
		"mcp":         {map[string]string{".mcp.json": "x"}, nil, "mcp"},
		"mcp plain":   {map[string]string{"mcp.json": "x"}, nil, "mcp"},
		"lsp":         {map[string]string{".lsp.json": "x"}, nil, "lsp"},
		"settings":    {map[string]string{"settings.json": "x"}, nil, "settings"},
		"keys":        {nil, []string{"hooks", "mcpServers", "lspServers", "interface"}, "hooks,lsp,mcp"},
		"all":         {map[string]string{"hooks/a": "x", ".mcp.json": "x", ".lsp.json": "x", "settings.json": "x"}, nil, "hooks,lsp,mcp,settings"},
	}
	for name, c := range cases {
		var got []string
		for _, capability := range Detect(tree(t, c.files), Manifest{Keys: c.keys}) {
			got = append(got, string(capability))
		}
		if strings.Join(got, ",") != c.want {
			t.Errorf("%s: Detect = %v, want %s", name, got, c.want)
		}
	}
}

func TestParseCapability(t *testing.T) {
	for _, capability := range Capabilities() {
		if got, err := ParseCapability(string(capability)); err != nil || got != capability {
			t.Errorf("ParseCapability(%q) = %v, %v", capability, got, err)
		}
	}
	if _, err := ParseCapability("shell"); err == nil {
		t.Error("an unknown capability parsed")
	}
}

func TestAuthorizeNamesEveryMissingGrant(t *testing.T) {
	if err := Authorize(nil, nil); err != nil {
		t.Errorf("no capability needs no grant: %v", err)
	}
	if err := Authorize([]Capability{CapabilityHooks}, []Capability{CapabilityHooks, CapabilityMCP}); err != nil {
		t.Errorf("a granted capability failed: %v", err)
	}
	err := Authorize([]Capability{CapabilityHooks, CapabilityMCP, CapabilityLSP}, []Capability{CapabilityMCP})
	if err == nil || !strings.Contains(err.Error(), "hooks, lsp") || strings.Contains(err.Error(), "mcp,") {
		t.Errorf("err = %v", err)
	}
}

func TestCheckApproval(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	other := "sha256:" + strings.Repeat("b", 64)
	if err := CheckApproval(digest, digest); err != nil {
		t.Errorf("the matching digest failed: %v", err)
	}
	if err := CheckApproval(digest, ""); err == nil || !strings.Contains(err.Error(), "--approve-digest "+digest) {
		t.Errorf("the preview step must name the digest: %v", err)
	}
	if err := CheckApproval(digest, "nonsense"); err == nil {
		t.Error("a malformed digest passed")
	}
	if err := CheckApproval(digest, other); err == nil || !strings.Contains(err.Error(), "differs") {
		t.Errorf("a different digest passed: %v", err)
	}
}

const claudeMarketplace = `{"name":"mk","plugins":[{"name":"b-plugin","source":"./plugins/b","description":"B"},{"name":"a-plugin","source":"./plugins/a/"},{"name":"remote","source":{"source":"github","repo":"x/y"}}]}`
const codexMarketplace = `{"name":"mk","plugins":[{"name":"a-plugin","source":"./plugins/a/"},{"name":"c-plugin","source":"./plugins/c"}]}`

func TestReadMarketplaceMergesBothFormats(t *testing.T) {
	marketplace, err := ReadMarketplace(tree(t, map[string]string{ClaudeMarketplacePath: claudeMarketplace, CodexMarketplacePath: codexMarketplace}))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range marketplace.Entries {
		names = append(names, entry.Name)
	}
	if marketplace.Name != "mk" || strings.Join(names, ",") != "a-plugin,b-plugin,c-plugin,remote" {
		t.Errorf("marketplace = %+v", marketplace)
	}
	if entry, ok := marketplace.Find("b-plugin"); !ok || entry.Description != "B" {
		t.Errorf("Find(b-plugin) = %+v, %v", entry, ok)
	}
	if _, ok := marketplace.Find("missing"); ok {
		t.Error("Find found a missing plugin")
	}
}

func TestReadMarketplaceReadsOneFormat(t *testing.T) {
	for _, path := range MarketplacePaths() {
		marketplace, err := ReadMarketplace(tree(t, map[string]string{path: codexMarketplace}))
		if err != nil || len(marketplace.Entries) != 2 {
			t.Errorf("%s: %+v, %v", path, marketplace, err)
		}
	}
}

func TestReadMarketplaceFailsClosed(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"none":          {},
		"not json":      {ClaudeMarketplacePath: "{"},
		"bad name":      {ClaudeMarketplacePath: `{"name":"Bad","plugins":[]}`},
		"bad plugin":    {ClaudeMarketplacePath: `{"name":"mk","plugins":[{"name":"Bad","source":"./a"}]}`},
		"other name":    {ClaudeMarketplacePath: claudeMarketplace, CodexMarketplacePath: `{"name":"other","plugins":[]}`},
		"other source":  {ClaudeMarketplacePath: claudeMarketplace, CodexMarketplacePath: `{"name":"mk","plugins":[{"name":"a-plugin","source":"./elsewhere"}]}`},
		"second broken": {ClaudeMarketplacePath: claudeMarketplace, CodexMarketplacePath: "{"},
	} {
		if _, err := ReadMarketplace(tree(t, files)); err == nil {
			t.Errorf("%s: ReadMarketplace passed", name)
		}
	}
}

func TestEntryPath(t *testing.T) {
	for source, want := range map[string]string{"./plugins/a": "plugins/a", "./plugins/a/": "plugins/a", "./a": "a"} {
		if got, err := (Entry{Name: "x", Source: source}).Path(); err != nil || got != want {
			t.Errorf("Path(%q) = %q, %v", source, got, err)
		}
	}
	for _, source := range []string{"", "plugins/a", "/abs", "./", "./../x", "./a/../b", "../x"} {
		if _, err := (Entry{Name: "x", Source: source}).Path(); err == nil {
			t.Errorf("Path(%q) passed", source)
		}
	}
}
