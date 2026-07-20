# Invocation, Execution, and Control

AGENTS.md, a skill, and a command are prose a model reads. A workflow runtime executes
plugins. Keep that difference explicit without turning a hook into a required architecture.

## First decide how work starts

- Use prose or a command when a person or agent may choose when to invoke the behavior.
- Use a trigger adapter when a native event must start it automatically.
- The trigger may be a harness hook, CI/CD hook, schedule, webhook, provider event, or any
  other authenticated source.

A hook is one possible trigger. It is not the executor and does not imply a control.

## Then select one executor

Choose the smallest executor that fits the work:

| Work                                                | Executor adapter          |
| --------------------------------------------------- | ------------------------- |
| Deterministic operation                             | Script                    |
| Deterministic collection plus one bounded inference | Script plus LLM           |
| Adaptive multi-step work using a specialized guide  | Agent plus workflow skill |
| Product-specific behavior                           | Custom plugin             |

This choice describes execution only. It does not add approval, evidence, escalation, or
maturity.

## Add controls only when requested

A selected control states:

- its registered plugin;
- whether it runs before, after, or both;
- whether its result is `observe` or `enforce`.

Automatic execution is not the same as enforcement. An observing control records a denial
without blocking. An enforcing before-control can prevent execution. An enforcing
after-control can fail the result but cannot undo an earlier side effect.

## Keep the remaining dimensions separate

- The host may be local, a harness, CI, Kubernetes, or another adapter.
- Each evidence sink independently chooses whether failure is ignored or fails the run.
- Maturity is an optional report from a caller-owned capability model. A1/A2/A3 have no
  built-in behavior.

## Do not overclaim prose or hooks

Prose raises the chance that a model follows a rule; it does not execute. A hook proves
only that its native handler ran on covered events. Before claiming a blocking guarantee,
verify the exact event, handler support, operation coverage, output behavior, ordering,
failure behavior, active configuration, and executable selected by the trusted template.
