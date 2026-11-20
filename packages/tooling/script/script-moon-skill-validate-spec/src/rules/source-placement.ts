import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import type {ReferenceDocument} from "@xonovex/script-moon-skill-catalog-common/reference-file-links";

const EXTERNAL_REFERENCES_HEADING_RE = /^## +External References\b/im;

// External provenance lives in SOURCES.md, which the source audit reads; a
// list of sources anywhere else escapes that audit.
export const checkSourcePlacement = (
  body: string,
  referenceDocuments: readonly ReferenceDocument[],
): readonly Finding[] => {
  const misplaced = [
    ...(EXTERNAL_REFERENCES_HEADING_RE.test(body) ? ["SKILL.md"] : []),
    ...referenceDocuments
      .filter(({text}) => EXTERNAL_REFERENCES_HEADING_RE.test(text))
      .map(({name}) => `references/${name}`),
  ];
  return misplaced.length === 0
    ? [pass("sources: external provenance is kept in SOURCES.md")]
    : [
        fail(
          "sources: '## External References' belongs in SOURCES.md, not " +
            misplaced.toSorted().join(", "),
        ),
      ];
};
