import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";

const UNSAFE_CREDENTIAL_PATTERNS: readonly (readonly [RegExp, string])[] = [
  [
    /\becho\s+(?:["']?\$[A-Z_]*(?:TOKEN|KEY|PAT|SECRET|PASSWORD)["']?|<token>)\s*\|/i,
    "pipes a secret through echo",
  ],
  [/<\s*token\.txt\b/i, "reads a secret from token.txt"],
  [
    /\bexport\s+[A-Z_]*(?:TOKEN|KEY|PAT|SECRET|PASSWORD)\s*=\s*(?!["']?\$)\S+/i,
    "assigns a secret in an export command",
  ],
];

/** One document a reader copies examples from, under the label a finding names. */
export interface LabelledDocument {
  readonly label: string;
  readonly text: string;
}

// A reader copies an example as written, so an example that leaks a secret
// into the shell history or a file teaches the leak.
export const checkCredentialExamples = (
  documents: readonly LabelledDocument[],
): readonly Finding[] => {
  const unsafe = documents.flatMap(({label, text}) =>
    UNSAFE_CREDENTIAL_PATTERNS.filter(([pattern]) => pattern.test(text)).map(
      ([, reason]) => fail(`credentials: ${label} ${reason}`),
    ),
  );
  return unsafe.length > 0
    ? unsafe
    : [pass("credentials: examples avoid plaintext token anti-patterns")];
};
