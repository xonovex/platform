# Scope and Model

1. Pin the subject and exact revision or state that it is unversioned.
2. Select concrete usage and administrative scenarios; avoid modeling an undefined
   whole enterprise at once.
3. Identify assets and security objectives: confidentiality, integrity, availability,
   authenticity, authorization, accountability, privacy, safety, and recoverability as
   applicable.
4. Inventory actors, processes, data stores, external systems, channels, protocols,
   identities, credentials, privileges, secrets, and operational controls.
5. Draw or describe data flows and state changes across trust boundaries.
6. Mark entry points, privileged transitions, parsing or deserialization, external
   effects, sensitive storage, logging, and recovery paths.
7. Record assumptions, inherited controls, dependencies, environments, exclusions, and
   evidence references.
8. Walk at least one normal, error, retry, administrative, and recovery scenario
   through the model.

The model must be detailed enough that another reviewer can locate where a threat acts
and which component can mitigate it. Do not infer a provider's internal boundary from
its URL or product name.
