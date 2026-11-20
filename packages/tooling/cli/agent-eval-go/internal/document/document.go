// Package document reads an eval document from the file system, checks it
// against its schema, resolves the prompt and fixture files it points at, and
// checks the references between its entries that a schema cannot express. A
// document that breaks a rule is an InvalidError, so a caller maps it to a
// usage failure and never scores half a document.
package document

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// InvalidError reports a document that breaks its schema or a reference rule.
type InvalidError struct {
	Path       string
	Violations []jsonschema.Violation
}

// Error lists every violation on its own line.
func (e *InvalidError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s is not a valid document", e.Path)
	for _, v := range e.Violations {
		pointer := v.Pointer
		if pointer == "" {
			pointer = "/"
		}
		fmt.Fprintf(&b, "\n  %s: %s", pointer, v.Message)
	}
	return b.String()
}

// Reader reads documents through the file-system and schema ports.
type Reader struct {
	FS        fs.FileSystem
	Validator jsonschema.Validator
}

// read loads the file, checks it against the schema and decodes it into v.
func (r Reader) read(schemaID, file string, v any) error {
	data, err := r.FS.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read %s: %w", file, err)
	}
	violations, err := r.Validator.Validate(schemaID, data)
	if err != nil {
		return fmt.Errorf("validate %s: %w", file, err)
	}
	if len(violations) > 0 {
		return &InvalidError{Path: file, Violations: violations}
	}
	return model.Decode(data, v)
}

func invalid(file string, problems []string) error {
	if len(problems) == 0 {
		return nil
	}
	violations := make([]jsonschema.Violation, len(problems))
	for i, p := range problems {
		violations[i] = jsonschema.Violation{Message: p}
	}
	return &InvalidError{Path: file, Violations: violations}
}

// resolve fills the text of a prompt, reading a prompt file that sits below the
// directory of the document.
func (r Reader) resolve(dir string, p *model.Prompt) error {
	if p.PromptFile == "" {
		p.Text = p.Prompt
		return nil
	}
	data, err := r.FS.ReadFile(path.Join(dir, p.PromptFile))
	if err != nil {
		return fmt.Errorf("read prompt file %s: %w", p.PromptFile, err)
	}
	p.Text = strings.TrimRight(string(data), "\n")
	return nil
}

// Trigger reads a trigger eval.
func (r Reader) Trigger(file string) (model.TriggerEval, error) {
	var doc model.TriggerEval
	if err := r.read(schemas.EvalTrigger, file, &doc); err != nil {
		return doc, err
	}
	var problems []string
	seen := map[string]bool{}
	for i := range doc.Queries {
		q := &doc.Queries[i]
		if seen[q.ID] {
			problems = append(problems, "duplicate query id "+q.ID)
		}
		seen[q.ID] = true
		if err := r.resolve(path.Dir(file), &q.Prompt); err != nil {
			return doc, err
		}
	}
	return doc, invalid(file, problems)
}

// Output reads an output eval.
func (r Reader) Output(file string) (model.OutputEval, error) {
	var doc model.OutputEval
	if err := r.read(schemas.EvalOutput, file, &doc); err != nil {
		return doc, err
	}
	var problems []string
	seen := map[string]bool{}
	for i := range doc.Cases {
		c := &doc.Cases[i]
		if seen[c.ID] {
			problems = append(problems, "duplicate case id "+c.ID)
		}
		seen[c.ID] = true
		if err := r.resolve(path.Dir(file), &c.Prompt); err != nil {
			return doc, err
		}
		for j := range c.Fixtures {
			f := &c.Fixtures[j]
			if f.ContentFile == "" {
				f.Text = f.Content
				continue
			}
			data, err := r.FS.ReadFile(path.Join(path.Dir(file), f.ContentFile))
			if err != nil {
				return doc, fmt.Errorf("read fixture file %s: %w", f.ContentFile, err)
			}
			f.Text = string(data)
		}
	}
	return doc, invalid(file, problems)
}

// Instruction reads an instruction eval.
func (r Reader) Instruction(file string) (model.InstructionEval, error) {
	var doc model.InstructionEval
	if err := r.read(schemas.EvalInstruction, file, &doc); err != nil {
		return doc, err
	}
	var problems []string
	seen := map[string]bool{}
	for i := range doc.Cases {
		c := &doc.Cases[i]
		if seen[c.ID] {
			problems = append(problems, "duplicate case id "+c.ID)
		}
		seen[c.ID] = true
		if err := r.resolve(path.Dir(file), &c.Prompt); err != nil {
			return doc, err
		}
	}
	return doc, invalid(file, problems)
}

