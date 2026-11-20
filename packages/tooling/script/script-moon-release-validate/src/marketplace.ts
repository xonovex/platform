import {basename} from "node:path";
import type {Marketplace} from "./release-inputs.js";
import type {Checker} from "./validate.js";

/** The plugin entries of a marketplace, by plugin name. */
export interface MarketplaceEntries {
  readonly sources: ReadonlyMap<string, string | undefined>;
  readonly descriptions: ReadonlyMap<string, string | undefined>;
}

export const marketplaceEntries = (
  marketplace: Marketplace,
): MarketplaceEntries => ({
  sources: new Map(
    marketplace.plugins.map((entry) => [
      entry.name,
      typeof entry.source === "string" ? entry.source : entry.source.path,
    ]),
  ),
  descriptions: new Map(
    marketplace.plugins.map((entry) => [entry.name, entry.description]),
  ),
});

/** The marketplace name of the plugin a package directory publishes. */
export const expectedPluginName = (pluginPackage: string): string => {
  const packageDirectory = basename(pluginPackage);
  return packageDirectory.startsWith("plugin-")
    ? `xonovex-${packageDirectory.slice("plugin-".length)}`
    : `xonovex-${packageDirectory}`;
};

/** Checks that a marketplace lists each expected plugin exactly once. */
export const validateMarketplaceInventory = (
  label: string,
  marketplace: Marketplace,
  expectedNames: readonly string[],
  check: Checker["check"],
): void => {
  const counts = marketplace.plugins.reduce<ReadonlyMap<string, number>>(
    (current, entry) => {
      const next = new Map(current);
      next.set(entry.name, (next.get(entry.name) ?? 0) + 1);
      return next;
    },
    new Map(),
  );
  const expected = new Set(expectedNames);
  const missing = expectedNames.filter((name) => !counts.has(name)).toSorted();
  const unexpected = [...counts.keys()]
    .filter((name) => !expected.has(name))
    .toSorted();
  const duplicates = [...counts.entries()]
    .filter(([, count]) => count > 1)
    .map(([name, count]) => `${name} (${String(count)})`)
    .toSorted();

  check(
    missing.length === 0,
    `${label} is missing plugin entries: ${missing.join(", ")}`,
  );
  check(
    unexpected.length === 0,
    `${label} has unexpected plugin entries: ${unexpected.join(", ")}`,
  );
  check(
    duplicates.length === 0,
    `${label} has duplicate plugin entries: ${duplicates.join(", ")}`,
  );
};
