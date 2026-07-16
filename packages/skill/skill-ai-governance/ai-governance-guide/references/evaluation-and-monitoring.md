# Evaluation, Monitoring, and Learning

## Define evaluation before running it

Pin the exact system revision, intended use, risk/control version, evaluator and dataset versions, environments, scenarios, thresholds, statistical method, independence, evidence origin, freshness, limitations, and failure behavior.

Select contextual measures across validity, accuracy, calibration, robustness, harmful outcomes, subgroup and edge-case performance, data leakage, privacy, cybersecurity, prompt injection, tool abuse, excessive agency, refusal and over-refusal, content safety, human oversight, accessibility, resource/cost controls, latency, and recovery as applicable.

Aggregate scores never erase scenario-level failures. State uncertainty, confidence, sample limitations, evaluator disagreement, distribution gaps, and untested risks.

## Exercise controls adversarially

Test intended allows and denials, malformed and adversarial input, indirect prompt injection, poisoned retrieval, tool-result injection, privilege expansion, unauthorized data flow, secret canaries, duplicate and concurrent execution, timeout, provider outage, stale versions, fallback routing, telemetry outage, human intervention, rollback, and emergency disable.

Use independent red-team or qualified assessment where impact requires it. Preserve the exact finding, reproduction, evidence, mitigation, residual risk, and retest result.

## Monitor production with bounded data collection

Monitor exact versions, input/output quality signals, harmful outcomes, overrides, appeals, user friction, denied actions, policy bypasses, data and concept drift, provider changes, cost/resource use, latency, errors, accessibility, control coverage, incident indicators, and evidence freshness.

Collect metadata by default. Raw prompts, outputs, customer content, source, secrets, personal data, or sensitive attributes require explicit purpose, authority, minimization, redaction, access, retention, residency, deletion, and external-transfer rules.

## Respond and reassess

Define thresholds for observe, investigate, constrain, fall back, disable, roll back, notify, declare an incident, and reassess. Preserve detection, containment, recovery, impact, notification, corrective action, and effectiveness evidence with accountable owners. Legal notification scope and timelines require current-text and qualified review.

## Promote learning through change control

Extract candidate lessons from evaluations, monitoring, incidents, overrides, appeals, denials, exceptions, drift, onboarding, and module failures. Each candidate records source evidence, scope, owner, proposed target, expected outcome, risks, privacy limits, conflicts, and expiry.

Promotion is reviewed, versioned, canaried where executable, authorized, measurable, reversible, and removable. Never auto-promote a prompt, policy, evaluator, threshold, model route, or global instruction directly from one incident or metric.
