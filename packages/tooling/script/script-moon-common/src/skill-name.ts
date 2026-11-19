// A skill name is a gerund and then its object, as in using-git. The gerund is
// what tells a skill name from another kebab-case token in prose.
const SKILL_NAME = "[a-z0-9]+ing(?:-[a-z0-9]+)+";

const SKILL_NAME_RE = new RegExp(`^${SKILL_NAME}$`, "u");
const MENTION_RE = new RegExp(
  `(?<![a-z0-9-])${SKILL_NAME}(?![a-z0-9-])`,
  "giu",
);
const BOLD_MENTION_RE = new RegExp(
  `\\*\\*(${SKILL_NAME})\\*\\*(?!\\s+operation\\b)`,
  "gu",
);

export const SKILL_NAME_RULE = `^${SKILL_NAME.replace("?:", "")}$`;

export const isSkillName = (name: string): boolean => SKILL_NAME_RE.test(name);

// skillNameMentions returns each token of the text that has the form of a
// skill name, in lower case and in the order of the text.
export const skillNameMentions = (text: string): readonly string[] =>
  [...text.matchAll(MENTION_RE)].map(([mention]) => mention.toLowerCase());

// boldSkillNameMentions returns the skill names that the Markdown sets in bold,
// which is the notation of a handoff to a skill by name. A bold name before the
// word operation names an operation of a skill, and an operation can have a
// gerund name too, so it is not a handoff.
export const boldSkillNameMentions = (text: string): readonly string[] =>
  [...text.matchAll(BOLD_MENTION_RE)].flatMap(([, name]) =>
    name === undefined ? [] : [name],
  );
