# Documentation

Generate the website from the existing documents and component metadata, and use `DESIGN.md` for both themes.

- Edit repository source documents; keep generated pages, catalogs, and copied assets in the ignored `documentation-site/content/` directory.
- Restrict README writes to the marked installation section. Preserve authored prose and list only the harnesses that support a component.
- Discover component types from manifests and files. Keep an empty catalog when a type has no components.
- Keep light and dark colors in the root design table. Derive website variables and dark workflow SVG exports from those tokens.
- Run `npx moon run documentation-site:ci-check --force` after changing generation or presentation.
