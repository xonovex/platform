# Documentation Site

Run `npm run docs:dev` from the repository root to browse the Xonovex documentation and agent component catalogs. VitePress reads generated pages from the existing repository Markdown, plugin manifests, package metadata, and `DESIGN.md`.

## Commands

| Task | Command from the repository root |
| --- | --- |
| Generate pages and README installation sections | `npm run docs:generate` |
| Start the website with source watching | `npm run docs:dev` |
| Build the static website | `npm run docs:build` |
| Preview the production build | `npm run docs:preview` |
| Check installation sections for drift | `npm run docs:check` |
| Run the complete site gate | `npx moon run documentation-site:ci-check --force` |

## Source ownership

Edit the original documents and metadata. The `content/` directory contains generated pages, catalog data, styles, and copied assets; it is ignored by Git. Build output is under `dist/`. Generation writes only the installation block marked `xonovex:installation:start` and `xonovex:installation:end` in each plugin README. Other README prose remains authored in place.

Installation markers stay in the source Markdown and are hidden on the website. Inline and fenced code examples preserve literal marker text.

The marketplace lists the plugin packages. Plugin manifests identify their names, versions, dependencies, and supported harnesses. `SKILL.md` frontmatter identifies skills, and command frontmatter supplies descriptions. Each component type has a catalog. Agent, hook, and MCP server catalogs explain an empty state until bundles contain those components. Tool README files supply the agent CLI and operator documentation.

Skills and commands link to their owning plugin and get installation instructions on their generated detail pages. Commands are exposed for Claude Code; Codex uses the declared skills. Existing references and diagram pages stay linked to their source files. Local links are resolved to generated pages, copied assets, or the repository source.

## Appearance

The site follows [Xonovex design](../../../DESIGN.md). It reads light and dark colors from that document, follows the system theme initially, and remembers a manual choice. Workflow SVG images get generated dark variants. The original Graphviz sources and repository exports remain the editable source and light images.

The [Xonovex SVG logo](../../asset/asset-images/xonovex-logo.svg) appears in navigation and the homepage hero. The generator creates its light and dark variants from the shared text token and provides a favicon that follows system appearance.

## Validation

The site gate checks generated installation sections, TypeScript and Vue types, lint, tests, formatting, and the production build. VitePress checks its internal page links. Generation and the catalogs use the local repository; they do not need an account or a remote content service.

Use `DOCS_BASE=/platform/ npm run docs:build` for a host with a URL prefix. Publish the contents of `dist/` to a static host after reviewing the build.
