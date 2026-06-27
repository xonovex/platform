# decompose: Split a Multi-Concern Skill into Composable Skills

Break one oversized or multi-concern skill into several single-owner skills, each owning one concept and cross-referencing the others by name. The inverse of `merge`, and the single-skill version of the split move in [catalog-audit.md](catalog-audit.md). The target shape is [composability.md](composability.md): one concept, one owner, dependencies pointing upward through the general → language → framework tiers.

## When to decompose

- A `SKILL.md` mixes concerns that route differently — its description needs "and" to cover unrelated triggers.
- The body exceeds the budget (~500 lines / 5k tokens) because it carries several topics, not one deep one.
- The references split cleanly into groups that share no concept — each group is a latent skill.
- A general principle is trapped inside a language/framework skill and other skills would reuse it.

Decompose **splits**; it does not condense — reach for `simplify` when one skill is merely verbose, `merge` when two skills are really one.

## Core Workflow

1. **Map the concerns** — list the references and group them by concept; each cohesive group is a candidate skill. The description's distinct trigger clusters mark the seams.
2. **Assign tiers and owners** — for each group name the tier (general / language / framework) and the owning skill; a cross-cutting principle becomes its own general skill the others link up to.
3. **Choose the split** — one skill per concept, never a concept owned twice. Keep parallel-by-domain operations together (a create / merge / simplify trio for one domain is one skill, not three).
4. **Carve out each new skill** — move its references verbatim into a new skill directory; write its `SKILL.md` (name = dir, routing-first description, essentials, progressive-disclosure links) per the skill structure and [composability.md](composability.md).
5. **Cross-link, don't copy** — where a new skill needs another's concept, reference the owner by name (upward only); never duplicate content or use cross-package file links.
6. **Retire or slim the original** — if fully split, remove it and repoint every referrer; if a core concept remains, keep that core and replace the moved sections with by-name cross-references.
7. **Register** — add each new skill to the catalog / registry and packaging; tighten descriptions so triggers route to the new owners.
8. **Verify** — every `SKILL.md` → `references/` link resolves, no concept is owned twice, no referrer dangles, and trigger evals route to the right owner.

## Heuristics

- The description is the seam finder: if it needs "and" to join unrelated triggers, those are two skills.
- A reference group sharing no concept with the rest is already a separate skill — lift it whole.
- A principle reused across 3+ unrelated languages/frameworks belongs in a new general skill, linked from each consumer.
- Split along tier boundaries — a general principle, a language idiom, and framework specifics rarely belong in one skill.

## Pitfalls

- Splitting a cohesive single concept into fragments raises routing ambiguity and load overhead — decompose by concern, not by size alone.
- Leaving the original pointing at moved references dangles the links — repoint or retire every referrer in the same change.
- Duplicating a shared concept into each new skill recreates the problem decompose solves — one owner, cross-link the rest.
- Collapsing parallel-by-domain operations while splitting destroys real structure — those are separate-by-domain, not duplication.
- A new skill is not auto-discovered — register it (catalog, packaging, lockfile) or it strands.
