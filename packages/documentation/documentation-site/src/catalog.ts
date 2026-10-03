import {basename, dirname} from "node:path";
import {z} from "zod";
import {
  documentTitle,
  exists,
  filesBelow,
  jsonFile,
  markdown,
  readText,
} from "./files.js";
import type {
  Catalog,
  Component,
  ComponentType,
  Document,
  Harness,
  Plugin,
} from "./model.js";

const manifestSchema = z.looseObject({
  name: z.string(),
  version: z.string(),
  description: z.string().default(""),
  skills: z.union([z.string(), z.array(z.string())]).optional(),
  hooks: z.union([z.string(), z.record(z.string(), z.unknown())]).optional(),
  mcpServers: z
    .union([z.string(), z.record(z.string(), z.unknown())])
    .optional(),
  dependencies: z.array(z.string()).default([]),
});
const marketplaceSchema = z.looseObject({
  name: z.string(),
  plugins: z.array(z.looseObject({name: z.string(), source: z.string()})),
});
const packageSchema = z.looseObject({
  name: z.string(),
  version: z.string(),
  description: z.string().optional(),
});
const configurationSchema = z.record(z.string(), z.unknown());

const firstParagraph = (body: string): string =>
  body
    .split(/\n\s*\n/u)
    .find(
      (paragraph) =>
        paragraph.trim() !== "" &&
        !paragraph.startsWith("#") &&
        !paragraph.startsWith("|"),
    )
    ?.replaceAll(/\s+/gu, " ")
    .trim() ?? "";

