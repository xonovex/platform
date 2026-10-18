import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";

export const ROOT = "/repo";

export interface PluginFixture {
  readonly name: string;
  readonly directory: string;
  readonly version: string;
  readonly dependencies?: readonly string[];
  readonly claude: boolean;
  readonly codex: boolean;
  readonly readme?: string;
}

const json = (value: unknown): string => JSON.stringify(value);

const marketplaceEntry = (plugin: PluginFixture) => ({
  name: plugin.name,
  source: "./" + plugin.directory,
});

export const repositoryFiles = (
  plugins: readonly PluginFixture[],
  options: {
    readonly marketplace?: string;
    readonly codexMarketplace?: string;
  } = {},
): Readonly<Record<string, string>> => {
  const marketplace = options.marketplace ?? "xonovex-marketplace";
  const codexMarketplace = options.codexMarketplace ?? marketplace;
  const files: Record<string, string> = {
    [ROOT + "/.claude-plugin/marketplace.json"]: json({
      name: marketplace,
      plugins: plugins.map(marketplaceEntry),
    }),
    [ROOT + "/.agents/plugins/marketplace.json"]: json({
      name: codexMarketplace,
      plugins: plugins.filter((plugin) => plugin.codex).map(marketplaceEntry),
    }),
  };
  for (const plugin of plugins) {
    const manifest = json({
      name: plugin.name,
      version: plugin.version,
      description: "Fixture plugin.",
      dependencies: plugin.dependencies ?? [],
    });
    files[ROOT + "/" + plugin.directory + "/plugin.json"] = manifest;
    if (plugin.claude)
      files[ROOT + "/" + plugin.directory + "/.claude-plugin/plugin.json"] =
        manifest;
    if (plugin.codex)
      files[ROOT + "/" + plugin.directory + "/.codex-plugin/plugin.json"] =
        manifest;
    if (plugin.readme !== undefined)
      files[ROOT + "/" + plugin.directory + "/README.md"] = plugin.readme;
  }
  return files;
};

export const repositoryFileSystem = (
  plugins: readonly PluginFixture[],
  options: {
    readonly marketplace?: string;
    readonly codexMarketplace?: string;
  } = {},
) => memoryFileSystem({files: repositoryFiles(plugins, options)});
