# Service and Control Objectives

## Define outcomes before indicators

Name the user, operator, control, or dependent-system outcome that must hold. Examples include completing a journey, receiving a correct authorized decision, applying a configuration exactly once, producing fresh evidence, recovering data, or safely denying a privileged request.

For each outcome record scope, criticality, exact service/control version, owners, consumers, dependencies, assumptions, indicator source, target, measurement window, exclusions, freshness, uncertainty, response threshold, and review date.

## Balance the dimensions

Select only dimensions that affect the outcome:

- availability and reachability;
- correctness, integrity, authorization, and policy freshness;
- latency, deadlines, queue time, and completion time;
- durability, data loss, evidence retention, and recovery;
- capacity, saturation, rate limits, cost, and resource budgets;
- security, privacy, accessibility, safe failure, and user friction;
- dependency health, version compatibility, drift, and evidence gaps.

An `allow` decision from stale policy or an unverified configuration mutation is not success merely because the service returned quickly.

## Set objectives and budgets

Choose targets from user impact, risk appetite, contractual needs, recovery capability, cost, and dependency limits. An error budget is a decision aid, not permission to harm users or bypass mandatory security, privacy, safety, accessibility, or legal controls.

Define what consumes the budget, how partial/degraded outcomes count, when releases slow or stop, who decides, and how emergency work is handled. Keep planned exclusions narrow and auditable.

## Validate measurements

Test known healthy, known failing, missing-data, delayed, duplicated, mis-correlated, and telemetry-outage cases. Preserve source versions, sampling, aggregation, retention, access, and known blind spots. A missing event is inconclusive unless the collection guarantee proves it should exist.

## Avoid surveillance and gaming

Measure system and team outcomes, not individual productivity. State purpose, audience, aggregation, minimum cohort, access, retention, deletion, privacy review, interpretation limits, and prohibited uses. Pair leading and lagging indicators plus qualitative user/operator evidence so one number cannot dominate behavior.
