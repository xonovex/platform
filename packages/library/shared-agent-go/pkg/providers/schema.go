package providers

import _ "embed"

// SchemaID identifies the JSON Schema of a provider file.
const SchemaID = "https://schemas.xonovex.com/agent-cli/provider.schema.json"

//go:embed provider.schema.json
var schema []byte

// Schema returns the JSON Schema of a provider file.
func Schema() []byte { return append([]byte(nil), schema...) }
