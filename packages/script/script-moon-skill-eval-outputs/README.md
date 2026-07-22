# moon-skill-eval-outputs

OUTPUT eval runs each eval twice — vanilla vs skill-augmented — in isolated model contexts, grades both with a binary reference-guided LLM-as-judge, and writes a `benchmark.json` with pass-rate / token / duration deltas and a tier-aware quality gate.

## Usage

```bash
npx moon-skill-eval-outputs [evals.json] [skill-name] [iteration] [--harness claude|codex] [--runs N] [--concurrency N] [--model M] [--judge-model M] [--plugin-dir PATH]
# evals defaults to ./evals.json; skill-name defaults to the name in ./SKILL.md
```

Requires the selected CLI on PATH and its API credential. Claude is the default harness.

The Claude adapter uses `--plugin-dir packages/skill/<package>` to load an unpublished repository plugin and its declared local plugin dependencies only in the with-skill arm. Plugin preflight requires skill-only manifests and rejects commands, agents, hooks, MCP, LSP, and settings components. That arm invokes the exact namespaced slash command; the baseline prompt stays unchanged and blocks the Skill tool. Generation runs expose only `Skill` plus `Read` in the activated arm and only `Read` in baseline. Judges have no tools. The Codex adapter uses separate ephemeral homes, stages only the target skill for the activated arm, explicitly invokes it with `$skill-name`, ignores inherited configuration and rules, and runs read-only.

Claude generation is limited to six turns, 10,000 streamed response characters, and $0.10; each Claude judge is limited to one turn and $0.10. Both spend caps can only be lowered. Codex has no CLI dollar-budget flag. Both adapters cap runs per arm at three, concurrency at two, one invocation at 24 model calls, output at 10,000 characters, and generation time at the configured timeout. Larger eval sets must be split into bounded batches. Calls are attempted once, and failed or unactivated generations are not sent to a judge.

Claude judge responses use a strict JSON schema; Codex judges return the same assertion-result shape in their final message. The runner publishes `benchmark.json` only when every generation and judge call succeeds and every activated arm confirms the target skill. A timeout, budget exhaustion, output-limit, process error, unparseable judge response, or missing activation stops scheduling new jobs for that skill, removes any stale benchmark, writes `invalid-run.json`, and exits `2`; only calls already occupying a configured concurrency slot may finish.

Every `evals.json` declares `skill_name`, `tier` (`aggressive`, `moderate`, or `conservative`), and `evals`. A valid run exits `0` only when every activated arm fired and its tier's absolute with-skill pass-rate and minimum-delta gates pass. Aggressive and moderate skills require at least a 0.05 delta with absolute pass rates of 0.75 and 0.80 respectively; conservative skills require a 0.10 delta and a 0.90 absolute pass rate. Valid evidence that misses a quality gate exits `1`.
