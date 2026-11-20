package rules

import (
	"fmt"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

const dateLayout = "2006-01-02"

// Windows are the review windows of an audit, in days.
type Windows struct {
	// Default applies to a source with no pinned version.
	Default int
	// Versioned applies to a source that pins a version, because a release moves
	// faster than a page.
	Versioned int
}

// Review returns the finding for a source that is due for a review: its last
// review is older than its window, or its own review date has come. A source
// that another one supersedes needs no review.
func Review(source sources.Source, windows Windows, now time.Time) []findings.Finding {
	if source.SupersededBy != "" {
		return nil
	}
	window := windows.Default
	if source.Version != "" {
		window = windows.Versioned
	}
	today := now.UTC().Format(dateLayout)

	var reasons []string
	if age, ok := ageDays(source.ReviewedAt, now); ok && age > window {
		reasons = append(reasons, fmt.Sprintf("reviewed %d days ago, the window is %d days", age, window))
	}
	if source.ReviewAfter != "" && source.ReviewAfter <= today {
		reasons = append(reasons, "a review was due on "+source.ReviewAfter)
	}
	if len(reasons) == 0 {
		return nil
	}
	return []findings.Finding{finding(findings.Drift, CodeReviewOverdue, source.File,
		fmt.Sprintf("source %q: %s", source.ID, strings.Join(reasons, "; ")))}
}

// ageDays returns the whole calendar days from the review date, written as
// YYYY-MM-DD, to the day of now. It returns false for a date it cannot read.
func ageDays(reviewed string, now time.Time) (int, bool) {
	from, err := time.Parse(dateLayout, reviewed)
	if err != nil {
		return 0, false
	}
	return int(now.UTC().Truncate(24*time.Hour).Sub(from).Hours() / 24), true
}
