# Moon Configuration

- Tasks inherit from `tasks/tag-*.yml` based on `tags` in `moon.yml`
- Tags: `cli`, `command`, `go`, `npm`, `shell`, `skill`, `tsconfig`, `typescript-config`, `typescript-script`, `typescript`
- `ci-check` gates only side-effect-free work: no network, no process or server spawning, and no writes outside a task's declared `outputs`. Anything else belongs to `ci-integration`, `ci-e2e*`, or `ci-publish`.
- `runInCI` is `true` only inside the `ci-check` dependency closure; every other task sets `runInCI: false`. Moon rejects a CI-enabled task that depends on a CI-disabled one (`run_in_ci_mismatch`), so a task pulled into `ci-check` must be enabled in the same change.
- `runInCI` defaults to `true`, and also suppresses `moon run` when `CI` is set, so a task left at the default runs in CI whether or not it is wanted there.
