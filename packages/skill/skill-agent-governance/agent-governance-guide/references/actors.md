# Actor Record, Role, and Independence

## The actor record

Every consequential decision — acceptance, authorization, exception approval, emergency-exception approval — records the actor that made it:

| Field         | Contract                                                                                                                                                                                   |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `identity`    | The actor's identifier as issued by the provider that authenticated it. Opaque and stable; never a display name, a self-asserted string, or a value the governed actor picks per decision. |
| `role`        | An open audit label naming the authority the actor acted under. Scoped to the declaring profile.                                                                                           |
| `type`        | The actor's executor class, drawn from the canonical vocabulary in [execution.md](execution.md). An accountable decision requires `human`.                                                 |
| `accountable` | Whether this actor bears the consequence of the decision. Assembling evidence or advising never makes an actor accountable.                                                                |

These are semantic requirements, not field names. A contract names the record for its own decision — `actor` on an acceptance or authorization, `approver` on an exception or emergency exception — and a Review carries only a bare assessor identity rather than a full record.

The record states who decided. What a profile requires before a decision counts is a separate declaration — see [policy-and-authority.md](policy-and-authority.md) and [profiles.md](profiles.md). Recording an actor is evidence about a decision, never a permission check on it.

## `role` is an open audit label

`role` names the authority an actor claims to act under so a later reader can audit the decision. Nothing resolves behavior from its value, and no closed vocabulary exists.

- **Who mints one** — the profile that declares the actor requirement, within its authority zone. A role string means what that profile says it means and nothing outside it.
- **Why no registry** — every closed vocabulary here (executor classes, policy outcomes, module classifications, adoption modes, authority zones, profile facets) is declared once by a machine-readable owner and validated against every machine-readable declaring site by a CI guard; module kinds are pinned to that owner through the conformance fixture, with the `modules.md` table a human-readable view. `role` has no such declaration: organizations name the same authority differently, and a portable contract cannot enumerate them.
- **Not a dispatch key** — capabilities are actor-neutral. No capability, command, or adapter selects behavior from a role string, and none may start; that would couple the machinery to job titles it cannot portably know.

The role strings that appear in fixtures — `pm`, `developer`, `release-approver`, `production-approver`, `transition-approver`, `retirement-approver`, `emergency-approver` — are illustrative labels, not a vocabulary. `release-approver` and `production-approver` are neither two authorities nor one authority named twice: in those fixtures `release-approver` accompanies an `integration` action while `production-approver` accompanies a `release` action, so the names do not track even the actions they sit beside. A profile that needs them distinct declares and defines both; a profile that needs one declares one. This plane neither distinguishes nor unifies them.

## Independence

Independence constrains who may decide about a subject relative to who produced it. It is a property of a single decision, and it can mean three different things:

| Level                   | The deciding actor differs from the subject's author by | Catches                                                      |
| ----------------------- | ------------------------------------------------------- | ------------------------------------------------------------ |
| `distinct-identity`     | identity                                                | An author clearing their own subject under the same identity |
| `distinct-team`         | identity and reporting line                             | A reviewer who shares the author's delivery incentive        |
| `distinct-organization` | identity and legal entity                               | A supplier certifying its own deliverable                    |

Every level above `none` requires the deciding actor to differ from the author by identity, and that inequality between two identity strings is all the code compares for the separation itself. `distinct-team` and `distinct-organization` demand reporting-line or legal-entity separation on top of that. That separation is not expressible in the actor record — it carries no team or organization field — and not resolvable here, because membership and legal entity are provider state. So for those two levels the record must carry `independenceEvidenceReference`, a reference to the provider evidence that the provider checked the separation: a distinct decider whose record omits or empties it fails closed as `independence-provider-evidence-required` rather than passing on the identity comparison alone. The plane therefore no longer silently downgrades a stronger level to identity-only — but the code still does not verify membership or legal entity, only that the provider's evidence was recorded, so a real team or organization resolution remains the provider's job. The self-grant identity inequality is evaluated before the evidence requirement, so an actor clearing their own subject is caught at every level whether or not evidence is present. None of this detects one person holding two identities, a shared or service account, a delegated approval, or teammates approving each other in turn.

**The profile elects the check, never the record under scrutiny.** The acceptance, Review, and emergency-access contracts in **workflow-guide** resolve the required level from the governing profile — one of the three levels above, or `none` to require no comparison — and ignore any independence field the record carries, so a record cannot waive, weaken, or silently skip the comparison applied to it. Resolution fails closed:

| Profile declares                                        | Record state                                                  | Result                                                                                                      |
| ------------------------------------------------------- | ------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| A level above `none`                                    | Deciding actor equals the author                              | `acceptance-independence-failed`, `assessor-independence-failed`, or `emergency-access-independence-failed` |
| A level above `none`                                    | Either identity missing or not a string                       | `independence-unverifiable`                                                                                 |
| `distinct-team` or `distinct-organization`              | Distinct decider, no `independenceEvidenceReference` recorded | `independence-provider-evidence-required`                                                                   |
| `none`                                                  | Any                                                           | No comparison                                                                                               |
| Nothing, or a value outside `none` and the three levels | Any                                                           | `independence-requirement-undeclared`                                                                       |

Waiving the comparison therefore takes an explicit `none` from the declaring profile — the authority that owns the requirement — and a profile that is silent fails rather than passes.

