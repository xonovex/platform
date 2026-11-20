// Package compare audits a source catalog against the same catalog at a base
// ref. It answers a question the plain audit cannot: did this change move a
// pin, drop a claim or edit a fed file without a new review of the source?
package compare

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/rules"
)

// Base is the catalog as it stood at the base ref.
type Base interface {
	// Catalog returns the catalog in dir at the base ref. A catalog that did not
	// exist yet is empty.
	Catalog(dir string) (sources.Catalog, error)
	// ChangedFiles returns the files below dir that differ between the base ref
	// and the working tree, as slash paths relative to dir.
	ChangedFiles(dir string) ([]string, error)
}

// Against returns the findings for the changes between the base catalog of dir
// and the head catalog. A source whose review date moved forward was reviewed
// in the change, so its other changes are expected.
func Against(base Base, dir string, head sources.Catalog) ([]findings.Finding, error) {
	before, err := base.Catalog(dir)
	if err != nil {
		return nil, err
	}
	changed, err := base.ChangedFiles(dir)
	if err != nil {
		return nil, err
	}
	return Diff(before, head, changed), nil
}

// Diff compares two catalogs. changed lists the files that differ between the
// base ref and the working tree, relative to the catalog directory.
func Diff(before, head sources.Catalog, changed []string) []findings.Finding {
	touched := map[string]bool{}
	for _, file := range changed {
		touched[file] = true
	}
	var out []findings.Finding
	for _, old := range before.Sources {
		now, ok := head.ByID(old.ID)
		if !ok {
			out = append(out, finding(findings.Warn, rules.CodeSourceRemoved, old.File,
				fmt.Sprintf("source %q was in the base catalog and is gone", old.ID)))
			continue
		}
		out = append(out, compareSource(old, now, touched)...)
	}
	return out
}

func compareSource(old, now sources.Source, touched map[string]bool) []findings.Finding {
	if now.ReviewedAt < old.ReviewedAt {
		return []findings.Finding{finding(findings.Fail, rules.CodeReviewRegressed, now.File,
			fmt.Sprintf("source %q was reviewed on %s at the base ref and is dated %s now", now.ID, old.ReviewedAt, now.ReviewedAt))}
	}
	if now.ReviewedAt > old.ReviewedAt {
		return nil
	}

	var out []findings.Finding
	if fields := pinChanges(old, now); len(fields) > 0 {
		out = append(out, finding(findings.Drift, rules.CodeMetadataChanged, now.File,
			fmt.Sprintf("source %q changed %s without a new review", now.ID, strings.Join(fields, ", "))))
	}
	if claims := claimChanges(old, now); len(claims) > 0 {
		out = append(out, finding(findings.Drift, rules.CodeMetadataChanged, now.File,
			fmt.Sprintf("source %q changed its claims without a new review: %s", now.ID, strings.Join(claims, ", "))))
	}
	var edited []string
	for _, feed := range now.Feeds {
		if touched[feed] {
			edited = append(edited, feed)
		}
	}
	if len(edited) > 0 {
		out = append(out, finding(findings.Drift, rules.CodeGitFileChanged, now.File,
			fmt.Sprintf("source %q was not reviewed again, but the files it feeds changed: %s", now.ID, strings.Join(edited, ", "))))
	}
	return out
}

// pinChanges names the fields that say where a source is read from and which
// state of it was reviewed.
func pinChanges(old, now sources.Source) []string {
	var fields []string
	differ := func(name string, same bool) {
		if !same {
			fields = append(fields, name)
		}
	}
	differ("urls", slices.Equal(old.URLs, now.URLs))
	differ("path", old.Path == now.Path)
	differ("ref", old.Ref == now.Ref)
	differ("checkout", old.Checkout == now.Checkout)
	differ("version", old.Version == now.Version)
	differ("commit", old.Commit == now.Commit)
	differ("watch", reflect.DeepEqual(old.Watch, now.Watch))
	differ("contentSha256", old.ContentSHA256 == now.ContentSHA256)
	differ("outputSha256", old.OutputSHA256 == now.OutputSHA256)
	return fields
}

// claimChanges describes the claims that were dropped or edited. A new claim is
// not a change that needs a review of the source.
func claimChanges(old, now sources.Source) []string {
	current := map[string]sources.Claim{}
	for _, claim := range now.Claims {
		current[claim.ID] = claim
	}
	var out []string
	for _, claim := range old.Claims {
		updated, ok := current[claim.ID]
		switch {
		case !ok:
			out = append(out, claim.ID+" removed")
		case updated != claim:
			out = append(out, claim.ID+" edited")
		}
	}
	sort.Strings(out)
	return out
}

func finding(level findings.Level, code, file, message string) findings.Finding {
	return findings.Finding{Check: rules.CheckName, Level: level, Code: code, Path: file, Message: message}
}
