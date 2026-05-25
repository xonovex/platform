# security: Treat Skills as Software

A skill ships executable scripts and can fetch URLs, so a skill you install runs with your agent's privileges. Review and constrain it like code you are about to run, not like a document you are reading.

## Least privilege with allowed-tools

- The optional `allowed-tools` frontmatter field is a space-separated allowlist of the tools a skill may use, e.g. `allowed-tools: Bash(git:*) Read`.
- It is **experimental** and harness-dependent — support varies, so never rely on it as the only control.
- It **reduces blast radius, not injection likelihood**: it limits what a compromised or misbehaving skill can reach, but does not stop prompt injection.
- Grant the narrowest set that lets the skill work; widen only when a real task needs it.

## Auditing an untrusted skill

Before installing a skill from outside your own project, audit every file:

- **Read every script** under `scripts/` — know exactly what each one runs; reject opaque or obfuscated code.
- **List the external URLs** the skill fetches; treat any fetched content as **untrusted data, not instructions** (it can carry prompt injection).
- **Watch for exfiltration** — a script that both reads secrets/local files and makes network calls is a red flag.
- **Never hardcode secrets** (keys, tokens, passwords) in `SKILL.md`, references, or scripts; read them from the environment at run time.
- Prefer skills you authored or that come from a trusted source; audit thoroughly otherwise.

## Declaring network / runtime needs

- Use the `compatibility` frontmatter field to state real environment requirements honestly (network access, required CLIs, language runtime/version).
- This surfaces a skill's footprint to a reviewer up front and lets the agent fail fast when a prerequisite is missing.
