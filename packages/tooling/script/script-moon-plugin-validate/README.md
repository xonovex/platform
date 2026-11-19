# moon-plugin-validate

Use this validator to reject plugin packages that break the Xonovex skill catalog, link, dependency, or prose contracts.

For one plugin package directory it checks that:

- the README tables under `## Skills` list exactly the skills of the plugin: each skill is a directory under `skills/` that holds `SKILL.md`, and each table row starts with a link such as `[using-git](skills/using-git/SKILL.md)`
- every local link in the README and in `docs/**/*.md` resolves to a file, a directory, or a heading inside the package
- for every skill plugin that `.claude-plugin/plugin.json` or `package.json` declares as a dependency, a `SKILL.md` of the plugin names one of the skills of that dependency
- the package has no `commands` directory, because a plugin holds skills only
- no prose file in the package contains an em dash, an ellipsis character, or a typographic quote, in literal or Unicode-escaped form

The result line counts the skills of the plugin. The validator exits with 0 on a pass, 1 on a validation failure, and 2 on an input error.
