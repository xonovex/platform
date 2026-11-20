import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {loadPlugins} from "./catalog.js";
import {inside, readText} from "./files.js";
import {replaceInstallation} from "./installation-section.js";
import {installation} from "./installation.js";
import type {ReadmeUpdate} from "./model.js";

const readmeUpdates = (
  root: string,
  fs: FileSystem = nodeFileSystem,
): readonly ReadmeUpdate[] => {
  const catalog = loadPlugins(root, fs);
  return catalog.plugins.map((plugin) => {
    const path = plugin.directory + "/README.md";
    const previous = readText(root, path, fs);
    return {
      path,
      previous,
      content: replaceInstallation(
        previous,
        installation(plugin, catalog.marketplace),
      ),
    };
  });
};

const driftedReadmes = (
  root: string,
  fs: FileSystem = nodeFileSystem,
): readonly ReadmeUpdate[] =>
  readmeUpdates(root, fs).filter(
    (update) => update.previous !== update.content,
  );

export const syncInstallations = (
  root: string,
  fs: FileSystem = nodeFileSystem,
): readonly string[] => {
  const drifted = driftedReadmes(root, fs);
  for (const update of drifted) {
    fs.writeFile(inside(root, update.path), update.content);
  }
  return drifted.map((update) => update.path);
};

export const checkInstallations = (
  root: string,
  fs: FileSystem = nodeFileSystem,
): readonly string[] => driftedReadmes(root, fs).map((update) => update.path);
