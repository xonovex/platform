# Executable Module Trust and Provenance

An executable module is software running with agent, user, repository, organization, CI, or provider authority. Installation, discovery, or a valid manifest does not grant trust.

## Trust decision

Before load or execution, compare declared and observed source, version, compatibility, permissions, tools, filesystem, network, secrets, data flows, side effects, and lifecycle behavior.

- **Organization-managed** — require organization-controlled provenance, version pinning, change control, and verified distribution. User consent is not a substitute for managed provenance, and local configuration cannot silently weaken the module.
- **Project/repository** — require repository trust plus review of executable content and requested authority at the pinned revision. A trusted repository does not automatically approve newly expanded permissions.
- **User** — require informed user consent for the exact source/version and authority request. Consent does not override mandatory organization controls.
- **Session/runtime** — grants expire with the runtime and cannot establish durable trust or higher authority.
- **External** — accept evidence only for the provider's declared subject, scope, version, and authority.

Re-review when source, version, digest, signature identity, compatibility, permissions, data flow, side effects, failure behavior, or ownership changes.

## Provenance verification

Select one or more mechanisms appropriate to the package and threat model:

- package-manager metadata and locked package identity;
- cryptographic signature plus signer/identity policy;
- checksum or content digest from an independently trusted channel;
- SLSA provenance or another attestation with verified subject and builder claims;
- provider-native immutable revision, verified publisher record, transparency entry, or protected release evidence.

No mechanism is universally mandatory. Record the chosen method, trusted root/policy, verified subject and digest, result, limitations, and verification time. A moving tag, branch, unbounded version range, missing subject binding, failed verification, or mismatched digest fails before activation.

## Runtime and lifecycle probes

- Compare observed permissions and network/secret/data access with the approved declaration; stop on expansion.
- Inject concurrent duplicate invocation and retry after timeout. Side-effecting modules must be idempotent or reject/reconcile duplicates explicitly.
- Probe ordering and reentrancy without assuming serial hooks.
- Force timeout, partial application, provider outage, and rollback failure; verify the declared fail-closed, fail-visible, or advisory behavior.
- Preview and pin upgrades, re-run trust and capability probes, preserve a verified rollback target, and verify disable/removal without erasing retained evidence.
