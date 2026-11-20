import {
  fail,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";

/** A catalog file that exists, as its text or as the error that reading it raised. */
export type CatalogFile =
  | {readonly kind: "read"; readonly text: string}
  | {readonly kind: "unreadable"; readonly message: string};

export type ParsedJson =
  | {readonly kind: "parsed"; readonly value: unknown}
  | {readonly kind: "invalid"; readonly finding: Finding};

const invalid = (name: string, message: string): ParsedJson => ({
  kind: "invalid",
  finding: fail(`catalog: ${name} is not valid JSON: ${message}`),
});

export const parseCatalogJson = (
  name: string,
  file: CatalogFile,
): ParsedJson => {
  if (file.kind === "unreadable") return invalid(name, file.message);
  try {
    return {kind: "parsed", value: JSON.parse(file.text) as unknown};
  } catch (error) {
    return invalid(
      name,
      error instanceof Error ? error.message : String(error),
    );
  }
};
