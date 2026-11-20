# App Window

## Inspect before changing

Use `herdr-app` only to change what the person sees; use `hc` for work that need not change the screen. Load `herdr-app --skill`, the authority for the window, and confirm each command shape with `herdr-app commands`.

```bash
herdr-app ls
herdr-app status
herdr-app commands
```

`ls` lists spaces, their tabs, and terminals with pane IDs. `status` shows the layout, focus, and errors. A terminal address such as `t7` is not a pane ID. Use IDs from `ls`; do not act on the focused pane by guess.

## Show something when asked

Create spaces and tabs only when the user asks; they appear at once. Give each an explicit name, and pass `--focus false` to keep the user's current view.

```bash
herdr-app do space.new --name api --folder "$PWD" --argv '["claude"]' --task "Look at the login bug" --focus false
herdr-app do tab.new --name "fix login" --argv '["claude"]' --task "Look at the login bug" --focus false
herdr-app do pane.split '{"direction":"right","pane":"p12"}'
herdr-app do agent.focus --terminal t7
herdr-app do layout.clone --name <layout>
```

`do` takes `--key value` arguments, with JSON for arrays and objects, or one JSON object. Supply every required argument. It prints the created IDs, or `{ "error": ... }` with exit status 1. A successful reply means the terminal was created, not that its program started correctly: confirm with `ls` or `status`.

## Place an hc terminal in the window

The label `herdr.space=<space-id>` shows a terminal in that space, and `herdr.tab=<tab-id>` chooses its tab; without `herdr.tab`, the app creates a tab. Set the labels at start with `hc run --label`, or later with `hc label <t>`. A terminal without the labels stays hidden, so do not label a background worker unless the user wants to see it.

## Protect the window

Never start bare `herdr-app`: it opens or raises the window. Never stop, signal, or restart the app or the core. For settings, layouts, keys, or mods, follow `herdr-app --skill` and the installed guide that it names.
