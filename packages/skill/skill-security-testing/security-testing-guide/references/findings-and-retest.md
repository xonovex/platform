# Findings and Retest

## Finding Contract

```text
Identifier and status:
Affected subject, component, and exact revision:
Requirement or threat:
Preconditions and required access:
Minimal reproduction:
Expected and observed behavior:
Affected asset and realistic impact:
Exposure, likelihood, and existing controls:
Severity rationale and confidence:
Redacted evidence:
Remediation options and tradeoffs:
Retest scope:
Disclosure and owner:
```

Separate confirmed vulnerabilities, likely findings needing validation, informational
weaknesses, false positives, and accepted residual risk. Deduplicate by root cause
without hiding distinct affected assets or authorization paths.

## Remediation and Retest

Prefer fixing the root cause and adding durable prevention and detection. Do not expose
secrets, exploit details, or sensitive evidence beyond authorized recipients.

Retest the exact remediation revision using the original reproduction, adjacent bypass
paths, negative cases, and relevant regression checks. A changed response code alone
does not prove the control. Record fixed, partially fixed, not fixed, not reproducible,
or risk accepted with new evidence and limitations.

Only the accountable owner accepts residual risk or approves release. Preserve the
provider-native finding, decision, and remediation references.
