import type {ProsePunctuationFinding} from "@xonovex/script-moon-common/prose-punctuation";
import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";

export const checkProsePunctuation = (
  findings: readonly ProsePunctuationFinding[],
): readonly Finding[] =>
  findings.length === 0
    ? [pass("punctuation: no em dash or ellipsis character")]
    : findings.map(({excerpt, hint, label, line, path}) =>
        fail(
          `punctuation: ${label} in ${path} line ${String(line)}, ${hint}: ${excerpt}`,
        ),
      );
