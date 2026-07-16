# ddd: Strategic Design — Contexts, Language, Maps

Strategic design decides _where_ a domain's boundaries fall and _what words mean inside each one_ before any class is written.

## Ubiquitous language

One rigorous language per context, built jointly with domain experts: a term means exactly one thing in the model AND the code (`place(order)` reads as "place an order"). It only counts when the terms reach the code — a wiki glossary or a review-only convention is not a ubiquitous language, it is documentation that drifts. When the model changes, change the words everywhere, including the code.

## Bounded context

The explicit boundary within which one domain model and its ubiquitous language stay internally consistent — the model _is_ the language. Draw the boundary where the language stops being consistent (a word starts meaning something subtly different, or a different team owns the meaning). Total unification across a large system is not cost-effective; split into multiple contexts, each with its own model, and translate explicitly at each boundary.

## Polysemy: one model does not unify

The same word ("Customer") is polysemic across contexts. Each context keeps its own minimal model, translated at the boundary — not coordinated through one shared type.

- GOOD: Ordering's `Customer`, Shipping's `Recipient` (name, address, phone), Billing's `AccountHolder` (billing address, tax id, credit terms) — three minimal models of one person, translated at the boundary.

## Context map and context-mapping patterns

A context map pictures how the contexts relate; choose a pattern per edge:

- **Anti-corruption layer** — an isolating translation layer (below).
- **Conformist** — downstream adopts the upstream model wholesale, no translation.
- **Shared kernel** — two contexts share a small explicitly-agreed subset, changed only by joint agreement.
- **Customer–supplier** — upstream delivers what downstream needs, downstream's needs feed upstream's plan.
- **Open-host service** — upstream publishes a defined protocol for any number of consumers.
- **Published language** — a documented shared interchange format contexts translate in and out of.

## Anti-corruption layer

An ACL provides an upstream/external system's functionality expressed in _your own_ domain model. It talks through the other system's interface and translates in **both** directions, so foreign concepts never leak in. A plain upstream-API client that forwards calls without bidirectional translation is not an ACL.

- BAD: `UpstreamApiClient` returning the vendor's raw `PartyDTO` into your domain logic.
- GOOD: `ShippingRecipientTranslator` calls the address service and returns your own `Recipient`, never exposing `PartyDTO` past the layer.

## Subdomains: core, supporting, generic

A subdomain partitions the _problem space_:

- **Core domain** — the competitive edge; invest your best modelling here.
- **Supporting subdomain** — necessary but not a differentiator; model adequately.
- **Generic subdomain** — a solved problem (notifications, auth); buy or adopt, don't build.

## Bounded context vs subdomain vs service

Three different lenses; conflating them is a common error:

- **Subdomain** — a problem-space area of the business.
- **Bounded context** — a solution-space boundary of one consistent model and language.
- **Service** — an operational deployment/packaging unit.

They often align one-to-one but are not the same; a bounded context applies inside a single monolith just as well as across services.

## Cross-references

- Coupling/cohesion metrics (connascence) — **connascence-guide**; here "coupling" is only the business reason to draw a boundary.
- Class/inheritance mechanics — **oop-guide**.
- In-memory object/property database (cross-references, change notification, undo, serialization) — **data-model-guide**; distinct from a repository over aggregate roots.
- **orthogonal-pattern-guide** borrows "bounded context"/"ubiquitous language" as a seam lens; definitions live here.
- Tactical blocks modelled _inside_ a context are in [tactical-building-blocks.md](tactical-building-blocks.md).
- Behaviour as agreed examples — **bdd-guide**; story container — **user-stories-guide**; FDD's overall-model step defers here — **fdd-guide**.
