# Workflows, Pipelines, and Steps

## Composition

Treat each native layer separately:

- a **Step** is one versioned executable unit with inputs, outputs, source/provenance, tools, filesystem/network/secret/data access, side effects, timeout, and failure behavior;
- a **Workflow** orders Steps and may reuse other Workflows;
- a **Pipeline** orchestrates Workflows and stage dependencies;
- a **stack/runner** supplies the execution image, machine boundary, toolchain, caches, and network.

Detect the workspace plan, app configuration source, stack, runner type, and supported Pipeline features before proposing composition.

## Pinning and trust

Select exact reviewed Step versions and record source/publisher/verification status separately. For custom Steps, pin the repository revision and dependency/image versions. A Verified Step reduces publisher ambiguity but does not grant trust, least privilege, or compatibility.

Prevent untrusted code from altering secret-bearing or target-changing Workflow/Step inputs. Bound time, concurrency, retries, artifacts, and network destinations. Review caches and shared runner state for cross-build leakage.

## Evidence

Preserve app slug, pipeline/workflow IDs, build slug/number, source commit, trigger, stack/runner, Step IDs/versions/outcomes, timing, artifact references/digests, status/deployment references, and external target evidence. Logs remain access-controlled native references and are not copied by default.

Verify the effective configuration and executed Step versions from the build, not only the intended `bitrise.yml`.
