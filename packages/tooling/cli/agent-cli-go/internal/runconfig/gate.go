package runconfig

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
)

// Gate lets each pending content through the trust gate in order. The flag
// approves the one content whose digest it names; a flag that names none of
// them is an error, so a stale digest never passes silently.
func Gate(env Env, pending []Pending, flag string) error {
	if err := checkFlag(pending, flag); err != nil {
		return err
	}
	ledger := Ledger(env.Host)
	for _, p := range pending {
		given := ""
		if p.Subject.Digest == flag {
			given = flag
		}
		if err := trust.Decide(p.Subject, given, ledger, env.Host, env.Out); err != nil {
			return err
		}
	}
	return nil
}

func checkFlag(pending []Pending, flag string) error {
	if flag == "" {
		return nil
	}
	for _, p := range pending {
		if p.Subject.Digest == flag {
			return nil
		}
	}
	return fmt.Errorf("%w: --%s %s is the digest of no project file this command reads", trust.ErrNotApproved, trust.ApproveFlag, flag)
}

// Approved reports whether the ledger holds the digest of the content.
func Approved(env Env, p Pending) (bool, error) {
	return Ledger(env.Host).Approved(p.Subject.Digest)
}

// Revoke removes the approval of each content from the ledger and returns the
// content whose approval it removed.
func Revoke(env Env, pending []Pending) ([]Pending, error) {
	var revoked []Pending
	for _, p := range pending {
		held, err := Ledger(env.Host).Revoke(p.Subject.Digest)
		if err != nil {
			return revoked, err
		}
		if held {
			revoked = append(revoked, p)
		}
	}
	return revoked, nil
}
