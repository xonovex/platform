# The plan inputs and front matter

## Contents

- Reading the inputs
- Front matter

## Reading the inputs

- **Research files:** read each in full. A file that another lists in `corrects` is superseded for the points the correction restates, so read the correcting file for those points. The `questions` of the files are the decisions the change still needs.
- **Decisions:** group them by `state` as `decisions/index.md` does. A decision whose `status` is `deprecated` is history; follow its `superseded_by` to the current one. `settled` states a fact of the target. `open`, `paused` and `partly-settled` are linked, never described. A `partly-settled` decision contributes only the part its `# Decision` states as settled.
- **Questions with no decision file:** a research question that no decision file covers is unrecorded. Do not turn it into a decision here, because **settling-questions** owns decision files. List it in the reply, and write no work file that depends on it.
- **Owners, repositories, skills:** take `owners` from the decision owners and the people or teams the request names, `repositories` from the repositories the research read and the work will touch, and `skills_to_consult` from the union of the research files' lists, keeping only installed skill names.

## Front matter

Start from the shared plan template in **recording-handoff** and fill the common header as that skill describes.

| Key | Value |
| --- | --- |
| `generated` | `{by: writing-plan/<version>, at: <now with offset>}` |
| `status` | `draft`. No `verified` entry: nobody has confirmed the plan, and the skill never adds one for its own content |
| `owners` | Include when the inputs name them |
| `repositories` | `{ name, resource }` with the name from the project's repository map; include when the change writes to a repository |
| `skills_to_consult` | The union described above; leave the key out when empty |
| `sources` | One entry for each research file the plan relies on, with `id`, `title` and a bundle-relative `resource` like `/research/x` |

The `title` names the change as a short noun phrase, and the `description` is one sentence that states what the change delivers.
