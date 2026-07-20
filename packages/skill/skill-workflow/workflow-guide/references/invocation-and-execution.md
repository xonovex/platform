# Invocation and Execution

Invocation answers how a call starts. Execution answers who or what performs it. Both remain independent from the selected operation and from each other.

## Invocation

A person, harness hook, CI/CD event, webhook, or schedule may initiate the same operation contract. The trigger supplies context but does not add a lifecycle stage, change the operation, select a provider, or authorize a side effect.

## Execution

A human, deterministic tool, model, or agent may execute the call. Every executor must honor the same inputs, capability checks, provider boundary, output contract, and side-effect rules.

Executor labels or maturity assessments are metadata only. They never turn `review` into `validate`, make `execute` publish, give `decide` authority, or select required stages. Any control or enforcement belongs to the invoking host or provider and leaves operation semantics unchanged.
