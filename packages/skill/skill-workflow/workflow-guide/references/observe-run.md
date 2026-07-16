# observe-run: Publish Operational Observation

## Core workflow

1. Resolve the exact released/deployed subject or other observable resource, environment,
   time window, baselines, service/user objectives, data classifications, and current
   monitoring policy.
2. Collect provider-native signals from system monitoring, users and support, security,
   AI/model behavior, cost/resource use, accessibility, and delivery/outcome measures when
   applicable. Preserve source, revision/configuration, query, window, freshness, sampling,
   redaction, and limitations.
3. Correlate signals without inventing a workflow identity or copying sensitive raw data
   unnecessarily. Bounded agents/models may cluster and summarize only with cited source
   evidence and advisory origin.
4. Compare observations to pinned baselines and thresholds. Distinguish finding, anomaly,
   regression, unknown, missing telemetry, stale evidence, and changed instrumentation.
5. Escalate through the selected provider when an Incident, Assessment, Transition,
   corrective action, or new Development result is warranted; do not let an agent suppress
   a required human/external notification.
6. Publish an Observation result with exact subject/environment, window, signals, baselines,
   findings, evidence freshness, gaps, privacy/retention handling, and follow-up references.
