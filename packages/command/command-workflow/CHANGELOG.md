# @xonovex/command-workflow

## 7.0.0 - 2026-07-20

### Breaking changes

- Replaced the former lifecycle- and artifact-specific command catalog with eight
  composable operations and four explicit workspace utilities. Removed command
  aliases are not retained.
- Removed profile, approval-gate, role-specific API, central-reference, and implicit
  trigger/executor semantics from the command contract.
- Removed the operator-owned `AgentTrigger` and `AgentSchedule` Kubernetes APIs.
  External schedulers, webhook handlers, CI/CD systems, and manual tools now submit
  `AgentRun` resources directly.
