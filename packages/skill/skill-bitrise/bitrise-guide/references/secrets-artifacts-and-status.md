# Secrets, Artifacts, and Build Status

## Secrets

Discover workspace/app secret scopes, protected flags, pull-request and fork behavior, contributors with access, self-hosted runner exposure, external vault integrations, and every Workflow/Step that receives the value. Previews contain secret identifiers and flows, never values.

Separate validation of untrusted changes from secret-bearing builds. A fork or untrusted pull request defaults to no protected secrets. Rotation and revocation are explicit lifecycle operations with evidence.

## Artifacts

Record build/Step origin, source commit, artifact identity/type, content digest where available, access policy, retention/expiry, download/audit reference, and downstream deployment/status relationship. Do not treat a mutable URL or filename as immutable evidence.

Preview artifact volume, sensitive content, network destination, retention, access, and cost. Verify redaction and that forbidden files/secrets are absent.

## Git-provider status and deployment evidence

Bind the reported build status to the exact repository and commit. Preserve status key/context, provider/app identity, build URL/reference, conclusion, and update time. Test missing, renamed, spoofed, stale, duplicate, cancelled, timed-out, and failed status outcomes against the source host's merge policy.

Bitrise status publication is evidence, not the source host's merge authority. The repository provider owns whether a status is required and who may bypass it. Keep build, artifact, source-host status, deployment, and cloud-runtime evidence as separate related references.
