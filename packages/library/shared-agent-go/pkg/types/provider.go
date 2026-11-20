package types

// ModelSlot is one named model of a provider and where its id came from.
type ModelSlot struct {
	// Name is the slot, such as the model tier the agent asks for.
	Name string
	// ID is the model id the provider route receives.
	ID string
	// Effort is the reasoning effort the route appends to the id, empty for none.
	Effort string
	// Source names the layer that decided the id: the provider file or the
	// override that replaced it.
	Source string
}

// ModelProvider represents a model provider configuration
type ModelProvider struct {
	Name                string
	DisplayName         string
	AgentType           AgentType
	CredentialSourceEnv string            // Host environment variable containing the credential
	CredentialTargetEnv string            // Agent environment variable that receives the credential
	Portable            bool              // Whether the preset is safe outside the local host environment
	Environment         map[string]string // Environment variables to set when using this provider
	CliArgs             []string          // CLI arguments to add when using this provider
	Slots               []ModelSlot       // The models the route uses, sorted by slot name
}
