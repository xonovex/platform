import {parse as parseYaml} from "yaml";

const FRONTMATTER_RE = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/;

export type Frontmatter = Readonly<Record<string, unknown>>;

export interface SkillDocument {
  readonly kind: "parsed";
  readonly frontmatter: Frontmatter;
  /** The frontmatter as written, which the loader-quoting rule reads. */
  readonly frontmatterSource: string;
  readonly body: string;
}

export interface InvalidSkillDocument {
  readonly kind: "invalid";
  readonly message: string;
}

// A SKILL.md without a frontmatter block parses as an empty mapping, so the
// frontmatter rules report each missing field.
export const parseSkillDocument = (
  content: string,
): SkillDocument | InvalidSkillDocument => {
  const match = FRONTMATTER_RE.exec(content);
  if (!match) {
    return {
      kind: "parsed",
      frontmatter: {},
      frontmatterSource: "",
      body: content,
    };
  }
  const frontmatterSource = match[1] ?? "";
  const body = match[2] ?? "";
  let frontmatter: unknown;
  try {
    frontmatter = parseYaml(frontmatterSource) ?? {};
  } catch (error) {
    return {
      kind: "invalid",
      message: `Invalid YAML frontmatter: ${String(error)}`,
    };
  }
  if (
    frontmatter === null ||
    typeof frontmatter !== "object" ||
    Array.isArray(frontmatter)
  ) {
    return {kind: "invalid", message: "Frontmatter is not a mapping"};
  }
  return {
    kind: "parsed",
    frontmatter: frontmatter as Frontmatter,
    frontmatterSource,
    body,
  };
};
