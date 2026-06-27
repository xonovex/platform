# ddd: Strategic Design — Contexts, Language, Maps

Strategic design decides _where_ the boundaries in a domain fall and _what words mean inside each one_ before any class is written. It is the part of Domain-Driven Design that scales: it keeps a large system from collapsing into one tangled model where every word means several things at once.

## Contents

- Ubiquitous language
- Bounded context
- Polysemy and why one model does not unify
- Context map and context-mapping patterns
- Anti-corruption layer
- Subdomains: core, supporting, generic
- Bounded context vs subdomain vs service

## Ubiquitous language

A ubiquitous language is a common, rigorous language built up jointly by developers and domain experts and grounded in the domain model that the software uses. It is rigorous because software does not cope well with ambiguity: a term that means two things in conversation must mean exactly one thing in the model and the code.

Rationale: when the words on the whiteboard, in the conversation, and in the source diverge, every handoff pays a translation tax and bugs hide in the gap. Binding one set of agreed terms across speech, diagrams, and code removes that gap.

How to: agree the term with the domain expert, name the model element after it, and name the code element after the model element — so `place(order)` in code reads the same as "place an order" at the business's desk. Use the language pervasively until it flows; when the model changes, change the words everywhere, including the code.

It only counts if the terms reach the code. A wiki glossary that nobody compiles against, or a "naming convention" enforced only in review, is not a ubiquitous language — it is documentation that will drift.

BAD: the spec says "shopper", the UML says "Account", the database column is `usr_party`, and the class is `CustomerEntity`. Four words for one concept; every reader re-translates.

GOOD: the term is "Customer" in the conversation, the model, and the code (`Customer`, `customer.placeOrder(...)`). One word, no translation.

## Bounded context

A bounded context is the explicit boundary within which a particular domain model — and its ubiquitous language — applies and stays internally consistent. One ubiquitous language per bounded context: the model _is_ the language.

Rationale: total unification of one model across a large system is not cost-effective. Past a certain size the cost of keeping every team agreeing on one meaning for every word exceeds the benefit, so you deliberately split the system into multiple contexts, each with its own consistent model.

How to: draw the boundary where the language stops being consistent — where the same word starts meaning something subtly different, or where a different team owns the meaning. Inside the boundary, one model; at the boundary, an explicit translation to the next context.

## Polysemy and why one model does not unify

The same word is polysemic across contexts. "Customer", "Product" — or "meter" across the departments of a utility — means subtly different things depending on who is speaking. Such differences can be smoothed over in conversation but not inside a computer, so each context models the word its own way and you map between them.

This is why forcing one canonical "Customer" across the whole system is a mistake, not DRY hygiene: a shared "Customer" type that Ordering, Shipping, and Billing all import couples those contexts together, and every field one context needs becomes a field the others must carry and keep consistent. The bounded context exists precisely to remove that coupling.

BAD: a single shared `Customer` class with `cart`, `wishlist`, `deliveryAddress`, `taxId`, `creditTerms`, and `paymentMethod` — every context drags fields it does not use and breaks when another context changes one.

GOOD: Ordering's `Customer` (cart, wishlist, order history), Shipping's `Recipient` (name, address, phone), and Billing's `AccountHolder` (billing address, tax id, credit terms) are three models of one real-world person, each minimal for its context, translated at the boundary.

## Context map and context-mapping patterns

A context map is the picture of how the bounded contexts in a system relate, and the relationship patterns name the kinds of edge between them:

- **Anti-corruption layer** — an isolating translation layer (detailed below).
- **Conformist** — the downstream context adopts the upstream model wholesale with no translation, accepting its terms.
- **Shared kernel** — two contexts share a small, explicitly agreed subset of the model, changed only by joint agreement.
- **Customer–supplier** — an upstream context delivers what a downstream context needs, with the downstream's needs feeding the upstream's plan.
- **Open-host service** — an upstream context publishes a defined protocol for any number of downstream consumers.
- **Published language** — a well-documented shared interchange format that contexts translate into and out of.

How to: draw the map first to see the integration surface, then choose the pattern per edge — an anti-corruption layer where you must protect your model from a messy or external upstream, a shared kernel only where two teams genuinely co-own a concept and will coordinate every change.

## Anti-corruption layer

An anti-corruption layer (ACL) is an isolating layer that provides your system with the functionality of an upstream or external system expressed in terms of _your own_ domain model. It talks to the other system through that system's existing interface and translates in both directions, so foreign concepts never leak into your model.

Rationale: when you call another system's model directly, its vocabulary and quirks seep into yours; over time your model corrodes into a mirror of theirs. The ACL is the firewall that keeps your ubiquitous language intact.

How to: define the translation in terms of your concepts, call the foreign interface inside the layer, and map the foreign response onto your own types on the way back. It is bidirectional translation, not a forwarding shim.

BAD: an `UpstreamApiClient` that returns the vendor's raw `PartyDTO` straight into your domain logic — now `PartyDTO`'s fields and meanings are part of your model.

GOOD: a `ShippingRecipientTranslator` that calls the upstream address service and returns your own `Recipient`, never exposing the vendor's `PartyDTO` past the layer.

A plain upstream-API client is not an ACL: forwarding calls without translating concepts both directions leaves your model exposed.

## Subdomains: core, supporting, generic

A subdomain is a partition of the _problem space_ — an area of the business — and they come in three kinds:

- **Core domain** — the part that gives the business its competitive edge; invest your best modelling effort here.
- **Supporting subdomain** — necessary to the business but not a differentiator; model it adequately.
- **Generic subdomain** — a solved problem common to many businesses (e.g. notifications, authentication); buy or adopt rather than build.

How to: distil the core so you know where the deep modelling pays off, and resist lavishing core-domain effort on generic subdomains.

## Bounded context vs subdomain vs service

These three are different lenses and conflating them is a common error:

- A **subdomain** is a problem-space area of the business.
- A **bounded context** is a solution-space boundary of one consistent model and language.
- A **service** (deployment unit) is an operational packaging boundary.

They frequently align one-to-one — a core subdomain modelled in one bounded context deployed as one service — but they are not the same, and a bounded context applies just as well inside a single monolith as across services.

## Cross-references

- Code-level coupling and cohesion vocabulary (connascence, coupling ladders) is owned by **connascence-guide**; this file uses "coupling" only as the business reason to draw a boundary, not as a metric.
- General class/inheritance mechanics are owned by **oop-guide**.
- An in-memory object/property database with cross-references, change notification, undo, and serialization is owned by **data-model-guide** — distinct from a repository over aggregate roots.
- **orthogonal-pattern-guide** borrows "bounded context" and "ubiquitous language" as a lens for where a code seam falls; the definitions live here.
- The tactical blocks modelled _inside_ a context — entity, value object, aggregate, repository, domain event — are in [tactical-building-blocks.md](tactical-building-blocks.md).
- Eliciting behaviour inside a context as agreed examples is owned by **bdd-guide**; the story container by **user-stories-guide**; FDD's overall-model step defers domain modelling here (**fdd-guide**).
