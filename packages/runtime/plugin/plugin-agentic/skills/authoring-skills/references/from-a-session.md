# from-a-session: Fold the Lessons of a Session into a Skill

Turn the reusable lessons of a session into guidance of the skill that owns their domain. **authoring-instructions** owns the extraction, the stored lesson format and the governed promotion; a one-off project lesson goes through it to AGENTS.md.

## Inputs

- **Category** (optional): one lesson category; the default takes every reusable lesson of the session
- **From reflections** (optional): read stored `reflections/*.md` files in place of the session
- **Overwrite** (optional): replace an existing skill in place of a merge
- **Output path** (optional): the path of a new skill
- **Preview** (optional): show the result without a write

## Workflow

1. **Gather**: extract the lessons as **authoring-instructions** does, and keep those that hold beyond this project.
2. **Route**: merge each lesson into the existing skill that owns its domain, into its Essentials, its Gotchas or the matching `references/{topic}.md`, even across categories. A general lesson goes to its general owner, never to a project file.
3. **Create only with mass**: write a new skill only when a category has 3 to 7 lessons worth keeping and no skill owns it; one category becomes one skill with a gerund name, and each topic with enough mass one reference.
4. **Merge**: remove duplicates, keep 3 to 7 Essentials, and give each new reference a "Load when" line.
5. **Mark**: set `applied: true` on each stored lesson after the write.

A managed or executable target, such as a skill with scripts in a published plugin, follows the governed promotion of **authoring-instructions**.

## Gotchas

- One lesson does not make a skill: it goes to AGENTS.md
- A skill with only one or two Essentials is filler: wait for 3 to 7
- A change to a published skill needs the lockstep version bump of its plugin and both marketplaces
