---
name: connascence-guide
description: 'Use when grading or loosening the coupling between two pieces of code, or judging how cohesive a module is. Triggers on connascence, coupling vs cohesion, "too tightly coupled", decouple/loosen, the coupling ladder (content/common/control/stamp/data), the cohesion ladder, Law of Demeter / train wrecks like a.b().c(), positional vs named parameters, or reviewing a module boundary for coupling strength — even when the user only says ''tightly coupled''. Skip where a module/axis boundary should go (see orthogonal-pattern-guide), isolating a domain behind ports-and-adapters (see hexagonal-pattern-guide), and plugin registries (see microkernel-pattern-guide).'
---

# Connascence: a Precise Coupling Vocabulary

Grade a seam by naming its worst coupling and its strongest connascence, then weaken or localize it — "tight/loose" is unactionable.

## Essentials

- **Name the connascence, don't say "tight"** - classify the form and rate strength/locality/degree, see [references/connascence.md](references/connascence.md)
- **Two moves: weaken and localize** - rule of degree (strong→weak forms) and rule of locality (distance→weakness), see [references/connascence.md](references/connascence.md)
- **Push every seam down the coupling ladder** - content→common→control→stamp→data, worst to best, see [references/coupling-ladder.md](references/coupling-ladder.md)
- **Aim for functional cohesion** - one job per module; temporal cohesion is the package-by-layer trap, see [references/cohesion-ladder.md](references/cohesion-ladder.md)
- **Don't reach through strangers** - Law of Demeter as the call-site detector for content coupling, see [references/law-of-demeter.md](references/law-of-demeter.md)

## Gotchas

- "Tightly coupled" is not actionable — name the connascence form and the coupling rung, then you know the fix.
- Strong connascence is fine _inside_ one module; it's only a problem _across_ a boundary (rule of locality).
- Control coupling hides as configuration — a behavior flag passed into a shared function recouples its callers; split into two functions.
- Law of Demeter is a heuristic, not a law — fluent builders and pure data pipelines are deliberate exceptions; it targets reaching through a stranger to depend on or mutate its internals.
- Coupling and cohesion trade off — splitting a module to "loosen coupling" can shred its cohesion; grade both before cutting.

## Example

```
SEAM   a billing function reaches into an Order's private `total` field
GRADE  content coupling (reaches internals) + connascence of meaning across a boundary
FIX    Order exposes a neutral `amountDue()` value (rule of degree → data coupling);
       billing reads the value — agreement is now named, local, and weak
```

## Progressive Disclosure

- Read [references/connascence.md](references/connascence.md) - Load when classifying a coupling (static/dynamic forms, strength/locality/degree, the two rules)
- Read [references/coupling-ladder.md](references/coupling-ladder.md) - Load when grading or improving a seam by the content→data ladder
- Read [references/cohesion-ladder.md](references/cohesion-ladder.md) - Load when judging whether a module's parts belong together, or spotting temporal cohesion
- Read [references/law-of-demeter.md](references/law-of-demeter.md) - Load when a call site reaches through returned objects (train wrecks)
