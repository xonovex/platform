// Package check defines the port that every check implements and the context a
// check runs in. A check reads one repository through the context and returns
// findings; it never reads the host, so a unit test drives it with the memory
// file system.
package check

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// Repository answers the questions that only the version control system can.
type Repository interface {
	// TrackedFiles lists the committed files, relative to the repository root.
	TrackedFiles() ([]string, error)
	// IgnoredPaths lists the paths that the ignore rules exclude, with a
	// trailing slash on a directory.
	IgnoredPaths() ([]string, error)
}

// Context is everything a check reads.
type Context struct {
	Tree       tree.Tree
	Paths      []string
	Options    json.RawMessage
	Validator  jsonschema.Validator
	Repository Repository
}

// Check is one rule set over a repository.
type Check interface {
	// Name is the identifier of the check on the command line and in the
	// project file.
	Name() string
	// Summary says in one line what the check settles.
	Summary() string
	// DefaultPaths are the repository paths the check reads when the project
	// file names none.
	DefaultPaths() []string
	// Run returns the findings. A broken document is a finding; only a fault
	// of the host or of the options is an error.
	Run(ctx Context) ([]findings.Finding, error)
}

// DecodeOptions reads the options of a check onto the defaults in into. An
// unknown field is an error, so a misspelt option never passes silently.
func DecodeOptions(raw json.RawMessage, into any) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

// Reporter collects the findings of one check.
type Reporter struct {
	name string
	list []findings.Finding
}

// NewReporter starts the findings of the named check.
func NewReporter(name string) *Reporter { return &Reporter{name: name} }

func (r *Reporter) add(level findings.Level, code, path string, line int, format string, args []any) {
	r.list = append(r.list, findings.Finding{
		Check:   r.name,
		Level:   level,
		Code:    code,
		Path:    path,
		Line:    line,
		Message: fmt.Sprintf(format, args...),
	})
}

// Fail records a failure.
func (r *Reporter) Fail(code, path string, line int, format string, args ...any) {
	r.add(findings.Fail, code, path, line, format, args)
}

// Warn records an authoring warning.
func (r *Reporter) Warn(code, path string, line int, format string, args ...any) {
	r.add(findings.Warn, code, path, line, format, args)
}

// Drift records a measurement that the catalog has not yet been corrected for.
func (r *Reporter) Drift(code, path string, line int, format string, args ...any) {
	r.add(findings.Drift, code, path, line, format, args)
}

// Merge appends findings that another component produced.
func (r *Reporter) Merge(all []findings.Finding) { r.list = append(r.list, all...) }

// Findings returns what the reporter collected.
func (r *Reporter) Findings() []findings.Finding { return r.list }
