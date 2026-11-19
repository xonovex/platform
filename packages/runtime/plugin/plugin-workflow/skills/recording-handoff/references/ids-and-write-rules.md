# IDs, names and write rules

## Names

A file name without `.md` is `<slug>-<suffix>`:

- The slug is lowercase ASCII kebab-case, at most 48 characters. Derive it from the title: transliterate to ASCII, lowercase, turn every run of other characters into one hyphen, trim hyphens at both ends, cut at 48 and trim again.
- The suffix is 4 random characters from `[a-z2-7]`: only the letters `a` to `z` and the digits `2` to `7`, so `0`, `1`, `8` and `9` are never used. The same holds for run ids. Generate one with `LC_ALL=C tr -dc 'a-z2-7' </dev/urandom | head -c4`. Draw a new suffix when the directory already holds a file with that name. Randomness keeps parallel branches from colliding on a name.
- A log entry is `<YYYY-MM-DD>-<skill>-<suffix>`.

Display numbers (1, 2, 3 in index order) exist only in generated indexes. They never appear in a file name or in frontmatter, because they change when the order does.

## IDs are permanent

A concept ID is the file path without `.md`. A file is never moved or renamed, because other files cite it. A change of mind is a new file that supersedes the old one, never an edit. Titles may change in files that are edited in place; the ID does not follow them.

## Append-only and edited in place

- Append-only, so a change is a new file: decisions, research, critiques, log entries. The one permitted edit: a superseded decision gets `status: deprecated` and `superseded_by`. The edit leaves `generated` untouched, because the content did not change, and nothing else in the file changes.
- Edited in place, because git keeps the history: work files and `plan.md` outside its generated block.
- Edited only by the generator: `index.md` and `log.md`.

## Who may change what

- Only writing-plan and revising-plan change `plan.md`. Another skill that finds the plan wrong records the finding in its own file and leaves the plan alone.
- The plan holds only what no other file holds: objective, scope, success criteria, and the block between `<!-- generated:begin -->` and `<!-- generated:end -->` that links decisions and work. Rewrite the block as a whole and edit nothing inside it by hand.
- Only the owner skill of a type creates files of that type.

## The log entry of a run

- `action` is `Creation` when the run created a concept file other than the log entry itself, otherwise `Update`. `Deprecation` and `Initialization` keep their own meaning.
- `touched` lists the concept files the run created or changed and never the log entry itself. Omit the key when the list is empty.
- A string value that contains a colon followed by a space, a quote, or a leading special character is written in single quotes in the frontmatter, as in `authorization: 'The invoker said: go ahead'`.

## Every edit

- Preserve unknown frontmatter keys.
- Update `generated.at` when the content changes, and `generated.by` to the editing skill unless a person wrote the change.
- Write one log entry for each run that changes the bundle. A blocked run that changed nothing writes none.
- Never commit or stage. Leave the working tree for the invoker to review.