// Orchestration reads an orchestration eval and checks that every step names a
// known agent and known earlier steps, and that the steps form no cycle.
func (r Reader) Orchestration(file string) (model.OrchestrationEval, error) {
	var doc model.OrchestrationEval
	if err := r.read(schemas.EvalOrchestration, file, &doc); err != nil {
		return doc, err
	}
	for i := range doc.Steps {
		if err := r.resolve(path.Dir(file), &doc.Steps[i].Prompt); err != nil {
			return doc, err
		}
	}
	return doc, invalid(file, FlowProblems(doc))
}

// FlowProblems lists the reference problems of an orchestration: a duplicate
// agent or step id, an unknown agent, an unknown or self-referencing dependency
// and a dependency cycle.
func FlowProblems(doc model.OrchestrationEval) []string {
	var problems []string
	agents := map[string]bool{}
	for _, a := range doc.Agents {
		if agents[a.ID] {
			problems = append(problems, "duplicate agent id "+a.ID)
		}
		agents[a.ID] = true
	}
	steps := map[string]model.Step{}
	for _, s := range doc.Steps {
		if _, dup := steps[s.ID]; dup {
			problems = append(problems, "duplicate step id "+s.ID)
		}
		steps[s.ID] = s
	}
	for _, s := range doc.Steps {
		if !agents[s.Agent] {
			problems = append(problems, fmt.Sprintf("step %s names unknown agent %s", s.ID, s.Agent))
		}
		for _, dep := range s.After {
			if _, ok := steps[dep]; !ok {
				problems = append(problems, fmt.Sprintf("step %s waits on unknown step %s", s.ID, dep))
			}
		}
	}
	if len(problems) == 0 {
		if _, err := Order(doc.Steps); err != nil {
			problems = append(problems, err.Error())
		}
	}
	return problems
}

// ErrCycle reports steps that wait on each other.
var ErrCycle = errors.New("steps form a dependency cycle")

// Order returns the steps in an order in which each step comes after the steps
// it waits on. Steps that do not depend on each other keep their document
// order. A cycle returns ErrCycle.
func Order(steps []model.Step) ([]model.Step, error) {
	done := map[string]bool{}
	index := map[string]model.Step{}
	for _, s := range steps {
		index[s.ID] = s
	}
	var out []model.Step
	for len(out) < len(steps) {
		progressed := false
		for _, s := range steps {
			if done[s.ID] {
				continue
			}
			ready := true
			for _, dep := range s.After {
				if _, known := index[dep]; known && !done[dep] {
					ready = false
					break
				}
			}
			if ready {
				done[s.ID] = true
				out = append(out, s)
				progressed = true
			}
		}
		if !progressed {
			return nil, ErrCycle
		}
	}
	return out, nil
}

// Project reads the eval project file and checks that its default profile
// exists.
func (r Reader) Project(file string) (model.Project, error) {
	var doc model.Project
	if err := r.read(schemas.EvalProject, file, &doc); err != nil {
		return doc, err
	}
	var problems []string
	if doc.DefaultProfile != "" {
		if _, ok := doc.Profiles[doc.DefaultProfile]; !ok {
			problems = append(problems, "defaultProfile "+doc.DefaultProfile+" names no profile")
		}
	}
	return doc, invalid(file, problems)
}

// Results reads a results document.
func (r Reader) Results(file string) (model.Results, error) {
	var doc model.Results
	err := r.read(schemas.Results, file, &doc)
	return doc, err
}

// Baseline reads a baseline document.
func (r Reader) Baseline(file string) (model.Baseline, error) {
	var doc model.Baseline
	err := r.read(schemas.Baseline, file, &doc)
	return doc, err
}

// CostHistory reads a cost history document.
func (r Reader) CostHistory(file string) (model.CostHistory, error) {
	var doc model.CostHistory
	err := r.read(schemas.CostHistory, file, &doc)
	return doc, err
}

// Validate checks a document that the reader cannot type, such as a file that
// is about to be written, and returns its violations.
func (r Reader) Validate(schemaID string, data []byte) ([]jsonschema.Violation, error) {
	return r.Validator.Validate(schemaID, data)
}
