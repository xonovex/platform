// Package record is the run record: the JSON file every run leaves behind that
// says what ran and under which guarantees. The model and its schema are pure;
// the Store writes and reads the records through a file port.
package record

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"
)

// SchemaVersion is the version of the record layout.
const SchemaVersion = 1

// SchemaID identifies the JSON Schema of a run record.
const SchemaID = "https://schemas.xonovex.com/agent-cli/run-record.schema.json"

//go:embed run-record.schema.json
var schema []byte

// Schema returns the JSON Schema (draft 2020-12) of a run record.
func Schema() []byte { return append([]byte(nil), schema...) }

// Status is how a run ended.
type Status string

const (
	// StatusExited marks a run that started the agent, or its terminal wrapper, and
	// that ended with the exit code in the record.
	StatusExited Status = "exited"
	// StatusFailed marks a run that planned but failed before the agent started.
	StatusFailed Status = "failed-to-start"
)

// Workspace modes.
const (
	ModeInPlace  = "in-place"
	ModeCheckout = "checkout"
)

// Record is one run. It holds credential names and never their values, and it
// holds the layer that decided each configuration key rather than the values.
type Record struct {
	SchemaVersion int       `json:"schemaVersion"`
	ID            string    `json:"id"`
	StartedAt     time.Time `json:"startedAt"`
	EndedAt       time.Time `json:"endedAt"`
	Status        Status    `json:"status"`
	// ExitCode is the exit code of the agent, or of the terminal wrapper when the
	// run used one. It is absent for a run that failed to start.
	ExitCode    *int        `json:"exitCode,omitempty"`
	Error       string      `json:"error,omitempty"`
	Agent       Agent       `json:"agent"`
	Axes        Axes        `json:"axes"`
	Guarantees  []Guarantee `json:"guarantees"`
	Overlay     Overlay     `json:"overlay"`
	Credentials []string    `json:"credentials"`
	WorkDir     string      `json:"workDir"`
	Config      Config      `json:"config"`
}

// Agent is the agent the run launched.
type Agent struct {
	Name string `json:"name"`
	// VersionProbe is the command line that prints the agent version.
	VersionProbe string `json:"versionProbe"`
	Provider     string `json:"provider,omitempty"`
}

// Axes are the choices the run made on each axis.
type Axes struct {
	Isolation string    `json:"isolation"`
	Provision string    `json:"provision"`
	Network   string    `json:"network"`
	Workspace Workspace `json:"workspace"`
	Terminal  string    `json:"terminal,omitempty"`
}

// Workspace is where the run worked: the source directory in place, or a checkout.
type Workspace struct {
	Mode   string `json:"mode"`
	VCS    string `json:"vcs,omitempty"`
	Branch string `json:"branch,omitempty"`
}

// Guarantee is the evaluation of one isolation guarantee.
type Guarantee struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Provided bool   `json:"provided"`
}

// Overlay lists what the overlay of the run drew on, each with its content digest.
type Overlay struct {
	Plugins     []Digest `json:"plugins"`
	Directories []Digest `json:"directories"`
}

// Digest names an overlay source and the digest of its content.
type Digest struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
}

// Config says which configuration files took part and which layers decided each
// key.
type Config struct {
	Sources []Source            `json:"sources"`
	Decided map[string][]string `json:"decided"`
}

// Source is one configuration file, or the command line.
type Source struct {
	Layer string `json:"layer"`
	Path  string `json:"path,omitempty"`
}

// normalize replaces every nil collection with an empty one, so a record always
// holds each member the schema requires.
func (r Record) normalize() Record {
	if r.Guarantees == nil {
		r.Guarantees = []Guarantee{}
	}
	if r.Overlay.Plugins == nil {
		r.Overlay.Plugins = []Digest{}
	}
	if r.Overlay.Directories == nil {
		r.Overlay.Directories = []Digest{}
	}
	if r.Credentials == nil {
		r.Credentials = []string{}
	}
	if r.Config.Sources == nil {
		r.Config.Sources = []Source{}
	}
	if r.Config.Decided == nil {
		r.Config.Decided = map[string][]string{}
	}
	r.SchemaVersion = SchemaVersion
	return r
}

// Marshal encodes the record as indented JSON that ends with a newline.
func Marshal(r Record) ([]byte, error) {
	data, err := json.MarshalIndent(r.normalize(), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode the run record: %w", err)
	}
	return append(data, '\n'), nil
}

// Unmarshal decodes a record and refuses a layout of another version.
func Unmarshal(data []byte) (Record, error) {
	var r Record
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&r); err != nil {
		return Record{}, fmt.Errorf("decode the run record: %w", err)
	}
	if r.SchemaVersion != SchemaVersion {
		return Record{}, fmt.Errorf("the run record has schema version %d; this build reads version %d", r.SchemaVersion, SchemaVersion)
	}
	return r, nil
}
