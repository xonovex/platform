import {basename} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {pluginSkillDirectories} from "@xonovex/script-moon-common/package-layout";
import {pluginHandoffNames} from "@xonovex/script-moon-common/skill-handoffs";
import {z} from "zod";
import {exists, inside, jsonFile} from "./files.js";
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

// withCompanions adds to each plugin the other plugins that own a skill which
// a skill of this plugin hands off to, without the declared dependencies.
const withCompanions = (
  root: string,
  plugins: readonly Plugin[],
  fs: FileSystem,
): readonly Plugin[] => {
  const owners = new Map(
    plugins.flatMap((plugin) =>
      pluginSkillDirectories(inside(root, plugin.directory), fs).map(
        (directory) => [basename(directory), plugin.name] as const,
      ),
    ),
  );
  return plugins.map((plugin) => {
    const handoffs = pluginHandoffNames(inside(root, plugin.directory), fs);
    const companions = new Set(
      [...handoffs].flatMap((skill) => {
        const owner = owners.get(skill);
        return owner === undefined ? [] : [owner];
      }),
    );
    return {
      ...plugin,
      companions: [...companions]
        .filter(
          (name) => name !== plugin.name && !plugin.dependencies.includes(name),
        )
        .toSorted(),
    };
  });
};

export const loadPlugins = (root: string, fs: FileSystem): Catalog => {
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
        companions: [],
        harnesses,
      };
    })
    .toSorted((a, b) => a.name.localeCompare(b.name));
  return {
    marketplace: marketplace.name,
    plugins: withCompanions(root, plugins, fs),
  };
};
