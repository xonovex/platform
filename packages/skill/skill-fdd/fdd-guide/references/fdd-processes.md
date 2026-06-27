# fdd-processes: The Five Processes, Feature Template, and Milestones

FDD organises delivery as five named processes, a strict feature-naming template, a three-level feature hierarchy, six weighted build milestones, and a role structure with dynamically formed feature teams. This file is the FDD-specific residue: the domain model itself belongs to **ddd-guide** and the meaning of "client value" / how to slice it belongs to **user-stories-guide**.

## Contents

- The five processes
- The feature-naming template
- Feature, feature set, major feature set
- Feature sizing
- The six weighted milestones
- Roles and feature teams
- What FDD does not own

## The five processes

FDD defines exactly five processes, run in order:

1. **Develop an Overall Model** — build a high-level object model of the domain with domain experts.
2. **Build a Features List** — decompose the model's functionality into a categorised list of client-valued features.
3. **Plan by Feature** — sequence the feature sets, assign feature sets to Chief Programmers and classes to Class Owners.
4. **Design by Feature** — for a selected group of features, produce the design (sequence diagrams, class/method skeletons) and inspect it.
5. **Build by Feature** — implement, unit-test, inspect, and promote the designed features to the build.

Processes 1-3 run once at the start of a project to establish the model, the list, and the plan. Processes 4-5 form the iterative inner loop: they repeat per feature set until the list is built. The rationale is that a stable shape (model + complete feature list + plan) is established up front, then small client-valued increments are designed and built against it — short iterations against a known whole, rather than discovering scope as you go.

Process 1 is domain modelling: do not redefine it here. Hand the object model, its ubiquitous language, and its boundaries to **ddd-guide**; FDD's contribution is placing modelling first and feeding the model into the feature list.

## The feature-naming template

A feature is a small, client-valued function. Name it with the template:

```
<action> the <result> by|for|of|to a(n) <object>
```

The connector is one of **by / for / of / to** — chosen to read naturally, not fixed to "of". Each feature names an action, the result that action produces, and the object it acts on.

Worked neutral examples:

- Calculate the **total** _of_ an **order**
- Apply the **discount** _to_ an **order**
- Validate the **shipping address** _of_ an **order**
- Reserve the **inventory** _for_ an **order line**
- Authorize the **payment** _for_ an **order**
- Assign the **nearest taxi** _to_ a **ride request**
- Send the **confirmation** _to_ a **customer**

BAD (narrowed to one connector, awkward, or not client-valued):

```
Order total of           <- no action, no readable result/object
Build the OrderService    <- a component/task, not a client-valued function
Calculate the total of a sale of an order  <- forcing "of" where "for"/"to" reads better
```

GOOD:

```
Calculate the total of an order
Authorize the payment for an order
```

The verb-result-object shape keeps features small, concrete, and phrased in client terms. The unit of work and of progress is the client-valued function — not a layer, a class, or a component. The notion of "client value" and how to slice for it is owned by **user-stories-guide**; FDD only fixes the naming convention and the granularity.

## Feature, feature set, major feature set

Three levels of decomposition:

- **Feature** — one client-valued function named with the template above.
- **Feature set** — a group of features that support a particular business activity (e.g. "Processing an order").
- **Major feature set** — a subject area grouping several feature sets (e.g. "Order Management").

The features list (Process 2) is this three-level catalogue. Plan by Feature (Process 3) then sequences the feature sets and assigns ownership.

## Feature sizing

A feature must take **no more than two weeks**. In practice features run 1-10 days of effort, with most at 1-3 days. A step of a business activity that would take longer than two weeks is decomposed into smaller features until each fits.

This two-week figure is a _maximum size for a single feature_, not an iteration length. FDD predates and does not use fixed sprints; do not equate the two-week cap with a two-week Scrum timebox. Sizing for client value (INVEST-style) is owned by **user-stories-guide**; FDD only states the hard upper bound and the decomposition rule.

## The six weighted milestones

Build-by-Feature progress is tracked per feature with six milestones whose weights sum to 100%:

| Milestone          | Weight |
| ------------------ | ------ |
| Domain Walkthrough | 1%     |
| Design             | 40%    |
| Design Inspection  | 3%     |
| Code               | 45%    |
| Code Inspection    | 10%    |
| Promote to Build   | 1%     |

The weights are exact and deliberately uneven: Design (40%) and Code (45%) dominate because they carry the work; the walkthrough, inspections, and promotion are lighter checkpoints. Reporting percent-complete is just summing the weights of the milestones a feature has reached — e.g. a feature past Design Inspection but mid-Code reads 44%.

BAD (inventing an even split):

```
Each of 6 milestones = 16.6%   <- folklore; loses the design/code emphasis
```

GOOD:

```
Walkthrough 1 / Design 40 / Design Insp 3 / Code 45 / Code Insp 10 / Promote 1 = 100
```

Inspections (design and code) are first-class FDD practices, not optional extras; they hold weight in the milestone set.

## Roles and feature teams

FDD names six key project roles:

- **Project Manager**
- **Chief Architect**
- **Development Manager**
- **Chief Programmer**
- **Class Owner**
- **Domain Expert**

plus supporting and additional roles. Two are load-bearing for the inner loop:

- A **Class Owner** owns the classes touched by a feature. Ownership is single-owner-per-class — one person is responsible for each class. This is in deliberate tension with collective code ownership; it is an FDD design choice, not an accident.
- A **Chief Programmer** leads a **feature team**: a small team formed _dynamically per feature set_, pulling in the Class Owners of the classes that feature set touches. The feature team is not a fixed role and is not the same thing as a "feature"; it dissolves and reforms as feature sets change.

In Design by Feature and Build by Feature, the Chief Programmer convenes the Class Owners whose classes are involved, the team designs and inspects, each owner implements changes to their own classes, and the work is inspected and promoted.

The quality bar that design and inspection aim at — what "good design" means — is owned by **oop-guide** and **connascence-guide**; FDD prescribes that the inspections happen and who attends, not the design rules they apply.

## What FDD does not own

- The object model produced by Process 1 and its vocabulary -> **ddd-guide**.
- The definition of client value and how to slice work into client-valued increments -> **user-stories-guide**.
- The design and coupling standards the inspections enforce -> **oop-guide**, **connascence-guide**.
