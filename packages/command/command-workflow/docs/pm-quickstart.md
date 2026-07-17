# Product Owner Quickstart

A lens over the same actor-neutral commands — not a different set of them. No command knows what a
product owner is: `plan-accept` asks for authority over an exact revision, never for a job title. A
role is a way to _read_ the command set — which of the 59 you actually run, and which gates you
personally answer for. The [README](../README.md) documents all 59 commands and their contracts; the
[Developer Quickstart](developer-quickstart.md) is the sibling lens for the seat that writes the code,
and it owns the shared vocabulary (native references, `--revision`, `--profile`, `--method`).

`bdd-guide` makes the same move for discovery: the
[Three Amigos](../../../skill/skill-bdd/bdd-guide/references/discovery-three-amigos.md) are three
perspectives, not three people. Perspectives are what matter; who holds them is a staffing question
the machinery never asks.

## What this lens carries

The developer lens writes code. The QA lens assembles evidence. This lens carries **authority** — the
points where the workflow stops and waits for an accountable human, and where no model or agent may
answer in your place.

## The happy path

Four stages: **Frame → Decide → Authorize → Sign off.** None of them writes code.

### 1 · Frame

```
/xonovex-workflow:discovery-run <problem>       # observations, assumptions, unknowns
/xonovex-workflow:research-run <question>       # cited evidence, provenance, uncertainty
/xonovex-workflow:formulation-run <subject>     # candidate behavior, examples, constraints
```

- **`discovery-run` does not force stories**, and `formulation-run` defaults to `neutral`. Select
  `--method user-stories` or `--method bdd` only when that is genuinely your team's method; an
  explicitly selected method that is not installed fails visibly rather than substituting another.
- **There is no `story-refine`** — it was removed in 6.0.0. `discovery-run` then `formulation-run`
  replace it, with the method selected rather than assumed. See [MIGRATION.md](../MIGRATION.md).
- `research-run` keeps evidence and provenance distinct from synthesis and exposes uncertainty
  explicitly. Read the uncertainty — pricing it is the part of the job that is yours.

### 2 · Decide — one fork at a time

```
/xonovex-workflow:decision-create <question> --authority-reference <actor>
/xonovex-workflow:decision-critique <decision-ref> --revision <rev>    # FRESH session
/xonovex-workflow:decision-revise <decision-ref> --revision <rev> --feedback <critique-ref>
/xonovex-workflow:decision-accept <decision-ref> --revision <rev> --authority-reference <actor>
```

Key rules:

- **`decision-create` keeps evidence, recommendation, and authority separate** — deliberately. A model
  or agent may synthesize the options and recommend one; only the required human or qualified actor
  supplies the authority. A recommendation is not a decision.
- **`decision-critique` runs in a fresh, independent session**, and publishes findings without making
  or revising the decision. Same reason as `plan-critique`: the session that wrote the brief defends it.
- **`decision-revise` reopens authority when the material meaning changed**, and preserves prior
  authority and evidence rather than rewriting them.
- Reach for this family only on a genuine fork — a choice you would want recorded and would regret
  re-litigating.

### 3 · Authorize — the plan gate

```
/xonovex-workflow:plan-accept <plan-ref> --revision <rev>                  # sets status: approved
/xonovex-workflow:plan-reject <plan-ref> --revision <rev> --reason <text>
```

- **`plan-accept` is required before `plan-subplans-create`**, which needs `status: approved`. It is
  the gate between "someone wrote a plan" and "the team builds it".
- **Approval binds to one exact revision and never carries forward.** Every `plan-revise` resets
  status to pending-approval, so re-run `plan-accept` against the new revision. If you approved
  revision 3 and revision 4 exists, you have approved nothing.
- **`plan-reject` requires `--reason` and preserves history.** Rejecting is not deleting — the
  rejected revision and your rationale stay readable.
- A model may summarize the plan for you. It cannot fabricate the approval.

### 4 · Sign off — acceptance

