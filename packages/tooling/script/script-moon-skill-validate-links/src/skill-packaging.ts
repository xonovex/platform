import {join, relative} from "node:path";
import {
  fail,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import type {PluginPackage} from "./catalog-scan.js";
import {
  parseSkillPluginManifest,
  type SkillPluginManifest,
} from "./plugin-manifest.js";

const EXPECTED_SKILLS = ["./skills/"];

export interface PackageCheck {
  readonly findings: readonly Finding[];
  readonly valid: boolean;
  /** Whether both manifests parsed and agree on the plugin name. */
  readonly pair: boolean;
  /** The plugin manifest, when the pair agrees on its name. */
  readonly manifest: SkillPluginManifest | undefined;
}

export const sameStrings = (
  left: readonly string[],
  right: readonly string[],
): boolean => {
  const leftSorted = left.toSorted();
  const rightSorted = right.toSorted();
  return (
    leftSorted.length === rightSorted.length &&
    leftSorted.every((value, index) => value === rightSorted[index])
  );
};

const rejected = (
  findings: readonly Finding[],
  pair: boolean,
): PackageCheck => ({findings, valid: false, pair, manifest: undefined});

const agreementFindings = (
  claude: SkillPluginManifest,
  codex: SkillPluginManifest,
  plugin: PluginPackage,
  displayDirectory: string,
): readonly Finding[] => [
  ...(sameStrings(claude.dependencies, codex.dependencies)
    ? []
    : [
        fail(
          `skill dependencies: manifest dependencies differ for ${claude.name}`,
        ),
      ]),
  ...(sameStrings(claude.skills, codex.skills)
    ? []
    : [
        fail(`skill packaging: manifest skill paths differ for ${claude.name}`),
      ]),
  ...(plugin.skills.length === 0
    ? [fail(`skill packaging: ${displayDirectory} contains no skills`)]
    : []),
  ...(
    [
      ["Claude", claude],
      ["Codex", codex],
    ] as const
  ).flatMap(([kind, manifest]) =>
    sameStrings(manifest.skills, EXPECTED_SKILLS)
      ? []
      : [
          fail(
            `skill packaging: ${kind} manifest in ${displayDirectory} must point directly to ${EXPECTED_SKILLS.join(", ")}`,
          ),
        ],
  ),
];

/**
 * Checks that a plugin package ships a matching Claude and Codex manifest pair
 * that points at its skills, or returns undefined for a directory that is no
 * skill package.
 */
export const checkSkillPackage = (
  plugin: PluginPackage,
  repoRoot: string,
): PackageCheck | undefined => {
  const {claudeManifest, codexManifest, directory} = plugin;
  if (
    claudeManifest === undefined &&
    codexManifest === undefined &&
    !plugin.hasSkillsDirectory
  ) {
    return undefined;
  }
  const displayDirectory = relative(repoRoot, directory);
  if (claudeManifest === undefined || codexManifest === undefined) {
    return rejected(
      [
        fail(
          `skill dependencies: ${displayDirectory} needs both Claude and Codex manifests`,
        ),
      ],
      false,
    );
  }
  const claude = parseSkillPluginManifest(
    claudeManifest,
    relative(repoRoot, join(directory, ".claude-plugin", "plugin.json")),
  );
  const codex = parseSkillPluginManifest(
    codexManifest,
    relative(repoRoot, join(directory, ".codex-plugin", "plugin.json")),
  );
  if (claude.kind === "invalid" || codex.kind === "invalid") {
    return rejected(
      [claude, codex].flatMap((parsed) =>
        parsed.kind === "invalid" ? [fail(parsed.message)] : [],
      ),
      false,
    );
  }

  const expectedName = `xonovex-${plugin.name.replace(/^plugin-/u, "")}`;
  const nameFindings =
    claude.manifest.name === expectedName &&
    codex.manifest.name === expectedName
      ? []
      : [
          fail(
            `skill dependencies: manifests in ${displayDirectory} must be named ${expectedName}`,
          ),
        ];
  if (claude.manifest.name !== codex.manifest.name) {
    return rejected(
      [
        ...nameFindings,
        fail(
          `skill dependencies: manifest names differ in ${displayDirectory} (${claude.manifest.name} != ${codex.manifest.name})`,
        ),
      ],
      true,
    );
  }
  const findings = [
    ...nameFindings,
    ...agreementFindings(
      claude.manifest,
      codex.manifest,
      plugin,
      displayDirectory,
    ),
  ];
  return {
    findings,
    valid: findings.length === 0,
    pair: true,
    manifest: codex.manifest,
  };
};
