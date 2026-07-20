# Architecture and Composition

The executable core is a small plugin runtime. It knows how to resolve one executor, run
selected controls before and after it, and publish to selected evidence sinks. It does not
know about shell scripts, LLMs, agents, hooks, CI products, Kubernetes, or A1/A2/A3.

```text
native event -> trigger adapter -> invocation
                                  |
                    registry -> composition root
                                  |
        before controls -> executor -> after controls -> evidence sinks
```

## Stable ports

- A trigger adapter normalizes an event and binds it to trusted wiring.
- An executor returns a neutral outcome and references.
- A control returns `allow`, `deny`, or `abstain` at a declared phase.
- An evidence sink records an event and returns a reference.

The registry owns executable paths and adapter configuration. The invocation selects
registered plugin names and explicit modes; untrusted event data does not define commands
or silently add controls.

## Intrinsic enforcement

The runtime enforces only:

- invocation schema validity;
- plugin registration;
- explicitly required capabilities;
- denials from controls selected as `enforce`;
- evidence sink failures selected as `fail`.

Everything else is an optional plugin or host concern. Retries, budgets, token limits,
approval, provenance, protected targets, escalation, and maturity are not global defaults.

Persistence and provider handoffs remain caller-owned. The runtime executes a selected
composition without requiring a universal result envelope or storage model.
