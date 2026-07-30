# Assessment and Evidence

## Resolve the target before testing

Pin the standard, version, conformance level, organizational additions, exact subject revision, scope, complete processes, supported environments, dependencies, and applicable criteria. Record exclusions and unresolved applicability explicitly.

For each criterion, use one outcome:

- `pass`: sufficient current evidence supports the criterion for the declared scope;
- `fail`: a reproducible barrier or unmet requirement exists;
- `not-applicable`: the criterion does not apply, with a criterion-specific reason and review;
- `not-tested`: evidence is missing, stale, inconclusive, inaccessible, or outside the completed assessment.

Never turn missing evidence into `pass` or `not-applicable`.

## Compose independent evidence

Prefer the least adaptive reliable evaluator for each fact:

1. deterministically inspect authoritative source, configuration, structure, and exact revision;
2. run pinned automated rules and preserve tool, ruleset, environment, coverage, and raw native references;
3. exercise keyboard, focus, zoom/reflow, motion, timing, input, errors, status, and complete journeys where applicable;
4. test representative assistive-technology and platform combinations selected by the support profile;
5. use qualified human review for semantics, usability, context, applicability, and claims automation cannot establish.

Models do not invent passes, criteria, test execution, user evidence, or assessor authority.

## Make tests reproducible

Each test records subject revision, criterion, preconditions, data, environment, viewport or display settings, platform, browser or runtime, assistive technology, tool/ruleset version, action, expected outcome, observed outcome, evidence origin, time, and limitations.

Use representative fixtures without sensitive production data. Keep the value under test visible. A failure should name the user impact and reproducible path, not merely a tool rule identifier.

## Publish an assessment result

Publish through the selected native provider with:

- exact subject/scope, standard/version/level, applicable criteria, profile and evaluator versions;
- outcome per criterion, finding identity/severity, affected journey and user impact;
- deterministic, automated, human, and assistive-technology evidence origins;
- native references, timestamps, freshness, coverage, limitations, conflicts, and gaps;
- remediation and retest requirements plus any separate exception reference;
- assessor role, independence, qualification, and accountable decision boundary.

Preserve criterion-level outcomes and evidence, not only a summary count.

## Validate the assessment itself

Probe a known accessible case and a known failing case for each critical evaluator. Test scanner outage, unsupported content, inaccessible environment, stale revision, conflicting evaluator results, and evidence-provider failure. Mandatory profiles fail visibly when required evidence is absent or stale.
