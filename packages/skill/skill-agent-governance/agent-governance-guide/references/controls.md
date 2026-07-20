# Controls

A control plugin declares an identifier, capabilities, supported phases, and one evaluation function. Its result is `allow`, `deny`, or `abstain`, plus optional reason, data, and evidence references.

The composition, not the plugin, chooses the effect:

- `observe` records every result and never blocks.
- `enforce` stops on `deny`.

This allows one policy implementation to be evaluated in shadow mode before the owner changes the selection to enforcement. It also prevents the same policy from needing separate advisory and mandatory wrappers.

Use `before` for gates that can prevent the executor. Use `after` for result validation or reporting. An after-phase denial changes the composition outcome but does not reverse work already performed. Put irreversible-operation controls at a native before-action enforcement point such as a harness hook, CI required check, admission webhook, or provider gate.

Controls should own one decision concern. Protected targets, approvals, provenance, escalation, budgets, and independent critique are separate plugins when selected; none is silently implied by an executor, trigger, host, maturity label, or evidence sink.
