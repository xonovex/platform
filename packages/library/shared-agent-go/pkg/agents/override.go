package agents

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// OverrideDir is the directory of the override files beside each agentic.json,
// and OverrideKind names them in the approval prompt of a project file.
const (
	OverrideDir  = "harnesses.d"
	OverrideKind = "harness"
)

// Override is one harnesses.d/<name>.json file: the facts of one harness that it
// changes.
type Override struct {
	// Name is the agent type the file is named after.
	Name types.AgentType
	// Origin names the file in messages.
	Origin string
	Data   []byte
}

// Apply returns the catalog with the overrides applied in order, so a later
// override wins over an earlier one. An override names only the members it
// changes: an object member merges into the current object member by member,
// and any other member replaces the current value whole. Each override is
// checked against the override schema, and the facts it produces against the
// harness schema, so an override can neither add an unknown member nor leave a
// required one out.
func (c Catalog) Apply(validator jsonschema.Validator, overrides ...Override) (Catalog, error) {
	next := Catalog{facts: make(map[types.AgentType]types.AgentConfig, len(c.facts))}
	for agentType, facts := range c.facts {
		next.facts[agentType] = clone(facts)
	}
	for _, override := range overrides {
		facts, err := next.apply(validator, override)
		if err != nil {
			return Catalog{}, fmt.Errorf("harness override %s: %w", override.Origin, err)
		}
		next.facts[override.Name] = facts
	}
	return next, nil
}

func (c Catalog) apply(validator jsonschema.Validator, override Override) (types.AgentConfig, error) {
	current, ok := c.facts[override.Name]
	if !ok {
		return types.AgentConfig{}, fmt.Errorf("no harness %q to override; harnesses: %v", override.Name, c.Types())
	}
	if err := check(validator, schemas.HarnessOverride, override.Data); err != nil {
		return types.AgentConfig{}, err
	}
	var patch map[string]any
	// The schema accepted the document, so it is a JSON object.
	_ = json.Unmarshal(override.Data, &patch)
	delete(patch, "$schema")
	if name, set := patch["name"]; set && name != string(override.Name) {
		return types.AgentConfig{}, fmt.Errorf("the file of harness %s names %v", override.Name, name)
	}
	// Facts hold only strings, booleans, lists and objects, which always encode.
	data, _ := json.Marshal(current)
	var base map[string]any
	_ = json.Unmarshal(data, &base)
	merged, _ := json.Marshal(merge(base, patch))
	if err := check(validator, schemas.Harness, merged); err != nil {
		return types.AgentConfig{}, fmt.Errorf("the facts it produces: %w", err)
	}
	return decode(merged)
}

// merge applies patch to base: an object merges member by member, and any
// other value replaces the value of base.
func merge(base, patch map[string]any) map[string]any {
	for key, value := range patch {
		object, isObject := value.(map[string]any)
		current, hasObject := base[key].(map[string]any)
		if isObject && hasObject {
			base[key] = merge(current, object)
			continue
		}
		base[key] = value
	}
	return base
}

// check validates document against a schema and lists every violation.
func check(validator jsonschema.Validator, schemaID string, document []byte) error {
	violations, err := validator.Validate(schemaID, document)
	if err != nil {
		return err
	}
	if len(violations) == 0 {
		return nil
	}
	parts := make([]string, len(violations))
	for i, v := range violations {
		parts[i] = v.Pointer + ": " + v.Message
	}
	return fmt.Errorf("invalid harness facts: %s", strings.Join(parts, "; "))
}
