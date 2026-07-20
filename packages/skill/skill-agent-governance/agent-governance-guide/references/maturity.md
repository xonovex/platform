# Derived Maturity

Maturity is an optional report over available capabilities. The caller supplies ordered names and requirements; the assessment returns the highest satisfied level and the missing capabilities for every level.

```json
{
  "name": "team-agent-maturity",
  "levels": [
    {"name": "A1", "requiredCapabilities": ["oversight:critique"]},
    {
      "name": "A2",
      "requiredCapabilities": ["oversight:critique", "oversight:approval"]
    },
    {
      "name": "A3",
      "requiredCapabilities": [
        "oversight:critique",
        "oversight:approval",
        "oversight:escalation"
      ]
    }
  ]
}
```

These names have no built-in meaning. Another organization can define different requirements or omit maturity entirely. Changing a reported level does not add a control, change a control from observe to enforce, select an executor, or move execution to a different host.

If a capability must be present for execution, also list it in the invocation's `requiredCapabilities`. Maturity assessment alone is descriptive.
