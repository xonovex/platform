# Sources

## OWASP Secrets Management

- **URLs:**
  - https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html
  - https://cheatsheetseries.owasp.org/cheatsheets/CI_CD_Security_Cheat_Sheet.html
- **Last reviewed:** 2026-07-22
- **Used for:** `SKILL.md`, `references/credential-selection.md`, `references/automation-and-response.md`
- **References:** references/credential-selection.md, references/automation-and-response.md
- **Aspects extracted:** Secret lifecycle, least privilege, CI/CD isolation, dynamic creation, rotation, revocation, auditing, and exposure response.

## Platform secret stores and CI federation

- **URLs:**
  - https://specifications.freedesktop.org/secret-service/latest/
  - https://learn.microsoft.com/en-us/powershell/utility-modules/secretmanagement/get-started/using-secretstore?view=ps-modules
  - https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets
  - https://docs.github.com/en/actions/how-tos/secure-your-work/security-harden-deployments/oidc-with-reusable-workflows
- **Last reviewed:** 2026-07-22
- **Used for:** `references/local-storage.md`, `references/automation-and-response.md`
- **References:** references/local-storage.md, references/automation-and-response.md
- **Aspects extracted:** Secret Service storage, PowerShell vault behavior, CI secret injection, OIDC job identity, and trust-claim constraints.

## Guide-level synthesis

- **Provenance:** Repository-original provider-neutral lifecycle and ownership synthesis based on the sources above
- **Last reviewed:** 2026-07-22
- **Used for:** all
- **References:** all
- **Aspects extracted:** Decision order, provider ownership boundary, retrieval discipline, rotation sequence, and containment-first incident workflow.
