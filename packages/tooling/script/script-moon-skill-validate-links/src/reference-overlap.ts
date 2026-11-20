import {
  fail,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import type {MarkdownDocument} from "./catalog-scan.js";

// A pair of files shorter than this many shingles is too small to call a copy.
const MIN_SHINGLES = 20;
const DUPLICATE_CONTAINMENT = 0.3;

const contentShingles = (text: string): ReadonlySet<string> => {
  const words = text
    .toLowerCase()
    .replaceAll(/```[\s\S]*?```/g, " ")
    .replaceAll(/[^a-z0-9_]+/g, " ")
    .trim()
    .split(/\s+/)
    .filter(Boolean);
  return new Set(
    words
      .slice(0, Math.max(0, words.length - 4))
      .map((_, index) => words.slice(index, index + 5).join(" ")),
  );
};

const shingleContainment = (
  left: ReadonlySet<string>,
  right: ReadonlySet<string>,
): number => {
  const denominator = Math.min(left.size, right.size);
  if (denominator < MIN_SHINGLES) return 0;
  return [...left].filter((shingle) => right.has(shingle)).length / denominator;
};

export interface NamedReferences {
  readonly name: string;
  /** Reference files with paths relative to the plugin package. */
  readonly references: readonly MarkdownDocument[];
}

/** Fails each reference of a dependent plugin that copies one of its dependency. */
export const checkReferenceOverlap = (
  source: NamedReferences,
  target: NamedReferences,
): readonly Finding[] =>
  source.references.flatMap((sourceReference) => {
    const sourceShingles = contentShingles(sourceReference.text);
    return target.references.flatMap((targetReference) => {
      const containment = shingleContainment(
        sourceShingles,
        contentShingles(targetReference.text),
      );
      return containment < DUPLICATE_CONTAINMENT
        ? []
        : [
            fail(
              `skill ownership: ${source.name} ${sourceReference.path} duplicates ${target.name} ${targetReference.path} (${String(Math.round(containment * 100))}% five-word-shingle containment); keep only the dependent skill's specialization`,
            ),
          ];
    });
  });
