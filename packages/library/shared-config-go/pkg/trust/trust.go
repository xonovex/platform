// Package trust is the one gate that content a user did not write passes
// before it takes effect: a project file, an override shard beside it or a
// plugin. The user approves the sha256 digest of exactly the content that was
// reviewed, at a terminal prompt or with a flag, and the approval is kept in a
// ledger. Content that is not approved fails closed.
package trust

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
)

// ApproveFlag is the name of the flag that approves a digest without a
// terminal.
const ApproveFlag = "approve-config-digest"

// ErrNotApproved reports content whose digest the user has not approved.
var ErrNotApproved = errors.New("not approved")

// Approval is one approved digest.
type Approval struct {
	Digest string `json:"digest"`
	// Path names the content the digest was approved for, kept for a reader of
	// the ledger. The digest alone decides.
	Path string `json:"path"`
}

// Approvals is the port that keeps the approved digests.
type Approvals interface {
	Approved(digest string) (bool, error)
	Approve(approval Approval) error
}

// Prompter is the port to the terminal of the user.
type Prompter interface {
	IsTerminal() bool
	// ReadLine returns one line of input without its line ending, or io.EOF on
	// a closed input.
	ReadLine() (string, error)
}

// Subject is content that takes effect only after the user approves its
// digest.
type Subject struct {
	// Digest is the sha256 digest of exactly the content that Review shows or
	// summarises.
	Digest string
	// Path names the content in the ledger and in the messages.
	Path string
	// Notice says what the content does and that it waits for the approval, as
	// one sentence without its full stop.
	Notice string
	// Review is the text a person reads before answering.
	Review string
}

// Decide lets the subject through when its digest is approved, approved by the
// flag, or approved at the terminal, and records a new approval in the
// ledger; otherwise it fails with ErrNotApproved.
func Decide(subject Subject, flag string, approvals Approvals, prompter Prompter, out io.Writer) error {
	approved, err := approvals.Approved(subject.Digest)
	if err != nil || approved {
		return err
	}
	approval := Approval{Digest: subject.Digest, Path: subject.Path}
	switch {
	case flag != "":
		if flag != subject.Digest {
			return fmt.Errorf("%w: --%s %s differs from the digest %s of %s", ErrNotApproved, ApproveFlag, flag, subject.Digest, subject.Path)
		}
		return approvals.Approve(approval)
	case prompter.IsTerminal():
		if _, err := fmt.Fprintf(out, "%s.\n\n%s\n\nDigest: %s\n", subject.Notice, strings.TrimRight(subject.Review, "\n"), subject.Digest); err != nil {
			return err
		}
		yes, err := ask(prompter, out)
		if err != nil {
			return err
		}
		if !yes {
			// A character device that is not a terminal, such as /dev/null, also
			// reaches the prompt, so the refusal names the flag as well.
			return fmt.Errorf("%w: %s was declined; to approve it without the prompt, run again with --%s %s", ErrNotApproved, subject.Path, ApproveFlag, subject.Digest)
		}
		return approvals.Approve(approval)
	default:
		return fmt.Errorf("%w: %s: review it, then run again with --%s %s", ErrNotApproved, subject.Notice, ApproveFlag, subject.Digest)
	}
}

// Request is one gate check of a project file.
type Request struct {
	Document layer.Document
	// Sections are the gated sections the document sets that the tool reads.
	Sections []string
	// Flag is the digest the user approved on the command line, empty when
	// none.
	Flag string
}

// Gate decides on the file of the request, whose whole content is the
// review. A request without sections always passes.
func Gate(request Request, approvals Approvals, prompter Prompter, out io.Writer) error {
	if len(request.Sections) == 0 {
		return nil
	}
	return Decide(FileSubject(request.Document, request.Sections), request.Flag, approvals, prompter, out)
}

// FileSubject is the subject of a project file that sets gated sections: the
// review is the file itself.
func FileSubject(document layer.Document, sections []string) Subject {
	return Subject{
		Digest: document.Digest,
		Path:   document.Path,
		Notice: fmt.Sprintf("%s sets the %s configuration, which takes effect only after you approve the digest of the file", document.Path, strings.Join(sections, ", ")),
		Review: string(document.Data),
	}
}

// ask asks the user to approve the digest just shown. Only y and yes approve; an
// empty answer, any other answer or the end of the input decline.
func ask(prompter Prompter, out io.Writer) (bool, error) {
	if _, err := fmt.Fprint(out, "Approve this digest? [y/N] "); err != nil {
		return false, err
	}
	answer, err := prompter.ReadLine()
	if err != nil {
		// The question stays unanswered on a closed input, so nothing is approved.
		_, _ = fmt.Fprintln(out)
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, fmt.Errorf("read the answer: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
