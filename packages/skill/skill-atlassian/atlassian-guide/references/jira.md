# jira: Jira work-item operations with acli

The first Jira capabilities under an authenticated `acli` (auth is in [auth.md](auth.md)). All commands live under `acli jira workitem`; add `--json` for machine-readable output on the read and mutate commands. Related trees not covered in depth here: `acli jira project|board|sprint|filter|dashboard|field` (same shape — `acli jira <thing> --help`).

## Search

```bash
acli jira workitem search --jql "project = TEAM AND statusCategory != Done"
acli jira workitem search --jql "assignee = currentUser() AND statusCategory != Done" --json
acli jira workitem search --jql "project = TEAM" --fields "key,summary,assignee,status" --csv
acli jira workitem search --jql "project = TEAM" --count           # just the total
acli jira workitem search --jql "project = TEAM" --paginate        # all pages, not just the first
acli jira workitem search --filter 10001                            # by saved-filter id
acli jira workitem search --jql "project = TEAM" --web             # open the search in a browser
```

- `--jql` takes any valid JQL; `currentUser()` resolves to the authenticated account.
- Default fields are `issuetype,key,assignee,priority,status,summary`; override with `-f/--fields`.
- Output is a table by default; `--json` for scripting, `--csv` for spreadsheets. Use `--limit N` or `--paginate` to control result volume (`--paginate` fetches every page).

## View

```bash
acli jira workitem view KEY-123
acli jira workitem view KEY-123 --json
acli jira workitem view KEY-123 --fields summary,comment          # subset
acli jira workitem view KEY-123 --fields "*all"                    # everything
acli jira workitem view KEY-123 --fields "*navigable,-comment"     # navigable minus one
acli jira workitem view KEY-123 --web
```

`--fields` accepts `*all`, `*navigable`, an explicit comma list, or a field prefixed with `-` to exclude it.

## Create

```bash
# minimal: summary + project + type
acli jira workitem create --summary "New task" --project TEAM --type Task

# fuller: assignee (@me self-assigns), labels (comma-separated), parent, description
acli jira workitem create --project PROJ --type Bug \
  --summary "Login button misaligned" \
  --description "Repro: ... Expected: ..." \
  --assignee @me --label bug,ui --parent PROJ-100

# description from a file (plain text or Atlassian Document Format)
acli jira workitem create --project PROJ --type Story --summary "..." --description-file body.adf

# author interactively in $EDITOR
acli jira workitem create --project PROJ --type Task --editor

# JSON-driven: generate a template, fill it, create from it
acli jira workitem create --generate-json                 # writes a template JSON
acli jira workitem create --from-json workitem.json
acli jira workitem create --summary "..." --project TEAM --type Task --json   # JSON output of the created item
```

- `--assignee` takes an email or account id; `@me` self-assigns, `default` uses the project default assignee.
- `--type` is the work-item type name (`Task`, `Bug`, `Story`, `Epic`, …) as configured in the project.
- `--label` / `-l` is comma-separated; `--parent` sets the parent (e.g. an Epic or the parent of a subtask).
- For repeatable / complex creation, `--generate-json` → edit → `--from-json` is the sturdiest path. `create-bulk` exists for many at once.

## Transition (change status)

```bash
acli jira workitem transition --key "KEY-1,KEY-2" --status "Done"
acli jira workitem transition --jql "project = TEAM AND status = 'To Do'" --status "In Progress"
acli jira workitem transition --filter 10001 --status "To Do" --yes        # skip the confirm prompt
```

- Target items with `--key` (comma list), `--jql`, or `--filter`.
- `--status` is the destination status **name** and must be a currently-available transition for the item's workflow — an invalid target errors.
- `--yes` skips confirmation (needed for non-interactive runs); `--ignore-errors` continues past items that fail.

## Comment

```bash
acli jira workitem comment list KEY-123
acli jira workitem comment create KEY-123 --body "..."     # see `... comment create --help` for exact flags
acli jira workitem comment update ...
acli jira workitem comment visibility KEY-123              # available visibility roles/groups
```

Comment bodies accept plain text or Atlassian Document Format. Check `acli jira workitem comment create --help` for the current flag names before scripting (comment flags evolve across `acli` releases).

## Notes

- Every mutating command supports `--json`; capture it and parse a real field rather than scraping the table.
- Work-item **keys** (`KEY-123`) are what these commands take, not numeric ids.
- When a command's exact flags matter for a script, confirm against `acli jira workitem <cmd> --help` on the installed version — this reference tracks `1.3.x`.
