import {readdirSync, readFileSync} from "node:fs";
import {extname, join, relative} from "node:path";

interface ProsePunctuationRule {
  readonly hint: string;
  readonly label: string;
  readonly pattern: RegExp;
}

// Repository prose uses a comma, colon, or full stop instead of an em dash,
// three periods instead of the single ellipsis character, and straight quotes
// instead of typographic ones. Every one of these also hides behind a JSON or JS
// unicode escape, which a search for the literal character misses.
const RULES: readonly ProsePunctuationRule[] = [
  {
    hint: "use a comma, colon, or full stop",
    label: "em dash",
    pattern: /\u2014/u,
  },
  {
    hint: "use a comma, colon, or full stop",
    label: "escaped em dash",
    pattern: /\\u2014/iu,
  },
  {hint: "use three periods", label: "ellipsis", pattern: /\u2026/u},
  {
    hint: "use three periods",
    label: "escaped ellipsis",
    pattern: /\\u2026/iu,
  },
  {
    hint: "use a straight apostrophe",
    label: "smart single quote",
    pattern: /[\u2018\u2019]/u,
  },
  {
    hint: "use a straight apostrophe",
    label: "escaped smart single quote",
    pattern: /\\u201[89]/iu,
  },
  {
    hint: "use straight double quotes",
    label: "smart double quote",
    pattern: /[\u201C\u201D]/u,
  },
  {
    hint: "use straight double quotes",
    label: "escaped smart double quote",
    pattern: /\\u201[cd]/iu,
  },
];

// Text file types that carry prose in skill and command packages: guides and
// references, manifests and eval fixtures, bundled scripts, and asset templates.
export const PROSE_FILE_EXTENSIONS: readonly string[] = [
  ".json",
  ".jsonc",
  ".markdown",
  ".md",
  ".py",
  ".sh",
  ".template",
  ".txt",
  ".yaml",
  ".yml",
];

const SKIPPED_DIRECTORIES = new Set([
  ".git",
  ".moon",
  "coverage",
  "dist",
  "node_modules",
]);

export interface ProsePunctuationFinding {
  /** Trimmed offending line, capped for report output. */
  readonly excerpt: string;
  /** What to write instead. */
  readonly hint: string;
  /** What was found, for example `em dash`. */
  readonly label: string;
  /** 1-indexed line number. */
  readonly line: number;
  /** Path as supplied, relative to the root when scanning a tree. */
  readonly path: string;
}

export const prosePunctuationFindings = (
  path: string,
  text: string,
): readonly ProsePunctuationFinding[] => {
  const findings: ProsePunctuationFinding[] = [];
  for (const [index, line] of text.split(/\r\n|\r|\n/u).entries()) {
    for (const {hint, label, pattern} of RULES) {
      if (!pattern.test(line)) continue;
      findings.push({
        excerpt: line.trim().slice(0, 80),
        hint,
        label,
        line: index + 1,
        path,
      });
    }
  }
  return findings;
};

const proseFiles = (root: string, directory: string): readonly string[] => {
  const files: string[] = [];
  for (const entry of readdirSync(directory, {withFileTypes: true})) {
    const absolute = join(directory, entry.name);
    if (entry.isDirectory()) {
      if (SKIPPED_DIRECTORIES.has(entry.name)) continue;
      files.push(...proseFiles(root, absolute));
      continue;
    }
    if (!entry.isFile()) continue;
    if (!PROSE_FILE_EXTENSIONS.includes(extname(entry.name).toLowerCase())) {
      continue;
    }
    files.push(relative(root, absolute));
  }
  return files;
};

/** Every finding under `root`, with paths relative to `root`. */
export const scanProsePunctuation = (
  root: string,
): readonly ProsePunctuationFinding[] =>
  proseFiles(root, root)
    .toSorted()
    .flatMap((path) =>
      prosePunctuationFindings(path, readFileSync(join(root, path), "utf8")),
    );
