# Instrumentation and Quality

## Data Contract

Specify event name and semantic version, trigger, producer, required and optional
properties, types and allowed values, identity and session rules, timestamp semantics,
consent basis, sensitivity classification, retention, consumers, owner, and change
policy. Keep environment and synthetic-traffic markers explicit.

Collect the minimum data needed for the named decision. Review privacy, security,
consent, purpose limitation, retention, deletion, access, and cross-border obligations
with the accountable specialists. Do not put secrets or unnecessary free text into
analytics events.

Implementation belongs to the relevant platform analytics skill. Preserve the semantic
contract independently from a vendor SDK or dashboard.

## Quality Controls

Validate expected event sequences and properties before release and reconcile client,
server, transaction, support, or finance sources where possible. Monitor:

- completeness and missingness;
- schema validity and allowed values;
- duplicate, retry, and ordering behavior;
- identity stitching and eligibility;
- freshness, delay, and backfill;
- test, staff, bot, and fraud traffic;
- historical discontinuities and metric drift.

Version material definition changes. Run old and new definitions in parallel when
comparability matters, annotate breaks, and recompute only with a documented method.
