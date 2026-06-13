# moon-skill-eval-outputs

OUTPUT eval (ported from `eval-outputs.py`): runs each eval twice — vanilla vs skill-augmented — in isolated `claude -p` contexts, grades both with a binary reference-guided LLM-as-judge, and writes a `benchmark.json` with pass-rate / token / duration deltas.

## Usage

```bash
npx moon-skill-eval-outputs [evals.json] [skill-name] [iteration] [--runs N] [--concurrency N] [--model M] [--judge-model M]
# evals defaults to ./evals.json; skill-name defaults to the name in ./SKILL.md
```

Requires the `claude` CLI on PATH.
