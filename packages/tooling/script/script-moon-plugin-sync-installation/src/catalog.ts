import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {z} from "zod";
import {exists, jsonFile} from "./files.js";
import type {Catalog, Harness, Plugin} from "./model.js";

const manifestSchema = z.looseObject({
  name: z.string(),
  version: z.string(),
  dependencies: z.array(z.string()).default([]),
});
const marketplaceSchema = z.looseObject({
  name: z.string(),
  plugins: z.array(z.looseObject({name: z.string(), source: z.string()})),
});

export const loadPlugins = (
  root: string,
  fs: FileSystem = nodeFileSystem,
): Catalog => {
  const marketplace = jsonFile(
    root,
    ".claude-plugin/marketplace.json",
    marketplaceSchema,
    fs,
  );
  const codex = jsonFile(
    root,
    ".agents/plugins/marketplace.json",
    marketplaceSchema,
    fs,
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
        fs,
      );
      if (manifest.name !== entry.name)
        throw new Error("Marketplace and plugin names differ: " + entry.name);
      const harnesses: Harness[] = [];
      if (exists(root, directory + "/.claude-plugin/plugin.json", fs))
        harnesses.push("Claude Code");
      if (
        codex.plugins.some(
          (item) => item.name === entry.name && item.source === entry.source,
        ) &&
        exists(root, directory + "/.codex-plugin/plugin.json", fs)
      )
        harnesses.push("Codex");
      return {
        directory,
        name: manifest.name,
        version: manifest.version,
        dependencies: manifest.dependencies.toSorted(),
        harnesses,
      };
    })
    .toSorted((a, b) => a.name.localeCompare(b.name));
  return {marketplace: marketplace.name, plugins};
};
