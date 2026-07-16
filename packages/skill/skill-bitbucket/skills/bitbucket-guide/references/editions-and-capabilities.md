# Editions and Capabilities

## Detection

Resolve the canonical base URL and perform a read-only product probe. Record:

- Bitbucket Cloud workspace/account plan and repository UUID; or
- Bitbucket Data Center deployment version/build, project/repository keys, installed apps, authentication method, and cluster/base URL;
- effective actor, groups/permissions, repository settings, branch restrictions, checks/hooks, runners, and applicable organization controls;
- advertised REST/API versions and observation time.

If the probe cannot distinguish product or version, return `unknown`. Do not route Cloud operations to a self-managed host or Data Center operations to Cloud.

## Capability matrix

| Capability                          | Cloud baseline                              | Data Center 10.3 baseline                                            |
| ----------------------------------- | ------------------------------------------- | -------------------------------------------------------------------- |
| Repositories and pull requests      | candidate                                   | candidate                                                            |
| Branch/repository permissions       | Cloud branch restrictions/permissions       | versioned repository/project permissions                             |
| Build/deployment status             | native Cloud status/deployment records      | native status APIs; exact behavior version/app dependent             |
| Pipelines and shared configurations | candidate; plan/runner constraints          | unsupported unless a separately identified installed app provides it |
| OIDC for pipelines                  | candidate; claim/trust probe required       | unsupported by Cloud semantics                                       |
| Custom merge checks                 | candidate; plan/app callback probe required | do not infer; detect native/plugin policy separately                 |
| REST/webhooks                       | Cloud REST 2.0 and Cloud delivery semantics | pinned Data Center REST and deployment/plugin semantics              |

`candidate` means documentation-conformant, not live-passed. Return support state, product/version, plan/tier, installed dependencies, API, source snapshot, limitations, live-probe reference, and native reference kinds per capability.
