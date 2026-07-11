# fdd-processes: The Five Processes, Feature Template, and Milestones

FDD-specific residue only: the domain model belongs to **ddd-guide**, "client value" and how to slice it to **user-stories-guide**, and the design/coupling standards the inspections enforce to **oop-guide** / **connascence-guide**.

## The five processes

Run in order:

1. **Develop an Overall Model** — high-level object model of the domain with domain experts (this is domain modelling → **ddd-guide**).
2. **Build a Features List** — decompose the model into a categorised list of client-valued features.
3. **Plan by Feature** — sequence feature sets, assign feature sets to Chief Programmers and classes to Class Owners.
4. **Design by Feature** — design (sequence diagrams, class/method skeletons) a selected group of features and inspect it.
5. **Build by Feature** — implement, unit-test, inspect, and promote to the build.

Processes 1-3 run once at project start; 4-5 are the iterative inner loop, repeating per feature set against the fixed model + list + plan.

## The feature-naming template

```
<action> the <result> by|for|of|to a(n) <object>
```

The connector is one of **by / for / of / to** — chosen to read naturally, not fixed to "of". A feature is a small, client-valued function — not a layer, class, or component.

```
GOOD:  Calculate the total of an order
       Authorize the payment for an order
       Assign the nearest taxi to a ride request
BAD:   Build the OrderService            <- a component/task, not client-valued
       Calculate the total of a sale of an order  <- forcing "of" where "for" reads better
```

## Feature, feature set, major feature set

- **Feature** — one client-valued function named with the template.
- **Feature set** — features supporting one business activity (e.g. "Processing an order").
- **Major feature set** — a subject area grouping feature sets (e.g. "Order Management").

The features list (Process 2) is this three-level catalogue.

## Feature sizing

Max **two weeks** per feature; most run 1-3 days (1-10 days effort). Decompose any business step larger than two weeks. This is a max size for one feature, not an iteration length — FDD predates and does not use fixed sprints.

## The six weighted milestones

Per-feature build progress, weights summing to 100%:

| Milestone          | Weight |
| ------------------ | ------ |
| Domain Walkthrough | 1%     |
| Design             | 40%    |
| Design Inspection  | 3%     |
| Code               | 45%    |
| Code Inspection    | 10%    |
| Promote to Build   | 1%     |

Weights are exact and deliberately uneven (Design + Code dominate); an even 16.6% split is folklore. Percent-complete is the sum of reached-milestone weights — a feature past Design Inspection but mid-Code reads 44%. Design and code inspections are first-class practices, not optional.

## Roles and feature teams

Six key roles: **Project Manager, Chief Architect, Development Manager, Chief Programmer, Class Owner, Domain Expert** (plus supporting roles). Two are essential for the inner loop:

- **Class Owner** — single-owner-per-class: one person responsible for each class (deliberate tension with collective code ownership).
- **Chief Programmer** — leads a **feature team**: a small team formed dynamically per feature set, pulling in the Class Owners of the classes that set touches. Not a fixed role and not a "feature"; it dissolves and reforms as feature sets change.
