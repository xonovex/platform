# Sources

> The "application / class / method-level" smell grouping (as in the Wikipedia "Code smell" article) is editorial and weakly sourced — it is **not** used as a citation of record here. Smell membership follows the peer-reviewed design-problem taxonomy below; class-vs-method is cited only as a detection-granularity dimension.

## Code-smell taxonomy (design-problem families)

- **Title:** Mäntylä, Vanhanen & Lassenius — "A Taxonomy and an Initial Empirical Study of Bad Smells in Code" (IEEE ICSM 2003; extended in Empirical Software Engineering 11(3), 2006), the five-family grouping (Bloaters / OO-Abusers / Change-Preventers / Dispensables / Couplers) over Fowler & Beck's "Refactoring" Ch. 3 "Bad Smells in Code"
- **URLs:**
  - https://mmantyla.github.io/BadCodeSmellsTaxonomy
  - https://martinfowler.com/books/refactoring.html
  - https://refactoring.guru/refactoring/smells
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/smell-catalog.md` (family grouping + member smells, routed to owners)
- **Aspects extracted:**
  - The five design-problem families and their members; the catalog routes each smell to its owning skill

## Detection granularity (class vs method)

- **Title:** Arcelli Fontana et al. — "Comparing and Experimenting Machine Learning Techniques for Code Smell Detection" (Empirical Software Engineering, 2016); Lanza & Marinescu — "Object-Oriented Metrics in Practice" (Springer, 2006) metric-based detection strategies
- **URLs:**
  - https://link.springer.com/article/10.1007/s10664-015-9378-4
  - https://link.springer.com/book/10.1007/3-540-39538-5
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/smell-catalog.md`
- **Aspects extracted:**
  - Class-level vs method-level smell detection (a detection dimension, not a taxonomy); "design disharmony" strategies bound to method / class / hierarchy entities

## Complexity metrics

- **Title:** Cyclomatic Complexity (Thomas McCabe); Cognitive Complexity (G. Ann Campbell, SonarSource, 2017)
- **URLs:**
  - https://www.sonarsource.com/resources/cognitive-complexity/
  - https://en.wikipedia.org/wiki/Cyclomatic_complexity
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/robustness.md` (code smells)
- **Aspects extracted:**
  - Decision-point count and warn thresholds; understandability via flow-breaks + nesting; complexity is branching/nesting, not line count

## Robustness & secure boundaries

- **Title:** "Parse, don't validate" (Alexis King); OWASP Cheat Sheets (Input Validation, Error Handling, Logging); "Fail Fast" (Martin Fowler, IEEE Software); "Make illegal states unrepresentable" (DevIQ)
- **URLs:**
  - https://lexi-lambda.github.io/blog/2019/11/05/parse-don-t-validate/
  - https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html
  - https://martinfowler.com/ieeeSoftware/failFast.pdf
  - https://deviq.com/principles/make-illegal-states-unrepresentable/
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/robustness.md`
- **Aspects extracted:**
  - Parse at the boundary so the type is earned; validate at trust boundaries; fail-fast; preserve error cause + context; log at boundaries/error paths; never log secrets/PII

## Supplementary smells (outside the original catalog)

- **Title:** Robert C. Martin — "Clean Code" Ch. 17 "Smells and Heuristics" (G5 Duplication, G25 Magic Numbers, F1 Too Many Arguments); Liskov & Wing — "A Behavioral Notion of Subtyping" (ACM TOPLAS, 1994), the Liskov Substitution Principle behind Refused Bequest and downcasting
- **URLs:**
  - https://www.oreilly.com/library/view/clean-code-a/9780136083238/chapter17.xhtml
  - https://dl.acm.org/doi/10.1145/197320.197383
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `references/robustness.md` (magic numbers), `references/smell-catalog.md` (provenance of supplementary smells)
- **Aspects extracted:**
  - Magic numbers / too-many-arguments / duplication heuristics; LSP as the principle behind refused bequest and downcasting (owned by oop-guide)

## Grading & severity

- **Title:** SonarQube issue severities (Blocker/Critical/Major/Minor/Info); Maintainability Index (Microsoft); CVSS qualitative severity scale (FIRST); SQALE technical-debt / remediation-effort (Jean-Louis Letouzey)
- **URLs:**
  - https://docs.sonarsource.com/sonarqube-server/10.3/user-guide/issues
  - https://learn.microsoft.com/en-us/visualstudio/code-quality/code-metrics-maintainability-index-range-and-meaning
  - https://www.first.org/cvss/v3.1/specification-document
- **Last reviewed:** 2026-06-27
- **Used for:**
  - `SKILL.md` → Essentials (grade by severity); `references/robustness.md`, `references/smell-catalog.md`
- **Aspects extracted:**
  - Severity tiers; grade by blast radius × likelihood and remediation effort, not finding count

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added smells/metrics)
3. Update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
