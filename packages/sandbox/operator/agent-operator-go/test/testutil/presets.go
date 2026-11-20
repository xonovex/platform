package testutil

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
)

// Preset names of the catalog NewPresets builds.
const (
	// PresetPortable is a portable provider of the claude agent.
	PresetPortable = "portable-route"
	// PresetLocal is a claude provider that is not portable.
	PresetLocal = "local-route"
	// PresetArgs is a portable provider of the opencode agent that sets arguments.
	PresetArgs = "args-route"
)

// PresetArgsModel is the model id the PresetArgs provider passes to its agent.
const PresetArgsModel = "vendor/model-a"

var presetFiles = []string{
	`{"name":"portable-route","agent":"claude","portable":true,
 "credential":{"sourceEnv":"HOST_KEY","targetEnv":"ANTHROPIC_AUTH_TOKEN"},
 "models":{"main":{"id":"model-a"}},
 "environment":{"ANTHROPIC_BASE_URL":"https://models.example.test/api","API_TIMEOUT_MS":"3000000","ANTHROPIC_DEFAULT_SONNET_MODEL":"{model:main}"}}`,
	`{"name":"local-route","agent":"claude","environment":{"ANTHROPIC_BASE_URL":"http://127.0.0.1:8317"}}`,
	`{"name":"args-route","agent":"opencode","portable":true,"models":{"main":{"id":"vendor/model-a"}},"args":["--model","{model:main}"]}`,
}

// NewPresets returns the provider catalog the operator tests resolve a presetRef
// against. The documents are constants of this file, so one that fails to parse
// is a defect of the file and panics.
func NewPresets() providers.Catalog {
	var documents []providers.Document
	for _, file := range presetFiles {
		preset, err := providers.ParsePreset([]byte(file))
		if err != nil {
			panic("testutil: provider document does not parse: " + err.Error())
		}
		documents = append(documents, providers.Document{Preset: preset, Layer: "test", Origin: preset.Name + providers.Extension, Raw: []byte(file)})
	}
	return providers.NewCatalog(documents...)
}
