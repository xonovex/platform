# Integration Lifecycle

Each harness fills these four phases with its own commands, paths and probes; the order, the evidence and the stop conditions stay the same.

## Discover

- Probe the runtime version; record `not-installed` when the binary is absent.
- Inventory every configuration layer, plugin, hook or extension source, trust decision, and enabled state that can affect the requested behavior.
- Resolve each requested behavior and guarantee against the capability matrix before recommending a handler.
- Review every command, URL, tool, prompt, agent definition, environment variable, transcript or path input, and executable source.

## Preview

- Show the exact merge into the selected layer; never overwrite the whole file, and keep unrelated entries.
- State the source scope, source version or digest, handler type, event and matcher, command and arguments, timeout, and concurrency.
- State the permissions, files, network, secrets, and content exposure, including what model input leaves the process.
- State the expected evidence, the verification probes, the disable step, and the rollback step.
- Pin a plugin or package preview to its source and version, and review its code without loading it.

## Apply and verify

- Apply only after explicit authorization for the exact diff and the authority it requests, through the native mechanism, and idempotently.
- Reload or restart where the harness requires it, then verify source attribution, trust, and enabled state with the harness's own diagnostics.
- Probe allow, deny, context output, malformed output, timeout, and concurrent siblings; add the probes the capability matrix flags for the harness.
- Never present a result from one surface as evidence for another.

## Lifecycle

- Disable only the owned entry through the native switch, or remove it, and verify it no longer fires.
- Roll back to the recorded configuration digest, and any separately delivered script digest, then rerun the probes.
- Treat an update as a fresh preview and review.
- Treat runtime version, schema, source precedence, command digest, permissions, data flow, and observed event behavior as drift dimensions.
