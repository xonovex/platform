// Package trust is the gate a project file passes before its gated sections
// take effect: the user approves the sha256 digest of exactly the file that was
// reviewed, at a terminal prompt or with a flag, and the approval is kept in a
// ledger. A file that is not approved fails closed.
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

// ErrNotApproved reports a project file whose gated sections the user has not
// approved.
var ErrNotApproved = errors.New("project configuration not approved")

// Approval is one approved file digest.
type Approval struct {
	Digest string `json:"digest"`
	// Path is the file the digest was approved for, kept for a reader of the
	// ledger. The digest alone decides.
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

// Request is one gate check.
type Request struct {
	Document layer.Document
	// Sections are the gated sections the document sets that the tool reads.
	Sections []string
	// Flag is the digest the user approved on the command line, empty when
	// none.
	Flag string
}

// Gate lets the request through when the digest of the document is approved,
// approved by the flag, or approved at the terminal; otherwise it fails with
// ErrNotApproved. A request without sections always passes.
func Gate(request Request, approvals Approvals, prompter Prompter, out io.Writer) error {
	if len(request.Sections) == 0 {
		return nil
	}
	document := request.Document
	approved, err := approvals.Approved(document.Digest)
	if err != nil || approved {
		return err
	}
	approval := Approval{Digest: document.Digest, Path: document.Path}
	sections := strings.Join(request.Sections, ", ")
	switch {
	case request.Flag != "":
		if request.Flag != document.Digest {
			return fmt.Errorf("%w: --%s %s differs from the digest %s of %s", ErrNotApproved, ApproveFlag, request.Flag, document.Digest, document.Path)
		}
		return approvals.Approve(approval)
	case prompter.IsTerminal():
		if _, err := fmt.Fprintf(out, "%s sets the %s configuration, which takes effect only after you approve the digest of the file.\n\n%s\n\nDigest: %s\n", document.Path, sections, strings.TrimRight(string(document.Data), "\n"), document.Digest); err != nil {
			return err
		}
		yes, err := ask(prompter, out)
		if err != nil {
			return err
		}
		if !yes {
			// A character device that is not a terminal, such as /dev/null, also
			// reaches the prompt, so the refusal names the flag as well.
			return fmt.Errorf("%w: %s was declined; to approve it without the prompt, run again with --%s %s", ErrNotApproved, document.Path, ApproveFlag, document.Digest)
		}
		return approvals.Approve(approval)
	default:
		return fmt.Errorf("%w: %s sets the %s configuration, which takes effect only after you approve the digest of the file: review it, then run again with --%s %s", ErrNotApproved, document.Path, sections, ApproveFlag, document.Digest)
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
