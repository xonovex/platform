# Validation

Validate each boundary independently.

| Area                      | Required coverage                                                                          | Executable evidence                              |
| ------------------------- | ------------------------------------------------------------------------------------------ | ------------------------------------------------ |
| Workflow contracts        | Open result kinds, opaque handles, provider capabilities, explicit errors                  | Skill validation                                 |
| Composition runtime       | Schema, registry resolution, required capabilities, before/after ordering, observe/enforce | `shared-workflow-runtime:test`                   |
| Executor adapters         | Script, script-LLM, agent-skill, timeouts, output bounds, failure propagation              | `shared-workflow-runtime:test`                   |
| Trigger adapters          | Open trigger kinds, template binding, untrusted data isolation                             | `shared-workflow-runtime:test` and harness tests |
| Evidence sinks            | Ignore/fail selection and reference publication                                            | `shared-workflow-runtime:test`                   |
| Maturity                  | Caller-defined levels and no execution side effects                                        | `shared-workflow-runtime:test`                   |
| Kubernetes host           | API, admission, reconciliation, CRDs, deployment rendering                                 | operator Go tests and Kustomize build            |
| Documentation and release | Links, manifests, formatting, package lock, deleted-path checks                            | command and release validation                   |

Useful checks:

```bash
npx moon run shared-workflow-runtime:test command-workflow:test
npx moon run command-workflow:cross-package-links
npx moon run agent-operator-go:test
npm run fmt:check
```

Passing a skill or documentation test proves only that repository contracts agree. It does
not prove that a hook, control, evidence sink, or platform restriction is active in a
particular environment.
