import {mkdirSync, readFileSync, rmSync, writeFileSync} from "node:fs";
import {dirname, join} from "node:path";
import MarkdownIt from "markdown-it";
import type Token from "markdown-it/lib/token.mjs";
import {stringify as stringifyYaml} from "yaml";
import {logo, logoSource} from "./brand.js";
import {createCatalog} from "./catalog.js";
import {exists, filesBelow, inside, readText} from "./files.js";
import {installation, readmeUpdates} from "./install.js";
import {resolveLink} from "./links.js";
import {
  componentTypes,
  type Catalog,
  type ComponentType,
  type Document,
} from "./model.js";
import {darkSvg, designPalette, paletteCss} from "./palette.js";

export const categoryNames: Readonly<Record<ComponentType, string>> = {
  plugins: "Plugins",
  skills: "Skills",
  commands: "Commands",
  agents: "Agents",
  hooks: "Hooks",
  "mcp-servers": "MCP servers",
  tools: "Agent tools",
};

const categoryDescriptions: Readonly<Record<ComponentType, string>> = {
  plugins:
    "Install related skills and commands as one plugin. Filter by supported harness or find a bundle by name.",
  skills:
    "Find the procedure that fits your task. Each skill links to its owning plugin, supporting references, and installation instructions.",
  commands:
    "Find a user-invoked command and its argument contract. Install the owning plugin in a harness that supports the command.",
  agents: "Find specialist agents declared by plugin bundles.",
  hooks: "Find lifecycle hooks declared by plugin bundles.",
  "mcp-servers":
    "Find Model Context Protocol servers declared by plugin bundles.",
  tools:
    "Run coding agents locally or as Kubernetes workloads. Open the tool documentation for setup and usage.",
};

const frontmatter = (
  metadata: Readonly<Record<string, unknown>>,
  body: string,
): string =>
  "---\n" +
  stringifyYaml(metadata, {lineWidth: 0}) +
  "---\n\n" +
  body.trim() +
  "\n";

const renderDocument = (document: Document, catalog: Catalog): string => {
  const component = catalog.entries.find(
    (item) => item.source === document.source,
  );
  const plugin =
    component === undefined
      ? undefined
      : catalog.plugins.find((item) => item.name === component.plugin);
  const install =
    plugin === undefined ||
    component === undefined ||
    ["plugins", "tools"].includes(component.type)
      ? ""
      : "\n\n" +
        installation(
          {
            ...plugin,
            harnesses: plugin.harnesses.filter((harness) =>
              component.harnesses.includes(harness),
            ),
          },
          catalog.marketplace,
        );
  return frontmatter(
    {
      title: document.title,
      description: document.description,
      sourcePath: document.source,
      ...(component === undefined ? {} : {component: component.id}),
    },
    document.body + install,
  );
};

const tokensIn = (tokens: readonly Token[]): readonly Token[] =>
  tokens.flatMap((token) => [token, ...tokensIn(token.children ?? [])]);

const sourceAssets = (root: string, catalog: Catalog): readonly string[] => {
  const parser = new MarkdownIt({html: false});
  return [
    ...new Set(
      catalog.documents.flatMap((document) =>
        tokensIn(parser.parse(document.body, {})).flatMap((token) => {
          const href = token.attrGet(token.type === "image" ? "src" : "href");
          if (href === null) return [];
          const asset = resolveLink(root, catalog, document.source, href).asset;
          return asset === undefined ? [] : [asset];
        }),
      ),
    ),
  ].toSorted();
};

const home = (catalog: Catalog): string =>
  frontmatter(
    {
      layout: "home",
      title: "Xonovex",
      description:
        "Agent tools, reusable skills, and development workflows from one repository.",
      hero: {
        name: "Xonovex",
        image: logo,
        text: "Build with the right agent tools.",
        tagline:
          "Find a skill, install its plugin, and follow a workflow from research to delivery.",
        actions: [
          {
            theme: "brand",
            text: "Browse the catalog",
            link: "/catalog/plugins.html",
          },
          {theme: "alt", text: "Get started", link: "/guide/platform.html"},
        ],
      },
      features: (["plugins", "skills", "commands", "tools"] as const).map(
        (type) => ({
          title:
            String(
              catalog.entries.filter((item) => item.type === type).length,
            ) +
            " " +
            categoryNames[type].toLowerCase(),
          details: categoryDescriptions[type],
          link: "/catalog/" + type + ".html",
        }),
      ),
    },
    "",
  );

