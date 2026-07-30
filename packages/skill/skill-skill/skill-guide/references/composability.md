# composability: Design Skills to Coexist

The depth behind the **Composable split** Core Principle. A skill is one capability among many that load together; it must work alongside others, not assume it is the only capability available.

## One concept, one owner

- Every concept has exactly **one owner skill**. Never duplicate it across skills.
- When another skill needs the concept, **cross-reference the owner by name** ("see zod-guide"), not by copying its content, and not by file link (cross-package file links don't resolve at load time).
- A skill may keep a short concept-specific note that points at the owner for the "why", never a copy of the owner's depth.

## Tiered model: general → language → framework

Concepts stack in three tiers, and dependencies point **upward only**:

- **General**: a paradigm or principle that is not tied to any language or API (functional style, object/data model, memory ownership, cache layout, concurrency).
- **Language**: idioms and tooling for one language.
- **Framework / opinionated**: a library, runtime, or house style layered on a language.

Specific tiers link **up** to the general tier for the rationale; the general tier **never depends** on a specific one. Reference model:

- `fp-guide` / `oop-guide` → `typescript-guide` / `python-guide` / `lua-guide` / `c99-guide` → `hono-guide` / `react-guide` / `c99-opinionated-guide`.

## Relating to another skill

A skill or command relates to another capability in three distinct ways. Never copy the other skill's content or use a cross-package file link:

- **Advisory handoff**: name the concept owner (`owned by **testing-guide**`, `see **bdd-guide**`) so the agent can continue there when it is available. The current skill remains useful without it, so this does not create an installation dependency and may point laterally.
- **Soft selection**: state the needed capability in the operation request and choose from installed skill names and routing descriptions. Prefer an explicitly named fitting skill; otherwise choose the narrowest unambiguous description match. Report missing preferred support and block on missing required support.
- **Hard dependency**: name the exact skill because the current skill's core workflow cannot complete without it, and add the bare plugin name to `dependencies` in every manifest supported by that package. Skill plugins pair Claude and Codex manifests; a Claude-only command plugin declares the edge only in its Claude manifest. Installation still does not put the skill in context, so load it explicitly when needed.

Hard dependencies point **upward only**: a specific skill may require a general one; the general tier never requires a specific one, and the manifest graph has no cycles. Advisory handoffs may be reciprocal because they do not promise installation or loading. Every named handoff and declared dependency must resolve to an existing registered skill. The composition validator enforces matching package-derived manifest names, matching dependency lists, named hard-dependency handoffs, no missing targets, dependency-first order, and no cycles. Whether a workflow truly cannot complete without a dependency and whether an edge points upward remain author-review decisions.

Soft selection is deliberately descriptive rather than another dependency graph. Do
not infer suitability from package order, reference filenames, or an "opinionated"
label. When two descriptions fit equally well, ask or report the ambiguity instead of
silently selecting one.

An opinionated or layered guide is not a separate overlay mechanism. It is an ordinary
specific skill whose routing description names the narrower context and whose manifest
hard-depends on its general foundation only when it cannot stand alone. Precedence
comes from the more specific applicable guidance, not separate overlay metadata.

## Generalize-or-link decision

For any concept appearing in a language/framework skill, ask: **is this concept inherently tied to this language or API?**

- **No** → it belongs in a general skill that the specific skill links to (e.g. immutability → `fp-guide`; cache layout → `data-oriented-design-guide`). Generalize it once; link from each consumer.
- **Yes** → keep it local to the owning skill (e.g. `LuaMultiReturn` → `typescript-to-lua-guide`; SoA `_simd` suffixes → `c99-opinionated-guide`).

## Rules

- Prefer many small single-concern skills over one large bundle: they mix and match per task.
- The general tier must stay self-contained: it explains the principle without referencing any language or framework skill.
- Mark advisory ownership handoffs by **skill name**, bolded in prose (`owned by **hono-guide**`): readers route by name, and names survive repackaging that file paths do not.
- A description routes by its own trigger words, not by naming other skills; keep cross-references in the body, which links the owning skill by name for depth.
- Do not encode optional runtime choices as manifest dependencies. A hard dependency installs one exact required implementation; a soft selection chooses among what is already installed.
