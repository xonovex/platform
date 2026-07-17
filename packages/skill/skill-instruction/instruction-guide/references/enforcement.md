# enforcement: Prose, Hook, Hook plus Model, or Agent

Deciding which surface carries a behavior. AGENTS.md, an Agent Skill, and a command are all prose — a model reads them. A hook is executed by the harness. That difference decides what a rule can promise.

## Advisory capability is not enforcement

Prose raises the odds. It guarantees nothing: the model may comply, may not, or may never load the file at all.

A hook differs in kind, not in degree. The harness runs it on a native event — whether or not a model read anything, and whether or not the model would have chosen to. Prose has no executor of its own; the model is the only thing that could act on the sentence.

So a behavior that must happen **every time, automatically** is a requirement prose cannot meet. Not because the wording is weak, and not because ALWAYS / MUST / NEVER in capitals would fix it. There is nothing to strengthen — the surface does not execute.

## Lowest surface that meets the requirement

**agent-governance-guide** owns the adoption ladder and the rule to choose the lowest-authority composition. That rule is least privilege and it holds. It is also only half a decision: the chosen rung must still **meet the requirement**. Reading down to the cheapest rung — rather than the cheapest rung that qualifies — is how "always do X" lands in AGENTS.md.

Knowledge-only is the correct default and covers most rules. It stops being correct the moment the requirement says _automatically_.

## Decision procedure

Ask in order. Stop at the first surface that meets the requirement.

1. **Must it happen every time, automatically — even in a session where the model never loads the instruction?**
   - No → prose. AGENTS.md, a skill, or a command. Done; do not escalate.
   - Yes → prose cannot meet it. Continue.
2. **Can a script decide it from the event payload alone — a path, a command string, an exit code, a diff?**
   - Yes → **hook**. The harness executes it on every covered event; no model is consulted.
   - No → continue.
3. **Does exactly one narrow judgement stand between the payload and the decision — one classification, one closed question over fixed input?**
   - Yes → **hook plus bounded model**. The hook is still the trigger and still runs every time; it delegates that one judgement to a model with a closed output schema, and validates the answer before consuming it.
   - No → continue.
4. **Is the work genuinely adaptive — branching, multi-step, the tool sequence decided as it goes?**
   - Yes → **agent**, launched explicitly and bounded.
   - No → the requirement is not understood well enough to build yet. Narrow it and start again.

Most rules stop at step 1. An agent is the answer to step 4 only — reaching for one earlier buys adaptivity that nothing asked for, at the cost of the guarantee steps 2 and 3 would have given.

## What each surface buys

| Surface                         | Executor class                             | Guarantee                                                                                      |
| ------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------------- |
| AGENTS.md, Agent Skill, command | none — a model reads text                  | Nothing executes; advisory capability only                                                     |
| Hook                            | `deterministic`                            | The harness runs the handler on every covered event; blocking only where the matrix says so    |
| Hook plus bounded model         | `deterministic` trigger, `model` judgement | The trigger is guaranteed; the judgement is validated inference, never authoritative by itself |
| Explicit launcher               | `agent`                                    | A bounded, attenuated attempt — not an outcome                                                 |

`human` and `external` are the other two executor classes. Neither is selected by this procedure: they answer who must own the outcome, not how automatic the behavior is. **agent-governance-guide** owns all five class definitions, the bounded-model and bounded-agent contracts, and the hook mechanics.

## The failure mode

Writing "always do X" in AGENTS.md and expecting enforcement.

The sentence is imperative, the file loads into every session, and the rule reads like a control — so the job feels done. It is not. Nothing runs. The failure is quiet: no audit shows the rule failing, because the rule never executed. It shows X simply not happening, some of the time, with no error and nothing to grep for.

Tells that a rule has outgrown prose:

- it says always, never, every, before every, after every;
- it exists because someone forgot X once, and the fix was to write it down harder;
- a violation should block, not be noticed later;
- you have already written it twice, in stronger words each time.

## Escalating a rule out of prose

Prose keeps a job once the hook exists. Split it:

- the hook enforces the mechanical part;
- AGENTS.md states the rule once and names the hook that enforces it, so a reader knows the guarantee is real and where it lives;
- the rationale — why the rule exists, what to do when it fires — stays prose. That is knowledge, and a hook cannot carry it.

Do not drop the rule when the hook lands, and do not restate the hook's matching logic in prose; the copy drifts and the prose loses.

## Naming a hook is not evidence it enforces

A hook is one enforcement point, not a boundary. Before AGENTS.md claims a rule is enforced, the hook behind it must be supported, blocking, and covering the operation set on the harness actually in use — and installing a skill or module is not evidence that anything executes or blocks. Matching native event names across harnesses do not imply matching guarantees. **agent-governance-guide** owns the capability matrix, the support states, and the external-enforcement path for operations a harness cannot cover.

A mandatory rule that names an advisory or unsupported hook as its only enforcement point is worse than the prose it replaced: it reads as settled.