Autonomy's oversight coupling depends on independence from `A1` upward — see [autonomy.md](autonomy.md). That dependency resolves to `distinct-identity` today, so it is only as strong as the identity comparison above.

**Provider contract — one stable principal per actor.** Even `distinct-identity` is only as strong as the identity provider beneath it. The code compares two `identity` strings; it cannot tell that both belong to the same person, that an account is shared, or that one actor authenticated under a second issued identity. Closing that ceiling is the identity provider's contract, not the plane's: the provider that issues and authenticates the `identity` field must guarantee a stable, unique principal per real actor — one principal per person, no shared accounts, and no one-person-many-identities — and must not reissue or alias a principal across actors. `distinct-identity`, and the identity component of every stronger level, is exactly as strong as that guarantee and no stronger; where the provider cannot make it, the profile must add a compensating control. This contract is stated here because no separate identity-provider port owns it: the plane records the resulting `identity` but never proves it, the same way the rows below attribute identity authenticity and uniqueness to that provider.

## Segregation of duties

Segregation of duties assigns conflicting duties to different actors across a lifecycle, so that no single actor can both create a risk and clear it. Independence is one instrument of it — one decider weighed against one other party inside a single record — not the whole of it. Duties a profile separates typically include:

- authoring a subject and accepting it;
- requesting a privileged action and authorizing it;
- authorizing an action and executing it;
- approving an exception and relying on it;
- operating a control and assessing its effectiveness.

Two pairs are enforced, each through the identity comparison above and each within a single record: authoring a subject and accepting it, and approving an exception and relying on it. The second compares an exception or emergency-exception approver's `identity` against the `owner` the access is held by, resolving the requirement from the profile like every other independence check.

That second comparison is weaker than its name suggests, though it now reads a validated shape. `owner` names the party the access is held by as a **party reference** — a `<kind>:<id>` string whose `kind` is one of a closed set of party kinds (`human`, `team`, `service`, `role`) that can be accountable for holding an access. It legitimately ranges over parties that are not people — `team:checkout` is a valid owner — but a malformed reference, one whose kind is outside that set (the advisory executor kinds `agent` and `model` are excluded), or a non-string owner such as an object now fails closed as `emergency-access-contract-incomplete` before any comparison, rather than passing on presence alone or being read as `independence-unverifiable`. The comparison itself still only catches a **person-owner** (`human:` kind) whose id equals the approver's identity — a self-grant — and nothing more. An owner naming a team or role the approver belongs to, or a second identity the same person holds, reads as a different party and passes: membership and identity linkage are not expressible in the record. The remaining pairs are contracts a profile declares and its provider enforces; this plane records the resulting actors but never compares them across decisions.

## What code enforces today

| Contract                                                                               | Status           | Rests on                                                                                                                                                                               |
| -------------------------------------------------------------------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A non-empty `identity` and `role` are present                                          | Enforced         | Acceptance, authorization, and emergency-access contracts reject missing or empty values                                                                                               |
| `type` is `human` and `accountable` is `true`                                          | Enforced         | Acceptance decisions, authorizations, and exception or emergency-exception approvers                                                                                                   |
| The **value** of `role`                                                                | **Not enforced** | Nothing reads it; any non-empty string passes                                                                                                                                          |
| The identity component of every independence level                                     | Enforced         | Inequality of two identity strings, resolved from the profile; fails closed when the profile is silent or an identity is missing                                                       |
| A recorded provider-evidence reference for `distinct-team` and `distinct-organization` | Enforced         | A distinct decider whose record omits or empties `independenceEvidenceReference` fails closed as `independence-provider-evidence-required`; the self-grant inequality is checked first |
| The team and organization separation of `distinct-team` and `distinct-organization`    | **Not enforced** | The code checks only that the provider's evidence was recorded, never membership or legal entity — that separation is not expressible in the actor record                              |
| One person holding two identities, or a shared account                                 | **Not enforced** | The provider that issues and authenticates identities                                                                                                                                  |
| Approving an exception and relying on it                                               | Enforced         | Inequality of an exception or emergency-exception approver's `identity` and the access `owner`, resolved from the profile                                                              |
| `owner` is a well-formed party reference (`<kind>:<id>`, `kind` in the party set)      | Enforced         | The emergency-access, incident, and corrective-action contracts each reject a malformed, unknown-kind, or non-string owner as their own contract-incomplete code                       |
| A person-owner (`human:` kind) whose id equals the approver's identity (self-grant)    | Enforced         | Inequality of the approver's `identity` and the validated `owner`, resolved from the profile                                                                                           |
| A team, group, or role owner the approver belongs to, or a second identity they hold   | **Not enforced** | Membership and identity linkage are not expressible in the record; the comparison reads only the two party-reference strings                                                           |
| Segregation of duties beyond those two pairs                                           | **Not enforced** | The profile and its provider                                                                                                                                                           |
| `identity` is authentic and belongs to the actor                                       | **Not enforced** | The provider that issues and authenticates it                                                                                                                                          |

Nothing gates on a role. A reader who sees `release-approver` on a decision learns what the actor claimed, not that a control tested the claim. Everything marked not enforced is a contract a profile and its provider must uphold: recording it proves the claim was made, not that a control verified it — the same rule that makes an installed module no evidence that a control executes, per [policy-and-authority.md](policy-and-authority.md).
