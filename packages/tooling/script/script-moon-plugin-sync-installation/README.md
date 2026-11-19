# script-moon-plugin-sync-installation

Use this command to keep the generated installation section of each plugin README in step with its marketplace entry and manifest. The section sits between the `xonovex:installation` markers and names the marketplace, the plugin version, its dependencies, the harness commands that install it, and the optional companions: the plugins that its skills hand off to in bold and that the manifest does not declare.

Run `npx moon run script-moon-plugin-sync-installation:plugin-installation-sync` after changing a plugin's version, name, dependencies, harness support, or the bold handoffs of its skills. It rewrites only the marked block of each README. Run `npx moon run script-moon-plugin-sync-installation:plugin-installation-check` to report READMEs that have drifted; it writes nothing and exits 1 on drift. The `ci-check` task runs the check.
