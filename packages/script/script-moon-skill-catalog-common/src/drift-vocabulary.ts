import {readFileSync} from "node:fs";
import {isFile} from "@xonovex/script-moon-common/fs";
import {z} from "zod";

export const VocabularyManifestSchema = z.record(z.string(), z.string());

export type VocabularyManifest = z.infer<typeof VocabularyManifestSchema>;

export interface VocabularyFile {
  // path is repository-relative and matches the manifest's owner values.
  readonly path: string;
  readonly text: string;
}

export interface VocabularyFinding {
  readonly path: string;
  readonly message: string;
}

const FENCED_BLOCK_RE = /^```[\s\S]+?^```[^\S\n]*$/gm;
const HEADING_RE = /^#{1,6}[^\S\n]+(.+?)[^\S\n]*$/gm;
const BOLD_DEFINITION_RE = /\*\*(.+?)\*\*\s*(?:—|-{1,2}|:|is\b|means\b)/g;
// A coined term of art is a phrase or hyphenated compound. A single bare token
// in backticks is an API identifier, which a skill may name freely.
const CODE_DEFINITION_RE =
  /`([A-Za-z][\w-]*(?:[ -][A-Za-z][\w-]*)+)`\s+(?:is|means|refers to)\s+(?:a|an|the)\b/g;

const normalizeTerm = (term: string): string =>
  term.replaceAll(/[`*_]/g, "").trim().toLowerCase();

const withoutCode = (text: string): string =>
  text.replaceAll(FENCED_BLOCK_RE, "");

// definedTerms collects the terms a file presents as its own: section headings
// and bold or backticked lead-ins that introduce a definition.
export const definedTerms = (text: string): ReadonlySet<string> => {
  const prose = withoutCode(text);
  const terms = new Set<string>();
  for (const match of prose.matchAll(HEADING_RE)) {
    const term = normalizeTerm(match[1] ?? "");
    if (term.length > 0) terms.add(term);
  }
  for (const match of prose.matchAll(BOLD_DEFINITION_RE)) {
    const term = normalizeTerm(match[1] ?? "");
    if (term.length > 0) terms.add(term);
  }
  return terms;
};

// coinedTerms collects backticked phrases a file defines without declaring them
// in the vocabulary manifest. A term of art is used after it is defined, so a
// phrase mentioned exactly once is an example or a command line, not vocabulary.
export const coinedTerms = (text: string): ReadonlySet<string> => {
  const prose = withoutCode(text);
  const terms = new Set<string>();
  for (const match of prose.matchAll(CODE_DEFINITION_RE)) {
    const phrase = match[1] ?? "";
    const term = normalizeTerm(phrase);
    if (term.length === 0) continue;
    const uses = prose.toLowerCase().split(term).length - 1;
    if (uses >= 2) terms.add(term);
  }
  return terms;
};

export const readVocabularyManifest = (
  path: string,
): {readonly manifest: VocabularyManifest; readonly error?: string} => {
  if (!isFile(path)) return {manifest: {}};
  let parsed: unknown;
  try {
    parsed = JSON.parse(readFileSync(path, "utf8"));
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    return {
      error: `vocabulary manifest is not valid JSON: ${detail}`,
      manifest: {},
    };
  }
  const result = VocabularyManifestSchema.safeParse(parsed);
  if (!result.success) {
    return {
      error: "vocabulary manifest entries must map a term to its owning file",
      manifest: {},
    };
  }
  return {manifest: result.data};
};

// evaluateVocabulary reports files that redefine a term another file owns, and
// definitions of terms the manifest never declared.
export const evaluateVocabulary = (
  files: readonly VocabularyFile[],
  manifest: VocabularyManifest,
): readonly VocabularyFinding[] => {
  const owners = new Map(
    Object.entries(manifest).map(([term, owner]) => [
      normalizeTerm(term),
      owner,
    ]),
  );
  return files.flatMap((file) => {
    const redefinitions = [...definedTerms(file.text)].flatMap((term) => {
      const owner = owners.get(term);
      if (owner === undefined || owner === file.path) return [];
      return [
        {
          message:
            `vocabulary: ${file.path} redefines '${term}', which ` +
            `${owner} owns; use the term and cross-reference the owner`,
          path: file.path,
        },
      ];
    });
    const undeclared = [...coinedTerms(file.text)].flatMap((term) =>
      owners.has(term)
        ? []
        : [
            {
              message:
                `vocabulary: ${file.path} defines the coined term '${term}' ` +
                `without declaring it in the vocabulary manifest`,
              path: file.path,
            },
          ],
    );
    return [...redefinitions, ...undeclared];
  });
};
