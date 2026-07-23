# composition-metadata: Classify and Resolve Skills

Use one harness-neutral catalog to describe semantic composition. Agent Skill frontmatter owns discovery, plugin manifests own installation, and the composition catalog owns classification plus semantic contracts.

Catalog contract v2 also defines the broad-to-narrow preference-overlay precedence:
global, organization, repository, language, framework, path, then explicit request.
The resolver uses this declared order; prompt or load order has no authority.

## Two independent classifications

Every guide has exactly one lifecycle and one primary functional role:

| Axis            | Value            | Meaning                                                                                        |
| --------------- | ---------------- | ---------------------------------------------------------------------------------------------- |
| Lifecycle       | `durable`        | Knowledge, context, preferences, or communication rules intended to remain stable across tasks |
| Lifecycle       | `procedural`     | Capability instructions, procedures, assurance, or recovery behavior selected for a task       |
| Functional role | `domain`         | Meaning, entities, invariants, and business or technical domain facts                          |
| Functional role | `context`        | Repositories, services, environments, topology, and ownership                                  |
| Functional role | `preference`     | Explicit conventions and narrower-over-broader overrides                                       |
| Functional role | `procedure`      | Preconditions, steps, decisions, outputs, and completion conditions                            |
| Functional role | `capability-use` | Reliable use, limits, and verification of an available mechanism                               |
| Functional role | `assurance`      | Questions, criteria, evidence, and verification                                                |
| Functional role | `recovery`       | Retry boundaries, alternatives, rollback, and escalation                                       |
| Functional role | `communication`  | Audience, terminology, sections, evidence, and handoff format                                  |

`mixed` is not a classification. Classify the guide by its primary contract during migration, then split it when independent concerns need separate routing or selection. The lifecycle describes the guidance, not the package release lifetime.

## Catalog entry

Keep identity, classification, semantic provisions, and semantic requirements together:

```json
{
  "name": "example-guide",
  "classification": {
    "lifecycle": "procedural",
    "functionalRole": "procedure"
  },
  "provisions": [{"id": "method:example-review", "version": "1.0.0"}],
  "requirements": [
    {
      "id": "assurance:evidence",
      "range": "^1.0.0",
      "strength": "preferred",
      "reason": "The review benefits from specialist evidence checks."
    }
  ]
}
```

- **Name one primary contract** - Use the guide name, one lifecycle, and one functional role
- **Declare only stable provisions** - Add a provision only when another skill can rely on its named, versioned semantic contract
- **Keep requirements invariant** - Catalog requirements apply whenever the skill runs; derive conditional needs such as a selected perspective in the operation request instead
- **Explain every requirement** - State why it is needed, not merely which provider happens to satisfy it
- **Sort entries by name** - Keep inventory diffs deterministic and reviewable

Provision identifiers use `<namespace>:<kebab-name>`. Useful namespaces include `method`, `perspective`, `procedure`, `assurance`, and `recovery`. The namespace describes semantics, never authority or tool access.

## Three relationship types

Do not encode all composition as a dependency:

- **Advisory handoff** - Name another guide in prose; the current skill remains useful without it
- **Semantic requirement** - Request a provision by identifier and compatible contract range; the installed snapshot resolves the provider
- **Exact hard dependency** - Name the plugin in both manifests when the core workflow requires that exact implementation

An exact manifest dependency remains the installation contract. Do not duplicate it as a semantic requirement merely to make it visible in the catalog.

## Deterministic installed-snapshot resolution

Resolve without network lookup or a newest-version fallback:

1. Honor an explicit exact guide and implementation version.
2. Honor exact plugin dependencies already installed by the manifests.
3. Match accepted or inherited required semantic requirements against installed provisions.
4. Resolve preferred requirements after required ones.
5. Select only when exactly one installed provision satisfies the identifier and SemVer range.

No compatible provider is unavailable or incompatible. More than one compatible provider is ambiguous until an explicit selection or policy resolves it. A missing required provision blocks composition; a missing preferred provision remains visible as non-blocking degradation. Never substitute a similarly named skill.

## Version and compatibility boundaries

Keep three versions separate:

- **Catalog contract version** - The schema understood by catalog readers; an unsupported major version fails closed
- **Provision contract version** - The semantics supplied by one provision and matched by a requirement range
- **Implementation version** - The installed package/plugin version recorded for reproducibility

The free-text Agent Skill `compatibility` field still documents runtime, operating-system, network, or tool prerequisites. It is not a semantic provision range and must not drive automatic provider selection.

## Selection provenance

Every resolved selection records:

- guide and plugin identifiers;
- exact installed implementation version;
- catalog contract version and content digest;
- matched provision identifier, provision version, and requested range when semantic;
- selection kind: explicit, exact dependency, semantic requirement, or policy;
- requesting skill or operation and human-readable reason;
- package path and the selected guide's `SOURCES.md` path when the installed runtime
  exposes local paths.

`SOURCES.md` remains the content-provenance authority. Do not copy its URLs, commits, review dates, or source versions into the composition catalog. Durable domain or context conflicts require explicit authority, scope, freshness, and evidence resolution; preference overlay order does not resolve factual conflicts.

## Preference overlays

Only a guide classified as `preference` may be requested as an overlay. Each request
names the guide, target, scope kind/value, and reason. Resolve applicable overlays
from broad to narrow catalog precedence. Two different overlays for the same target
at the same scope are a conflict until policy or an explicit narrower scope selects
one. Never pass a domain or context guide through overlay resolution.

## Distribution

Repository validation uses one canonical composition catalog. Package an exact generated snapshot inside the workflow skill that performs semantic selection:

- Keep the snapshot byte-identical to the canonical file.
- After changing the canonical catalog, run `npx moon run script-moon-skill-validate:composition-sync`.
- Compute the catalog digest from those exact bytes.
- Make the workflow guidance load its packaged snapshot rather than reaching outside its installed plugin directory.
- Fail validation when the packaged snapshot is absent or differs, so it can never become a second authored source.

Do not assume that a marketplace repository root remains readable after one plugin is installed. Per-skill frontmatter is not the fallback: duplicating structured requirements and provisions into string-only metadata recreates drift across the entire catalog.

## Failure checks

- Reject missing, duplicate, unexpected, or unsorted catalog entries.
- Reject `mixed`, unknown roles, malformed identifiers, invalid SemVer, and invalid ranges.
- Reject duplicate provisions or requirements within one guide.
- Reject required requirements that are missing, incompatible, or ambiguous in the installed snapshot.
- Reject cycles formed by deterministically selected required semantic requirements.
- Preserve preferred failures as visible results rather than silently dropping them.
- Reject non-preference overlays and equal-scope overlay conflicts.
- Reject a missing or non-identical packaged workflow snapshot.
