# moon-skill-eval-outputs

OUTPUT eval runs each eval twice — vanilla vs skill-augmented — in isolated `claude -p` contexts, grades both with a binary reference-guided LLM-as-judge, and writes a `benchmark.json` with pass-rate / token / duration deltas.

## Usage

```bash
npx moon-skill-eval-outputs [evals.json] [skill-name] [iteration] [--runs N] [--concurrency N] [--model M] [--judge-model M] [--plugin-dir PATH]
# evals defaults to ./evals.json; skill-name defaults to the name in ./SKILL.md
```

Requires the `claude` CLI on PATH.

Use `--plugin-dir packages/skill/<package>` to load an unpublished repository plugin and its declared local plugin dependencies only in the with-skill arm. Plugin preflight requires skill-only manifests and rejects commands, agents, hooks, MCP, LSP, and settings components. That arm invokes the exact namespaced slash command; the baseline prompt stays unchanged and blocks the Skill tool. Generation runs expose only `Skill` plus `Read` in the activated arm and only `Read` in baseline. Judges have no tools. Every subprocess excludes all user/project/local settings and MCP discovery and does not persist sessions.

Each generation is limited to six turns, 10,000 streamed response characters, and $0.10; each judge is limited to one turn and $0.10. Both spend caps can only be lowered. Runs per arm are capped at three, concurrency at two, and one invocation may launch at most 24 model calls across generation and judging. Larger eval sets must be split into bounded batches. Generation is instructed to stay below 1,000 words, and the runner kills a stream that crosses the hard character limit. Calls are attempted once, and failed or unactivated generations are not sent to a judge. Minimal fixed system prompts avoid paying for Claude Code's general-purpose coding context.

Judge responses use a strict JSON schema with exactly one verdict per assertion. The runner publishes `benchmark.json` only when every generation and judge call succeeds and every activated arm confirms the target skill is available. A timeout, budget exhaustion, output-limit, process error, unparseable judge response, or missing activation stops scheduling new jobs for that skill, removes any stale benchmark, writes `invalid-run.json`, and exits `2`; only calls already occupying a configured concurrency slot may finish. A valid non-positive quality delta exits `1`; a valid positive delta exits `0`.
