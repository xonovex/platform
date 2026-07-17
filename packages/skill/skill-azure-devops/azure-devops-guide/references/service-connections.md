# Service Connections

## Default identity design

Prefer workload identity federation or a managed identity path that issues temporary Azure credentials. Scope the service connection to the selected project and target resources, constrain the federated subject, and grant the minimum Azure role at the narrowest resource scope. Do not grant every pipeline access by default.

Avoid creating client secrets, certificates, or AWS access keys as an onboarding shortcut. If the selected Server/Services edition or target cannot federate, return the limitation and require an explicit alternative preview with credential owner, storage, rotation, expiry, revocation, and blast radius.

## Preview

Show tenant, issuer, audience, subject, application/managed identity, service-connection ID/name, project authorization, target subscription/resource scope, Azure role assignments, pipeline authorization, network destinations, secrets created or read, and exact rollback order. Redact credential values.

## Verification

1. Re-read service-connection type, authorization scope, readiness, creator/owner, and pipeline grants.
2. Run a least-privilege identity probe and record the temporary credential expiry without printing the token.
3. Prove an allowed action succeeds and an out-of-scope action fails.
4. Resolve provider-native pipeline, role-assignment/activity, and target-resource evidence.
5. Verify rollback removes only owned federation, grants, and connection resources without erasing retained evidence.

Treat service-connection authorization and Azure RBAC as independent controls; success in one does not prove the other is least privilege.
