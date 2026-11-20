package runconfig

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Legacy is the port to the configuration files of the previous release: it
// reads them beside agentic.json, and writes them as agentic.json.
type Legacy interface {
	// Read returns the legacy content of every layer, with the project files
	// in projectDir.
	Read(env Env, projectDir string) (LegacyContent, error)
	// Migrate writes the legacy content of the layers the request names as
	// agentic.json and providers.d shards, and returns what it wrote.
	Migrate(env Env, request Migration) ([]Migrated, error)
}

// LegacyContent is the legacy configuration of every layer.
type LegacyContent struct {
	// Documents are the run sections of the legacy files, one for each layer
	// at most, each with a path that ends in LegacySuffix.
	Documents []layer.Document
	// Providers are the legacy provider files, each with an origin that ends
	// in LegacySuffix.
	Providers []providers.Document
	// Files are the legacy files that hold content, for the notice of each one
	// and the refusal of a layer that also has an agentic.json.
	Files []LegacyFile
	// Pending is the legacy project content that neither ledger approves.
	Pending []Pending
}

// LegacySuffix marks a path of the previous release in the sources of a value.
const LegacySuffix = " (legacy)"

// LegacyFile is one legacy file or provider file of a layer.
type LegacyFile struct {
	Layer spec.Layer
	Path  string
}

// Migration names the layers to migrate.
type Migration struct {
	// ProjectDir is the directory of the legacy project files and of the
	// project agentic.json to write.
	ProjectDir string
	// Managed adds the managed layer, which only an administrator can write.
	Managed bool
}

// Migrated is one file Migrate wrote and the legacy files it came from.
type Migrated struct {
	Path string
	From []string
}

// readLegacy reads the legacy content when the environment has the port, and
// refuses a layer that has both an agentic.json and a legacy file, because
// neither may silently win.
func readLegacy(env Env, projectDir string, documents []layer.Document) (LegacyContent, error) {
	if env.Legacy == nil {
		return LegacyContent{}, nil
	}
	content, err := env.Legacy.Read(env, projectDir)
	if err != nil {
		return LegacyContent{}, err
	}
	for _, document := range documents {
		for _, file := range content.Files {
			if file.Layer == document.Layer {
				return LegacyContent{}, fmt.Errorf("the %s layer has %s and the legacy %s; move what the legacy file sets into %s, then remove the legacy file", document.Layer, document.Path, file.Path, document.Path)
			}
		}
	}
	return content, nil
}

// Notices are the warnings of the legacy files the configuration read.
func (l Loaded) Notices() []string {
	notices := make([]string, len(l.legacy))
	for i, file := range l.legacy {
		notices[i] = fmt.Sprintf("%s is a legacy %s configuration file, read for this release only; run agent-cli config migrate to move it to agentic.json", file.Path, file.Layer)
	}
	return notices
}