const loadPlugins = (
  root: string,
): {readonly marketplace: string; readonly plugins: readonly Plugin[]} => {
  const marketplace = jsonFile(
    root,
    ".claude-plugin/marketplace.json",
    marketplaceSchema,
  );
  const codex = jsonFile(
    root,
    ".agents/plugins/marketplace.json",
    marketplaceSchema,
  );
  if (codex.name !== marketplace.name)
    throw new Error("Harness marketplace names differ.");
  const plugins = marketplace.plugins
    .map((entry): Plugin => {
      const directory = entry.source.replace(/^\.\//u, "");
      const manifest = jsonFile(
        root,
        directory + "/plugin.json",
        manifestSchema,
      );
      if (manifest.name !== entry.name)
        throw new Error("Marketplace and plugin names differ: " + entry.name);
      const harnesses: Harness[] = [];
      if (exists(root, directory + "/.claude-plugin/plugin.json"))
        harnesses.push("Claude Code");
      if (
        codex.plugins.some(
          (item) => item.name === entry.name && item.source === entry.source,
        ) &&
        exists(root, directory + "/.codex-plugin/plugin.json")
      )
        harnesses.push("Codex");
      return {
        directory,
        name: manifest.name,
        description: manifest.description,
        version: manifest.version,
        dependencies: manifest.dependencies.toSorted(),
        harnesses,
      };
    })
    .toSorted((a, b) => a.name.localeCompare(b.name));
  return {marketplace: marketplace.name, plugins};
};

const entry = (
  plugin: Plugin,
  type: ComponentType,
  name: string,
  source: string,
  description: string,
  harnesses: readonly Harness[] = plugin.harnesses,
): Component => ({
  id: type + ":" + plugin.name + ":" + name,
  type,
  name,
  description,
  source,
  link:
    type === "plugins"
      ? "/plugins/" + name
      : "/" + type + "/" + plugin.name + "/" + name,
  plugin: plugin.name,
  version: plugin.version,
  harnesses,
  dependencies: plugin.dependencies,
});

const markdownComponents = (
  root: string,
  plugin: Plugin,
  type: "skills" | "commands" | "agents",
): readonly Component[] => {
  const paths = filesBelow(root, plugin.directory + "/" + type).filter(
    (path) =>
      type === "skills"
        ? path.endsWith("/SKILL.md")
        : path.endsWith(".md") && basename(path) !== "README.md",
  );
  return paths.map((source) => {
    const metadata = markdown(readText(root, source));
    const name =
      metadata.name ??
      (type === "skills" ? basename(dirname(source)) : basename(source, ".md"));
    const harnesses =
      type === "skills"
        ? plugin.harnesses
        : plugin.harnesses.filter((harness) => harness === "Claude Code");
    return entry(
      plugin,
      type,
      name,
      source,
      metadata.description ?? firstParagraph(metadata.body),
      harnesses,
    );
  });
};

const configuredComponents = (
  root: string,
  plugin: Plugin,
): readonly Component[] => {
  const configurations = plugin.harnesses.flatMap((harness) => {
    const path =
      plugin.directory +
      (harness === "Codex"
        ? "/.codex-plugin/plugin.json"
        : "/.claude-plugin/plugin.json");
    const manifest = jsonFile(root, path, manifestSchema);
    return (["hooks", "mcpServers"] as const).flatMap((field) => {
      const declaration = manifest[field];
      const fallback = field === "hooks" ? "hooks/hooks.json" : ".mcp.json";
      const source =
        typeof declaration === "string"
          ? plugin.directory + "/" + declaration.replace(/^\.\//u, "")
          : plugin.directory + "/" + fallback;
      const fileConfiguration = exists(root, source)
        ? jsonFile(root, source, configurationSchema)
        : undefined;
      const raw =
        typeof declaration === "object" ? declaration : fileConfiguration;
      if (raw === undefined) return [];
      const map = configurationSchema.parse(raw[field] ?? raw);
      const type = field === "hooks" ? "hooks" : "mcp-servers";
      return Object.keys(map)
        .toSorted()
        .map((name) =>
          entry(
            plugin,
            type,
            name,
            typeof declaration === "object" ? path : source,
            field === "hooks"
              ? "Lifecycle hook for " + name + "."
              : "Model Context Protocol server: " + name + ".",
            [harness],
          ),
        );
    });
  });
  return configurations.reduce<readonly Component[]>((result, component) => {
    const previous = result.find((item) => item.id === component.id);
    return previous === undefined
      ? [...result, component]
      : result.map((item) =>
          item.id === component.id
            ? {
                ...item,
                harnesses: [
                  ...new Set([...item.harnesses, ...component.harnesses]),
                ],
              }
            : item,
        );
  }, []);
};

const toolComponents = (
  root: string,
  sources: readonly string[],
): readonly Component[] =>
  sources
    .filter((source) =>
      /^packages\/(?:tooling\/cli|sandbox\/operator)\/[^/]+\/README\.md$/u.test(
        source,
      ),
    )
    .map((source) => {
      const directory = dirname(source);
      const body = markdown(readText(root, source)).body;
      const metadata = exists(root, directory + "/package.json")
        ? jsonFile(root, directory + "/package.json", packageSchema)
        : undefined;
      const name = documentTitle(body, basename(directory));
      return {
        id: "tools:" + basename(directory),
        type: "tools",
        name,
        description: metadata?.description ?? firstParagraph(body),
        source,
        link: "/tools/" + basename(directory),
        plugin: "",
        version: metadata?.version ?? "",
        harnesses: [],
        dependencies: [],
      };
    });

export const routeFor = (
  source: string,
  entries: readonly Component[],
): string => {
  const component = entries.find((item) => item.source === source);
  if (component !== undefined) return component.link;
  if (source === "README.md") return "/guide/platform";
  if (source === "CONTRIBUTING.md") return "/guide/contributing";
  if (source === "DESIGN.md") return "/guide/design";
  return "/reference/" + source.replace(/\.md$/u, "");
};

export const createCatalog = (root: string): Catalog => {
  const {marketplace, plugins} = loadPlugins(root);
  const sources = [
    ...filesBelow(root, "packages"),
    ...["README.md", "CONTRIBUTING.md", "DESIGN.md", "AGENTS.md"].filter(
      (path) => exists(root, path),
    ),
  ].filter(
    (path) =>
      path.endsWith(".md") &&
      !path.endsWith("/CLAUDE.md") &&
      !path.endsWith("/CHANGELOG.md"),
  );
  const entries = [
    ...plugins.flatMap((plugin) => [
      entry(
        plugin,
        "plugins",
        plugin.name,
        plugin.directory + "/README.md",
        plugin.description,
      ),
      ...markdownComponents(root, plugin, "skills"),
      ...markdownComponents(root, plugin, "commands"),
      ...markdownComponents(root, plugin, "agents"),
      ...configuredComponents(root, plugin),
    ]),
    ...toolComponents(root, sources),
  ].toSorted((a, b) => a.id.localeCompare(b.id));
  if (new Set(entries.map((item) => item.id)).size !== entries.length)
    throw new Error("Duplicate component identifiers in the catalog.");
  const documents = sources.toSorted().map((source): Document => {
    const parsed = markdown(readText(root, source));
    return {
      source,
      route: routeFor(source, entries),
      title: documentTitle(parsed.body, basename(source, ".md")),
      description: parsed.description ?? firstParagraph(parsed.body),
      body: parsed.body,
    };
  });
  if (new Set(documents.map((item) => item.route)).size !== documents.length)
    throw new Error("Duplicate documentation routes.");
  return {
    marketplace,
    repository: "https://github.com/xonovex/platform",
    plugins,
    entries,
    documents,
  };
};
