# ddd: Tactical Building Blocks

Tactical design fills in a single bounded context with a small, named set of modelling blocks. Each block is a tool for keeping the domain's invariants honest while staying ignorant of infrastructure. Strategic design (in [strategic-design.md](strategic-design.md)) draws the boundaries; these are what you build inside one.

## Contents

- Entity vs value object
- Aggregate and aggregate root
- The four aggregate-design rules of thumb
- Domain event
- Repository
- Domain service vs application service
- The anemic-domain-model anti-pattern

## Entity vs value object

An **entity** has an identity that persists over time and across changes to its attributes: an Order is the same Order tomorrow even after its lines change. Equality is by identity, not by field values.

A **value object** is defined entirely by its attributes, has no conceptual identity, and is immutable: a Money of 12.50 EUR is interchangeable with any other 12.50 EUR; an Address is just its fields. Equality is by value; to "change" one you create a new one.

How to: ask "do I care _which_ one this is, or only _what_ it is?" If you track it through a lifecycle, it is an entity; if two instances with the same attributes are indistinguishable and substitutable, it is a value object. Default to value objects — they are immutable, freely shareable, and cheap to reason about.

BAD: modelling money as a raw `number`, or an address as four loose strings threaded through every method — a domain concept carried as primitives, with no place for its rules.

GOOD: `Money { amount, currency }` and `Address { ... }` as immutable value objects with their own validation, and `Order` as an entity with a stable id.

## Aggregate and aggregate root

An **aggregate** is a cluster of domain objects treated as a single unit for data changes. One entity is designated the **aggregate root**: external references point only to the root, and outside code reaches the inner objects only through it. The aggregate is the basic unit of storage transfer — you load and save the whole thing.

Rationale: the root is the guardian of the aggregate's invariants. Funnelling all changes through it means the cluster can never be left in a half-valid state, and a clear boundary tells you exactly how much must change together in one transaction.

How to: pick the root as the entity that owns the invariant, keep the inner objects (often value objects) reachable only via the root, and let other aggregates hold the root's id rather than a direct object reference.

Example: an Order is the root; its OrderLine value objects live inside it; the invariant "total equals the sum of line subtotals" is enforced by the Order, and nothing edits an OrderLine except through the Order.

## The four aggregate-design rules of thumb

These are heuristics, with documented reasons to break them — not laws:

1. **Model true invariants in consistency boundaries** — put inside one aggregate only what must stay consistent together within a single transaction.
2. **Design small aggregates** — often just the root plus value-typed properties; large aggregates lose to performance, scalability, and transaction contention.
3. **Reference other aggregates by identity** — hold the other root's id, not its object, so each aggregate stays a small, independently loadable unit.
4. **Use eventual consistency outside the boundary** — one transaction modifies one aggregate; consistency _across_ aggregates is reached afterward (e.g. via a domain event), not in the same transaction.

This is canonical, not folklore: one transaction modifies only one aggregate, and cross-aggregate consistency is eventual — applied as a rule of thumb with stated reasons to break it (for example, a small, tightly-coupled invariant that genuinely must be immediate).

BAD: a `Customer` aggregate that directly contains every `Order` object, every `OrderLine`, and every `Shipment`, all loaded and locked together to add one line.

GOOD: `Order` and `Shipment` are separate aggregates; `Order` holds `shipmentId`; placing an order emits an event that Shipping consumes, reaching consistency eventually.

## Domain event

A **domain event** captures something meaningful that happened in the domain, named in the past tense in the ubiquitous language (OrderPlaced, PaymentCaptured). It carries the identifiers and data describing the occurrence, and lets other aggregates or contexts react without the originator depending on them.

How to: emit the event from the aggregate root after the change commits, carry identifiers (not whole foreign objects) in its payload, and let interested contexts translate it through their own anti-corruption layer into their own concepts.

## Repository

A **repository** is a persistence-ignorant, collection-like abstraction that gives access to aggregate roots: you ask it for an aggregate by identity and it loads the whole aggregate; you hand it an aggregate and it saves the whole aggregate. It decouples the domain layer from infrastructure, so domain code reads as if working with an in-memory collection.

Rationale: the domain should express _what_ it needs ("give me the order with this id") without knowing _how_ it is stored. One repository per aggregate root, returning whole roots, keeps the aggregate boundary intact.

BAD: a `OrderDao` exposing `selectOrderLineRows()`, `updateTotalColumn()`, and SQL fragments — the domain now knows the table layout and can mutate an inner row without going through the root.

GOOD: `orders.byId(id): Order` and `orders.save(order)` — collection-like, root-only, persistence-ignorant. Whether it is backed by SQL, a document store, or memory is invisible to the domain.

A thin ORM/DAO wrapper that leaks rows and columns is not a repository: a repository hands out whole aggregate roots and hides persistence entirely.

## Domain service vs application service

A **domain service** holds domain logic that does not naturally belong to a single entity or value object — a calculation or policy spanning several aggregates, expressed in the ubiquitous language (e.g. a `FareCalculator` for a taxi dispatcher, or a transfer that touches two accounts).

An **application service** is a thin orchestration layer: it has no domain rules of its own, it loads aggregates via repositories, invokes domain behaviour, and coordinates the transaction and events. It is the entry point a controller calls.

How to: keep domain rules in entities, value objects, and domain services; keep the application service free of business logic — if it starts making domain decisions, push that logic down into the model.

## The anemic-domain-model anti-pattern

An **anemic domain model** is a model whose entities are bags of getters and setters with no behaviour, while all the logic lives in services that reach in and manipulate that data. It has the _shape_ of a domain model — named classes, fields — but none of the substance, because behaviour is divorced from the data it guards.

It is the documented anti-pattern, not a style of DDD: putting all logic in services with data-only entities loses the central benefit of binding rules to the data they protect, and invariants end up scattered and unenforced.

BAD: `Order` exposes `getLines()`/`setTotal()`; an `OrderService.recalculate(order)` reads the lines, sums them, and calls `setTotal` — any caller can set an inconsistent total.

GOOD: `Order.addLine(line)` updates the lines and recomputes the total internally, so the "total equals sum of subtotals" invariant cannot be violated from outside.

## Cross-references

- Where these blocks live and how words stay consistent — bounded contexts, ubiquitous language, anti-corruption layer — is in [strategic-design.md](strategic-design.md).
- Class and inheritance mechanics (how to structure the classes themselves) are owned by **oop-guide**.
- Coupling and cohesion metrics for the resulting code are owned by **connascence-guide**.
- An in-memory object/property database with stable cross-references, change notification, undo, and serialization is owned by **data-model-guide** — a different concept from a repository over aggregate roots.
- Driving an aggregate's behaviour from agreed examples is owned by **bdd-guide**; the unit of client-valued work that exercises it is owned by **user-stories-guide** and **fdd-guide**.
