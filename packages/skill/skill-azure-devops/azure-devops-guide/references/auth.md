# Azure DevOps CLI Authentication

Load **credential-management-guide** for provider-neutral storage, secret-manager retrieval, CI boundaries, rotation, and leak response. This reference owns Azure DevOps identity choices, PAT scopes, transport, and verification.

## Credential choice

Prefer the short-lived cache created by `az login` for interactive use and a managed identity, service principal, or workload-identity-federated service connection for automation. Use a PAT only when the operation cannot use those paths.

Create, rotate, and revoke a PAT at `https://dev.azure.com/<org>/_usersSettings/tokens`. Limit it to one organization, the shortest practical expiry, and the operations required: pull-request work generally needs Code read/write; work-item links need Work Items read/write. A PAT never exceeds its owner's access, and organization policy may prohibit PATs.

## Transport

With `az login`, `az repos` and `az boards` use the cached identity. The Azure DevOps extension reads a PAT from `AZURE_DEVOPS_EXT_PAT`; raw REST uses HTTP Basic auth with an empty username:

```bash
export AZURE_DEVOPS_EXT_PAT="$(<provider-neutral-secret-store-read>)"
az repos pr list --org https://dev.azure.com/<org> --project <project>

PAT="$(<provider-neutral-secret-store-read>)"
curl -s -u ":$PAT" "https://dev.azure.com/<org>/_apis/projects?api-version=7.1"
```

Replace the placeholders with a retrieval command selected by **credential-management-guide**; never substitute a literal token.

## Verification

```bash
az devops project list --org https://dev.azure.com/<org> >/dev/null
curl -s -o /dev/null -w '%{http_code}\n' -u ":$PAT" \
  "https://dev.azure.com/<org>/_apis/connectionData?api-version=7.1"
```

Azure DevOps may return `203` or redirect to an HTML sign-in page for an unauthenticated request. Validate the content type as well as status. `401` indicates missing or invalid credentials; `403` indicates insufficient permission; `404` often indicates a wrong organization, project, or repository.

## Automation

For Azure Pipelines, prefer an `AzureCLI@2` task backed by a workload-identity-federated service connection. For another CI provider, bind its OIDC claims to an Entra federated credential. A PAT fallback belongs in a protected CI variable or Key Vault-backed variable group and is exported only for the consuming step.
