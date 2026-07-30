# Cross-Cutting Concerns Are Not Axes

An axis partitions choices; a cross-cutting concern applies to every selected combination. Logging, telemetry, authorization, and a universal policy check usually belong around the composed operation, not beside selectable leaves.

```text
BAD  channel/{email,sms,push,logging}/
BAD  identical audit check copied into email/, sms/, and push/
GOOD compose selected channel, then wrap it once with audit and policy
```

Use three shapes:

- **Axis**: one mutually exclusive choice from a family.
- **Bridge**: one localized constraint between two specific variants.
- **Cross-cutting concern**: one rule applied uniformly across all choices.

Centralize homogeneous behavior once. When each variant contributes different details, let the composition mechanism invoke one narrow, uniformly defined hook; each variant supplies only its own data. Registry and capability mechanics belong to **microkernel-pattern-guide**.
