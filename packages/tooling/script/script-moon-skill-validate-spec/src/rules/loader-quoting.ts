import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import {splitLines} from "../text-lines.js";

const QUOTED_SCALAR_RE = /^[ \t]*[a-z_][\w-]*:[ \t]*(.+?)[ \t]*$/i;

const checkDescriptionQuoting = (
  frontmatterSource: string,
): readonly Finding[] => {
  const descriptionLine = splitLines(frontmatterSource).find((line) =>
    /^description\s*:/.test(line),
  );
  if (descriptionLine === undefined) return [];
  const rawDescription = descriptionLine
    .slice(descriptionLine.indexOf(":") + 1)
    .trim();
  return rawDescription.length >= 2 &&
    rawDescription.startsWith('"') &&
    rawDescription.endsWith('"')
    ? [pass("frontmatter: description uses a double-quoted scalar")]
    : [
        fail(
          "frontmatter: 'description' must use one double-quoted scalar; " +
            "use single quotes for phrases inside it",
        ),
      ];
};

const quotedScalarFinding = (rawLine: string): Finding | undefined => {
  const match = QUOTED_SCALAR_RE.exec(rawLine);
  if (!match) return undefined;
  const rest = match[1] ?? "";
  if (rest.length < 2 || (!rest.startsWith("'") && !rest.startsWith('"')))
    return undefined;
  const quote = rest.startsWith("'") ? "'" : '"';
  // A multiline or unterminated scalar cannot be judged from one line.
  if (!rest.endsWith(quote)) return undefined;
  const key = (rawLine.split(":", 1)[0] ?? "").trim();
  const count = rest.split(quote).length - 1;
  if (quote === "'" && rest.includes("''")) {
    return fail(
      `frontmatter: '${key}' uses single-quote '' escaping, which the skill ` +
        "loader's parser does not support (it reads the tail as an unknown " +
        "attribute). Use a double-quoted scalar with literal apostrophes, e.g. " +
        "\"… doesn't say 'TDD'.\"",
    );
  }
  if (count !== 2) {
    return fail(
      `frontmatter: '${key}' has an inner ${quote} inside a ${quote}-quoted value; ` +
        "the skill loader stops at the first inner quote. Use a double-quoted scalar " +
        "with single quotes for inner phrases (apostrophes stay literal).",
    );
  }
  return undefined;
};

// The skill loader parses frontmatter more naively than YAML does, so a scalar
// that YAML accepts can still be cut short at an inner quote.
export const checkLoaderQuoting = (
  frontmatterSource: string,
): readonly Finding[] => {
  const quoting = splitLines(frontmatterSource).flatMap((line) => {
    const finding = quotedScalarFinding(line);
    return finding === undefined ? [] : [finding];
  });
  return [
    ...checkDescriptionQuoting(frontmatterSource),
    ...(quoting.length > 0
      ? quoting
      : [
          pass(
            "frontmatter: quoting is loader-safe (no '' escapes or inner delimiter quotes)",
          ),
        ]),
  ];
};
