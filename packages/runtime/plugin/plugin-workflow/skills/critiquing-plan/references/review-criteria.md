# Review criteria

## Contents

- When to use the defaults
- The defaults
- Adding criteria

## When to use the defaults

Use the invoker's criteria when they give any. Otherwise give the reviewer the defaults below, word for word, so a second critique of a revised plan applies the same standard.

## The defaults

Questions about the plan as a document and as a promise:

- **Objective:** does it describe only the settled end state, in a way a reader can picture, without a mechanism nobody agreed to?
- **Scope:** are in and out both stated, and is anything that the objective needs missing from both?
- **Success criteria:** is each one observable by a stated means, such as a metric, a request path, a behavior or a check, and free of implementation detail? Could the plan be followed to the letter and a criterion still fail?
- **Dependencies:** does every assumption about another team, system, version or data source have a named owner and a way to confirm it?
- **Work:** does every piece of work have an owner team, and does each waiting piece name the open decision it waits on?
- **Sequence and recovery:** is the order of work safe at every step, and is there a way back from a bad step?
- **Variants:** where the change differs by platform, client or environment, does the plan say which difference is decided and by whom?
- **Consistency:** does the plan contradict a decision, copy a decision's text so the two can drift, or leave a decision unreferenced?
- **Sensitive content:** does the plan hold a secret or detail that should not be in a shared document?

## Adding criteria

Add a criterion for the kind of change when it applies: a migration needs a rollback and a coexistence window; a new feature needs the neighbors it touches; a quality refactor needs the measurement that shows the improvement; an investigation needs the stop condition. State each added criterion as a question the reviewer can answer from the plan.
