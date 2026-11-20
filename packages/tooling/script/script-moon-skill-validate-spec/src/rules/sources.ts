import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";

const URL_FIELD_RE = /\*\*URLs?:\*\*/i;
const HTTP_URL_RE = /https?:\/\/\S+/i;
const PROVENANCE_RE = /\*\*Provenance:\*\*\s*\S+/i;
const REVIEWED_RE = /\*\*Last reviewed:\*\*\s*\d{4}-\d{2}-\d{2}/i;
const VERSION_FIELD_RE = /\*\*Version:\*\*\s*\S+/i;
const CONTENT_SHA256_FIELD_RE = /\*\*Content SHA256:\*\*\s*[0-9a-f]{64}/i;
const CHECKOUT_FIELD_RE = /\*\*Checkout:\*\*\s*\S+/i;
const COMMIT_FIELD_RE = /\*\*Commit:\*\*\s*[0-9a-f]{7,40}/i;
const WATCH_FIELD_RE = /\*\*Watch:\*\*\s*\S+/i;
const VERSIONED_DESCRIPTION_RE = /\b\d+(?:\.\d+){0,2}\+/;

// A versioned web source needs an anchor the source audit can compare against:
// a hash of the fetched content, or a checkout with a commit and watched paths.
const checkDriftAnchors = (text: string): readonly Finding[] =>
  text
    .split(/^##\s+/m)
    .slice(1)
    .filter((block) => VERSION_FIELD_RE.test(block) && HTTP_URL_RE.test(block))
    .filter(
      (block) =>
        !CONTENT_SHA256_FIELD_RE.test(block) &&
        !(
          CHECKOUT_FIELD_RE.test(block) &&
          COMMIT_FIELD_RE.test(block) &&
          WATCH_FIELD_RE.test(block)
        ),
    )
    .map((block) => {
      const title = block.split(/\r?\n/, 1)[0]?.trim() ?? "unnamed source";
      return fail(
        `catalog: versioned web source '${title}' needs Content SHA256 or Checkout + Commit + Watch drift fields`,
      );
    });

/** The provenance, review date and drift anchors that SOURCES.md records. */
export const checkSources = (
  text: string | undefined,
  description: unknown,
): readonly Finding[] => {
  if (text === undefined) return [fail("catalog: missing SOURCES.md")];
  const hasProvenance =
    (URL_FIELD_RE.test(text) && HTTP_URL_RE.test(text)) ||
    PROVENANCE_RE.test(text);
  const reviewed = REVIEWED_RE.test(text);
  return [
    ...(hasProvenance
      ? []
      : [
          fail(
            "catalog: SOURCES.md needs an HTTP URL/URLs field or an explicit Provenance field",
          ),
        ]),
    ...(reviewed
      ? []
      : [fail("catalog: SOURCES.md needs at least one Last reviewed date")]),
    ...(typeof description === "string" &&
    VERSIONED_DESCRIPTION_RE.test(description) &&
    !VERSION_FIELD_RE.test(text)
      ? [
          fail(
            "catalog: version-pinned skill needs a Version field in SOURCES.md",
          ),
        ]
      : []),
    ...checkDriftAnchors(text),
    ...(hasProvenance && reviewed
      ? [pass("catalog: source provenance and review date are present")]
      : []),
  ];
};
