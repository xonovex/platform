import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import type {Frontmatter} from "../skill-document.js";

/** The files of a skill that hold code rather than prose. */
export interface ImplementationFiles {
  /** Whether scripts/ holds at least one file. */
  readonly hasScripts: boolean;
  /** The TypeScript and MJS files below the skill, relative to it. */
  readonly forbidden: readonly string[];
}

const isNonEmptyString = (value: unknown): value is string =>
  typeof value === "string" && value.trim().length > 0;

const checkFormats = (forbidden: readonly string[]): readonly Finding[] =>
  forbidden.length === 0
    ? []
    : [
        fail(
          `catalog: skills must not contain TypeScript or MJS implementation files: ${forbidden.join(", ")}`,
        ),
      ];

// A skill that ships scripts runs code, so it states what it needs to run and
// which tools it may call.
const checkScriptMetadata = (frontmatter: Frontmatter): readonly Finding[] => {
  const hasCompatibility = isNonEmptyString(frontmatter.compatibility);
  const hasAllowedTools = isNonEmptyString(frontmatter["allowed-tools"]);
  return [
    ...(hasCompatibility
      ? []
      : [
          fail(
            "catalog: scripted skills need compatibility runtime/network metadata",
          ),
        ]),
    ...(hasAllowedTools
      ? []
      : [
          fail(
            "catalog: scripted skills need a non-empty allowed-tools policy",
          ),
        ]),
    ...(hasCompatibility && hasAllowedTools
      ? [
          pass(
            "catalog: scripted skill declares compatibility and allowed-tools",
          ),
        ]
      : []),
  ];
};

export const checkImplementationFiles = (
  files: ImplementationFiles,
  frontmatter: Frontmatter,
): readonly Finding[] => [
  ...checkFormats(files.forbidden),
  ...(files.hasScripts ? checkScriptMetadata(frontmatter) : []),
];