```
/xonovex-workflow:acceptance-validate <criteria-ref> --subject <deliverable-ref> --revision <rev>
/xonovex-workflow:acceptance-decide <deliverable-ref> --revision <rev> --target <target> --evidence <ref>
```

**The split is the point.** `acceptance-validate` assembles evidence and explicitly cannot sign off;
`acceptance-decide` is where you, identified through the provider's own authoritative mechanism,
record the decision. An agent may assemble; only a human decides. The evidence half belongs to the
[QA Quickstart](qa-quickstart.md).

- `acceptance-validate` reports each criterion as satisfied, failed, partial, stale, not applicable,
  unknown, or not evaluated, and **never coerces missing, skipped, cancelled, or timed-out evidence to
  passing**. Gaps arrive as gaps. Bounded model or agent help is marked advisory.
- `acceptance-decide` records `accept`, `accept-with-conditions`, or `reject` bound to the exact
  subject, target, evidence, policy version, and expiry. Conditions are first-class — you never have
  to choose between blocking and waving it through.
- **Acceptance grants no target mutation by itself.** `integration-validate` and `integration-run`
  revalidate every binding at an external enforcement point. Your sign-off authorizes; it does not ship.
- Changed bindings or a passed expiry invalidate the decision for future use.

## Gates where a human is mandatory

Four commands wait for an accountable actor, and nothing advances past them on a model's say-so:

- **`decision-accept`** — supplies authority for a fork. A model may prepare the brief; it cannot
  supply the authority action.
- **`plan-accept` / `plan-reject`** — approves or rejects one exact Planning revision.
- **`experience-design-accept`** — accepts an exact design revision where that authority is yours; see
  the [UX Quickstart](ux-quickstart.md).
- **`acceptance-decide`** — records accountable sign-off, and rejects a script, model, agent,
  unqualified actor, self-reviewer where independence is required, or copied provider identity.

`--authority-reference` is a claim to be verified, not a credential: actor, qualification, scope, and
freshness are checked before the decision binds to the revision.

## Executors and autonomy

Two models appear throughout the argument help. Both are owned elsewhere and deliberately not
restated here:

- **Executor classes** (`--executor`) —
  [execution.md](../../../skill/skill-agent-governance/agent-governance-guide/references/execution.md).
  The class that matters to this lens is `human`: it is selected because accountable judgment must own
  the outcome, never because the work is hard. An `--executor` value is a requested ceiling that may be
  rejected, never an escalation.
- **Autonomy levels `A0`–`A3`** —
  [autonomy.md](../../../skill/skill-agent-governance/agent-governance-guide/references/autonomy.md).
  Raising a level changes who is watching and when. It never converts a `human` executor into a `model`
  one, and high-impact gates stay human at every level. Yours are high-impact: a wrong approval is not
  both detectable without a human and reversible without an external system of record.

At the unattended levels a run that needs you raises a bounded escalation with a declared window and
safe default. **Silence is never approval** — an unanswered escalation falls back to pause or abandon
when the window expires, and a timeout never advances a gate.

## What you can skip

- **`discovery-run` / `research-run`** — for a well-understood ask you can state in a sentence.
- **The `decision-*` family** — for a reversible choice nobody would want recorded.
- **The critique/revise loops** — for a brief you accept on first draft, at the cost of the
  adversarial reader.
- **Everything in Build and Ship** — the [Developer Quickstart](developer-quickstart.md) owns it; you
  meet it again at `plan-accept` and `acceptance-decide`.
- **`experience-design-*` and `solution-design-*`** — the UX and developer lenses own them.
- **The governed tail** (`integration-*`, `release-run`, `observe-run`) — ops owns it.

## See also

- [README](../README.md) — the full command table
- [Developer Quickstart](developer-quickstart.md) — the sibling lens, and the jargon decoder
- [QA Quickstart](qa-quickstart.md) · [UX Quickstart](ux-quickstart.md) — the other lenses
- [`plan-guide`](../../../skill/skill-plan/plan-guide/SKILL.md) and
  [`workflow-guide`](../../../skill/skill-workflow/workflow-guide/SKILL.md) — the behavior each
  command runs
