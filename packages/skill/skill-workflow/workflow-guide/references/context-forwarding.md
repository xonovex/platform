# Context Forwarding

Carry the reasoning a later role needs to understand why the subject has its present
shape: decisions, constraints, assumptions, rejected alternatives, tradeoffs. Forward
the minimum that role needs, not everything the current one knows.

Inside one session, context travels with the work and needs no format. Across a cold
boundary it travels as the Decisions section of a handoff, whose shape
[handoffs.md](handoffs.md) defines: what was decided, why, and where in the code.

## Accepting Context

1. Collect inline context and every `--context` reference.
2. For an opaque reference, select the provider capability from explicit provider,
   repository, or destination metadata. Never infer a provider from a reference's
   shape.
3. Resolve it. An unresolved reference is a blocker before effects: report the
   provider, the reference, and what was attempted rather than proceeding without it.
4. Check that the context still applies to the current subject revision. Report stale
   or uncertain applicability instead of silently reusing it.

## Producing Context

Capture a material choice when Create, Revise, Decide, Execute, Abandon, or a
workspace operation makes one. Anchor it to a code location so the next session can
find what the decision is about. Strip secrets, personal data, and privileged security
detail before the context leaves its current audience.

Publishing context to a provider note or comment is a separate Publish operation with
its own preview and apply. Carrying context never publishes it.

## Superseding

A changed decision is a new decision that names the one it replaces. Do not rewrite or
delete an already-published note; add the successor and let the history stand.

Authority and evidence rules are in [governance.md](governance.md).
