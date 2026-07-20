# Host and Platform Adapters

Local processes, harnesses, CI runners, Kubernetes, and provider automation are peer hosts.
Each adapter owns its native authentication, configuration, isolation, retries, and
operational evidence.

Before enabling an adapter, inspect the exact executable or image, permissions, secrets,
network access, filesystem access, native event mapping, failure behavior, and rollback.
Pin artifacts where the host requires reproducibility.

Host restrictions stay host-local. For example, the Kubernetes operator can require a
digest-pinned image and constrain runtime classes through an optional `AgentPolicy`.
Those constraints do not create workflow controls or imply maturity.

CI required checks, repository rules, deployment approvals, admission controls, and cloud
identity policy may be selected as native controls. Register and select each independently;
do not convert platform presence into a universal governance profile.
