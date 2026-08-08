import {join, relative, resolve, sep} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {scanProsePunctuation} from "@xonovex/script-moon-common/prose-punctuation";
import {checkCommandBudgets} from "./command-budgets.js";
import {
  parseCommandDocument,
  type CommandDocument,
} from "./command-document.js";
import {checkDependencyUsage} from "./dependency-usage.js";
import {checkReadmeCatalog} from "./readme-catalog.js";
import {
  issue,
  type ValidationIssue,
  type ValidationReport,
} from "./validation.js";

interface PackageManifest {
  readonly dependencies?: Readonly<Record<string, string>>;
}

interface PluginManifest {
  readonly dependencies?: readonly string[];
  readonly name?: string;
}

export interface CommandPackageValidation {
  readonly documents: readonly CommandDocument[];
  readonly report: ValidationReport;
}

const readJson = (
  path: string,
  issues: ValidationIssue[],
  fs: FileSystem,
): unknown => {
  try {
    return JSON.parse(fs.readText(path)) as unknown;
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    issues.push(
      issue("manifest.invalid", path, `cannot parse JSON: ${message}`),
    );
    return undefined;
  }
};

const expectedNpmDependency = (plugin: string): string | undefined => {
  const match = /^xonovex-(skill-[a-z0-9-]+)$/u.exec(plugin);
  return match?.[1] === undefined ? undefined : `@xonovex/${match[1]}`;
};

const operationReference = (operation: string): string =>
  `references/${operation
    .trim()
    .toLowerCase()
    .replaceAll(/[^a-z0-9]+/gu, "-")
    .replaceAll(/^-|-$/gu, "")}.md`;

const validateArgumentParity = (
  document: CommandDocument,
  issues: ValidationIssue[],
): void => {
  const argumentKeys = document.arguments.map(
    ({kind, name}) => `${kind}:${name}`,
  );
  for (const duplicate of new Set(
    argumentKeys.filter(
      (argument, index) => argumentKeys.indexOf(argument) !== index,
    ),
  )) {
    const name = duplicate.slice(duplicate.indexOf(":") + 1);
    issues.push(
      issue(
        "command.argument-duplicate",
        document.path,
        `${duplicate.startsWith("flag:") ? "flag" : "positional argument"} '${name}' appears more than once in argument-hint`,
      ),
    );
  }
  const hintedNames = document.arguments.map(({name}) => name);
  const hinted = new Set(hintedNames);
  for (const name of [...hinted].filter(
    (argument) => !document.documentedArguments.has(argument),
  )) {
    issues.push(
      issue(
        "command.argument-undocumented",
        document.path,
        `argument '${name}' appears in argument-hint but not the Arguments section`,
      ),
    );
  }
  for (const name of document.documentedArguments.difference(hinted)) {
    issues.push(
      issue(
        "command.argument-missing-hint",
        document.path,
        `argument '${name}' appears in the Arguments section but not argument-hint`,
      ),
    );
  }
};

const markdownFilesBelow = (
  directory: string,
  fs: FileSystem,
): readonly string[] => {
  if (!fs.isDirectory(directory)) return [];
  return fs.readDirectory(directory).flatMap((name) => {
    const path = join(directory, name);
    if (fs.isDirectory(path)) return markdownFilesBelow(path, fs);
    return fs.isFile(path) && name.endsWith(".md") ? [path] : [];
  });
};

