# Xonovex Design

Use a white canvas, black text, pale gray surfaces, and restrained blue and green accents for Xonovex documentation and diagrams. Use this document as the visual reference when creating or changing an asset.

## Reference

The [Xonovex staging site](https://staging.xonovex.com/) is the brand reference, inspected on October 3, 2026. Its page and stylesheet use monochrome navigation and buttons, rounded cards, a subtle dot grid, blue heading accents, and blue-to-green section accents. The site requests Inter and Space Grotesk from Google Fonts; its hero headings explicitly use Space Grotesk.

The palette below records the site's colors as hexadecimal values for static rendering. Dark green and pale green are diagram extensions of the site's green accent. They provide readable text and a quiet completion surface.

## Colors

| Role | Color | Use |
| --- | --- | --- |
| Canvas | `#ffffff` | Page and diagram background. |
| Surface | `#f9fafb` | Phase panels and secondary sections. |
| Text | `#000000` | Titles, labels, and primary prose. |
| Secondary text and connections | `#374151` | Supporting text and normal flow arrows. |
| Muted text | `#6b7280` | Metadata on white surfaces. |
| Border | `#d1d5db` | Cards, nodes, and phase boundaries. |
| Strong surface | `#111827` | Selected stages, handoffs, and terminal outcomes, with white text. |
| Blue accent | `#3b82f6` | Small brand accents and emphasis. |
| Dark blue | `#2563eb` | Decision borders and retry arrows. |
| Pale blue | `#eff6ff` | Decision surfaces. |
| Green accent | `#22c55e` | Small positive accents. |
| Dark green | `#15803d` | Borders for improved guidance or a completed result. |
| Pale green | `#f0fdf4` | Quiet positive surfaces, with black text. |

Keep most of each asset white or gray. Use accents to explain a state or relationship. Blue and green gradients may appear as small decorative rules in web layouts; diagram connections and labels use solid colors. Use black or dark gray text on pale surfaces. Use white text on strong surfaces. Keep bright blue and green out of small body text.

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
| Phase panel | Pale gray fill, gray border, black heading. | Groups related steps. |
| Ordinary step | White fill, gray border, black label. | An action or operation. |
| Decision | Pale blue fill, dark blue border, diamond shape. | Selects a labeled outcome. |
| Normal connection | Solid dark gray arrow. | Advances through the flow. |
| Retry connection | Dashed dark blue arrow with a label. | Returns to an earlier step. |
| Handoff or selected stage | Strong surface with white text. | Names another phase or emphasizes a stage. |
| Positive result | Pale green fill, dark green border, black text. | Records improved guidance or a completed result. |
| Supporting note | White fill, gray border, black text. | Explains context or a constraint. |

Use shape, line style, and explicit labels with color. A reader must be able to follow the process in grayscale. Keep command ownership visible where a diagram crosses plugin boundaries. Link every diagram to its explanatory document.

## Sources and exports

Keep editable Graphviz sources with their committed SVG and PNG exports. Use SVG for scalable documentation and PNG where vector images are unavailable. Render a white background at 150 dots per inch for PNG outputs.

The [workflow diagrams](packages/runtime/plugin/plugin-workflow/diagrams/README.md) apply this theme. After editing their sources, run:

```bash
nix develop --no-update-lock-file --command npx moon run plugin-workflow:graph-build --force
nix develop --no-update-lock-file --command npx moon run plugin-workflow:ci-check --force
```

Before delivering an update, inspect the rendered layout, check labels and contrast, and confirm that source and export changes describe the same workflow. Apply these rules when changing other repository diagrams; screenshots retain the appearance of the product they record.
