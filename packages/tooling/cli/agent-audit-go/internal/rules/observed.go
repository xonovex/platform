package rules

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
)

// Observed returns the findings about what the transports saw of a source: the
// providers that failed, a digest that no longer matches, claims whose anchors
// are gone, and the drift of a checkout. A source with no observation has
// nothing to report.
func Observed(source sources.Source, observations []observe.Observation) []findings.Finding {
	var documents []observe.Document
	var out []findings.Finding
	for _, observation := range observations {
		documents = append(documents, observation.Documents...)
		if observation.Git != nil {
			out = append(out, gitFindings(source, *observation.Git)...)
		}
	}
	out = append(out, providerFindings(source, documents)...)
	out = append(out, contentFindings(source, documents)...)
	out = append(out, anchorFindings(source, documents)...)
	return out
}

// ProviderFault is the finding for a transport that could not run at all.
func ProviderFault(source sources.Source, transport string, err error) findings.Finding {
	return finding(findings.Fail, CodeProviderError, source.File,
		fmt.Sprintf("source %q: the %s transport failed: %v", source.ID, transport, err))
}

func providerFindings(source sources.Source, documents []observe.Document) []findings.Finding {
	var out []findings.Finding
	for _, d := range documents {
		switch d.Status {
		case observe.Missing, observe.Failed:
			out = append(out, finding(findings.Fail, CodeProviderError, source.File,
				fmt.Sprintf("source %q: %s is %s (%s)", source.ID, d.Locator, d.Status, d.Detail)))
		case observe.Restricted:
			out = append(out, finding(findings.Warn, CodeProviderRestricted, source.File,
				fmt.Sprintf("source %q: %s refused to answer (%s), so its content was not checked", source.ID, d.Locator, d.Detail)))
		}
	}
	return out
}

func contentFindings(source sources.Source, documents []observe.Document) []findings.Finding {
	if source.ContentSHA256 == "" || len(documents) == 0 {
		return nil
	}
	got, ok := observe.Aggregate(documents)
	switch {
	case !ok:
		return []findings.Finding{finding(findings.Warn, CodeContentUnverified, source.File,
			fmt.Sprintf("source %q records a content digest, but not every document could be read", source.ID))}
	case got != source.ContentSHA256:
		return []findings.Finding{finding(findings.Drift, CodeMetadataChanged, source.File,
			fmt.Sprintf("source %q: the content changed, recorded %s, observed %s", source.ID, source.ContentSHA256, got))}
	}
	return nil
}

func anchorFindings(source sources.Source, documents []observe.Document) []findings.Finding {
	var bodies [][]byte
	for _, d := range documents {
		if d.Body != nil {
			bodies = append(bodies, d.Body)
		}
	}
	if len(bodies) == 0 {
		return nil
	}
	var out []findings.Finding
	for _, claim := range source.Claims {
		if claim.Anchor == "" || presentInAny(bodies, claim.Anchor) {
			continue
		}
		out = append(out, finding(findings.Drift, CodeAnchorMissing, source.File,
			fmt.Sprintf("source %q: the anchor %q of claim %q is not in the content", source.ID, claim.Anchor, claim.ID)))
	}
	return out
}

func presentInAny(bodies [][]byte, anchor string) bool {
	for _, body := range bodies {
		if anchorPresent(body, anchor) {
			return true
		}
	}
	return false
}

func gitFindings(source sources.Source, g observe.Git) []findings.Finding {
	problem := func(level findings.Level, code, message string) []findings.Finding {
		return []findings.Finding{finding(level, code, source.File, fmt.Sprintf("source %q: %s", source.ID, message))}
	}
	switch {
	case !g.Resolved:
		return problem(findings.Fail, CodeProviderError, fmt.Sprintf("the checkout %s is missing or is not a git repository", g.Checkout))
	case g.PullFailed:
		return problem(findings.Fail, CodeProviderError, "refreshing the tags of "+g.Checkout+" failed")
	}

	var out []findings.Finding
	if g.Behind {
		out = append(out, problem(findings.Drift, CodeBehindLatestTag,
			fmt.Sprintf("pinned %s, the latest release is %s", g.PinnedVersion, g.LatestVersion))...)
	}
	switch {
	case g.NoRef:
		out = append(out, problem(findings.Fail, CodePinMissing, "it pins neither a commit nor a version to diff from")...)
	case g.RefMissing:
		out = append(out, problem(findings.Fail, CodeProviderError, fmt.Sprintf("the pinned ref %s is not in %s", g.Ref, g.Checkout))...)
	case g.CommitCount > 0:
		message := fmt.Sprintf("%d commits changed the watched paths since %s", g.CommitCount, g.Ref)
		if len(g.ReviewFeeds) > 0 {
			message += "; review " + strings.Join(g.ReviewFeeds, ", ")
		}
		out = append(out, problem(findings.Drift, CodeGitFileChanged, message)...)
	}
	return out
}
