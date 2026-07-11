# ddd: Tactical Building Blocks

Tactical design fills a single bounded context with a small named set of modelling blocks that keep invariants honest while staying ignorant of infrastructure. Strategic design ([strategic-design.md](strategic-design.md)) draws the boundaries.

## Entity vs value object

- **Entity** — an identity that persists over time and across attribute changes; equality by identity. An Order is the same Order tomorrow after its lines change.
- **Value object** — defined entirely by its attributes, no identity, immutable; equality by value; to "change" it you create a new one. 12.50 EUR is interchangeable with any other 12.50 EUR.

Ask "do I care _which_ one this is, or only _what_ it is?" Default to value objects. Don't carry a domain concept as primitives — model `Money { amount, currency }` and `Address { ... }` as immutable value objects with their own validation.

## Aggregate and aggregate root

A cluster of domain objects treated as one unit for data changes. One entity is the **aggregate root**: external references point only to the root, inner objects are reached only through it, and you load/save the whole aggregate. The root guards the aggregate's invariants — funnelling all changes through it prevents half-valid states.

Example: `Order` is the root; its `OrderLine` value objects live inside; the invariant "total equals the sum of line subtotals" is enforced by `Order`; nothing edits an `OrderLine` except through `Order`.

## Four aggregate-design rules of thumb

Heuristics with documented reasons to break them, not laws:

1. **Model true invariants in consistency boundaries** — put in one aggregate only what must stay consistent together within a single transaction.
2. **Design small aggregates** — often just the root plus value-typed properties; large aggregates lose to performance, scalability, transaction contention.
3. **Reference other aggregates by identity** — hold the other root's id, not its object.
4. **Use eventual consistency outside the boundary** — one transaction modifies one aggregate; cross-aggregate consistency is reached afterward (e.g. via a domain event).

Break rule 4 only for a small, tightly-coupled invariant that genuinely must be immediate.

- BAD: a `Customer` aggregate directly containing every `Order`, `OrderLine`, and `Shipment`, all loaded and locked to add one line.
- GOOD: `Order` and `Shipment` are separate aggregates; `Order` holds `shipmentId`; placing an order emits an event Shipping consumes, reaching consistency eventually.

## Domain event

Something meaningful that happened, named past-tense in the ubiquitous language (`OrderPlaced`, `PaymentCaptured`). Emit it from the root after the change commits; carry identifiers (not whole foreign objects) in the payload; let other contexts translate it through their own ACL.

## Repository

A persistence-ignorant, collection-like abstraction over aggregate **roots**: ask it for an aggregate by identity (loads the whole root), hand it an aggregate (saves the whole root). One repository per aggregate root. A thin ORM/DAO wrapper that leaks rows/columns is not a repository.

- BAD: `OrderDao` exposing `selectOrderLineRows()`, `updateTotalColumn()`, SQL fragments — the domain knows the table layout and can mutate an inner row without the root.
- GOOD: `orders.byId(id): Order` and `orders.save(order)` — root-only, persistence-ignorant; SQL/document/memory backing is invisible.

## Domain service vs application service

- **Domain service** — domain logic that fits no single entity/value object: a calculation or policy spanning several aggregates, named in the ubiquitous language (`FareCalculator`, an account-to-account transfer).
- **Application service** — thin orchestration with no domain rules: loads aggregates via repositories, invokes domain behaviour, coordinates the transaction and events; the entry point a controller calls. If it starts making domain decisions, push that logic down into the model.

## Anemic-domain-model anti-pattern

Entities that are bags of getters/setters with no behaviour, while all logic lives in services that reach in and manipulate the data. It has the shape of a domain model but binds no rules to the data they protect, so invariants scatter and go unenforced.

- BAD: `Order.getLines()`/`setTotal()` plus `OrderService.recalculate(order)` — any caller can set an inconsistent total.
- GOOD: `Order.addLine(line)` updates the lines and recomputes the total internally, so "total equals sum of subtotals" cannot be violated from outside.

## Cross-references

- Strategic context — bounded contexts, ubiquitous language, ACL — in [strategic-design.md](strategic-design.md).
- Class/inheritance mechanics — **oop-guide**. Coupling/cohesion metrics — **connascence-guide**.
- In-memory object/property database — **data-model-guide**; distinct from a repository over aggregate roots.
- Driving behaviour from agreed examples — **bdd-guide**; the client-valued unit of work — **user-stories-guide**, **fdd-guide**.
