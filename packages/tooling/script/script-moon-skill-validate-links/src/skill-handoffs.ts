import {basename, dirname, relative, sep} from "node:path";
import {boldSkillNameMentions} from "@xonovex/script-moon-common/skill-name";
import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import type {MarkdownDocument} from "./catalog-scan.js";

/**
 * Fails each bold skill name that names no skill among the documents, and
 * passes the set when every named handoff resolves.
 */
export const checkNamedSkillHandoffs = (
  documents: readonly MarkdownDocument[],
  repoRoot: string,
): readonly Finding[] => {
  const knownSkillNames = new Set(
    documents
      .filter(({path}) => path.endsWith(`${sep}SKILL.md`))
      .map(({path}) => basename(dirname(path))),
  );
  const mentions = documents.flatMap((document) =>
    boldSkillNameMentions(document.text).map((name) => ({document, name})),
  );
  const broken = mentions
    .filter(({name}) => !knownSkillNames.has(name))
    .map(({document, name}) =>
      fail(
        `skill handoffs: ${relative(repoRoot, document.path)} names missing **${name}**`,
      ),
    );
  return broken.length === 0 && mentions.length > 0
    ? [
        pass(
          `skill handoffs: ${String(mentions.length)}/${String(mentions.length)} named handoff(s) resolve`,
        ),
      ]
    : broken;
};
