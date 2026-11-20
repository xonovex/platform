import {
  fail,
  pass,
  warn,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";

const HEADING_RE = /^#{1,6} /m;
const CODE_FENCE_OPEN_RE = /^```([\w+-]+)?[^\S\n]*$/gm;

const checkLength = (body: string): Finding => {
  const lineCount = (body.match(/\n/g) ?? []).length;
  if (lineCount > 500) {
    return fail(`body: ${String(lineCount)} lines (>500 spec target)`);
  }
  if (lineCount > 400) {
    return warn(`body: ${String(lineCount)} lines (>80% of 500-line target)`);
  }
  return pass(`body: ${String(lineCount)} lines (under 500)`);
};

const checkCodeFences = (body: string): readonly Finding[] => {
  const fences = [...body.matchAll(CODE_FENCE_OPEN_RE)].map(
    (match) => match[1],
  );
  if (fences.length === 0) return [];
  const opensNoLang = fences.filter((fence) => !fence).length;
  const opensWithLang = fences.length - opensNoLang;
  // Each ``` is either an opening or a closing, and a bare one is either an
  // opening without a language or a closing. Blocks usually come in pairs, so
  // the unmarked openings are about the bare count less the tagged count.
  const unmarked = opensNoLang - opensWithLang;
  return unmarked > 0
    ? [
        warn(
          `body: ${String(unmarked)} code block(s) appear to lack a language marker`,
        ),
      ]
    : [
        pass(
          `body: all code block(s) have language markers ` +
            `(${String(opensWithLang)} language-tagged opening fence(s) found)`,
        ),
      ];
};

/** The SKILL.md body: its length, headings, code fences and gotchas. */
export const checkBody = (body: string): readonly Finding[] => [
  checkLength(body),
  HEADING_RE.test(body)
    ? pass("body: has at least one heading")
    : fail("body: no headings found"),
  ...checkCodeFences(body),
  /^## *Gotchas/m.test(body)
    ? pass("content: '## Gotchas' section present")
    : warn(
        "content: no '## Gotchas' section — non-obvious env-specific facts have no home",
      ),
];
