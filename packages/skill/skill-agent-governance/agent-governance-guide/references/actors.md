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
- **Why no registry** — every closed vocabulary here (executor classes, policy outcomes, module kinds, module classifications) is declared once and validated against every declaring site. `role` has no such declaration: organizations name the same authority differently, and a portable contract cannot enumerate them.
- **Not a dispatch key** — capabilities are actor-neutral. No capability, command, or adapter selects behavior from a role string, and none may start; that would couple the machinery to job titles it cannot portably know.

The role strings that appear in fixtures — `pm`, `developer`, `release-approver`, `production-approver`, `transition-approver`, `retirement-approver`, `emergency-approver` — are illustrative labels, not a vocabulary. `release-approver` and `production-approver` are neither two authorities nor one authority named twice: in those fixtures `release-approver` accompanies an `integration` action while `production-approver` accompanies a `release` action, so the names do not track even the actions they sit beside. A profile that needs them distinct declares and defines both; a profile that needs one declares one. This plane neither distinguishes nor unifies them.

## Independence

Independence constrains who may decide about a subject relative to who produced it. It is a property of a single decision, and it can mean three different things:

| Level                   | The deciding actor differs from the subject's author by | Catches                                                      |
| ----------------------- | ------------------------------------------------------- | ------------------------------------------------------------ |
| `distinct-identity`     | identity                                                | An author clearing their own subject under the same identity |
| `distinct-team`         | identity and reporting line                             | A reviewer who shares the author's delivery incentive        |
| `distinct-organization` | identity and legal entity                               | A supplier certifying its own deliverable                    |

Every level requires the deciding actor to differ from the author by identity; the stronger two add a requirement on top of that. **Only the identity component is enforced, and only as an inequality between two identity strings.** It does not detect one person holding two identities, a shared or service account, a delegated approval, or teammates approving each other in turn. The reporting-line and legal-entity components of `distinct-team` and `distinct-organization` are not expressible in the actor record at all — it carries no team or organization field — so a profile requiring either gets the identity comparison from code and must enforce the remainder in its provider and record the native evidence.

**The profile elects the check, never the record under scrutiny.** The acceptance, Review, and emergency-access contracts in **workflow-guide** resolve the required level from the governing profile — one of the three levels above, or `none` to require no comparison — and ignore any independence field the record carries, so a record cannot waive, weaken, or silently skip the comparison applied to it. Resolution fails closed:

| Profile declares                                        | Record state                            | Result                                                                                                      |
| ------------------------------------------------------- | --------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| A level above `none`                                    | Deciding actor equals the author        | `acceptance-independence-failed`, `assessor-independence-failed`, or `emergency-access-independence-failed` |
| A level above `none`                                    | Either identity missing or not a string | `independence-unverifiable`                                                                                 |
| `none`                                                  | Any                                     | No comparison                                                                                               |
| Nothing, or a value outside `none` and the three levels | Any                                     | `independence-requirement-undeclared`                                                                       |

Waiving the comparison therefore takes an explicit `none` from the declaring profile — the authority that owns the requirement — and a profile that is silent fails rather than passes.

Autonomy's oversight coupling depends on independence from `A1` upward — see [autonomy.md](autonomy.md). That dependency resolves to `distinct-identity` today, so it is only as strong as the identity comparison above.

## Segregation of duties

Segregation of duties assigns conflicting duties to different actors across a lifecycle, so that no single actor can both create a risk and clear it. Independence is one instrument of it — one decider weighed against one other party inside a single record — not the whole of it. Duties a profile separates typically include:

- authoring a subject and accepting it;
- requesting a privileged action and authorizing it;
- authorizing an action and executing it;
- approving an exception and relying on it;
- operating a control and assessing its effectiveness.

Two pairs are enforced, each through the identity comparison above and each within a single record: authoring a subject and accepting it, and approving an exception and relying on it. The second compares an exception or emergency-exception approver's `identity` against the `owner` the access is held by, resolving the requirement from the profile like every other independence check.

That second comparison is weaker than its name suggests. `owner` names the party the access is held by and carries no declared shape: it is one unstructured value, checked for presence, and it legitimately ranges over parties that are not people — `team:checkout` is a valid owner. So the comparison catches an owner that repeats the approver's identity verbatim, and nothing else. An owner naming a team or group the approver belongs to, or a second identity the approver holds, reads as a different party and passes; an owner that is not a string at all fails closed as `independence-unverifiable` rather than being interpreted. The remaining pairs are contracts a profile declares and its provider enforces; this plane records the resulting actors but never compares them across decisions.

## What code enforces today

| Contract                                                                            | Status           | Rests on                                                                                                                         |
| ----------------------------------------------------------------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| A non-empty `identity` and `role` are present                                       | Enforced         | Acceptance, authorization, and emergency-access contracts reject missing or empty values                                         |
| `type` is `human` and `accountable` is `true`                                       | Enforced         | Acceptance decisions, authorizations, and exception or emergency-exception approvers                                             |
| The **value** of `role`                                                             | **Not enforced** | Nothing reads it; any non-empty string passes                                                                                    |
| The identity component of every independence level                                  | Enforced         | Inequality of two identity strings, resolved from the profile; fails closed when the profile is silent or an identity is missing |
| The team and organization components of `distinct-team` and `distinct-organization` | **Not enforced** | Not expressible in the actor record                                                                                              |
| One person holding two identities, or a shared account                              | **Not enforced** | The provider that issues and authenticates identities                                                                            |
| Approving an exception and relying on it                                            | Enforced         | Inequality of an exception or emergency-exception approver's `identity` and the access `owner`, resolved from the profile        |
| An `owner` naming a team, a group, or any other non-person party                    | **Not enforced** | `owner` has no declared shape; only an owner repeating the approver's identity verbatim is caught                                |
| Segregation of duties beyond those two pairs                                        | **Not enforced** | The profile and its provider                                                                                                     |
| `identity` is authentic and belongs to the actor                                    | **Not enforced** | The provider that issues and authenticates it                                                                                    |

Nothing gates on a role. A reader who sees `release-approver` on a decision learns what the actor claimed, not that a control tested the claim. Everything marked not enforced is a contract a profile and its provider must uphold: recording it proves the claim was made, not that a control verified it — the same rule that makes an installed module no evidence that a control executes, per [policy-and-authority.md](policy-and-authority.md).
