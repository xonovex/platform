# Harness Capabilities

A harness hook is one trigger adapter and, when the native surface can block, one possible
enforcement point. It is not a required runtime layer.

For each harness integration, record:

- native event and normalized trigger kind;
- input authentication and data minimization;
- whether the event can block, only observe, or inject context;
- ordering, concurrency, timeout, and retry behavior;
- exact executable and trusted template selected;
- unsupported operations and tested product version.

The bundled Claude pre-tool-use adapter is deliberately thin. It forwards native standard
input to the exact executable named by `XONOVEX_WORKFLOW_HOOK_EXECUTABLE`. It does not
select a policy, control mode, evidence sink, maturity level, or workflow template.

Keep control configuration in trusted composition wiring. Do not allow an untrusted hook
payload to choose an executable or turn an observing control into enforcement.
