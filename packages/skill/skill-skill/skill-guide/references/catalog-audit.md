# catalog-audit: De-duplicate and Tier a Set of Skills

The repeatable method for auditing an existing catalog and refactoring it onto the composable split. Use when a set of skills has grown overlapping concepts, missing cross-links, or language/API coupling — not when authoring one skill (see [composability.md](composability.md) for the target shape).

## Audit method

1. **Inventory** each skill: its `description`, its `references/` filenames, its exact manifest dependencies, and its composition-catalog entry. The filenames are the concept map.
2. **Classify** the guide with one lifecycle and one primary functional role; use [composition-metadata.md](composition-metadata.md), and record a split candidate instead of inventing `mixed`.
3. **Find shared concepts** — any concept appearing in 2+ skills (same topic, often the same reference filename).
4. **Name the rightful owner** for each shared concept using the tiered model (general → language → framework). The most general skill that the concept naturally belongs to is the owner.
5. **Resolve** each overlap with one move:
   - **Extract** a cross-cutting concept into a general skill, then link from each consumer.
   - **Cross-link** — delete the duplicate, keep the owner, reference it by name.
   - **Generalize** — rewrite a language-coupled reference as language-neutral in a general skill.
   - **Split** a bundle whose concerns belong to different tiers.
   - **Merge** two skills that are really one concept under different names.
6. **Declare composition conservatively** — add a semantic provision only when its contract is stable enough for another skill to depend on; add an invariant semantic requirement only when it applies on every activation.
7. **Verify**: no concept owned twice; each guide appears exactly once in the catalog; required semantic selection is deterministic and acyclic; every `SKILL.md` → `references/` link resolves; no author/company/talk names leaked into prose (provenance lives in `SOURCES.md`).

## Heuristics

- **Duplicated reference filenames across two skills = a tiered-split smell.** Two skills shipping `error-handling.md` (or `validation.md`, `middleware.md`…) almost always means one should own it and the other should cross-link.
- A concept that recurs across **3+ unrelated languages/frameworks** is a general-skill candidate — extract it once.
- Sibling skills with overlapping trigger words are a routing gap — tighten each description's triggers so the distinguishing keywords route to the right owner.

## Per-overlap execution

1. Pick the **owner** skill for the concept.
2. **Keep** the canonical reference in the owner; **delete** the duplicate from the other skill.
3. In the de-duplicated skill, **replace** the removed bullet/reference with a by-name cross-reference in the `SKILL.md` bullet and any `references/` "Related" lines; tighten the description's trigger words so the concept routes to the owner.
4. **Fix links** — repoint anything that pointed at the removed file; confirm no dangling links.
5. **Move provenance** to the owner's `SOURCES.md`; re-scan prose for leaked attributions.
6. **Update composition metadata** — retain one classification for each resulting guide, move or retire provisions with their owner, and re-resolve every requirement that used the old provision.
7. **Validate** — JSON valid, links resolve, formatter clean; spot-check trigger evals (positives for the skill, a negative pointing at the now-cross-linked sibling).
8. Do **one cluster per commit** so each refactor is reviewable and revertible.

## Not every overlap is duplication

Parallel operations that each act on a different domain are **correct**, not duplication to collapse. A `merge` / `create` / `simplify` trio existing separately for instructions, prompts, and skills is one concept _per domain_ — each owns its own subject. Only collapse overlaps where the **same** concept has two owners; leave parallel-by-domain structure intact.
