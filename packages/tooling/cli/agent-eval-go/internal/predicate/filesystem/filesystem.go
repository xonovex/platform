// Package filesystem is the predicate leaf that checks the files a run left in
// its workspace.
package filesystem

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
)

// Evaluator checks filesystem predicates against the workspace snapshot.
type Evaluator struct{}

// New returns the leaf.
func New() Evaluator { return Evaluator{} }

// Kind is the predicate kind the leaf checks.
func (Evaluator) Kind() string { return model.KindFilesystem }

// Evaluate checks each predicate.
func (Evaluator) Evaluate(_ context.Context, predicates []model.Predicate, subject predicate.Subject) ([]predicate.Verdict, model.Usage, error) {
	verdicts := make([]predicate.Verdict, len(predicates))
	for i, p := range predicates {
		verdicts[i] = check(p, subject.Obs.Files)
	}
	return verdicts, model.Usage{}, nil
}

func check(p model.Predicate, files map[string]string) predicate.Verdict {
	content, exists := files[path.Clean(p.Path)]
	switch p.Check {
	case "exists":
		return verdict(exists, "file "+p.Path+" is missing")
	case "absent":
		return verdict(!exists, "file "+p.Path+" exists")
	}
	if !exists {
		return predicate.Verdict{Detail: "file " + p.Path + " is missing"}
	}
	switch p.Check {
	case "equals":
		return verdict(content == p.Value, "file "+p.Path+" differs from the expected content")
	case "contains":
		return verdict(strings.Contains(content, p.Value), fmt.Sprintf("file %s does not contain %q", p.Path, p.Value))
	case "matches":
		re, err := regexp.Compile(p.Value)
		if err != nil {
			return predicate.Verdict{Detail: "invalid pattern: " + err.Error()}
		}
		return verdict(re.MatchString(content), fmt.Sprintf("file %s does not match %q", p.Path, p.Value))
	}
	return predicate.Verdict{Detail: "unknown check " + p.Check}
}

func verdict(ok bool, failure string) predicate.Verdict {
	if ok {
		return predicate.Verdict{Passed: true}
	}
	return predicate.Verdict{Detail: failure}
}
