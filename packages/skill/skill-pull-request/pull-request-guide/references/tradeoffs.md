# tradeoffs: Surface Limitations and Decisions Early

**Guideline:** State known limitations, risks, and non-obvious decisions in the description, before a reviewer has to find them.

**Rationale:** Naming a tradeoff up front signals technical honesty and saves a back-and-forth in review comments. Hidden limitations get discovered late or in production. A "why I chose A over B" lets reviewers engage with your reasoning instead of guessing at it.

**How to Apply:**

1. Add a short "Tradeoffs / risks" section when the change has any.
2. State each known limitation and why it is acceptable for now.
3. Note alternatives you considered and why you did not take them.
4. Flag anything that affects backwards compatibility, data, config, or other consumers.
5. If the change diverges from a related convention or another branch, explain why.

**Example:**

```text
// Bad
(no mention of the gap)

// Good
## Tradeoffs / risks
- We create the queue names the service actually reads, the legacy script created a differently-named, unused queue.
- Local-dev tooling only, production images are unaffected.
```

**Counter-Example:** A change with no limitations or risky decisions does not need the section - do not invent risks.

**Related:** [description.md](./description.md)
