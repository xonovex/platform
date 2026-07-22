# commit: Auto-Commit with Conventional Messages

Auto-commit with a conventional message `<type>: <description>` inferred from changed files. **Default: commit immediately without prompting**; show suggestions and prompt only when interactive mode is explicitly requested.

## Type detection

Infer `<type>` from changed files (default `chore` if nothing matches):

| Changed files                             | Type       |
| ----------------------------------------- | ---------- |
| `*.test.ts`, `*.spec.ts`                  | `test`     |
| `*.md` under `docs/`                      | `docs`     |
| CI config files                           | `ci`       |
| `package.json`, lockfiles, project config | `chore`    |
| New files in `src/`                       | `feat`     |
| Small modifications in `src/`             | `fix`      |
| Large changeset across many files         | `refactor` |

Types: `chore`, `feat`, `fix`, `docs`, `refactor`, `test`, `ci`. Format `<type>: <description>` — lowercase, no footers. Precedence: user-specified type > detected > `chore`.

## Commit and push

```bash
git add -A && git commit -m "<type>: <description>"
# push (optional): remote from `git config branch.<branch>.remote` (default origin), branch from `git branch --show-current`
git push -o ci.skip <remote> HEAD:<branch>
```

When the staged change needs more context than the subject can carry, use a concise
multi-line body derived from the diff and the user's stated intent:

```
feat: implement email+password flow with TOTP

Add LoginFlow with email validation and TOTP verification.
```

## Gotchas

- Auto-detecting `refactor` on a wide changeset often misses a more specific intent (`feat` / `fix`) — interactive mode is safer for very large diffs
- `-o ci.skip` skips CI on the push; drop it when you actually want CI to run
- Default mode commits without asking — risky on dirty trees with mixed-intent changes; pre-stage or use interactive mode
