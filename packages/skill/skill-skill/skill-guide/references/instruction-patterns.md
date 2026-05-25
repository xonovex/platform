# instruction-patterns: Structural Patterns for Skill Bodies

## Contents

[Templates for Output Format](#templates-for-output-format) · [Checklists for Multi-Step Workflows](#checklists-for-multi-step-workflows) · [Validation Loops](#validation-loops) · [Plan-Validate-Execute](#plan-validate-execute) · [When to Use Which](#when-to-use-which) · [Gotchas](#gotchas)

Reusable structural patterns for instruction sections inside a SKILL.md body. Use the ones that fit the task — most skills won't need all of them.

## Templates for Output Format

Provide a concrete markdown template when the agent must produce structured output. Pattern-matching against a sample is more reliable than prose descriptions.

- **Short templates** — inline in SKILL.md
- **Long / conditional templates** — store in `assets/` and reference from SKILL.md so they load only when needed

````markdown
## Report Structure

Use this template, adapting sections as needed:

```markdown
# [Analysis Title]

## Executive summary

[One-paragraph overview]

## Key findings

- Finding 1 with supporting data
- Finding 2 with supporting data

## Recommendations

1. Specific actionable recommendation
2. Specific actionable recommendation
```
````

## Checklists for Multi-Step Workflows

Use markdown checkboxes so the agent tracks progress and doesn't skip steps. Best when steps have dependencies, validation gates, or ordering constraints.

```markdown
## Workflow

Progress:

- [ ] Step 1: Analyze input (run `scripts/analyze.py`)
- [ ] Step 2: Create mapping (edit `mapping.json`)
- [ ] Step 3: Validate mapping (run `scripts/validate.py`)
- [ ] Step 4: Execute (run `scripts/execute.py`)
- [ ] Step 5: Verify output (run `scripts/verify.py`)
```

## Validation Loops

Do work → run validator → fix issues → repeat until passing. More reliable than hoping the agent gets it right first try. The validator can be a script, a reference checklist, or a self-check.

```markdown
## Editing Workflow

1. Make your edits
2. Run validation: `python scripts/validate.py output/`
3. If validation fails:
   - Review the error message
   - Fix the issues
   - Run validation again
4. Only proceed when validation passes
```

A reference document can serve as the validator — instruct the agent to check its work against the reference before finalizing.

## Plan-Validate-Execute

For batch or destructive operations: agent produces a structured plan → validates against a source of truth → executes only on pass. The validate step gives the agent feedback to self-correct before destructive action.

```markdown
## PDF Form Filling

1. Extract form fields: `python scripts/analyze_form.py input.pdf` → `form_fields.json`
2. Create `field_values.json` mapping each field name to its intended value
3. Validate: `python scripts/validate_fields.py form_fields.json field_values.json`
   (checks that every field name exists, types compatible, required fields present)
4. If validation fails, revise `field_values.json` and re-validate
5. Fill the form: `python scripts/fill_form.py input.pdf field_values.json output.pdf`
```

Key ingredient: step 3's validator returns errors that include the available options (e.g. `"Field 'signature_date' not found — available fields: customer_name, order_total, signature_date_signed"`), giving the agent enough information to self-correct.

## When to Use Which

- **Template** — output format must be consistent (reports, structured docs, code scaffolds)
- **Checklist** — multi-step workflow with dependencies or validation gates
- **Validation loop** — fragile edits where the agent often gets it wrong on first try
- **Plan-validate-execute** — batch / destructive ops where mistakes are costly

These compose: a multi-step workflow can use a checklist _and_ end each step with a validation loop, with plan-validate-execute as the spine.

## Gotchas

- Templates that prescribe every detail eliminate useful agent judgment — leave room for context-dependent decisions
- Checklists with too many steps (>10) signal the skill is too broad; consider splitting
- Validation loops without a clear "pass" criterion can loop indefinitely — make the validator decisive
- Plan-validate-execute is overkill for low-risk operations; reserve for destructive / batch work
