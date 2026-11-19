# Labels and content

What a source names and what it leaves out. The files, the diagram kinds and the layout are in [drawing.md](drawing.md).

## Participant names

- Name a participant by the role of the component, using the name the decision files and the plan use for it, so the picture and the text match. A client is named by its platform, such as the web app or the mobile app.
- Draw several clients as separate participants only when the options differ for them. When they behave the same, one participant named for all of them keeps the picture short.
- Show a gateway or a proxy as its own participant only when a decision changes what passes through it.
- A path that a client calls is a label on an arrow, not a participant.
- Write labels in the language of the plan.

## What stays out

- Sensitive findings: no token, key, password, internal hostname, address, ticket key or person's name in a label, note or comment. Name the component by its role.
- A remote `!include`. Sources stay self-contained so they render offline.
- Payload shapes, gates and anything a table says better.
- The reasoning behind a rejected option. A diagram shows a flow, the decision file holds the reason.
- Display numbers.

The diagrams carry working options that were never decided, so do not copy an image into a published documentation page.
