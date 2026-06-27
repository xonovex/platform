# user-stories: Appendix — User-Story Best Practices

A single-page, project-agnostic checklist for the whole craft. Each item links to the reference that develops it. Use it as a final pass over a story or a backlog.

## The story is a placeholder for a conversation about value

A user story is not a specification handed down; it is a small token that promises a conversation about user value. Keep it lightweight and resolve the detail by talking, not by writing more. The three components are the Card, the Conversation, and the Confirmation; see [three-cs.md](three-cs.md).

## Check the story against INVEST

Independent, Negotiable, Valuable, Estimable, Small, Testable. **Small** means a few person-weeks at most — small enough to finish within one sprint. A failed letter is a prompt to negotiate, split, or clarify. See [invest-and-smart.md](invest-and-smart.md).

## Write What and Why, not How

Use "As a [user], I want [capability], so that [value]" to capture who, what, and why while leaving the implementation open and Negotiable. The template is a conversation starter, not a mandatory format. See [template-and-acceptance-criteria.md](template-and-acceptance-criteria.md).

## Give every story clear acceptance criteria

Cover the happy path, boundaries, error/invalid cases, and UI changes — not the happy path alone. The criteria are the Confirmation from the 3 Cs. They can be written as Given-When-Then for automation, but that notation is owned by **bdd-guide**, and the test design behind them by **testing-guide**. See [template-and-acceptance-criteria.md](template-and-acceptance-criteria.md).

## Slice vertically

Every story is a thin end-to-end slice that delivers usable value, never a horizontal layer ("the DB", "the API", "the UI"). The first slice of a new capability is the walking skeleton: production code with tests linking the main components, not a throwaway prototype or a spike. See [splitting-flowchart.md](splitting-flowchart.md).

## Split a too-big story with SPIDR or the flowchart

SPIDR (Spike, Path, Interface, Data, Rules) is the compact starter set; the splitting flowchart is the fuller decision tree with preconditions and two evaluation rules (enable deprioritization; favor equal-sized splits). A Spike yields knowledge, not value. See [splitting-spidr.md](splitting-spidr.md) and [splitting-flowchart.md](splitting-flowchart.md).

## Refine the backlog continuously

Refinement is ongoing, not a fixed grooming event. The Developers size the items; an item is ready when it is doable within one sprint, and PM, Developers, and QA collaborate on it. See [backlog-refinement.md](backlog-refinement.md).

## Where this skill stops

- Given-When-Then notation and executable specs — **bdd-guide**.
- Test design, levels, and doubles — **testing-guide**.
- Domain modelling and ubiquitous language — **ddd-guide**.
- The FDD feature list and domain walkthrough — **fdd-guide**.
- The coupling theory behind "Independent" — **connascence-guide**.
