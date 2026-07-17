# retirement-run: Retire Lifecycle Resources Safely

## Core workflow

1. Resolve the exact subjects, owners, consumers, dependencies, current revisions, data and
   evidence retention duties, legal holds, replacement/migration state, residual risk, and
   requested retirement scope. Subjects may include models, datasets, credentials, features,
   APIs, infrastructure, dependencies, policies, modules, or provider configurations.
2. Plan archive, export, migration, disablement, revocation, deletion, communication,
   observation, rollback, and irreversible steps per resource. Require explicit disposition
   for every dependency, consumer, secret, data copy, backup, replica, and provider route.
3. Apply [operational-contracts.md](operational-contracts.md) before target-changing or
   irreversible work. Use the protected provider capability with exact authorization,
   external enforcement, least privilege, and separate exception/emergency-exception
   evidence.
4. Execute in dependency-safe stages. Record native mutation, deletion, revocation,
   migration, archive, notification, and failure references. Never infer deletion from a
   missing local file or a successful request alone.
5. Verify each selected resource through authoritative provider state, consumer and route
   checks, credential denial, data deletion/retention evidence, replacement health, and the
   declared observation window. Preserve required evidence without retaining prohibited
   content.
6. Publish a Retirement result with exact scope, authorization, actions, per-resource
   verification, residual risk, exceptions, irreversible outcomes, and follow-up. Partial
   retirement remains open and visible; rollback or recreation is a new authorized action.