const writeChanged = (
  directory: string,
  path: string,
  content: string | Buffer,
): void => {
  const destination = inside(directory, path);
  const expected = typeof content === "string" ? Buffer.from(content) : content;
  if (exists(directory, path) && readFileSync(destination).equals(expected))
    return;
  mkdirSync(dirname(destination), {recursive: true});
  writeFileSync(destination, expected);
};

export interface GenerationResult {
  readonly catalog: Catalog;
  readonly changedReadmes: readonly string[];
  readonly pages: number;
}

export const generateSite = (
  root: string,
  site: string,
  check: boolean,
): GenerationResult => {
  const initial = createCatalog(root);
  const updates = readmeUpdates(root, initial).filter(
    (update) => update.previous !== update.content,
  );
  if (check && updates.length > 0)
    throw new Error(
      "Generated installation sections are stale. Run npm run docs:generate.\n" +
        updates.map((update) => update.path).join("\n"),
    );
  for (const update of updates)
    writeFileSync(inside(root, update.path), update.content);
  const catalog = updates.length === 0 ? initial : createCatalog(root);
  const palette = designPalette(readText(root, "DESIGN.md"));
  const brand = readText(root, logoSource);
  const darkBrand = darkSvg(brand, palette);
  const textColor = palette.find((color) => color.token === "text");
  if (textColor === undefined) throw new Error("DESIGN.md has no text color.");
  const favicon = brand.replace(
    "</title>",
    "</title>\n  <style>@media (prefers-color-scheme: dark) { g { fill: " +
      textColor.dark +
      "; } }</style>",
  );
  const output = join(site, "content");
  const files = new Map<string, string | Buffer>([
    ["public" + logo.light, brand],
    ["public" + logo.dark, darkBrand],
    ["public/favicon.svg", favicon],
    ["index.md", home(catalog)],
    [
      "catalog.json",
      JSON.stringify(
        {entries: catalog.entries, repository: catalog.repository},
        null,
        2,
      ) + "\n",
    ],
    ["theme.css", paletteCss(palette)],
  ]);
  for (const document of catalog.documents)
    files.set(
      document.route.slice(1) + ".md",
      renderDocument(document, catalog),
    );
  for (const category of componentTypes) {
    files.set(
      "catalog/" + category + ".md",
      frontmatter(
        {
          title: categoryNames[category],
          catalog: category,
          outline: false,
          aside: false,
          prev: false,
          next: false,
        },
        "# " +
          categoryNames[category] +
          "\n\n" +
          categoryDescriptions[category],
      ),
    );
  }
  for (const component of catalog.entries.filter((item) =>
    ["hooks", "mcp-servers"].includes(item.type),
  )) {
    const definition = readText(root, component.source);
    files.set(
      component.link.slice(1) + ".md",
      frontmatter(
        {
          title: component.name,
          component: component.id,
          sourcePath: component.source,
        },
        "# " +
          component.name +
          "\n\n" +
          component.description +
          "\n\n```json\n" +
          definition.trim() +
          "\n```\n\n" +
          installation(
            catalog.plugins.find(
              (plugin) => plugin.name === component.plugin,
            ) ??
              (() => {
                throw new Error("Component has no owning plugin.");
              })(),
            catalog.marketplace,
          ),
      ),
    );
  }
  for (const source of sourceAssets(root, catalog)) {
    const content = readFileSync(inside(root, source));
    const destination = "public/repository-assets/" + source;
    files.set(destination, content);
    if (source.includes("plugin-workflow/diagrams/") && source.endsWith(".svg"))
      files.set(
        destination.replace(/\.svg$/u, ".dark.svg"),
        darkSvg(content.toString("utf8"), palette),
      );
  }
  for (const previous of filesBelow(output, "."))
    if (!files.has(previous)) rmSync(inside(output, previous));
  for (const [path, content] of files) writeChanged(output, path, content);
  return {
    catalog,
    changedReadmes: updates.map((update) => update.path),
    pages: catalog.documents.length,
  };
};
