export interface DuplicationFile {
  // path is repository-relative and identifies the file in findings.
  readonly path: string;
  readonly text: string;
}

export interface DuplicationFinding {
  readonly paths: readonly string[];
  readonly message: string;
  readonly sentence: string;
}

// A normative sentence states a rule; only those are worth de-duplicating.
const NORMATIVE_RE =
  /\b(?:must|never|always|require[sd]?|shall|may not|cannot|do not|don't)\b/i;
const FENCED_BLOCK_RE = /^```[\s\S]+?^```[^\S\n]*$/gm;
const SENTENCE_SPLIT_RE = /(?<=[.!?])\s+|\n{2,}/;

export const DUPLICATION_SIMILARITY_THRESHOLD = 0.75;
export const DUPLICATION_MIN_FILES = 3;
const MIN_SENTENCE_WORDS = 6;
// Bigrams tolerate the single-word substitutions a restated invariant picks up;
// trigrams miss them while offering no extra separation from unrelated rules.
const NGRAM_SIZE = 2;

const normalizeWords = (sentence: string): readonly string[] =>
  sentence
    .toLowerCase()
    .replaceAll(/[`*_[\]()]/g, "")
    .replaceAll(/[^\da-z\s-]/g, " ")
    .split(/\s+/u)
    .filter((word) => word.length > 0);

export const normativeSentences = (text: string): readonly string[] =>
  text
    .replaceAll(FENCED_BLOCK_RE, "")
    .split(SENTENCE_SPLIT_RE)
    .map((sentence) => sentence.replaceAll(/\s+/gu, " ").trim())
    .filter(
      (sentence) =>
        NORMATIVE_RE.test(sentence) &&
        normalizeWords(sentence).length >= MIN_SENTENCE_WORDS,
    );

const ngrams = (words: readonly string[]): ReadonlySet<string> => {
  const grams = new Set<string>();
  const size = Math.min(NGRAM_SIZE, words.length);
  for (let index = 0; index + size <= words.length; index += 1) {
    grams.add(words.slice(index, index + size).join(" "));
  }
  return grams;
};

// similarity is Jaccard overlap over word n-grams, which tolerates the small
// rewordings that restated invariants pick up as they spread.
export const similarity = (left: string, right: string): number => {
  const leftGrams = ngrams(normalizeWords(left));
  const rightGrams = ngrams(normalizeWords(right));
  if (leftGrams.size === 0 || rightGrams.size === 0) return 0;
  let shared = 0;
  for (const gram of leftGrams) if (rightGrams.has(gram)) shared += 1;
  return shared / (leftGrams.size + rightGrams.size - shared);
};

interface Cluster {
  readonly paths: Set<string>;
  readonly representative: string;
}

// evaluateDuplication clusters near-identical normative sentences and reports
// any invariant restated across at least DUPLICATION_MIN_FILES files.
//
// Pass skill prose only. A command document must restate its own argument
// contract — command-validate requires every flag in both the argument-hint and
// the Arguments section — so shared flag wording across commands is a required
// interface, not drift.
export const evaluateDuplication = (
  files: readonly DuplicationFile[],
): readonly DuplicationFinding[] => {
  const clusters: Cluster[] = [];
  for (const file of files) {
    for (const sentence of normativeSentences(file.text)) {
      const existing = clusters.find(
        (cluster) =>
          similarity(cluster.representative, sentence) >=
          DUPLICATION_SIMILARITY_THRESHOLD,
      );
      if (existing === undefined) {
        clusters.push({paths: new Set([file.path]), representative: sentence});
        continue;
      }
      existing.paths.add(file.path);
    }
  }
  return clusters
    .filter((cluster) => cluster.paths.size >= DUPLICATION_MIN_FILES)
    .map((cluster) => {
      const paths = [...cluster.paths].toSorted();
      return {
        message:
          `duplication: an invariant is restated in ${String(paths.length)} ` +
          `files (${paths.join(", ")}); keep one normative source and ` +
          `cross-reference it`,
        paths,
        sentence: cluster.representative,
      };
    });
};
