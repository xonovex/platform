# Xonovex Design

Use a monochrome canvas, clear text, quiet surfaces, and restrained blue and green accents for Xonovex documentation and diagrams. Provide matching light and dark modes for the documentation website. Use this document as the visual reference when creating or changing an asset.

## Reference

The [Xonovex staging site](https://staging.xonovex.com/) is the brand reference, inspected on October 3, 2026. Its page and stylesheet use monochrome navigation and buttons, rounded cards, a subtle dot grid, blue heading accents, and blue-to-green section accents. The site requests Inter and Space Grotesk from Google Fonts; its hero headings explicitly use Space Grotesk.

The light palette records the site's colors as hexadecimal values for static rendering. The dark palette adapts the same hierarchy to dark surfaces. Dark green and pale green are diagram extensions of the site's green accent. The documentation generator reads the token table below to create the website colors and dark workflow diagram variants.

## Logo

Use the [Xonovex SVG logo](packages/asset/asset-images/xonovex-logo.svg) for brand identity. Its vector path traces the [original staging logo](https://staging.xonovex.com/xonovex-logo-1.png), preserving the two triangles and three connected rings. Keep the transparent background and the `559:838` aspect ratio. Do not stretch, crop, or add a background to the mark.

Use the text token for the logo: black in light mode and near-white in dark mode. The website derives both SVG variants from this source and the color table below. Place the logo beside the Xonovex name in navigation at 36 pixels high and show it in the homepage hero. The SVG favicon follows the browser's system appearance. Give each logo image an accessible name.

## Colors

| Role | Token | Light | Dark | Use |
| --- | --- | --- | --- | --- |
| Canvas | `canvas` | `#ffffff` | `#0b0f19` | Page and diagram background. |
| Surface | `surface` | `#f9fafb` | `#111827` | Phase panels and secondary sections. |
| Text | `text` | `#000000` | `#f9fafb` | Titles, labels, and primary prose. |
| Secondary text and connections | `secondary` | `#374151` | `#d1d5db` | Supporting text and normal flow arrows. |
| Muted text | `muted` | `#6b7280` | `#9ca3af` | Metadata on the canvas. |
| Border | `border` | `#d1d5db` | `#374151` | Cards, nodes, and phase boundaries. |
| Strong surface | `strong` | `#111827` | `#f9fafb` | Selected stages, handoffs, and terminal outcomes, with canvas-colored text. |
| Blue accent | `blue` | `#3b82f6` | `#60a5fa` | Small brand accents and emphasis. |
| Dark blue | `blue-strong` | `#2563eb` | `#93c5fd` | Decision borders, links, focus outlines, and retry arrows. |
| Pale blue | `blue-soft` | `#eff6ff` | `#172554` | Decision surfaces. |
| Green accent | `green` | `#22c55e` | `#4ade80` | Small positive accents. |
| Dark green | `green-strong` | `#15803d` | `#86efac` | Borders for improved guidance or a completed result. |
| Pale green | `green-soft` | `#f0fdf4` | `#052e16` | Quiet positive surfaces, with primary text. |

Keep most of each asset monochrome. Use accents to explain a state or relationship. Blue and green gradients may appear as small decorative rules in web layouts; diagram connections and labels use solid colors. Use the text and secondary tokens on canvas and surface backgrounds. Use the canvas token for text on a strong surface. Keep bright accents out of small body text.

## Website modes and interaction

Follow the visitor's system appearance on the first visit. Provide a visible light and dark mode switch and retain an explicit choice. Apply the selected mode to navigation, catalogs, forms, search, code blocks, and generated workflow diagram images.

Use a responsive catalog grid with one column on small screens. Keep filter labels visible, announce the matching result count, and show a clear empty state. Use a visible blue focus outline, native controls, keyboard-accessible links, and sufficient contrast in both modes. Keep motion brief and honor reduced-motion preferences.

Use Space Grotesk for website headings and Inter for text and controls. Keep plugin ownership, version, and supported harnesses visible where they help readers select or install a component. Link to the source document from every generated documentation page.

## Typography

Use Space Grotesk for diagram titles and phase headings. Use Inter for node labels, connection labels, and supporting text. Use a monospace font for standalone code examples in documentation.

For the complete workflow diagram, use 22-point phase headings, 17-point node labels, and 15-point connection labels. Compact diagrams may use 17 to 19 points for nodes and 15 to 16 points for connections. Keep labels concise and break them at meaningful boundaries. Preserve exact command names.

The diagram render environment supplies Inter and Space Grotesk through the pinned Nix packages. Select the supplied Fontconfig configuration when rendering so a host font does not silently replace either family. SVG labels remain text; a viewer without these fonts can use its sans-serif fallback. PNG outputs preserve the rendered typography.

## Shape and spacing

Use rounded rectangular nodes and phase panels. Use diamonds for decisions, ovals for terminal states, and note shapes for supporting rules. Keep a visible gap between panels and enough padding to separate text from borders.

Use thin borders and a clear hierarchy. Keep shadows, dot grids, and decorative gradients out of information-dense diagrams. Preserve the site's quiet surface treatment without adding detail behind labels or connections.

## Diagram conventions

| Element | Treatment | Meaning |
| --- | --- | --- |
| Phase panel | Surface fill, border token, primary text. | Groups related steps. |
| Ordinary step | Canvas fill, border token, primary text. | An action or operation. |
| Decision | Pale blue fill, dark blue border, diamond shape. | Selects a labeled outcome. |
| Normal connection | Solid secondary-colored arrow. | Advances through the flow. |
| Retry connection | Dashed dark blue arrow with a label. | Returns to an earlier step. |
| Handoff or selected stage | Strong surface with canvas-colored text. | Names another phase or emphasizes a stage. |
| Positive result | Pale green fill, dark green border, primary text. | Records improved guidance or a completed result. |
| Supporting note | Canvas fill, border token, primary text. | Explains context or a constraint. |

Use shape, line style, and explicit labels with color. A reader must be able to follow the process in grayscale. Keep command ownership visible where a diagram crosses plugin boundaries. Link every diagram to its explanatory document.

## Sources and exports

Keep editable Graphviz sources with their committed SVG and PNG exports. Use SVG for scalable documentation and PNG where vector images are unavailable. Repository exports use the light palette and render a white background at 150 dots per inch for PNG outputs. The website generates dark SVG variants from the token table and selects the correct image when the visitor changes mode. Product screenshots retain their recorded appearance.

The [workflow diagrams](packages/runtime/plugin/plugin-workflow/diagrams/README.md) apply this theme. After editing their sources, run:

```bash
nix develop --no-update-lock-file --command npx moon run plugin-workflow:graph-build --force
nix develop --no-update-lock-file --command npx moon run plugin-workflow:ci-check --force
```

Before delivering an update, inspect the rendered layout, check labels and contrast, and confirm that source and export changes describe the same workflow. Apply these rules when changing other repository diagrams; screenshots retain the appearance of the product they record.
