# Sources

## Origin of BDD and Given-When-Then

- **Title:** Dan North — "Introducing BDD"
- **URL:** https://dannorth.net/introducing-bdd/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/appendix-discovery.md`
  - `SKILL.md` → Essentials
- **Aspects extracted:**
  - The origin of BDD; "should" test naming; JBehave and removing the word "test"; the Given-When-Then template; "a ubiquitous language for the analysis process"; the behaviour-not-testing reframing

## Gherkin notation

- **Title:** Cucumber — "Gherkin Reference"
- **URL:** https://cucumber.io/docs/gherkin/reference/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/gherkin-reference.md`
- **Aspects extracted:**
  - Authoritative keyword definitions: Feature, Rule, Scenario/Example, Scenario Outline/Template, Examples/Scenarios, Background, Given/When/Then/And/But; Then = assertion

## The three BDD practices and living documentation

- **Title:** Cucumber — "BDD"
- **URL:** https://cucumber.io/docs/bdd/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/specification-by-example.md`
  - `references/appendix-discovery.md`
- **Aspects extracted:**
  - The three practices (discovery / formulation / automation); examples as guide-rails; living documentation; "the code reflects the documentation, and the documentation reflects the team's shared understanding"

## Example Mapping

- **Title:** Matt Wynne — "Introducing Example Mapping" (Cucumber blog)
- **URL:** https://cucumber.io/blog/bdd/example-mapping-introduction/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/discovery-three-amigos.md`
- **Aspects extracted:**
  - The four card colours (yellow story / blue rules / green examples / red questions); the ~25-minute session; the thumb vote; the story-split signals; the minimum amigo group

## The collaboration-tool framing (Appendix A)

- **Title:** Aslak Hellesøy (creator of Cucumber) — comment on the "Cucumber: Behaviour-Driven Development" thread (Hacker News, item 10194242), corroborated by "The World's Most Misunderstood Collaboration Tool" (Cucumber blog)
- **URLs:**
  - https://news.ycombinator.com/item?id=10194242
  - https://cucumber.io/blog/collaboration/the-worlds-most-misunderstood-collaboration-tool/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/appendix-discovery.md`
  - `SKILL.md` → Gotchas
- **Aspects extracted:**
  - Cucumber tests people's understanding of how yet-to-be-written software should behave, not the software; most rework comes from misunderstandings
  - The ~20-minute Discovery Workshop with business + IT; concrete plain-language examples ("the one with five taxis in range"); the thumbs-up/down vote to accept or split a story
  - A developer fleshes out 2-5 Gherkin scenarios, the business confirms them, then regular TDD drives the core domain logic with external services / message queues / databases stubbed out (no UI / Selenium)
  - Far more low-level unit tests than Cucumber scenarios; "Cucumber makes sure you write the right code; unit tests make sure you write the code right"; the output is executable, living documentation

## Specification by Example

- **Title:** Gojko Adzic — "Specification by Example: How Successful Teams Deliver the Right Software" (Manning, 2011)
- **URL:** https://www.manning.com/books/specification-by-example
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/specification-by-example.md`
- **Aspects extracted:**
  - The seven key process patterns; collaborative specification; specs-as-tests; trustworthy living documentation; Specification by Example as equivalent to ATDD (the term was coined by Martin Fowler in 2002 and systematised with seven patterns and 50+ case studies in this book; pattern lists verified against Part 2)

## ATDD and the three amigos

- **Title:** Agile Alliance Glossary — "ATDD"
- **URL:** https://agilealliance.org/glossary/atdd/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/discovery-three-amigos.md`
  - `SKILL.md` → the ATDD-and-BDD-are-one-practice framing
- **Aspects extracted:**
  - The ATDD definition; the three amigos as three perspectives (business / development / testing); ATDD documented as synonymous with Specification by Example, Story-Test-Driven Development, and BDD

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull
3. Update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
