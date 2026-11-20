# Decision brief template

Copy the block into the message for one question, replace every angle-bracket placeholder and delete any comment lines. One brief for each message, and the message ends after the closing line.

```markdown
## Decision <N> of <M>: <title>

### Files involved

| Path or system | What it does today | Consumers or markers |
| -------------- | ------------------ | -------------------- |
| <path>         | <behavior>         | <who relies on it>   |

### Why this is not obvious

<The constraint that makes the question hard, what the source assumes, and whether the code or the design record agrees. Say which options were excluded and why.>

### Options

| Option                             | Gains           | Costs           |
| ---------------------------------- | --------------- | --------------- |
| A. <Literal reading of the source> | <What it gives> | <What it costs> |
| B. <Alternative>                   | <What it gives> | <What it costs> |

### Recommendation

<Option letter and one sentence of reason.> This accepts <the trade-off>.

You can pick an option, add your own, leave this open for <owner> to validate, or pause it until <the fact that would settle it> is known.
```

## Fill in

- `Decision <N> of <M>`: the display position in the queue and the queue size. When the queue grew, say so once.
- `Files involved`: every file, page or system the question touches, read before you write the row. Leave out files that are not affected. In a discovered question with no files, drop the section.
- `Why this is not obvious`: two to five sentences. State a fact only when you checked it, and say how.
- `Options`: two to four lettered options. One is the literal reading of the leading source. An option the operator adds later takes the next letter.
- `Recommendation`: one option, honest, with the trade-off named. Label a recommendation that rests on an inference.
- The closing line keeps the open and paused states on offer. Name the likely owner and the fact when you can.
- Do not include the next question, a record, or a consolidation in the same message.
