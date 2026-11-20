// Package git is the transport that compares a source with the local checkout
// of its upstream repository: whether a newer release exists, and which
// commits touched the watched paths since the pinned commit.
package git

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
)

// Name is the name of the transport in reports.
const Name = "git"

// maxListedCommits bounds the commits a report lists.
const maxListedCommits = 20

var (
	releaseTag = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
	semver     = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
)

// Transport reads a checkout through the file-system and process ports.
type Transport struct {
	FS     fs.FileSystem
	Runner process.Runner
}

// Name returns the name of the transport.
func (Transport) Name() string { return Name }

// Remote returns false: the transport reads a local checkout. Refreshing the
// tags is a separate step that the caller asks for with Request.Pull.
func (Transport) Remote() bool { return false }

// Applies reports whether the source names a checkout.
func (Transport) Applies(source sources.Source) bool { return source.Checkout != "" }

// Observe compares the pinned state of the source with the checkout.
//
// The diff starts at the pinned commit, or at the pinned version when there is
// no commit, and ends at the latest release tag, or at HEAD when the repository
// has none. Only the watched paths are compared.
func (t Transport) Observe(ctx context.Context, request observe.Request) (observe.Observation, error) {
	source := request.Source
	dir := path.Join(request.Dir, source.Checkout)
	state := &observe.Git{Checkout: source.Checkout, PinnedVersion: source.Version}
	result := observe.Observation{Transport: Name, Git: state}

	repository, err := fs.Exists(t.FS, path.Join(dir, ".git"))
	if err != nil {
		return observe.Observation{}, fmt.Errorf("stat %s: %w", dir, err)
	}
	if !repository {
		return result, nil
	}
	state.Resolved = true

	git := func(args ...string) (process.Captured, error) {
		return process.Capture(ctx, t.Runner, process.Request{Dir: dir, Name: "git", Args: args})
	}

	if request.Pull {
		fetched, err := git("fetch", "--tags", "--quiet")
		if err != nil {
			return observe.Observation{}, err
		}
		if !fetched.Succeeded() {
			state.PullFailed = true
			return result, nil
		}
	}

	latest, err := t.latestRelease(git)
	if err != nil {
		return observe.Observation{}, err
	}
	state.LatestVersion = latest.raw
	if pinned, ok := parseSemver(source.Version); ok && latest.raw != "" {
		state.Behind = compare(latest.version, pinned) > 0
	}

	state.Ref = source.Commit
	if state.Ref == "" {
		state.Ref = source.Version
	}
	if state.Ref == "" {
		state.NoRef = true
		return result, nil
	}
	known, err := git("rev-parse", "--verify", "--quiet", state.Ref+"^{commit}")
	if err != nil {
		return observe.Observation{}, err
	}
	if !known.Succeeded() {
		state.RefMissing = true
		return result, nil
	}

	end := "HEAD"
	if latest.raw != "" {
		end = latest.raw
	}
	return result, t.diff(git, source, state, state.Ref+".."+end)
}

type gitFunc func(args ...string) (process.Captured, error)

func (t Transport) diff(git gitFunc, source sources.Source, state *observe.Git, rng string) error {
	paths := make([]string, 0, len(source.Watch))
	for _, w := range source.Watch {
		paths = append(paths, w.Path)
	}
	suffix := []string{}
	if len(paths) > 0 {
		suffix = append([]string{"--"}, paths...)
	}

	log, err := git(append([]string{"log", "--oneline", rng}, suffix...)...)
	if err != nil {
		return err
	}
	commits := lines(log)
	state.CommitCount = len(commits)
	if len(commits) > maxListedCommits {
		commits = commits[:maxListedCommits]
	}
	state.Commits = commits

	names, err := git(append([]string{"diff", "--name-only", rng}, suffix...)...)
	if err != nil {
		return err
	}
	state.ChangedFiles = lines(names)
	state.ReviewFeeds = reviewFeeds(source.Watch, state.ChangedFiles)
	return nil
}

// reviewFeeds returns the sorted files that draw on a watched path that changed.
func reviewFeeds(watch []sources.Watch, changed []string) []string {
	set := map[string]struct{}{}
	for _, file := range changed {
		for _, w := range watch {
			if strings.HasPrefix(file, w.Path) {
				for _, feed := range w.Feeds {
					set[feed] = struct{}{}
				}
			}
		}
	}
	out := make([]string, 0, len(set))
	for feed := range set {
		out = append(out, feed)
	}
	sort.Strings(out)
	return out
}

type release struct {
	raw     string
	version [3]int
}

// latestRelease returns the highest tag of the form v1.2.3 or 1.2.3, or the
// zero release when the repository has none. A failing listing reads as none.
func (t Transport) latestRelease(git gitFunc) (release, error) {
	tags, err := git("tag", "--list")
	if err != nil {
		return release{}, err
	}
	var best release
	for _, line := range lines(tags) {
		match := releaseTag.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		candidate := release{raw: line, version: [3]int{atoi(match[1]), atoi(match[2]), atoi(match[3])}}
		if best.raw == "" || compare(candidate.version, best.version) > 0 {
			best = candidate
		}
	}
	return best, nil
}

func parseSemver(text string) ([3]int, bool) {
	match := semver.FindStringSubmatch(text)
	if match == nil {
		return [3]int{}, false
	}
	return [3]int{atoi(match[1]), atoi(match[2]), atoi(match[3])}, true
}

func compare(a, b [3]int) int {
	for i := range a {
		if a[i] != b[i] {
			return a[i] - b[i]
		}
	}
	return 0
}

// atoi converts digits the pattern already matched. A value too large for an
// int cannot be a real release number and reads as zero.
func atoi(digits string) int {
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0
	}
	return n
}

// lines returns the trimmed, non-empty lines of the output of a successful
// command, and none for a command that failed.
func lines(result process.Captured) []string {
	if !result.Succeeded() {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(result.Stdout), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}
