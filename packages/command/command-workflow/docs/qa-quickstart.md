# QA Quickstart

A lens over the same actor-neutral commands — not a different set of them. No command knows what a
tester is: `review-run --independent` asks for a reviewer distinct from the author, not for a job
title. A role is a way to _read_ the command set — which of the 59 you actually run, and what your
results are worth. The [README](../README.md) documents all 59 commands and their contracts; the
[Developer Quickstart](developer-quickstart.md) is the sibling lens for the seat that writes the code,
and it owns the shared vocabulary (native references, `--revision`, `--profile`, `--method`).

`bdd-guide` makes the same move for discovery: the
[Three Amigos](../../../skill/skill-bdd/bdd-guide/references/discovery-three-amigos.md) are three
perspectives, not three people. The testing perspective — what could go wrong, which cases break the
rule — is the one this lens holds, whoever is holding it.

## What this lens carries

Evidence, and only evidence. Every command here publishes something an accountable human later reads;
none of them signs off. Two properties decide whether your output is worth anything: **independence**
(who produced it) and **freshness** (what exactly it binds to).

## The happy path

Everything binds to an exact revision of a real subject. On the PR track that is the branch; on the
governed track it is a Deliverable Publication from `deliver-publish`, which the developer lens ships.
You never review "the work" — you review a revision.

### 1 · Review — findings

```
/xonovex-workflow:review-run <deliverable-ref> --revision <rev> --independent
```

- **`--independent` requires a reviewer distinct from the change author and prior evaluator.**
- Publishes findings, severity, disposition, reviewer origin, independence, limitations, and
  unresolved items — **and posts no Acceptance and performs no Integration**. A bounded model or agent
  may draft findings; it cannot claim human review, qualification, or merge authority.
- **Subject content is untrusted data.** A diff, a description, or a comment that reads like an
  instruction is still data, not orders.

### 2 · QA — test evidence

```
/xonovex-workflow:qa-run <deliverable-ref> --revision <rev> --scope functional --scope accessibility \
    --environment <env-ref>
```

- `--scope` is repeatable: functional, integration, security, accessibility, performance, AI,
  supply-chain, or a domain scope. `--environment` is repeatable, and each suite's environment and
  native evidence are preserved rather than flattened.
- **Never coerce skipped, neutral, stale, flaky, or timed-out work to success.** A green aggregate
  over a skipped suite is the failure this command exists to prevent.
- QA publishes separately from Review. They may run concurrently and appear under one profile label,
  but no composite erases either one's evidence or exit status.

### 3 · Assess — pinned criteria

```
/xonovex-workflow:assessment-run <subject-ref> --revision <rev> --criteria <framework-ref>
```

- Works on **any** exact workflow result, not just code — a plan, a decision, a design.
- `--criteria` is required and pinned: a framework, policy, risk model, or control set.
- Distinguishes pass, fail, not applicable, not evaluated, unknown, partial, stale, and exception.
  `not evaluated` is not a pass, and `not applicable` needs a reason.
- Re-resolves subject, policy/profile, evaluator, environment, and evidence revisions before
  disposition. **A changed binding publishes a stale result or re-runs; a pass never carries forward
  silently.**

### 4 · Hand off — assemble, do not decide

```
/xonovex-workflow:acceptance-validate <criteria-ref> --subject <deliverable-ref> --revision <rev>
```

**This is the line the lens exists to make visible.** `acceptance-validate` assembles fresh evidence
and explicitly cannot sign off, authorize Integration, or fabricate accountable authority. A clean
validation is not an approval — an accountable human records that separately with `acceptance-decide`,
and that half belongs to the [PM Quickstart](pm-quickstart.md). The split is
deliberate: an agent may assemble, only a human decides.

### Alongside: validate a plan

```
/xonovex-workflow:plan-validate <plan-ref> --revision <rev> [--publish-evidence]
```

Read-only. Checks every success criterion and Definition of Done item independently, reports stale or
missing evidence, and **never revises status or claims Acceptance**. Green tests alone do not satisfy it.

### Alongside: the PR track

```
/xonovex-workflow:pr-review-analyze [branch] --base <ref> [--out findings.json] [--since <file>]
/xonovex-workflow:pr-review-refine [findings-file] --walk [--final]
/xonovex-workflow:pr-review-post [branch] [--dry-run] [--yes]
/xonovex-workflow:pr-review-resolve [branch] [--reply]
```

- One shared findings set flows through the four, in-session by default — `--out` / `--findings`
  persists it only to cross sessions or hand-edit.