const headingAnchors = (path: string, fs: FileSystem): ReadonlySet<string> => {
  const anchors = new Set<string>();
  for (const match of fs.readText(path).matchAll(/^#{1,6}\s+(.+)$/gmu)) {
    const heading = match[1];
    if (heading === undefined) continue;
    anchors.add(
      heading
        .trim()
        .toLowerCase()
        .replaceAll(/[`*_~]/gu, "")
        .replaceAll(/[^\p{Letter}\p{Number}\s-]/gu, "")
        .replaceAll(/\s+/gu, "-"),
    );
  }
  return anchors;
};

// The scan covers the whole package: commands, docs, README, and the manifests
// whose descriptions reach a marketplace listing.
const validateProsePunctuation = (
  packageDir: string,
  repositoryRoot: string,
  fs: FileSystem,
): readonly ValidationIssue[] =>
  scanProsePunctuation(packageDir, fs).map(
    ({excerpt, hint, label, line, path}) =>
      issue(
        "prose.punctuation",
        relative(repositoryRoot, join(packageDir, path)),
        `${label} on line ${String(line)}, ${hint}: ${excerpt}`,
      ),
  );

const validateInternalLinks = (
  packageDir: string,
  repositoryRoot: string,
  fs: FileSystem,
): readonly ValidationIssue[] => {
  const issues: ValidationIssue[] = [];
  const readme = join(packageDir, "README.md");
  const files = [
    ...(fs.isFile(readme) ? [readme] : []),
    ...markdownFilesBelow(join(packageDir, "commands"), fs),
    ...markdownFilesBelow(join(packageDir, "docs"), fs),
  ];
  for (const file of files) {
    const source = fs.readText(file);
    for (const match of source.matchAll(/\[.*?\]\(([^)\s]+)(?:\s[^)]*)?\)/gu)) {
      const rawTarget = match[1];
      if (
        rawTarget === undefined ||
        /^(?:[a-z]+:|\/)/iu.test(rawTarget) ||
        rawTarget.startsWith("#")
      ) {
        continue;
      }
      const [rawPath = "", rawFragment] = rawTarget.split("#", 2);
      const target = resolve(file, "..", decodeURIComponent(rawPath));
      if (target !== packageDir && !target.startsWith(packageDir + sep)) {
        continue;
      }
      if (!fs.isFile(target) && !fs.isDirectory(target)) {
        issues.push(
          issue(
            "link.missing-target",
            relative(repositoryRoot, file),
            `local link target '${rawTarget}' does not exist`,
          ),
        );
        continue;
      }
      if (
        rawFragment !== undefined &&
        rawFragment.length > 0 &&
        fs.isFile(target) &&
        target.endsWith(".md") &&
        !headingAnchors(target, fs).has(
          decodeURIComponent(rawFragment).toLowerCase(),
        )
      ) {
        issues.push(
          issue(
            "link.missing-fragment",
            relative(repositoryRoot, file),
            `local link fragment '${rawTarget}' does not exist`,
          ),
        );
      }
    }
  }
  return issues;
};

export const validateCommandPackage = (
  packageDir: string,
  repositoryRoot = resolve(packageDir, "../../.."),
  fs: FileSystem = nodeFileSystem,
): CommandPackageValidation => {
  const issues: ValidationIssue[] = [];
  const documents: CommandDocument[] = [];
  const commandDir = join(packageDir, "commands");
  if (!fs.isDirectory(commandDir)) {
    return {
      documents,
      report: {
        commands: 0,
        issues: [
          issue(
            "package.commands-missing",
            relative(repositoryRoot, commandDir),
            "commands directory does not exist",
          ),
        ],
      },
    };
  }

  const packageManifestPath = join(packageDir, "package.json");
  const pluginManifestPath = join(packageDir, ".claude-plugin", "plugin.json");
  const packageManifest = readJson(packageManifestPath, issues, fs) as
    PackageManifest | undefined;
  const pluginManifest = readJson(pluginManifestPath, issues, fs) as
    PluginManifest | undefined;
  const commandFiles = fs
    .readDirectory(commandDir)
    .filter((entry) => entry.endsWith(".md"))
    .toSorted();
  const commandNames = commandFiles.map((entry) =>
    entry.slice(0, -".md".length),
  );
  const delegatedPlugins = new Set<string>();
  issues.push(
    ...validateInternalLinks(packageDir, repositoryRoot, fs),
    ...validateProsePunctuation(packageDir, repositoryRoot, fs),
    ...checkReadmeCatalog(packageDir, commandNames, repositoryRoot, fs),
  );

  for (const commandFile of commandFiles) {
    const absolutePath = join(commandDir, commandFile);
    const displayPath = relative(repositoryRoot, absolutePath);
    const parsed = parseCommandDocument(displayPath, fs.readText(absolutePath));
    issues.push(...parsed.issues);
    const document = parsed.document;
    if (document === undefined) continue;
    documents.push(document);
    validateArgumentParity(document, issues);
    if (document.fileName !== document.command) {
      issues.push(
        issue(
          "command.filename",
          displayPath,
          `filename '${document.fileName}' does not match command '${document.command}'`,
        ),
      );
    }
    if (
      pluginManifest?.name !== undefined &&
      document.namespace !== pluginManifest.name
    ) {
      issues.push(
        issue(
          "command.namespace",
          displayPath,
          `namespace '${document.namespace}' does not match plugin '${pluginManifest.name}'`,
        ),
      );
    }

    const delegation = document.delegation;
    if (delegation === undefined) continue;
    delegatedPlugins.add(delegation.plugin);
    if (!(pluginManifest?.dependencies ?? []).includes(delegation.plugin)) {
      issues.push(
        issue(
          "delegation.plugin-dependency",
          displayPath,
          `plugin dependency '${delegation.plugin}' is not declared`,
        ),
      );
    }
    const npmDependency = expectedNpmDependency(delegation.plugin);
    if (
      npmDependency !== undefined &&
      packageManifest?.dependencies?.[npmDependency] === undefined
    ) {
      issues.push(
        issue(
          "delegation.package-dependency",
          displayPath,
          `package dependency '${npmDependency}' is not declared`,
        ),
      );
    }
    const skillPackage =
      npmDependency === undefined
        ? undefined
        : join(
            repositoryRoot,
            "packages",
            "skill",
            npmDependency.slice("@xonovex/".length),
          );
    const dependencyLabel = npmDependency ?? delegation.plugin;
    if (
      skillPackage !== undefined &&
      !fs.isFile(join(skillPackage, delegation.skill, "SKILL.md"))
    ) {
      issues.push(
        issue(
          "delegation.skill-missing",
          displayPath,
          `delegated skill '${delegation.skill}' does not exist in '${dependencyLabel}'`,
        ),
      );
      continue;
    }
    if (skillPackage !== undefined) {
      const skillPath = join(skillPackage, delegation.skill, "SKILL.md");
      const reference = operationReference(delegation.operation);
      if (!fs.readText(skillPath).includes(`](${reference})`)) {
        issues.push(
          issue(
            "delegation.operation-missing",
            displayPath,
            `delegated operation '${delegation.operation}' is not registered as '${reference}' in '${delegation.skill}'`,
          ),
        );
      }
    }
  }

  issues.push(
    ...checkDependencyUsage({
      packageDirectory: packageDir,
      repositoryRoot,
      commandNames,
      pluginDependencies: pluginManifest?.dependencies ?? [],
      packageDependencies: Object.keys(packageManifest?.dependencies ?? {}),
      delegatedPlugins,
    }),
    ...checkCommandBudgets(packageDir, repositoryRoot),
  );

  return {
    documents,
    report: {commands: commandFiles.length, issues},
  };
};
