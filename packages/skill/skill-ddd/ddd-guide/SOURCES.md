# Sources

## Ubiquitous language

- **Title:** Martin Fowler — "UbiquitousLanguage" (bliki)
- **URL:** https://martinfowler.com/bliki/UbiquitousLanguage.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/strategic-design.md`
- **Aspects extracted:**
  - The definition; rigour because software cannot cope with ambiguity; grounded in the domain model; used pervasively (speech, diagrams, code) until it flows

## Bounded context and context map

- **Title:** Martin Fowler — "BoundedContext" (bliki)
- **URL:** https://martinfowler.com/bliki/BoundedContext.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/strategic-design.md`
- **Aspects extracted:**
  - Bounded context as a central DDD pattern; multiple canonical models; one ubiquitous language per context; the polysemy of "Customer" / "Product" / "meter"; the need for a context map

## Aggregates

- **Title:** Martin Fowler — "DDD_Aggregate" (bliki)
- **URL:** https://martinfowler.com/bliki/DDD_Aggregate.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/tactical-building-blocks.md`
- **Aspects extracted:**
  - Aggregate as a cluster treated as a single unit; the aggregate root; external references only to the root; transactions do not cross aggregate boundaries; the aggregate as the unit of storage transfer; the Order/line-items example

## DDD overview: strategic vs tactical

- **Title:** Martin Fowler — "DomainDrivenDesign" (bliki)
- **URL:** https://martinfowler.com/bliki/DomainDrivenDesign.html
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/strategic-design.md`
  - `references/tactical-building-blocks.md`
- **Aspects extracted:**
  - The strategic vs tactical split; the building-block list; entity vs value object; repository; the origin in the 2003 book

## Aggregate design rules

- **Title:** Vaughn Vernon — "Implementing Domain-Driven Design" (the Red Book), Ch. 10, via InformIT / ArchiLab
- **URL:** https://www.informit.com/articles/article.aspx?p=2020371&seqNum=3
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/tactical-building-blocks.md`
- **Aspects extracted:**
  - The four aggregate-design rules of thumb; the design-small-aggregates rationale; reference-by-identity; eventual consistency outside the boundary; documented reasons to break the rules

## Anti-corruption layer, entity/value object, repository

- **Title:** Eric Evans — "Domain-Driven Design Reference", corroborated via Wikipedia "Domain-driven design"
- **URL:** https://en.wikipedia.org/wiki/Domain-driven_design
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/strategic-design.md`
  - `references/tactical-building-blocks.md`
- **Aspects extracted:**
  - The anti-corruption layer; entity vs value object; the repository; subdomains; rejecting one unified model in favour of bounded contexts; the term coined by Evans in 2003

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull, watching for the recurring folklore claims (bounded context = microservice; bounded context = subdomain; ubiquitous language = glossary; repository = ORM wrapper; anemic-model-is-DDD) so the corrections in the references stay accurate
3. Update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
