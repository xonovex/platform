# Sources

The operation model is repository-authored. External sources support provider
separation and Git workspace mechanics; they do not define a universal lifecycle or
provider-neutral identifier format.

## Repository-authored symmetric operation contract

- **Provenance:** Maintained as the public contract for eight sibling workflow
  operations and four orthogonal workspace utilities.
- **Last reviewed:** 2026-07-21
- **Used for:** `SKILL.md` and all `references/`
- **Aspects extracted:** Stage-neutral operation semantics, independent selection
  dimensions, explicit side-effect boundaries, and the rule that workspace lifecycle
  is never implicit in a core operation.

## Hexagonal architecture

- **URL:** https://alistair.cockburn.us/hexagonal-architecture
- **Last reviewed:** 2026-07-21
- **Used for:** `references/composition.md`, `references/provider-native-references.md`
- **Aspects extracted:** Separation of operation semantics from technology-specific
  provider adapters and effects.

## Git worktree documentation

- **URL:** https://git-scm.com/docs/git-worktree
- **Last reviewed:** 2026-07-21
- **Used for:** `references/workspace-create.md`, `references/workspace-merge.md`,
  `references/workspace-abandon.md`, `references/workspace-cleanup.md`
- **Aspects extracted:** Worktree creation, listing, locking, removal, and pruning
  mechanics. The workflow safety and confirmation rules are repository-authored.
