package config

import "encoding/json"

// SchemaID identifies the JSON Schema of a configuration file.
const SchemaID = "https://schemas.xonovex.com/agent-cli/config.schema.json"

// Schema returns the JSON Schema (draft 2020-12) of a configuration file in its
// YAML or TOML form. It derives from Specs, so a key and its schema cannot drift
// apart. The layers that may set a key are in its x-layers member.
func Schema() []byte {
	properties := map[string]any{}
	for _, spec := range Specs {
		layers := make([]string, 0, len(spec.Layers))
		for _, layer := range spec.Layers {
			layers = append(layers, string(layer))
		}
		property := map[string]any{
			"description": spec.Doc + " Merge rule: " + string(spec.Rule) + ".",
			"x-layers":    layers,
			"x-rule":      string(spec.Rule),
		}
		switch spec.Rule {
		case RuleReplace:
			property["type"] = "string"
		case RuleAppend:
			property["type"] = "array"
			property["items"] = map[string]any{"type": "string", "minLength": 1}
		case RuleRequire:
			property["type"] = "boolean"
		}
		properties[string(spec.Key)] = property
	}
	// The document holds only strings, string slices and maps of them, which
	// always encode.
	data, _ := json.MarshalIndent(map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"$id":                  SchemaID,
		"title":                "agent-cli configuration",
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}, "", "  ")
	return append(data, '\n')
}
