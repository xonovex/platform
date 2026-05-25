# composability: Design Skills to Coexist

The depth behind the **Composable split** Core Principle. A skill is one capability among many that load together; it must work alongside others, not assume it is the only capability available.

## One concept, one owner

- Every concept has exactly **one owner skill**. Never duplicate it across skills.
- When another skill needs the concept, **cross-reference the owner by name** ("see zod-guide") — not by copying its content, and not by file link (cross-package file links don't resolve at load time).
- A skill may keep a short concept-specific note that points at the owner for the "why" — never a copy of the owner's depth.

## Tiered model: general → language → framework

Concepts stack in three tiers, and dependencies point **upward only**:

- **General** — a paradigm or principle that is not tied to any language or API (functional style, object/data model, memory ownership, cache layout, concurrency).
- **Language** — idioms and tooling for one language.
- **Framework / opinionated** — a library, runtime, or house style layered on a language.

Specific tiers link **up** to the general tier for the rationale; the general tier **never depends** on a specific one. Reference model:

- `general-fp-guide` / `general-oop-guide` → `typescript-guide` / `python-guide` / `lua-guide` / `c99-guide` → `hono-guide` / `react-guide` / `c99-opinionated-guide`.

## Generalize-or-link decision

For any concept appearing in a language/framework skill, ask: **is this concept inherently tied to this language or API?**

- **No** → it belongs in a general skill that the specific skill links to (e.g. immutability → `general-fp-guide`; cache layout → `data-oriented-design-guide`). Generalize it once; link from each consumer.
- **Yes** → keep it local to the owning skill (e.g. `LuaMultiReturn` → `typescript-to-lua-guide`; SoA `_simd` suffixes → `c99-opinionated-guide`).

## Rules

- Prefer many small single-concern skills over one large bundle — they mix and match per task.
- The general tier must stay self-contained: it explains the principle without referencing any language or framework skill.
- Cross-reference by **skill name**, bolded in prose (`**hono-guide**`) — readers route by name, and names survive repackaging that file paths do not.
- A "Skip" clause in the description routes a concept to its owner ("Skip framework work — use hono-guide"); the body cross-links the same owner for depth.
