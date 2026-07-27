import {basename} from "node:path";
import {parse as parseYaml} from "yaml";
import {parseArgumentHint, type CommandArgument} from "./argument-hint.js";
import {issue, type ValidationIssue} from "./validation.js";

const FRONTMATTER_RE = /^---\r?\n([\s\S]*?)\r?\n---\r?\n([\s\S]*)$/u;
const TITLE_RE =
  /^# \/([a-z0-9]+(?:-[a-z0-9]+)*):([a-z0-9]+(?:-[a-z0-9]+)*)\s+—\s+(.+)$/mu;
const DELEGATION_RE =
  /Load the `([^`]+)` skill \(plugin `([^`]+)`\)[\s\S]*?\*\*([^*]+)\*\*/u;
const HEADING_RE = /^## (.+)$/gmu;

export interface CommandFrontmatter {
  readonly allowedTools: readonly string[];
  readonly argumentHint: string;
  readonly description: string;
}

export interface CommandDelegation {
  readonly operation: string;
  readonly plugin: string;
  readonly skill: string;
}

export interface CommandDocument {
  readonly arguments: readonly CommandArgument[];
  readonly body: string;
  readonly command: string;
  readonly delegation: CommandDelegation | undefined;
  readonly documentedArguments: ReadonlySet<string>;
  readonly fileName: string;
  readonly frontmatter: CommandFrontmatter;
  readonly headings: readonly string[];
  readonly namespace: string;
  readonly path: string;
  readonly title: string;
}

export interface CommandDocumentResult {
  readonly document: CommandDocument | undefined;
  readonly issues: readonly ValidationIssue[];
}

const stringArray = (value: unknown): readonly string[] | undefined =>
  Array.isArray(value) && value.every((entry) => typeof entry === "string")
    ? value
    : undefined;

const section = (body: string, heading: string): string => {
  const startMatch = new RegExp(String.raw`^## ${heading}\s*$`, "mu").exec(
    body,
  );
  if (startMatch === null) return "";
  const rest = body.slice(startMatch.index + startMatch[0].length);
  const nextHeading = /^## /mu.exec(rest);
  return nextHeading === null ? rest : rest.slice(0, nextHeading.index);
};

const documentedArguments = (body: string): ReadonlySet<string> => {
  const names = new Set<string>();
  for (const line of section(body, "Arguments").split(/\r?\n/u)) {
    if (!/^\s*-\s+/u.test(line)) continue;
    const separator = line.indexOf(":");
    const modifier = /\s+\((?:optional|required|repeatable)\b/iu.exec(line);
    const boundaries = [separator, modifier?.index ?? -1].filter(
      (index) => index >= 0,
    );
    const end = boundaries.length === 0 ? undefined : Math.min(...boundaries);
    const declaration = end === undefined ? line : line.slice(0, end);
    for (const match of declaration.matchAll(/`([^`]+)`/gu)) {
      const raw = match[1]?.trim();
      if (raw === undefined) continue;
      const flags = [
        ...raw.matchAll(/--([a-z0-9]+(?:-[a-z0-9]+)*)/giu),
      ].flatMap((flagMatch) =>
        flagMatch[1] === undefined ? [] : [flagMatch[1]],
      );
      if (flags.length > 0) {
        for (const flag of flags) names.add(flag);
        continue;
      }
      const positional = raw
        .replace(/^\[/u, "")
        .replace(/\]$/u, "")
        .split(/[ <|]/u, 1)[0];
      if (
        positional !== undefined &&
        /^[a-z0-9][a-z0-9-]*$/iu.test(positional)
      ) {
        names.add(positional);
      }
    }
  }
  return names;
};

export const parseCommandDocument = (
  path: string,
  source: string,
): CommandDocumentResult => {
  const issues: ValidationIssue[] = [];
  const frontmatterMatch = FRONTMATTER_RE.exec(source);
  if (frontmatterMatch === null) {
    return {
      document: undefined,
      issues: [issue("frontmatter.missing", path, "missing YAML frontmatter")],
    };
  }

  let rawFrontmatter: unknown;
  try {
    rawFrontmatter = parseYaml(frontmatterMatch[1] ?? "");
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return {
      document: undefined,
      issues: [
        issue("frontmatter.invalid-yaml", path, `invalid YAML: ${message}`),
      ],
    };
  }
  if (
    typeof rawFrontmatter !== "object" ||
    rawFrontmatter === null ||
    Array.isArray(rawFrontmatter)
  ) {
    return {
      document: undefined,
      issues: [
        issue(
          "frontmatter.invalid-type",
          path,
          "frontmatter must be a mapping",
        ),
      ],
    };
  }

  const values = rawFrontmatter as Readonly<Record<string, unknown>>;
  const description = values.description;
  const argumentHint = values["argument-hint"];
  const allowedTools = stringArray(values["allowed-tools"]);
  if (typeof description !== "string" || description.trim().length === 0) {
    issues.push(
      issue(
        "frontmatter.description",
        path,
        "description must be a non-empty string",
      ),
    );
  }
  if (typeof argumentHint !== "string" || argumentHint.trim().length === 0) {
    issues.push(
      issue(
        "frontmatter.argument-hint",
        path,
        "argument-hint must be a non-empty string",
      ),
    );
  }
  if (allowedTools === undefined || allowedTools.length === 0) {
    issues.push(
      issue(
        "frontmatter.allowed-tools",
        path,
        "allowed-tools must be a non-empty string list",
      ),
    );
  } else if (new Set(allowedTools).size !== allowedTools.length) {
    issues.push(
      issue(
        "frontmatter.allowed-tools-duplicate",
        path,
        "allowed-tools must not contain duplicates",
      ),
    );
  }

  const body = frontmatterMatch[2] ?? "";
  const titleMatch = TITLE_RE.exec(body);
  if (titleMatch === null) {
    issues.push(
      issue(
        "command.title",
        path,
        "expected '# /<namespace>:<command> — <title>'",
      ),
    );
  }
  const headings = [...body.matchAll(HEADING_RE)].flatMap((match) =>
    match[1] === undefined ? [] : [match[1]],
  );
  if (headings.includes("Requirements")) {
    issues.push(
      issue(
        "command.requirements-unsupported",
        path,
        "soft requirements belong in delegation prose and runtime description matching, not a machine-readable Requirements section",
      ),
    );
  }
  for (const requiredHeading of ["Arguments", "Delegation"]) {
    if (
      headings.filter((heading) => heading === requiredHeading).length !== 1
    ) {
      issues.push(
        issue(
          "command.heading",
          path,
          `expected exactly one '${requiredHeading}' section`,
        ),
      );
    }
  }

  const delegationMatch = DELEGATION_RE.exec(body);
  const delegation =
    delegationMatch === null
      ? undefined
      : {
          skill: delegationMatch[1] ?? "",
          plugin: delegationMatch[2] ?? "",
          operation: (delegationMatch[3] ?? "").trim(),
        };
  if (delegation === undefined) {
    issues.push(
      issue(
        "command.delegation",
        path,
        "delegation must name a skill, plugin, and operation",
      ),
    );
  } else if (allowedTools !== undefined && !allowedTools.includes("Skill")) {
    issues.push(
      issue(
        "command.delegation-tool",
        path,
        "a delegated command must allow the Skill tool",
      ),
    );
  }

  if (
    titleMatch === null ||
    typeof description !== "string" ||
    typeof argumentHint !== "string" ||
    allowedTools === undefined
  ) {
    return {document: undefined, issues};
  }

  return {
    document: {
      arguments: parseArgumentHint(argumentHint),
      body,
      command: titleMatch[2] ?? "",
      delegation,
      documentedArguments: documentedArguments(body),
      fileName: basename(path, ".md"),
      frontmatter: {allowedTools, argumentHint, description},
      headings,
      namespace: titleMatch[1] ?? "",
      path,
      title: titleMatch[3] ?? "",
    },
    issues,
  };
};
