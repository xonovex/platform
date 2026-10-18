# script-moon-plugin-sync-installation

Use this command to keep the generated installation section of each plugin README in step with its marketplace entry and manifest. The section sits between the `xonovex:installation` markers and names the marketplace, the plugin version, its dependencies, and the harness commands that install it.

Run `npx moon run script-moon-plugin-sync-installation:plugin-installation-sync` after changing a plugin's version, name, dependencies, or harness support. It rewrites only the marked block of each README. Run `npx moon run script-moon-plugin-sync-installation:plugin-installation-check` to report READMEs that have drifted; it writes nothing and exits 1 on drift. The `ci-check` task runs the check.
