# Role Lenses

A role lens is a convenience mapping to suggested perspectives, criteria, evidence,
and communication depth. It is not a persona, executor identity, permission set,
owner, workflow, or stage sequence.

Prefer explicit repeatable `--perspective` selections. Use `--role` when a familiar
organizational lens is more convenient, then inspect the resolved suggestions.

## Suggested mappings

| Role lens   | Possible perspectives                       | Typical evidence questions                                             |
| ----------- | ------------------------------------------- | ---------------------------------------------------------------------- |
| product     | user value, scope, acceptance               | Does this solve the stated problem and cover the accepted outcome?     |
| design      | usability, accessibility, content           | Can intended users understand and operate it across relevant contexts? |
| engineering | correctness, maintainability, compatibility | Is the change bounded, supportable, and safe for consumers?            |
| quality     | behavior, evidence quality, regression      | Is each binding criterion reproducibly demonstrated?                   |
| security    | threat, privacy, abuse, auditability        | What trust boundary changed and what evidence supports the control?    |
| operations  | reliability, operability, recovery          | Can the change be observed, reversed, and supported?                   |

Mappings are suggestions, not exhaustive bundles. The runtime records:

- every resolved perspective;
- whether it was explicit, inherited, suggested, or inferred;
- the reason it applies;
- questions, evidence needs, and advisory criteria it added;
- any specialist skill selected to support it.

## Authority boundary

A role lens cannot:

- grant access or approval;
- determine executor identity;
- claim organizational ownership;
- force an operation order;
- select a provider;
- make an advisory criterion binding;
- authorize merge, release, deployment, publication, or cleanup.

The runtime or provider owns authority. Explicit caller input, an authoritative
artifact, or mandatory policy determines binding criteria.

## Example

```text
/xonovex-workflow:review ./checkout-design.md \
  --role design \
  --perspective privacy \
  --perspective localization
```

The design role may suggest usability and accessibility. Privacy and localization
remain explicit additional perspectives. The result reports the full resolved set;
no operation chain is implied.

## Related guides

- [Command inventory](../README.md)
- [Provider-native resource bindings](references.md)
- [Invocation, effects, and execution](invocation.md)
- [Contract migration](migration.md)
- [Operation model](../../../diagram/diagram-agent-workflow/operation-model.png)
