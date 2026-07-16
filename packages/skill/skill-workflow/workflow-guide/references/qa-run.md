# qa-run: Validate an Exact Deliverable

## Core workflow

1. Resolve the Deliverable Publication reference/revision, source revisions, intended target, selected test scopes, expected behavior, profile requirements, and required environments.
2. Record each environment identity and relevant operating system, runtime, dependency, configuration, data, tool, scanner, harness, and external-service versions.
3. Load the applicable testing, accessibility, security, performance, AI, supply-chain, platform, and domain skills. Prefer deterministic tests and scanners; bound any model or agent investigation and keep its evidence origin explicit.
4. Run independent suites according to declared concurrency, isolation, timeout, retry, cancellation, and partial-failure rules. Preserve each native CI/job/report reference and do not coerce skipped, neutral, stale, flaky, or timed-out work to success.
5. Re-resolve the exact subject and freshness bindings before disposition. A changed subject, required environment, evaluator, policy, or criteria version invalidates the affected evidence.
6. Publish a QA result with exact subject revision, environments, test scope, results, defects, coverage or assurance gaps, evidence origins, failure behavior, and follow-up capabilities.

QA may consume Review or Assessment findings and may trigger new Development work, but it retains its own result and publication boundary. Tests demonstrate the declared scope only; they do not grant Acceptance or prove an untested control.
