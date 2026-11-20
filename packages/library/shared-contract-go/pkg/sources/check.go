package sources

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
)

// Finding codes of the catalog checks.
const (
	CodeMissing         = "sources.missing"
	CodeSchema          = "sources.schema"
	CodeDuplicateID     = "sources.duplicate-id"
	CodeDuplicateClaim  = "sources.duplicate-claim"
	CodeSupersededSelf  = "sources.superseded-self"
	CodeSupersededUnset = "sources.superseded-unknown"
	CodeSupersededCycle = "sources.superseded-cycle"
	CodeReviewOrder     = "sources.review-order"
	CodeClaimUnanchored = "sources.claim-unanchored"
	CodeOK              = "sources.ok"
)

func finding(level findings.Level, code, file, message string) findings.Finding {
	return findings.Finding{Check: CheckName, Level: level, Code: code, Path: file, Message: message}
}

// Check returns the findings that hold across the sources of a catalog: ids
// and claim ids are unique, a supersession names a source that exists and does
// not loop, a review falls after the last one, and each claim has an anchor. A
// catalog with none of these problems yields one pass finding.
func Check(c Catalog) []findings.Finding {
	var out []findings.Finding
	out = append(out, checkIDs(c)...)
	out = append(out, checkSupersession(c)...)
	out = append(out, checkReviewOrder(c)...)
	out = append(out, checkClaims(c)...)
	if len(out) == 0 {
		out = append(out, finding(findings.Pass, CodeOK, "", fmt.Sprintf("%d sources hold together", len(c.Sources))))
	}
	return out
}

func checkIDs(c Catalog) []findings.Finding {
	var out []findings.Finding
	first := map[string]Source{}
	for _, s := range c.Sources {
		if earlier, dup := first[s.ID]; dup {
			out = append(out, finding(findings.Fail, CodeDuplicateID, s.File,
				fmt.Sprintf("source id %q repeats the one in %s", s.ID, earlier.File)))
			continue
		}
		first[s.ID] = s
	}
	return out
}

func checkSupersession(c Catalog) []findings.Finding {
	var out []findings.Finding
	next := map[string]string{}
	for _, s := range c.Sources {
		if s.SupersededBy == "" {
			continue
		}
		if s.SupersededBy == s.ID {
			out = append(out, finding(findings.Fail, CodeSupersededSelf, s.File,
				fmt.Sprintf("source %q supersedes itself", s.ID)))
			continue
		}
		if _, ok := c.ByID(s.SupersededBy); !ok {
			out = append(out, finding(findings.Fail, CodeSupersededUnset, s.File,
				fmt.Sprintf("source %q is superseded by %q, which is not in the catalog", s.ID, s.SupersededBy)))
			continue
		}
		next[s.ID] = s.SupersededBy
	}
	reported := map[string]bool{}
	for _, s := range c.Sources {
		if reported[s.ID] || !loops(next, s.ID) {
			continue
		}
		for id := s.ID; !reported[id]; id = next[id] {
			reported[id] = true
		}
		out = append(out, finding(findings.Fail, CodeSupersededCycle, s.File,
			fmt.Sprintf("source %q is part of a supersession cycle", s.ID)))
	}
	return out
}

// loops reports whether following next from start returns to start.
func loops(next map[string]string, start string) bool {
	for id, steps := next[start], 0; id != "" && steps <= len(next); id, steps = next[id], steps+1 {
		if id == start {
			return true
		}
	}
	return false
}

func checkReviewOrder(c Catalog) []findings.Finding {
	var out []findings.Finding
	for _, s := range c.Sources {
		if s.ReviewAfter != "" && s.ReviewAfter < s.ReviewedAt {
			out = append(out, finding(findings.Fail, CodeReviewOrder, s.File,
				fmt.Sprintf("source %q asks for a review after %s, before it was reviewed on %s", s.ID, s.ReviewAfter, s.ReviewedAt)))
		}
	}
	return out
}

func checkClaims(c Catalog) []findings.Finding {
	var out []findings.Finding
	for _, s := range c.Sources {
		seen := map[string]bool{}
		unanchored := 0
		for _, claim := range s.Claims {
			if seen[claim.ID] {
				out = append(out, finding(findings.Fail, CodeDuplicateClaim, s.File,
					fmt.Sprintf("source %q repeats the claim id %q", s.ID, claim.ID)))
			}
			seen[claim.ID] = true
			if claim.Anchor == "" {
				unanchored++
			}
		}
		if unanchored > 0 {
			out = append(out, finding(findings.Warn, CodeClaimUnanchored, s.File,
				fmt.Sprintf("%d of %d claims of source %q have no anchor", unanchored, len(s.Claims), s.ID)))
		}
	}
	return out
}
