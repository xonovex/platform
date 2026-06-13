# moon-skill-eval-triggers

TRIGGER eval (ported from `eval-triggers.py`): runs the queries in an `eval-queries.json` and reports whether the target **Skill fires** for each, killing the `claude` run on first match so non-triggering runs stay cheap.

## Usage

```bash
npx moon-skill-eval-triggers [eval-queries.json] [skill-name] [train|validation|all] [--runs N] [--threshold F] [--model M] [--max-budget-usd N]
# queries defaults to ./eval-queries.json; skill-name defaults to the name in ./SKILL.md
```

Requires the `claude` CLI on PATH.
