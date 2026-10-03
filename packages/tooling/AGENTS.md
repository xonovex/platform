# Tooling

Keep command-line tools, Moon scripts, and Moon extensions in their owning tooling directories.

- Keep engineer-run command line tools under `cli/`, Moon task binaries under `script/`, and Moon extensions and toolchains under `moon/`. Runtime plugin packages must not contain buildable tooling.
