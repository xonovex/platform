// Package providers defines the model provider axis as data. A provider is a
// JSON document that names an agent, a credential, the environment and the
// arguments of a route, and the models the route uses as named slots. The
// package ships no provider and names no model: a catalog is built from the
// documents the caller loads, and a model id is only ever a value in them.
package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// Preset is one provider document.
type Preset struct {
	Schema      string          `json:"$schema,omitempty"`
	Name        string          `json:"name"`
	DisplayName string          `json:"displayName,omitempty"`
	Agent       types.AgentType `json:"agent"`
	// Extends names a provider of the same agent that this document refines. The
	// refinement keeps every field it does not set, merges models and environment
	// entry by entry, and replaces the arguments when it sets any.
	Extends string `json:"extends,omitempty"`
	// Portable says that the route is safe outside the local host environment.
	Portable   *bool             `json:"portable,omitempty"`
	Credential *Credential       `json:"credential,omitempty"`
	Models     map[string]Slot   `json:"models,omitempty"`
	Env        map[string]string `json:"environment,omitempty"`
	Args       []string          `json:"args,omitempty"`
}

// Credential maps the host variable that holds the credential to the variable
// the agent reads it from. The document holds the names, never a value.
type Credential struct {
	SourceEnv string `json:"sourceEnv"`
	TargetEnv string `json:"targetEnv"`
}

// Slot is one named model of a route.
type Slot struct {
	ID     string `json:"id"`
	Effort string `json:"effort,omitempty"`
}

var (
	namePattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	slotPattern   = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	envKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	effortPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	// placeholderPattern matches {model:<slot>} and {model:<slot>:effort}.
	placeholderPattern = regexp.MustCompile(`\{model:([a-z][a-z0-9-]*)(:effort)?\}`)
)

// ParsePreset decodes one provider document. It rejects an unknown member, more
// than one document and a document that fails Validate, so a typo never changes
// a route silently.
func ParsePreset(data []byte) (Preset, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var preset Preset
	if err := decoder.Decode(&preset); err != nil {
		return Preset{}, fmt.Errorf("decode provider: %w", err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return Preset{}, errors.New("decode provider: the file holds more than one JSON document")
	}
	if err := preset.Validate(); err != nil {
		return Preset{}, err
	}
	return preset, nil
}

// Validate checks the fields of one document. A placeholder that names a slot the
// document does not define is checked when the catalog resolves the document,
// because a refining document takes its slots from the provider it extends.
func (p Preset) Validate() error {
	if !namePattern.MatchString(p.Name) {
		return fmt.Errorf("provider name %q must be lowercase letters, digits and hyphens", p.Name)
	}
	if !namePattern.MatchString(string(p.Agent)) {
		return fmt.Errorf("provider %q: agent %q must be lowercase letters, digits and hyphens", p.Name, p.Agent)
	}
	if p.Extends != "" && !namePattern.MatchString(p.Extends) {
		return fmt.Errorf("provider %q: extends %q must name a provider", p.Name, p.Extends)
	}
	if p.Credential != nil {
		if !envKeyPattern.MatchString(p.Credential.SourceEnv) || !envKeyPattern.MatchString(p.Credential.TargetEnv) {
			return fmt.Errorf("provider %q: credential needs the environment variable names sourceEnv and targetEnv", p.Name)
		}
	}
	for name, slot := range p.Models {
		if !slotPattern.MatchString(name) {
			return fmt.Errorf("provider %q: model slot %q must be lowercase letters, digits and hyphens", p.Name, name)
		}
		if err := slot.validate(); err != nil {
			return fmt.Errorf("provider %q: model slot %q: %w", p.Name, name, err)
		}
	}
	for key := range p.Env {
		if !envKeyPattern.MatchString(key) {
			return fmt.Errorf("provider %q: environment key %q is not a valid variable name", p.Name, key)
		}
	}
	return nil
}

func (s Slot) validate() error {
	if s.ID == "" || strings.IndexFunc(s.ID, func(r rune) bool { return r <= ' ' || r == 0x7f }) >= 0 {
		return fmt.Errorf("id %q must be non-empty and hold no whitespace or control character", s.ID)
	}
	if s.Effort != "" && !effortPattern.MatchString(s.Effort) {
		return fmt.Errorf("effort %q must be lowercase letters, digits and hyphens", s.Effort)
	}
	return nil
}

// refine returns the document p applied over base: p wins in every field it
// sets.
func refine(base, p Preset) Preset {
	merged := base
	merged.Schema = ""
	merged.Name = p.Name
	merged.Extends = ""
	if p.DisplayName != "" {
		merged.DisplayName = p.DisplayName
	}
	if p.Portable != nil {
		merged.Portable = p.Portable
	}
	if p.Credential != nil {
		merged.Credential = p.Credential
	}
	merged.Models = overlay(base.Models, p.Models)
	merged.Env = overlay(base.Env, p.Env)
	if p.Args != nil {
		merged.Args = p.Args
	}
	return merged
}

func overlay[V any](base, top map[string]V) map[string]V {
	if len(base) == 0 && len(top) == 0 {
		return nil
	}
	merged := make(map[string]V, len(base)+len(top))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range top {
		merged[key] = value
	}
	return merged
}

// checkPlaceholders fails when a value names a slot the provider does not define
// or holds a malformed {model:...} placeholder.
func (p Preset) checkPlaceholders() error {
	check := func(where, value string) error {
		_, err := expand(value, p.Models)
		if err != nil {
			return fmt.Errorf("provider %q: %s: %w", p.Name, where, err)
		}
		return nil
	}
	keys := make([]string, 0, len(p.Env))
	for key := range p.Env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := check("environment "+key, p.Env[key]); err != nil {
			return err
		}
	}
	for index, arg := range p.Args {
		if err := check(fmt.Sprintf("args[%d]", index), arg); err != nil {
			return err
		}
	}
	return nil
}

// expand replaces each {model:<slot>} with the slot id and each
// {model:<slot>:effort} with the id and the effort in parentheses, or the id
// alone when the slot has no effort.
func expand(value string, slots map[string]Slot) (string, error) {
	var failure error
	out := placeholderPattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := placeholderPattern.FindStringSubmatch(match)
		slot, ok := slots[parts[1]]
		if !ok {
			if failure == nil {
				failure = fmt.Errorf("placeholder %s names the model slot %q, which the provider does not define", match, parts[1])
			}
			return match
		}
		if parts[2] != "" && slot.Effort != "" {
			return slot.ID + "(" + slot.Effort + ")"
		}
		return slot.ID
	})
	if failure != nil {
		return "", failure
	}
	if strings.Contains(placeholderPattern.ReplaceAllString(value, ""), "{model:") {
		return "", fmt.Errorf("value %q holds a malformed {model:...} placeholder; write {model:<slot>} or {model:<slot>:effort}", value)
	}
	return out, nil
}
