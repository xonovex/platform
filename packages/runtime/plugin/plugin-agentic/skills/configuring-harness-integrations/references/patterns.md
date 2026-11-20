# Portable Patterns

Each portable pattern maps to one native capability per harness; the harness skill holds that mapping and its caveats.

## Patterns

- Protected path, secret, or tool policy: a deterministic pre-tool handler that matches the real tool and input.
- Formatting: an idempotent post-tool handler for covered write tools.
- Validation before completion: a stop handler with a bounded continuation count.
- Audit: a post-event handler that minimizes transcript, prompt, and file content.
- Context injection: model-visible context, never evidence.
- Model evaluator: a bounded model decision with a closed output schema.
- Specialist agent: a bounded child run with depth, authority, budget, and cancellation limits.
- Organization-managed adoption: managed configuration plus separately pinned scripts, tested for startup and non-weakening of lower scopes.

## Executor classes

- Declare the executor class of every handler: deterministic command, model evaluator, or specialist agent.
- Do not hide a model evaluator inside a deterministic template, and do not call a command that runs a model or child agent deterministic.
- A preview states when model input leaves the process, which content is sent, and what happens on timeout or invalid output.
- Where a harness parses but skips a handler type, use an explicit bounded command runner and report the skip.
