package agents

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	iofs "io/fs"
	"maps"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// harnessDir holds one facts file per harness, named <agent type>.json. A new
// harness adds one file there; no Go code names a harness fact.
const harnessDir = "harnesses"

//go:embed harnesses/*.json
var embedded embed.FS

// Catalog holds the facts of every harness by agent type. It is a value: Apply
// returns a new catalog and Get returns a copy, so no caller changes the facts
// another caller reads.
type Catalog struct {
	facts map[types.AgentType]types.AgentConfig
}

// embeddedCatalog parses the embedded facts once. A unit test proves that they
// parse, so a failure is a fault of the build.
var embeddedCatalog = sync.OnceValue(func() Catalog { return must(load(embedded, harnessDir)) })

// Embedded returns the facts the module ships.
func Embedded() Catalog { return embeddedCatalog() }

func must(catalog Catalog, err error) Catalog {
	if err != nil {
		panic(fmt.Sprintf("embedded harness facts: %v", err))
	}
	return catalog
}

// load reads every facts file of dir. The name of a file is the agent type it
// declares.
func load(fsys iofs.FS, dir string) (Catalog, error) {
	entries, err := iofs.ReadDir(fsys, dir)
	if err != nil {
		return Catalog{}, fmt.Errorf("list harness facts: %w", err)
	}
	catalog := Catalog{facts: map[types.AgentType]types.AgentConfig{}}
	for _, entry := range entries {
		name, ok := strings.CutSuffix(entry.Name(), ".json")
		if !ok {
			continue
		}
		file := path.Join(dir, entry.Name())
		data, err := iofs.ReadFile(fsys, file)
		if err != nil {
			return Catalog{}, fmt.Errorf("read harness facts %s: %w", file, err)
		}
		facts, err := decode(data)
		if err != nil {
			return Catalog{}, fmt.Errorf("harness facts %s: %w", file, err)
		}
		if string(facts.Type) != name {
			return Catalog{}, fmt.Errorf("harness facts %s declare %q; name the file %s.json", file, facts.Type, facts.Type)
		}
		catalog.facts[facts.Type] = facts
	}
	return catalog, nil
}

// document is a facts file: the facts and the optional $schema member.
type document struct {
	Schema string `json:"$schema,omitempty"`
	types.AgentConfig
}

// decode reads one complete facts document strictly and checks that every path
// it names stays inside the directory it belongs to, which the schema cannot
// express.
func decode(data []byte) (types.AgentConfig, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var doc document
	if err := decoder.Decode(&doc); err != nil {
		return types.AgentConfig{}, err
	}
	facts := doc.AgentConfig
	paths := append(slices.Clone(facts.UserConfigPaths), facts.InstructionFile, facts.SkillsDir,
		facts.Plugin.Manifest, facts.Plugin.Marketplace, facts.Plugin.SkillsDir)
	if facts.ConfigHome != nil {
		paths = append(paths, facts.ConfigHome.Dir)
	}
	for _, p := range paths {
		if p != "" && !local(p) {
			return types.AgentConfig{}, fmt.Errorf("path %q leaves the directory it belongs to", p)
		}
	}
	return facts, nil
}

// local reports whether p is a relative slash-separated path that stays below
// the directory it is joined to.
func local(p string) bool {
	clean := path.Clean(p)
	return !path.IsAbs(p) && !strings.Contains(p, `\`) && clean != ".." && !strings.HasPrefix(clean, "../")
}

// Get returns a copy of the facts of an agent type.
func (c Catalog) Get(agentType types.AgentType) (*types.AgentConfig, error) {
	facts, ok := c.facts[agentType]
	if !ok {
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}
	copied := clone(facts)
	return &copied, nil
}

// PluginLayout returns the plugin layout of an agent type, the zero layout for
// an agent type the catalog lacks.
func (c Catalog) PluginLayout(agentType types.AgentType) types.PluginLayout {
	return c.facts[agentType].Plugin
}

// Types returns the agent types of the catalog in sorted order.
func (c Catalog) Types() []types.AgentType {
	return slices.Sorted(maps.Keys(c.facts))
}

// clone copies every slice, map and pointer of facts.
func clone(facts types.AgentConfig) types.AgentConfig {
	facts.LaunchArgs = slices.Clone(facts.LaunchArgs)
	facts.SandboxArgs = slices.Clone(facts.SandboxArgs)
	facts.VersionArgs = slices.Clone(facts.VersionArgs)
	facts.UserConfigPaths = slices.Clone(facts.UserConfigPaths)
	if facts.ConfigHome != nil {
		home := *facts.ConfigHome
		facts.ConfigHome = &home
	}
	facts.Headless.Args = slices.Clone(facts.Headless.Args)
	facts.Headless.HermeticArgs = slices.Clone(facts.Headless.HermeticArgs)
	facts.Headless.Output = cloneArgs(facts.Headless.Output)
	facts.Headless.Options = cloneArgs(facts.Headless.Options)
	return facts
}

func cloneArgs(in map[string][]string) map[string][]string {
	if in == nil {
		return nil
	}
	out := make(map[string][]string, len(in))
	for key, args := range in {
		out[key] = slices.Clone(args)
	}
	return out
}

// GetAgent returns a copy of the shipped facts of an agent type.
func GetAgent(agentType types.AgentType) (*types.AgentConfig, error) {
	return Embedded().Get(agentType)
}

// GetAgentTypes returns the shipped agent types in sorted order.
func GetAgentTypes() []types.AgentType {
	return Embedded().Types()
}