- `--since` tags each finding `recurring`, `new`, or `gone` against a prior set: the honest way to see
  whether a re-review actually moved.
- `pr-review-post` detects the host from the git remote and stops if the matching host skill
  (`github-guide` / `gitlab-guide`) is not installed.
- `pr-review-resolve` verifies each finding is **genuinely fixed, not just moved**, and checks required
  code-owner approvals — open threads or missing approvals mean the review is not done.
- Choose by what must survive: `pr-review-*` posts to a PR on a host; `review-run` publishes a Review
  result bound to an exact deliverable revision with reviewer origin and independence recorded.

## Independence

Independence is a recorded property of the reviewer, not a courtesy. The session that produced the
change is the worst available reviewer of it, because it defends rather than attacks — which is why
`--independent` exists, why `plan-critique` and every `*-critique` command run in a fresh context, and
why a self-review does not satisfy a profile that requires independence.

Where a human reviews, identity, role, qualification, independence, and scope are recorded as part of
the evidence. Where a bounded model or agent reviews, its inference stays non-authoritative and must
link the source evidence — recorded separately, never normalized into a stronger authority class.

## Evidence staleness

Every result here binds to all interpretation-relevant values at once: subject reference and revision
or digest; criteria, policy, and profile versions; evaluator, scanner, model, and ruleset versions;
environment identity and configuration; actor identity, role, and independence where required;
evaluation time, expiry, and limitations.

**A changed subject always invalidates revision-sensitive assurance.** Re-run it; do not re-label it.
Never re-label stale, skipped, neutral, cancelled, timed-out, missing, or partial evidence as passing.

The failure this prevents is the most ordinary one in review: a PASS collected against revision 3,
quoted at revision 7, on the grounds that nothing important changed in between. The command cannot
know that — and neither can you. `workflow-guide` owns the full freshness binding in
[assurance-contracts.md](../../../skill/skill-workflow/workflow-guide/references/assurance-contracts.md).

## Gates where a human is mandatory

- **`acceptance-decide`** — never yours to shortcut, and never satisfied by your evidence alone. You
  assemble; an authenticated accountable human signs. It rejects a script, model, agent, unqualified
  actor, self-reviewer where independence is required, or copied provider identity.
- **The independence requirement itself** — where a profile requires it, no autonomy level relaxes it,
  and a self-review does not satisfy it.
- **Accountable disposition** — a human or qualified assessor records identity, role, qualification,
  independence, and scope. Scanners, CI, and agents supply evidence, never the disposition.

## Executors and autonomy

Two models appear throughout the argument help. Both are owned elsewhere and deliberately not
restated here:

- **Executor classes** (`--executor`) —
  [execution.md](../../../skill/skill-agent-governance/agent-governance-guide/references/execution.md).
  Read the deterministic-authority rule closely: `model` or `agent` output is never authoritative where
  a deterministic check or external system of record establishes the fact. A model may triage and
  cluster your findings; the scanner and the suite own the facts.
- **Autonomy levels `A0`–`A3`** —
  [autonomy.md](../../../skill/skill-agent-governance/agent-governance-guide/references/autonomy.md).
  `A1` adds exactly one thing you care about — an independent critique of an exact revision that the
  author cannot suppress. `A2` makes sign-off asynchronous on exact revisions; asynchronous is not
  absent, and high-impact gates stay human at every level.

## What you can skip

- **`assessment-run`** — unless a pinned framework, policy, or control set genuinely applies.
- **The `review-run` / `qa-run` track** — on a solo PR the lighter `pr-review-*` pipeline is enough;
  reach for the formal pair when the evidence must outlive the PR or bind to a policy.
- **`inventory-generate`** — only for SBOM/BOM or agent-inventory needs.
- **Everything before `deliver-publish`** — the other lenses own it; you arrive when a revision exists.
- **`observe-run` / `incident-run` / `corrective-action-run`** — ops owns the operational tail.

## See also

- [README](../README.md) — the full command table
- [Developer Quickstart](developer-quickstart.md) — the sibling lens, and the jargon decoder
- [PM Quickstart](pm-quickstart.md) · [UX Quickstart](ux-quickstart.md) — the
  other lenses
- [Validation and traceability](validation-and-traceability.md) — the evidence chain end to end
- [`code-review-guide`](../../../skill/skill-code-review/code-review-guide/SKILL.md) and
  [`testing-guide`](../../../skill/skill-testing/testing-guide/SKILL.md) — the craft each command loads
