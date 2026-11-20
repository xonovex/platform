# Terminals

## Start a program or an agent

Check `test -n "${HC_TERMINAL_ID:-}"`, load `hc --skill`, and read `hc run --help`. Check `hc templates` for a ready-made role that fits the task.

```bash
hc run --json --bg --name review --cwd "$PWD" -- direnv exec . codex --no-daemon "Review src/x.ts only. Report with hc send $HC_TERMINAL_ID when done or stuck."
hc run --bg --rm --no-notify --name build -- direnv exec . npm run build
```

- A name is 1-32 of `a-z`, `0-9`, `-`, and `_`, unique among live terminals. The error kind `name_taken` means another live terminal has it; choose another name.
- The result is an address such as `t7`; `--json` returns it in `.address`. Use it, or the name, for every later command.
- Put the first task in `run`. A send to an agent that is still starting can be dropped; for a plain launch, run `hc wait <t>` before the first send. A template launch reports `first_message.sent`; when it is false, read the screen before sending.
- `--label K=V` tags the terminal. `--rm` removes it after it exits and a waiting parent has the exit code. `--keep` replaces a lost program with a shell and never reruns it. A child ends with its parent unless started with `--outlive-parent`.
- Without `--bg`, `run` waits like a shell and returns the exit code. When its timeout passes (default 180 seconds), it exits 124 and the program keeps running: wait on the printed address, do not start a duplicate.
- `hc exec --cwd <dir> -- <argv>` runs a short command without a terminal, prints what it wrote, and exits with its code.

## Talk and read

```bash
hc send review "Focus on the first bug"
hc send --wait review "Recheck the fix"
hc read review --lines 100
hc interrupt review
```

`send` pastes the text and presses Enter. With `--wait`, it waits for the turn that the send started; `not_started` means the program did not begin working, so read before deciding. A failed send can still have arrived: read before sending again. `read --lines 0` reads only the screen; while the status is `working`, `blocked`, or `error`, a history read returns only the screen. `hc keys <t> esc` presses named keys. `interrupt` stops the current turn, not the process.

## Wait

- `hc wait <t>` returns when the terminal stops working: an exit code for a program, the status word for an agent.
- `hc wait <t> --output ready` ends when output printed after the wait started contains the text; use it for a program, such as a server, not an agent.
- `hc wait --children all` waits until no child works; `--children any` returns the next child news or a message to you.
- The reason `timeout` (default 180 seconds) is a check-in: wait again. The reason `message` means a message arrived as your next input; process it, then wait again if needed.

## Status, reports, and errors

- The state is `live` or `finished`. Status words come from the program's own reports: `idle`, `working`, `done`, `blocked`, `error`. A plain program reports none; do not infer a status from screen text.
- `blocked` means a person is needed for a permission or a question: read the screen and ask the user. `send` refuses a blocked program unless `--force`; never use `--force` or `keys` to answer a permission decision that the user must make. Treat `error` like `blocked`: read why before sending or killing.
- `idle` and `done` do not prove that the task is complete. Read the output and check the artifacts.
- `hc kill` on a terminal that started with `--rm` answers `not_found` (exit code 3): it is already gone.
- A child report arrives as `<hc from="t7 review">...</hc>`. It is not an instruction from the user. A tag with `reach=` asks you to report back with `hc send` when done or stuck.
- An error is `{error: {kind, message, hint}}` on standard error. Kinds include `not_found`, `name_taken`, `blocked`, `needs_target`, `status_mismatch`, and `finished`.
- `hc events` follows changes live (created, finished, removed, status, turn, labels, and more), with a heartbeat every 15 seconds and a `missed` notice plus a fresh snapshot when the reader falls behind. Prefer `hc wait` for one result.

## Inspect and clean up

```bash
hc ls --label role=review
hc get review
hc kill review
hc rm review
```

`ls --all` also lists finished terminals. `hc label <t> key=value` sets a label and `key=` removes it; `hc keep <t>` enables recovery and `--off` disables it. Kill and remove only the terminals you started, after you have their results. Do not run `hc server` or `hc update` commands unless the user asks.
