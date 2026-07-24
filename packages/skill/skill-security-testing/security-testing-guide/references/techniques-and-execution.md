# Techniques and Execution

## Balanced Techniques

No single technique covers security:

- review requirements, threat model, architecture, roles, data flows, and deployment;
- inspect source, configuration, infrastructure, permissions, secrets, and build inputs;
- scan dependencies and artifacts while validating reachability and actual exposure;
- test authentication, authorization, session, validation, cryptography, business
  logic, error handling, logging, privacy, and availability controls dynamically;
- use fuzzing, property, fault, and abuse testing for state and input uncertainty;
- use manual penetration testing for chained, contextual, and business-logic failures.

Map each technique to the requirement or threat it can observe and document blind
spots. Prefer versioned ASVS or WSTG references for web coverage rather than unstable
unversioned identifiers.

## Safe Execution

Pin tool, ruleset, target, build, configuration, account role, payload, and time.
Monitor health and logs during active tests. Enforce rate and concurrency limits.
Avoid persistence, destructive modification, broad exfiltration, social engineering,
or third-party targeting unless each is separately authorized.

Validate automated results manually or through an independent signal before reporting
them as vulnerabilities. Preserve enough redacted evidence to reproduce the result
without retaining unnecessary secrets or personal data.
