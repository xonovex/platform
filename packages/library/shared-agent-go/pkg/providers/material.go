package providers

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
)

// Material is the provider content a repository supplies: its provider files
// and the model overrides of its configuration file. A user approves the digest
// of the material before it takes effect.
type Material struct {
	Documents []Document
	// Models are the "slot=id" overrides of the repository configuration.
	Models []string
}

// Empty reports whether the repository supplies nothing.
func (m Material) Empty() bool { return len(m.Documents) == 0 && len(m.Models) == 0 }

// Digest returns the sha256 digest of the material: the bytes of each provider
// file and the model overrides. Any change to either changes the digest.
func (m Material) Digest() string {
	files := make([]artifact.File, 0, len(m.Documents)+1)
	for _, document := range m.Documents {
		files = append(files, artifact.File{Path: "providers/" + document.Preset.Name + Extension, Content: document.Raw})
	}
	if len(m.Models) > 0 {
		files = append(files, artifact.File{Path: "models", Content: []byte(strings.Join(m.Models, "\n") + "\n")})
	}
	// Names are validated lowercase words, so the paths are unique and confined.
	tree, err := artifact.NewTree(files)
	if err != nil {
		return ""
	}
	return tree.Digest()
}

// Describe returns what a person reviews before approving the material: for each
// provider file its agent, the hosts of its route and its model slots.
func (m Material) Describe() []Description {
	described := make([]Description, 0, len(m.Documents))
	for _, document := range m.Documents {
		p := document.Preset
		description := Description{Name: p.Name, Agent: string(p.Agent), Origin: document.Origin, Extends: p.Extends, Hosts: p.Hosts()}
		for _, name := range slices.Sorted(maps.Keys(p.Models)) {
			description.Slots = append(description.Slots, name+"="+p.Models[name].ID)
		}
		described = append(described, description)
	}
	return described
}

// Description is the reviewable summary of one provider file.
type Description struct {
	Name    string
	Agent   string
	Origin  string
	Extends string
	// Hosts are the hosts the route sends requests to.
	Hosts []string
	// Slots are the model slots as "slot=id".
	Slots []string
}

var hostPattern = regexp.MustCompile(`https?://[^\s/"'?#]+`)

// Hosts returns the hosts that the environment values and the arguments of the
// route name as URLs, sorted. A credential sent to a route reaches these hosts.
func (p Preset) Hosts() []string {
	found := map[string]bool{}
	scan := func(value string) {
		for _, match := range hostPattern.FindAllString(value, -1) {
			_, host, _ := strings.Cut(match, "://")
			found[host] = true
		}
	}
	for _, value := range p.Env {
		scan(value)
	}
	for _, arg := range p.Args {
		scan(arg)
	}
	return slices.Sorted(maps.Keys(found))
}
