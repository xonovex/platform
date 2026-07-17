# UX Quickstart

A lens over the same actor-neutral commands — not a different set of them. No command knows what a
designer is: `experience-design-create` asks for a subject and publishes a proposed result; it never
asks who you are. A role is a way to _read_ the command set — which of the 59 you actually run, and
where your work binds. The [README](../README.md) documents all 59 commands and their contracts; the
[Developer Quickstart](developer-quickstart.md) is the sibling lens for the seat that writes the code,
and it owns the shared vocabulary (native references, `--revision`, `--profile`, `--method`). See the
[ux-workflow diagram](../../../diagram/diagram-agent-workflow/ux-workflow.png) for the same flow at
a glance.

`bdd-guide` makes the same move for discovery: the
[Three Amigos](../../../skill/skill-bdd/bdd-guide/references/discovery-three-amigos.md) are three
perspectives, not three people. Design is not a fourth amigo bolted on — it is a way of reading the
same conversation and the same commands.

## What this lens carries

Proposals, and the evidence that the people using the thing can actually use it. The whole
`experience-design-*` family is **optional by design** — it publishes a _proposed_ result and is not
mandatory for any profile. Accessibility is the part that is not optional when a profile says so, and
it binds at four separate points rather than being a stage you pass.

## The happy path

Four stages: **Frame → Design → Critique → Accept.**

### 1 · Frame — shared, not owned

```
/xonovex-workflow:discovery-run <problem>              # observations, assumptions, unknowns
/xonovex-workflow:research-run <question> --scope <boundary>
/xonovex-workflow:formulation-run <subject>            # only if the behavior is yours to state
```

- These are the same commands the [PO lens](pm-quickstart.md) opens with — deliberately.
  Discovery and Research belong to no role. **If someone already ran them, read the result by its
  native reference instead of re-running them**; that handle is how you pick up their work rather than
  repeat it.
- **User research is a selectable method of `research-run`, not a separate command.** `--scope` sets
  the inclusion, exclusion, and stop boundaries; evidence and provenance stay distinct from synthesis,
  and uncertainty is stated rather than smoothed away.
- `discovery-run` deliberately forces no story, no Gherkin, and no artifact shape.

### 2 · Design

```
/xonovex-workflow:experience-design-create <subject-or-ref> [--method <selection>]
```

- Takes a Discovery, Research, Formulation, or Decision reference as input, and publishes a proposed
  Experience Design result **without making it mandatory for every profile**. Do not run it for a copy
  change.
- Soft-selects the installed research, interaction, content, prototyping, and accessibility
  capabilities the profile selects. None is a hard dependency — and an explicitly selected capability
  that is not installed **fails visibly and names what is missing**, rather than being quietly
  substituted.

### 3 · Critique → Revise

```
/xonovex-workflow:experience-design-critique <design-ref> --revision <rev>    # FRESH session
/xonovex-workflow:experience-design-revise <design-ref> --revision <rev> --feedback <critique-ref>
```

- **`experience-design-critique` runs in a fresh, independent context** and publishes separate
  findings. It does not revise and does not accept — the session that made the design defends it.
- **`experience-design-revise` resolves every feedback item**, publishes a new revision, preserves
  supersession, and **never overwrites or silently accepts the subject**. Iteration does not cost you
  the history.
- The pair loops. Re-critique the new revision until the design holds.

### 4 · Accept

```
/xonovex-workflow:experience-design-accept <design-ref> --revision <rev> --authority-reference <actor>
```

- Records an accept, reject, or conditional decision **against one exact revision**. A model may
  prepare the brief but must not fabricate the actor, qualification, or decision.
- Identical in shape to `plan-accept` and `decision-accept` — learn one, read all three. The
  [PO lens](pm-quickstart.md) covers the authority half.
- **A design accept is not a release accept.** `acceptance-decide` downstream is a separate,
  separately-bound sign-off.

## Where accessibility binds

Accessibility is not a stage in this path. It binds at four points, and **none of them substitutes for
another**:

