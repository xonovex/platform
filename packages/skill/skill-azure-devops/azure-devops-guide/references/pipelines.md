# Pipelines, Templates, Checks, and Artifacts

## Reuse and pinning

Prefer a repository-owned template with typed parameters and a protected immutable ref or otherwise change-controlled version. Use `extends` when the platform must constrain pipeline structure; disclose template-repository authorization and required-template approval behavior. Never pass secrets through template parameters or echo expanded configuration.

## Protected resources

Approvals and checks attach to resources such as environments, service connections, repositories, variable groups, secure files, or agent pools. Pipeline authors do not own those checks. Discover resource ownership, check order/category, approvers, self-approval, timeout, retry, exclusive lock, branch control, required template, business hours, and external checks before claiming a gate.

## Evidence

Preserve pipeline/definition ID, run/build ID, source commit, requested actor, timeline/stage/job/task records, check and approval references, artifact name/resource ID and content digest where available, environment/deployment record, and logs only by access-controlled reference.

## Transaction

Preview the template ref, YAML/resource changes, permissions, identities, variables/secrets, network destinations, expected checks, artifact retention/access, failure behavior, verification, and rollback. After apply, queue a pinned test revision, resolve its native check/build/artifact references, and run a negative probe that omits or violates the required control.

Skipped, cancelled, timed-out, partially succeeded, stale, missing, or similarly named runs are not success unless an explicit advisory policy says so. A mandatory gate fails closed on unavailable or unverified evidence.
