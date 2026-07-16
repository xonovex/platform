# Capability Execution Contract

## Executor classes

| Class           | Use                                                                    | Authority and evidence rule                                                                                 |
| --------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `deterministic` | Script, API call, scanner, provider operation, reproducible inspection | Preferred for authoritative facts and privileged actions; publish command/input version and native evidence |
| `model`         | Bounded interpretation or classification with fixed context            | Validate structured output; model output is not authoritative when deterministic/external evidence exists   |
| `agent`         | Adaptive multi-step investigation requiring branching tool use         | Explicit bounded launcher with attenuated authority and result boundary                                     |
| `human`         | Accountable judgment, approval, review, or manual action               | Record identity, role, independence, scope, decision, and native evidence                                   |
| `external`      | CI, deployment, identity, monitoring, GRC, or another system of record | Preserve native subject, revision, policy/version, decision, and freshness                                  |

## Base contract

Every capability execution declares:

- identity, version, purpose, and permitted/preferred executor classes;
- semantic input references and expected output/result contract;
- side effects, idempotency, reentrancy, and privileged-operation status;
- requested and effective tool, filesystem, network, secret, model, provider, and data authority;
- input validation, output validation, authoritative evidence origin, and freshness;
- timeout, retry, concurrency, ordering, cancellation, and kill behavior;
- failure behavior: deny, ask, advise, observe, or explicitly ignore;
- telemetry and audit references with sensitivity, redaction, retention, and access rules.

Select deterministic execution whenever authoritative inspection can establish the fact. A model or agent may interpret evidence, identify candidates, or investigate gaps, but cannot silently replace the authoritative check.

## Bounded model execution

Declare fixed input scope, model/provider constraints, data classification, structured output, validator, retry ceiling, time/token/cost budget, failure behavior, and evidence that distinguishes model inference from source facts.

## Bounded agent launch

Require all of:

- explicit purpose and result contract;
- maximum child depth and recursion behavior;
- selected model/provider and token, cost, and elapsed-time budgets;
- allowed tools, filesystem roots, network destinations, secrets, and data classes;
- authority attenuation: effective authority is no greater than the launcher and normally less;
- validation, evidence, failure/partial-result behavior, cancellation, and kill switch.

A child cannot broaden its own scope, launch beyond maximum depth, or treat inherited ambient access as authorization.
