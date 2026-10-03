import {watch, type FSWatcher} from "node:fs";
import {basename, dirname, extname, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {createServer, disposeMdItInstance} from "vitepress";
import {filesBelow} from "./files.js";
import {generateSite} from "./site.js";

const startDevelopment = async (root: string, site: string): Promise<void> => {
  const watchers = new Map<string, FSWatcher>();
  let server: Awaited<ReturnType<typeof createServer>> | undefined;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let pending = Promise.resolve();
  const restart = (): Promise<void> => {
    pending = pending
      .then(async () => {
        await server?.close();
        disposeMdItInstance();
        generateSite(root, site, false);
        server = await createServer(
          site,
          {
            watch: {
              ignored: [
                site + "/content/**",
                site + "/src/**",
                site + "/.vitepress/**",
              ],
            },
          },
          restart,
        );
        await server.listen();
        server.printUrls();
        refreshWatchers();
      })
      .catch((error: unknown) => {
        console.error(error instanceof Error ? error.message : String(error));
        process.exitCode = 1;
      });
    return pending;
  };
  const changed = (path: string, event: string): void => {
    if (path.includes(".timestamp-")) return;
    if (
      ["content", "dist", "target", "node_modules", "cache", ".temp"].includes(
        basename(path),
      )
    )
      return;
    if (
      !path.endsWith(".md") &&
      !path.endsWith("/plugin.json") &&
      !path.endsWith("/package.json") &&
      !path.endsWith("/marketplace.json") &&
      !path.endsWith(".json") &&
      ![
        ".svg",
        ".png",
        ".jpg",
        ".jpeg",
        ".gif",
        ".webp",
        ".dot",
        ".pdf",
      ].includes(extname(path)) &&
      !(event === "rename" && extname(path) === "") &&
      !path.startsWith(site + "/src/") &&
      !path.startsWith(site + "/.vitepress/")
    )
      return;
    clearTimeout(timer);
    timer = setTimeout(() => {
      void restart();
    }, 150);
  };
  const refreshWatchers = (): void => {
    const directories = new Set([
      root,
      resolve(root, "packages"),
      resolve(root, ".claude-plugin"),
      resolve(root, ".agents/plugins"),
      resolve(site, ".vitepress"),
      resolve(site, ".vitepress/theme"),
      ...filesBelow(root, "packages").flatMap((path) => {
        const parents: string[] = [];
        let directory = dirname(resolve(root, path));
        while (directory !== root) {
          parents.push(directory);
          directory = dirname(directory);
        }
        return parents;
      }),
    ]);
    for (const [directory, watcher] of watchers) {
      if (directories.has(directory)) continue;
      watcher.close();
      watchers.delete(directory);
    }
    for (const directory of directories) {
      if (watchers.has(directory)) continue;
      watchers.set(
        directory,
        watch(directory, (_event, filename) => {
          if (filename !== null) changed(resolve(directory, filename), _event);
        }),
      );
    }
  };
  const stop = (): void => {
    clearTimeout(timer);
    for (const watcher of watchers.values()) watcher.close();
    void pending.then(async () => {
      await server?.close();
    });
  };
  process.once("SIGINT", stop);
  process.once("SIGTERM", stop);
  await restart();
};

const site = resolve(dirname(fileURLToPath(import.meta.url)), "..");
await startDevelopment(resolve(site, "../../.."), site);
