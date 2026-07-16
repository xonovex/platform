# Sources

## NIST SP 800-53 Rev. 5, Release 5.2.0

- **URL:** https://csrc.nist.gov/pubs/sp/800/53/r5/upd1/final
- **Last reviewed:** 2026-07-16
- **Used for:** `SKILL.md`, `references/risk-and-controls.md`, `references/verification-and-operations.md`
- **Aspects extracted:** Flexible security/privacy control catalog, assurance and functionality perspectives, control families, tailoring, and explicit caution that mappings and crosswalks are not necessarily one-to-one or equivalent.

## NIST Cybersecurity Framework 2.0

- **URL:** https://www.nist.gov/cyberframework
- **Last reviewed:** 2026-07-16
- **Used for:** `references/risk-and-controls.md`, `references/verification-and-operations.md`
- **Aspects extracted:** Organization-wide cybersecurity risk outcomes, profiles, governance, and informative references.

## Secure Software Development Framework 1.1

- **URL:** https://csrc.nist.gov/pubs/sp/800/218/final
- **Last reviewed:** 2026-07-16
- **Used for:** `references/secure-development.md`, `references/supply-chain.md`
- **Aspects extracted:** Secure-development practices across preparation, protection, production, and vulnerability response.

## SLSA specification 1.2

- **URL:** https://slsa.dev/spec/v1.2/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/supply-chain.md`
- **Aspects extracted:** Build and source provenance concepts, tracks, levels, verification, and threat-oriented supply-chain assurance. Profile selection remains contextual.

## Sigstore documentation

- **URL:** https://docs.sigstore.dev/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/supply-chain.md`
- **Aspects extracted:** Artifact signing, identity-bound verification, transparency, and trust-policy inputs. No specific implementation is mandatory.

## Refresh Workflow

1. Re-check selected framework, control-catalog, secure-development, provenance, and signature versions and their current authoritative texts.
2. Re-run exact-subject, intended-denial, provenance mismatch, permission expansion, outage, concurrency, exception-abuse, rollback, incident, and retirement cases.
3. Keep control mappings, verification evidence, risk acceptance, certification, and legal compliance distinct and update **Last reviewed**.
