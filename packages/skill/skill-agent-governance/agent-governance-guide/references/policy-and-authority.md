# Policy, Enforcement, Actors, and Exceptions

## Policy decision contract

A policy decision records:

- decision reference, policy identity/version, evaluation time, and freshness;
- subject, requested action, actor/executor, authority zone, and contextual facts;
- applicable profile/control references and applicability status;
- outcome, reasons, required remediation or evidence, and expiry;
- evidence origin and limitations.

Supported outcomes are `allow`, `deny`, `ask`, `advise`, `observe`, `require-evidence`, `exception`, and `break-glass`.

## Decision and enforcement separation

The decision point evaluates facts. The enforcement point applies a decision at a harness hook, CI gate, repository rule, protected environment, admission webhook, provider operation, or human control. Record enforcement separately with its native event, subject/revision, applied decision reference, action, outcome, and evidence.

Do not claim a decision was enforced merely because a policy file, skill, or hook module was installed.

## Actors and evidence origin

Profiles declare the actor requirements a decision must satisfy, plus qualification and permitted evidence origin. [actors.md](actors.md) owns the actor record, what a role string may mean, independence, segregation of duties, and which of them code enforces; declaring a requirement there is not the same as enforcing it. Distinguish evidence origin:

- agent recommendation;
- model inference;
- deterministic automated evidence;
- human accountable decision;
- external authoritative evidence.

Evidence is authoritative only for the declared subject, revision, scope, source, policy/control version, and freshness. Qualified legal, security, privacy, accessibility, safety, or domain review cannot be replaced by an agent assertion.

## Exception contract

An exception includes scope, control, owner, authorized approver, rationale, compensating controls, start, expiry, evidence, affected subjects, and mandatory review. It cannot authorize outside the approver's authority or silently become a default.

## Break-glass contract

Break-glass additionally requires emergency reason, time-limited access, explicit invocation, authoritative access-system evidence, immediate notification where applicable, containment, revocation, and post-event review. Expired, ownerless, evidence-free, or unreviewed break-glass state fails closed for mandatory controls.
