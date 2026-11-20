// Package rules judges a source: it takes the source and what the transports
// observed and returns findings. Every function is pure. The clock, the files
// and the network come in as values, so a rule never reaches for them.
package rules

import "github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

// CheckName is the name findings of this package carry.
const CheckName = "audit"

// Finding codes of the audit rules.
const (
	// CodeReviewOverdue marks a source whose review window has run out.
	CodeReviewOverdue = "audit.review-overdue"
	// CodeMetadataChanged marks a source whose recorded digest or pin no longer
	// matches what is observed or what the base ref recorded.
	CodeMetadataChanged = "audit.metadata-changed"
	// CodeAnchorMissing marks a claim whose anchor is absent from the source.
	CodeAnchorMissing = "audit.anchor-missing"
	// CodeGitFileChanged marks watched files that changed since the pin.
	CodeGitFileChanged = "audit.git-file-changed"
	// CodeBehindLatestTag marks a pin older than the latest release.
	CodeBehindLatestTag = "audit.behind-latest-tag"
	// CodeProviderError marks a source that its provider could not deliver.
	CodeProviderError = "audit.provider-error"
	// CodeProviderRestricted marks a source that its provider refused to show.
	CodeProviderRestricted = "audit.provider-restricted"
	// CodeContentUnverified marks a recorded digest that could not be checked.
	CodeContentUnverified = "audit.content-unverified"
	// CodeReferenceMissing marks a source that maps to no reference file.
	CodeReferenceMissing = "audit.reference-missing"
	// CodeReferenceDangling marks a mapping to a file that does not exist.
	CodeReferenceDangling = "audit.reference-dangling"
	// CodeReferenceUncovered marks a reference file that no source backs.
	CodeReferenceUncovered = "audit.reference-uncovered"
	// CodePinMissing marks a pinned version that nothing can verify.
	CodePinMissing = "audit.pin-missing"
	// CodeNoSources marks a catalog that lists no source.
	CodeNoSources = "audit.no-sources"
	// CodeSourceRemoved marks a source that the head catalog dropped.
	CodeSourceRemoved = "audit.source-removed"
	// CodeReviewRegressed marks a review date that moved back.
	CodeReviewRegressed = "audit.review-regressed"
	// CodeOK marks a catalog with no problem.
	CodeOK = "audit.ok"
)

func finding(level findings.Level, code, file, message string) findings.Finding {
	return findings.Finding{Check: CheckName, Level: level, Code: code, Path: file, Message: message}
}
