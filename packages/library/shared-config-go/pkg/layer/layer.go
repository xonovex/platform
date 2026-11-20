// Package layer reads one agentic.json file: it checks the envelope that every
// tool shares, splits the file into its sections, and checks the sections a
// tool owns against their schemas. A tool only type-checks the sections of
// another tool, so a fault in one of them never stops it.
package layer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Version is the only version of the file.
const Version = 1

// SectionNames lists the sections a file may hold, in the order of the file
// schema.
func SectionNames() []string {
	return []string{"run", "providers", "validate", "audit", "eval"}
}

// Document is one parsed file of one layer.
type Document struct {
	Layer spec.Layer
	Path  string
	// Data is the content of the file, exactly as read.
	Data []byte
	// Digest is the sha256 digest of Data, the value a user approves.
	Digest string
	// Sections holds each section the file sets, by name.
	Sections map[string]json.RawMessage
}

// Digest returns the sha256 digest of data as sha256: and 64 hexadecimal digits.
func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Parse checks the envelope of a file: a JSON object with version 1, an
// optional $schema string and only known sections, each of them an object.
func Parse(layer spec.Layer, path string, data []byte) (Document, error) {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(data, &members); err != nil || members == nil {
		return Document{}, fmt.Errorf("%s: not a JSON object", path)
	}
	known := map[string]bool{}
	for _, name := range SectionNames() {
		known[name] = true
	}
	document := Document{Layer: layer, Path: path, Data: data, Digest: Digest(data), Sections: map[string]json.RawMessage{}}
	names := make([]string, 0, len(members))
	for name := range members {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		raw := members[name]
		switch {
		case name == "$schema":
			var schema string
			if err := json.Unmarshal(raw, &schema); err != nil {
				return Document{}, fmt.Errorf("%s: $schema is not a string", path)
			}
		case name == "version":
			var version int
			if err := json.Unmarshal(raw, &version); err != nil || version != Version {
				return Document{}, fmt.Errorf("%s: version is %s, want %d", path, raw, Version)
			}
		case known[name]:
			if !isObject(raw) {
				return Document{}, fmt.Errorf("%s: the %s section is not an object", path, name)
			}
			document.Sections[name] = raw
		default:
			return Document{}, fmt.Errorf("%s: unknown member %q; a file holds $schema, version and the sections %s", path, name, strings.Join(SectionNames(), ", "))
		}
	}
	if _, ok := members["version"]; !ok {
		return Document{}, fmt.Errorf("%s: version is missing", path)
	}
	return document, nil
}

func isObject(raw json.RawMessage) bool {
	var object map[string]json.RawMessage
	return bytes.HasPrefix(bytes.TrimSpace(raw), []byte("{")) && json.Unmarshal(raw, &object) == nil
}

// Check validates the sections the document sets among sections against their
// schemas and lists every violation.
func (d Document) Check(validator jsonschema.Validator, sections []spec.Section) error {
	for _, section := range sections {
		raw, ok := d.Sections[section.Name]
		if !ok {
			continue
		}
		violations, err := validator.Validate(section.SchemaID, raw)
		if err != nil {
			return fmt.Errorf("%s: check the %s section: %w", d.Path, section.Name, err)
		}
		if len(violations) > 0 {
			parts := make([]string, len(violations))
			for i, v := range violations {
				parts[i] = "/" + section.Name + v.Pointer + ": " + v.Message
			}
			return fmt.Errorf("%s: invalid %s section: %s", d.Path, section.Name, strings.Join(parts, "; "))
		}
	}
	return nil
}
