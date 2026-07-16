---
name: ddd-guide
description: "Use when finding and naming domain boundaries and modelling inside them — establishing a ubiquitous language, drawing bounded contexts and a context map, protecting a model with an anti-corruption layer, and applying the tactical building blocks (entity, value object, aggregate + root, domain event, repository, domain/application service). Triggers on ubiquitous language, bounded context, context map, anti-corruption layer, aggregate / aggregate root, entity vs value object, domain event, repository, strategic vs tactical design, anemic domain model — even when the user doesn't say 'DDD'."
---

# Domain-Driven Design

Find the boundaries in a domain, name everything inside them rigorously, and model each boundary with building blocks that keep business invariants honest. Strategic design draws the map; tactical design fills it in.

## Essentials

- **Build one ubiquitous language per context** - bind agreed terms into both the model and the code, see [references/strategic-design.md](references/strategic-design.md)
- **Draw bounded contexts before classes** - each context is a boundary of model consistency where words mean exactly one thing, see [references/strategic-design.md](references/strategic-design.md)
- **Map contexts and shield with an anti-corruption layer** - translate foreign models both directions into your own terms, see [references/strategic-design.md](references/strategic-design.md)
- **Distinguish entity from value object** - identity-over-time versus defined-by-attributes-and-immutable, see [references/tactical-building-blocks.md](references/tactical-building-blocks.md)
- **Keep aggregates small with one root** - protect true invariants in a transaction, reference other aggregates by id, see [references/tactical-building-blocks.md](references/tactical-building-blocks.md)
- **Hand whole aggregates through a repository** - persistence-ignorant, collection-like access to roots only, see [references/tactical-building-blocks.md](references/tactical-building-blocks.md)

## Gotchas

- A bounded context is not a microservice or a subdomain: it is a solution-space model boundary; a subdomain is a problem-space area; a service is a deployment unit. They often align but are not the same thing.
- Ubiquitous language is not "a glossary". It only counts when the agreed terms live in the model AND the code; a wiki nobody compiles against is not a ubiquitous language.
- A thin ORM/DAO wrapper is not a repository, and an upstream-API client is not an anti-corruption layer: a repository hands out whole aggregate roots and is persistence-ignorant; an ACL translates foreign concepts both directions, not just forwards calls.
- Code-level coupling metrics belong to **connascence-guide**, class/inheritance mechanics to **oop-guide**, and undo/serialization machinery to **data-model-guide** — DDD names the domain boundaries, not those.

## Example

```text
Bounded context: Ordering
  Ubiquitous language: Customer, Order, OrderLine, place, fulfil
  Aggregate Order (root)
    invariant: total == sum(orderLines.subtotal)   // enforced in one transaction
    orderLines: OrderLine[]                          // value objects, no identity
    references Shipment by id, never by object       // cross-aggregate = eventual consistency
  emits OrderPlaced (domain event, carries identifiers)
  Repository: orders.save(order) / orders.byId(id)   // whole aggregate, persistence-ignorant

Bounded context: Shipping
  "Customer" here means Recipient: name + address + phone
  anti-corruption layer translates OrderPlaced -> Recipient   // does not import the Ordering model
```

## Progressive Disclosure

- Read [references/strategic-design.md](references/strategic-design.md) - Load when finding boundaries: ubiquitous language, bounded contexts, context maps, context-mapping patterns, anti-corruption layer, subdomains
- Read [references/tactical-building-blocks.md](references/tactical-building-blocks.md) - Load when modelling inside a context: entity, value object, aggregate + root, domain event, repository, domain/application service
