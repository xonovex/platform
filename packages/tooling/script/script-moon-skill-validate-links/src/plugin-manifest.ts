import {isRecord} from "@xonovex/script-moon-common/records";
import {parseJsonFile, type FileRead} from "./file-read.js";

export interface SkillPluginManifest {
  readonly name: string;
  readonly dependencies: readonly string[];
  readonly skills: readonly string[];
}

const isStringArray = (value: unknown): value is readonly string[] =>
  Array.isArray(value) && value.every((entry) => typeof entry === "string");

export type ParsedManifest =
  | {readonly kind: "parsed"; readonly manifest: SkillPluginManifest}
  | {readonly kind: "invalid"; readonly message: string};

/** Parses a Claude or Codex plugin manifest; displayPath names it in a message. */
export const parseSkillPluginManifest = (
  file: FileRead,
  displayPath: string,
): ParsedManifest => {
  const parsed = parseJsonFile(file);
  if (parsed.kind === "invalid") {
    return {
      kind: "invalid",
      message: `skill dependencies: invalid JSON in ${displayPath}: ${parsed.message}`,
    };
  }
  const {value} = parsed;
  if (!isRecord(value) || typeof value.name !== "string") {
    return {
      kind: "invalid",
      message: `skill dependencies: ${displayPath} needs a string name`,
    };
  }
  const dependencies = value.dependencies ?? [];
  if (!isStringArray(dependencies)) {
    return {
      kind: "invalid",
      message: `skill dependencies: ${displayPath} dependencies must be strings`,
    };
  }
  if (typeof value.skills !== "string") {
    return {
      kind: "invalid",
      message: `skill packaging: ${displayPath} skills must be a string`,
    };
  }
  return {
    kind: "parsed",
    manifest: {
      name: value.name,
      dependencies,
      skills: [value.skills],
    },
  };
};