- **Design** — `experience-design-create` soft-selects an installed accessibility capability when the
  profile selects it. This produces requirements and intent, not evidence.
- **Test** — `qa-run <deliverable-ref> --revision <rev> --scope accessibility --environment <env-ref>`
  publishes test evidence in recorded environments. See the [QA Quickstart](qa-quickstart.md).
- **Assess** — `assessment-run <subject-ref> --revision <rev> --criteria <standard-ref>` grades an
  exact revision against a pinned standard, version, and level, reporting pass, fail, not applicable,
  and not tested with origins and limitations.
- **Observe** — `observe-run <subject-ref> --window <range>` publishes accessibility evidence from
  production. Telemetry can identify a possible barrier; it does not replace direct evaluation.

Authority binds separately from evidence, and this is where the lens earns its keep:

- **A scanner is authoritative only for its declared scope, version, subject, and environment.** A
  clean report is not proof that every applicable criterion passes, and `not-applicable` needs a
  criterion-specific reason and a reviewer — it is never a synonym for untested or inconvenient.
- **An exception preserves a known gap and a remediation duty. It never turns a failed criterion into
  a pass**, and never establishes conformance by itself.
- Where a profile makes an accessibility criterion mandatory, it needs a protected gate or an
  accountable human control **that the same actor doing the work cannot bypass**.
- Conformance applies to the declared subject, scope, version, level, and dependencies. A changed UI,
  content path, component, platform, or assistive-technology environment may invalidate the evidence —
  the same freshness rule the [QA lens](qa-quickstart.md) lives by.

`accessibility-guide` owns the model — pinning the standard, version, level, applicable criteria,
exact subject revision, environments, and journeys, and layering deterministic inspection, automated
checks, keyboard and zoom testing, representative assistive technology, and qualified human review.
Read it there:
[accessibility-guide](../../../skill/skill-accessibility/accessibility-guide/SKILL.md).

## Gates where a human is mandatory

- **`experience-design-accept`** — the authority action, bound to an exact revision, never fabricated
  from a model's brief.
- **`acceptance-decide`** — the accountable human sign-off downstream, recorded through the provider's
  authoritative identity mechanism.
- **Qualified human review inside accessibility evidence** — scanners and agents supply evidence but
  never accountable conformance by themselves.

## Executors and autonomy

Two models appear throughout the argument help. Both are owned elsewhere and deliberately not
restated here:

- **Executor classes** (`--executor`) —
  [execution.md](../../../skill/skill-agent-governance/agent-governance-guide/references/execution.md).
  `human` is selected because accountable judgment must own the outcome, never because the work is
  hard — which is exactly why a qualified reviewer, and not a better scanner, is what an accessibility
  conformance claim needs.
- **Autonomy levels `A0`–`A3`** —
  [autonomy.md](../../../skill/skill-agent-governance/agent-governance-guide/references/autonomy.md).
  Raising a level changes who is watching and when. It never converts a `human` executor into a
  `model` one, and high-impact gates stay human at every level.

## What you can skip

- **The whole `experience-design-*` family** — it is optional. Skip it for anything without a real
  interaction surface to design.
- **The critique/revise loop** — for a design accepted on first draft, at the cost of the adversarial
  reader.
- **`formulation-run`** — when the PO owns the behavior and examples.
- **`solution-design-*`** — architecture belongs to the developer lens.
- **Everything after accept** — the [Developer Quickstart](developer-quickstart.md) builds it; you
  return at `qa-run --scope accessibility` and `observe-run`.

## See also

- [README](../README.md) — the full command table
- [Developer Quickstart](developer-quickstart.md) — the sibling lens, and the jargon decoder
- [PM Quickstart](pm-quickstart.md) · [QA Quickstart](qa-quickstart.md) — the
  other lenses
- [`accessibility-guide`](../../../skill/skill-accessibility/accessibility-guide/SKILL.md) and
  [`plan-guide`](../../../skill/skill-plan/plan-guide/SKILL.md) — the behavior each command runs
