# Choosing what gets a picture

## Contents

- The rule
- Draw
- Do not draw
- The choice line
- Worked example

## The rule

A decision gets a picture when its options change the topology: who calls whom, who owns a step, where a message or a request travels. A reader sees that faster than in prose. Anything else is shown better by a table or a sentence, and a picture of it hides the difference behind boxes.

## Draw

- **The lead diagram**, once something is settled: the flow or the structure as it will be, with settled elements only.
- **One diagram for each set of options** of an open, paused or partly settled decision whose options differ in topology. One picture holds all options of the decision, one branch for each.
- **A system model** (Structurizr) when several views need the same elements, such as a context view and a view of one container, and keeping the elements in step by hand would drift.
- **The current flow**, when the research shows it and the options only make sense against it. Treat it as an element state, "current", in the lead or the option diagram, not as a separate picture.

## Do not draw

- A payload shape, a field list or a data format: use a table of fields.
- A gate, a rule or a condition on rollout: use a table of the conditions and what each one switches.
- A decision whose options all sit inside one component with no change in who calls whom.
- A decision that is settled and does not alter the lead diagram.
- Anything that restates a table the plan already holds.

When asked for one of these, say in one line why a table is better and offer to draw the topology it sits in, if there is one.

## The choice line

The reply lists every decision once, before any drawing:

- `<decision title>`: option diagram `<topic>-option`, because the options move the call from the app to the service.
- `<decision title>`: no picture, a payload shape; the research file holds the field table.
- `<decision title>`: settled, shown in the lead diagram.

## Worked example

A change moves reminders from a nightly job to an event-driven sender. Three decisions:

1. Who triggers the reminder: the booking service on a state change, or a scheduler reading the booking store. The options change who calls whom: an option diagram.
2. Which fields the reminder message carries. A payload shape: a table, no picture.
3. Whether the sender is switched on per market or everywhere. A gate: a table, no picture.

The decision on the transport is settled (a message queue), so the lead diagram shows the booking service, the queue and the sender, and nothing about the two options of decision 1.
