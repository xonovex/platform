# Design and Implementation

## Establish the user and product context

Start with affected users, access needs, assistive strategies, devices, environments, languages, content formats, and complete journeys. Define the capability a person must complete, including authentication, errors, help, consent, payment, interruption, recovery, and support; do not assess only isolated screens.

Record which parts are authored, embedded, provider-controlled, generated, or user supplied. Ownership may be distributed, but the user journey still needs an explicit accessible outcome and escalation path.

## Turn criteria into testable requirements

For each selected criterion or organizational requirement, state:

- the user outcome and applicable content, component, state, and journey;
- design and content acceptance examples plus known non-examples;
- the platform implementation owner and native technique source;
- deterministic, automated, human, and assistive-technology checks;
- expected evidence, freshness, release behavior, and remediation owner.

Keep the requirement semantic. A web, Android, iOS, desktop, document, and kiosk implementation may use different APIs while satisfying the same selected outcome. Platform skills own API-level techniques; this skill owns assurance scope and evidence.

## Design the full interaction

Evaluate at least the applicable facets:

- text alternatives, captions, transcripts, structure, relationships, sequence, orientation, and sensory instructions;
- contrast, color-independent meaning, text resize, reflow, spacing, responsive layout, and non-text contrast;
- keyboard and switch access, focus order and visibility, target size, pointer alternatives, motion, timing, and seizure risk;
- navigation, headings, labels, instructions, errors, prevention, help, authentication, and consistent identification;
- programmatic name, role, value, state, status messages, live updates, and compatibility with assistive technology;
- plain language, localization, cognitive load, interruption, recovery, and accessible support channels where selected by the profile.

Do not infer semantics from visual appearance. Do not assume a component library makes every composition accessible; test effective labels, states, order, focus, scaling, and interaction in context.

## Include affected people without shifting assurance onto them

Representative user research and usability evaluation can reveal barriers that criteria and scanners miss. Define recruitment, consent, accommodations, compensation, privacy, safe handling of disability-related information, and how findings change design.

Participation complements, but does not replace, deterministic checks or qualified assessment. A participant is not responsible for proving conformance or discovering every defect.

## Handoff and change control

A handoff records the selected criteria, exact designs/content, interaction states, platform assumptions, native implementation owner, test plan, unresolved risks, exceptions, and evidence requirements. Reassess when a component, content source, journey, dependency, platform version, or assistive-technology assumption changes materially.
