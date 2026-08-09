import {readFileSync} from "node:fs";
import {join, relative, sep} from "node:path";
import type {Linter} from "eslint";

const ALLOWLIST_FILE = "filesystem-allowlist.json";

const WORKSPACE_ROOT = import.meta.dirname;

const MESSAGE =
  "reach the filesystem through the FileSystem port from " +
  "@xonovex/script-moon-common/file-system, taken as a defaulted last " +
  `parameter, or add this file to ${ALLOWLIST_FILE} with the reason it cannot`;

// The same manifest script-moon-release-validate reads, so the editor rule and
// the release gate cannot disagree about which modules may reach node:fs.
const allowedPaths = (): readonly string[] => {
  const parsed: unknown = JSON.parse(
    readFileSync(join(WORKSPACE_ROOT, ALLOWLIST_FILE), "utf8"),
  );
  return typeof parsed === "object" && parsed !== null && !Array.isArray(parsed)
    ? Object.keys(parsed)
    : [];
};

/**
 * The rules this repository adds on top of the published config: a module under
 * src reaches the filesystem through the FileSystem port rather than node:fs.
 *
 * The exemptions are scoped to the package being linted, because eslint runs
 * inside each package and a bare `src/index.ts` pattern would otherwise excuse
 * that path in every package rather than the one the allowlist names.
 */
export const workspaceRules = (
  packageDirectory: string,
): readonly Linter.Config[] => {
  const packagePath = relative(WORKSPACE_ROOT, packageDirectory)
    .split(sep)
    .join("/");
  const exemptions = allowedPaths()
    .filter((path) => path.startsWith(`${packagePath}/`))
    .map((path) => path.slice(packagePath.length + 1));

  return [
    {
      files: ["src/**/*.ts"],
      rules: {
        "no-restricted-imports": [
          "error",
          {
            paths: [
              {name: "node:fs", message: MESSAGE},
              {name: "node:fs/promises", message: MESSAGE},
            ],
          },
        ],
      },
    },
    ...(exemptions.length === 0
      ? []
      : [{files: exemptions, rules: {"no-restricted-imports": "off"}}]),
  ] as readonly Linter.Config[];
};
