# Capability Execution Contract

## Executor classes

| Class           | Use                                                                    | Authority and evidence rule                                                                                 |
| --------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `deterministic` | Script, API call, scanner, provider operation, reproducible inspection | Preferred for authoritative facts and privileged actions; publish command/input version and native evidence |
| `model`         | Bounded interpretation or classification with fixed context            | Validate structured output; model output is not authoritative when deterministic/external evidence exists   |
| `agent`         | Adaptive multi-step investigation requiring branching tool use         | Explicit bounded launcher with attenuated authority and result boundary                                     |
| `human`         | Accountable judgment, approval, review, or manual action               | Record identity, role, independence, scope, decision, and native evidence                                   |
| `external`      | CI, deployment, identity, monitoring, GRC, or another system of record | Preserve native subject, revision, policy/version, decision, and freshness                                  |

The identity, role, and independence the `human` row records are defined in [actors.md](actors.md), which also states which of them code enforces.

## Workflow execution families

The runtime exposes three concrete compositions. Trigger source is orthogonal to all three.

| Family                 | Runtime order                                                                                                       | Boundary                                                                                           |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| `workflow-script`      | Run one registered deterministic command                                                                            | Timeout, closed JSON result, durable execution evidence                                            |
| `workflow-script-llm`  | Establish deterministic facts, then pass only that closed result to a registered model evaluator                    | Overall timeout, retry ceiling, token/cost budgets, closed model result                            |
| `agent-workflow-skill` | Verify observed `A1`/`A2`/`A3` oversight, then launch a registered agent with one explicit workflow-skill reference | Overall timeout, retry/child-depth ceilings, token/cost budgets, cancellation, closed agent result |

Registered commands execute directly with `shell: false`, minimized environment, bounded output, abort propagation, and terminate/kill behavior. Success and failure both produce content-addressed JSONL evidence bound to the trigger reference and exact subject revision.

A workflow script is executable deterministic software with a versioned input/output contract. A prompt or slash-command file is harness context: it may instruct a model, but it is not a deterministic workflow script merely because a hook or CI job invokes it. Likewise, a workflow skill constrains an adaptive agent; the skill is not itself the trigger.

## Selecting a class

`deterministic`, `model`, and `agent` answer how much adaptivity the work itself needs, so a work shape selects among them — see **workflow-guide** for the work-shape literals and the classification procedure. `human` and `external` answer who must own the outcome: select them when accountable judgment or a system of record holds the authority, never because the work is hard. They are orthogonal to work shape, so no work shape selects them.

Select deterministic execution whenever authoritative inspection can establish the fact. A model or agent may interpret evidence, identify candidates, or investigate gaps, but cannot silently replace the authoritative check.

## Executor requests

An executor request — an `--executor` argument, profile field, or equivalent — is a requested ceiling, not a decision. Resolve its value through the caller's axis precedence, which the calling lifecycle guide states, then validate the resolved request against:

- **least-adaptive** — reject a `model` or `agent` request where a `deterministic` executor can authoritatively establish the same result;
- **deterministic-authority** — never treat `model` or `agent` output as authoritative where a deterministic check or an external system of record establishes the fact.

A request only narrows or confirms what work shape and these rules already permit; it never widens them. Where a narrower request cannot complete the work, reject the request rather than escalate to a more adaptive class. Record requested and effective class whenever they differ.

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
