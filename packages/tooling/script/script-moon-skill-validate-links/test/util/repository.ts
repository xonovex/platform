import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";

export const REPO = "/repo";

// A plugin without its own package.json gets one that pins each manifest
// dependency at the version every generated package declares.
const generatedPackageFiles = (
  files: Readonly<Record<string, string>>,
): Record<string, string> =>
  Object.fromEntries(
    Object.entries(files).flatMap(([path, content]) => {
      const match =
        /^(packages\/runtime\/plugin\/plugin-[^/]+)\/\.claude-plugin\/plugin\.json$/u.exec(
          path,
        );
      if (match?.[1] === undefined) return [];
      const packagePath = `${match[1]}/package.json`;
      if (files[packagePath] !== undefined) return [];
      const plugin = JSON.parse(content) as {
        readonly dependencies?: readonly string[];
        readonly name: string;
      };
      return [
        [
          packagePath,
          JSON.stringify({
            name: `@xonovex/${plugin.name.replace(/^xonovex-/u, "plugin-")}`,
            version: "7.0.0",
            dependencies: Object.fromEntries(
              (plugin.dependencies ?? []).map((dependency) => [
                `@xonovex/${dependency.replace(/^xonovex-/u, "plugin-")}`,
                "7.0.0",
              ]),
            ),
          }),
        ],
      ];
    }),
  );

/** A repository in memory whose files sit at the given repository-relative paths. */
export const makeRepo = (files: Readonly<Record<string, string>>): FileSystem =>
  memoryFileSystem({
    files: Object.fromEntries(
      Object.entries({...generatedPackageFiles(files), ...files}).map(
        ([path, content]) => [join(REPO, path), content],
      ),
    ),
  });

export const manifest = (
  name: string,
  dependencies: readonly string[] = [],
  skills = "./skills/",
): string => JSON.stringify({name, dependencies, skills});

/** A plugin package with one skill and an agreeing manifest pair. */
export const plugin = (
  name: string,
  dependencies: readonly string[] = [],
  skillText = `# ${name}\n`,
): Record<string, string> => {
  const base = `packages/runtime/plugin/plugin-${name}`;
  return {
    [`${base}/skills/using-${name}/SKILL.md`]: skillText,
    [`${base}/.claude-plugin/plugin.json`]: manifest(
      `xonovex-${name}`,
      dependencies,
    ),
    [`${base}/.codex-plugin/plugin.json`]: manifest(
      `xonovex-${name}`,
      dependencies,
    ),
  };
};
