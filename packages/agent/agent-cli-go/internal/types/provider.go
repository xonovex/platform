package types

// ModelProvider represents a model provider configuration
type ModelProvider struct {
	Name         string
	DisplayName  string
	AgentType    AgentType
	AuthTokenEnv string            // Environment variable name containing the auth token
	Environment  map[string]string // Environment variables to set when using this provider
	CliArgs      []string          // CLI arguments to add when using this provider
}
