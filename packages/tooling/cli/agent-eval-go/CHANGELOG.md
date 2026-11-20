# @xonovex/agent-eval-go

## 0.1.0

### Minor Changes

- first release

### Notes

- No model id ships. Without `--model`, a profile model or a model variable, the harness runs with its own default, and the report says `harness default`.
- The `sandbox` member of the `eval` section, or the `--sandbox*` flags, runs every harness inside an agent-cli isolation (`none`, `bwrap`, `docker` or `seatbelt`). A docker sandbox needs an image: set `eval.sandbox.image` in a user or managed `agentic.json`, or pass `--sandbox-image`.
- The `eval` section of a project `agentic.json` takes effect only after you approve its digest, at the terminal prompt or with `--approve-config-digest`.
