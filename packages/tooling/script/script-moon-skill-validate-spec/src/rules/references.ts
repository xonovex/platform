import {join} from "node:path";
import {
  fail,
  pass,
  warn,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import type {ReferenceDocument} from "@xonovex/script-moon-skill-catalog-common/reference-file-links";
import {KEBAB_CASE_RE, splitLines} from "../text-lines.js";

const REF_LINK_RE = /references\/([\w./<>{}-]+\.md)/g;
const REF_LINE_RE = /references\/[\w./<>{}-]+\.md/;
const PROGRESSIVE_DISCLOSURE_RE =
  /^## *Progressive Disclosure\s*\n([\s\S]*?)(?=^## |(?![\s\S]))/im;
const LOAD_WHEN_RE = /load when/i;
const TOC_RE = /^## *(?:Contents|Table of Contents)\b/im;

const isPlaceholder = (ref: string): boolean =>
  ref.includes("<") || ref.includes("{");

const checkReferenceLink = (
  ref: string,
  skillDir: string,
  isFile: (path: string) => boolean,
): readonly Finding[] => {
  const lastSegment = ref.includes("/")
    ? ref.slice(ref.lastIndexOf("/") + 1)
    : ref;
  const fileName = lastSegment.endsWith(".md")
    ? lastSegment.slice(0, -".md".length)
    : lastSegment;
  return [
    ...(isFile(join(skillDir, "references", ref))
      ? []
      : [fail(`references: broken link → references/${ref}`)]),
    ...(ref.includes("/")
      ? [
          warn(
            `references: 'references/${ref}' is nested deeper than one level`,
          ),
        ]
      : []),
    ...(KEBAB_CASE_RE.test(fileName)
      ? []
      : [warn(`references: 'references/${ref}' filename is not kebab-case`)]),
  ];
};

// Only the Progressive Disclosure section tells an agent when to load a file.
const checkLoadTriggers = (body: string): readonly Finding[] => {
  const section = PROGRESSIVE_DISCLOSURE_RE.exec(body)?.[1];
  if (section === undefined) return [];
  const links = splitLines(section).filter((line) => REF_LINE_RE.test(line));
  if (links.length === 0) return [];
  const missing = links.filter((line) => !LOAD_WHEN_RE.test(line)).length;
  return missing > 0
    ? [
        warn(
          `references: ${String(missing)} of ${String(links.length)} link(s) in Progressive ` +
            "Disclosure lack a 'Load when…' trigger",
        ),
      ]
    : [
        pass(
          `references: all ${String(links.length)} link(s) in Progressive Disclosure ` +
            "carry a 'Load when…' trigger",
        ),
      ];
};

/** The references/ links of the SKILL.md body. */
export const checkReferences = (
  body: string,
  skillDir: string,
  isFile: (path: string) => boolean,
): readonly Finding[] => {
  const refs = [
    ...new Set(
      [...body.matchAll(REF_LINK_RE)].flatMap((match) =>
        match[1] === undefined ? [] : [match[1]],
      ),
    ),
  ].toSorted();
  if (refs.length === 0) return [];

  const real = refs.filter((ref) => !isPlaceholder(ref));
  const linkFindings = real.flatMap((ref) =>
    checkReferenceLink(ref, skillDir, isFile),
  );
  const broken = linkFindings.filter(({level}) => level === "fail").length;
  return [
    ...linkFindings,
    ...(broken === 0 && real.length > 0
      ? [
          pass(
            `references: ${String(real.length)}/${String(real.length)} link(s) resolve`,
          ),
        ]
      : []),
    // An @references/ prefix makes a harness load the file eagerly.
    body.includes("@references/")
      ? fail(
          "references: '@references/' prefix found — strips progressive disclosure " +
            "(use plain 'references/')",
        )
      : pass("references: no '@references/' prefix"),
    ...checkLoadTriggers(body),
  ];
};

/** A long reference file opens with a table of contents. */
export const checkReferenceTocs = (
  documents: readonly ReferenceDocument[],
): readonly Finding[] => {
  const long = documents
    .map(({name, text}) => ({name, lines: splitLines(text)}))
    .filter(({lines}) => lines.length > 200);
  if (long.length === 0) return [];
  const missing = long
    .filter(({lines}) => !TOC_RE.test(lines.slice(0, 15).join("\n")))
    .map(({name, lines}) => `${name} (${String(lines.length)} lines)`);
  return missing.length > 0
    ? [
        warn(
          "references: file(s) >200 lines lack a '## Contents' table of contents — " +
            missing.join(", "),
        ),
      ]
    : [
        pass(
          `references: all ${String(long.length)} reference file(s) >200 lines open with a ` +
            "'## Contents' table of contents",
        ),
      ];
};
