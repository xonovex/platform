# Scanning the content

Scan every file written under `deliveries/<work-id>/` before reporting done, and the worktree diff too, because the diff is what the files describe.

## What to look for

- Secrets: private key blocks, access key IDs, passwords, tokens and API keys assigned in text, and URLs that carry credentials.
- Anything the plan marks sensitive. Read the plan for a statement of what is sensitive, such as a named system, an account, a personal data field or a finding that must stay out of files, and search the content for each term it names.

## Commands

From the bundle root, search the delivery directory with patterns for the usual shapes:

```bash
rg -n -i \
  -e '-----BEGIN [A-Z ]*PRIVATE KEY-----' \
  -e 'AKIA[0-9A-Z]{16}' \
  -e '(password|passwd|secret|token|api[_-]?key)[a-z_]*\s*[:=]' \
  -e '://[^/[:space:]:@]+:[^/[:space:]@]+@' \
  deliveries/<work-id>
```

Run the same patterns over the diff by piping it in, and add one `-e` for each sensitive term the plan names. A pattern hit is a lead, not a verdict: read the line, and treat a placeholder or an invented example in test data as a finding to mention, not a secret.

## When something is found

- In a delivery file: remove the content from the file, and rerun the scan.
- In the diff: the files must not carry it, and the worktree is not yours to change. Do not report done. Report the path and line, never the value, tell the invoker to remove it from the worktree and to revoke the credential if it was real, then rerun the skill.
- In the reply: never repeat a secret, even to point at it.

Report the scan as a command with its outcome, and name any pattern the project's check may need that the list above lacks.
