import {readdirSync, readFileSync, statSync} from "node:fs";
import {join} from "node:path";

export const MD_LINK_RE = /\[[^\]]*\]\(([^)\s]+)\)/g;
const EXTERNAL_LINK_RE = /^(?:https?|mailto):/i;

// Minimal sink the check reports through; the CLI's Report satisfies it.
export interface LinkReport {
  addFail(message: string): void;
  addPass(message: string): void;
}

export const isFile = (path: string): boolean => {
  try {
    return statSync(path).isFile();
  } catch {
    return false;
  }
};

const isDir = (path: string): boolean => {
  try {
    return statSync(path).isDirectory();
  } catch {
    return false;
  }
};

// The file a markdown link points at, with any in-page fragment stripped, or
// null when the link names no real sibling file: an external scheme
// (http/https/mailto), a placeholder or elided illustrative form (<topic>.md,
// {topic}.md, the …/pull/PR and …/commit/hash ellipsis links documented in the
// versioning changelog guide), or a pure #anchor.
export const relativeLinkTarget = (raw: string): string | null => {
  if (EXTERNAL_LINK_RE.test(raw)) return null;
  if (raw.includes("<") || raw.includes("{") || raw.includes("…")) return null;
  const target = raw.split("#", 1)[0] ?? "";
  return target === "" ? null : target;
};

// Resolve every markdown link inside each references/*.md relative to that
// file's own directory. The body-scan check only sees links in SKILL.md, so a
// broken sibling link written inside a reference file — e.g.
// [references/foo.md](references/foo.md), which resolves to the non-existent
// references/references/foo.md instead of the sibling references/foo.md —
// would otherwise pass unnoticed. A broken link is a fact, so each is a FAIL.
export const checkReferenceFileLinks = (
  skillDir: string,
  report: LinkReport,
): void => {
  const refsDir = join(skillDir, "references");
  if (!isDir(refsDir)) {
    return;
  }
  const mdFiles = readdirSync(refsDir)
    .filter((entry) => entry.endsWith(".md"))
    .filter((entry) => isFile(join(refsDir, entry)))
    .toSorted();

  let resolved = 0;
  let broken = 0;
  for (const entry of mdFiles) {
    const text = readFileSync(join(refsDir, entry), "utf8");
    for (const m of text.matchAll(MD_LINK_RE)) {
      const raw = m[1];
      if (raw === undefined) continue;
      const target = relativeLinkTarget(raw);
      if (target === null) continue;
      if (isFile(join(refsDir, target))) {
        resolved += 1;
      } else {
        report.addFail(
          `reference links: broken link in references/${entry} → ${raw}`,
        );
        broken += 1;
      }
    }
  }

  if (resolved > 0 && broken === 0) {
    report.addPass(
      `reference links: ${String(resolved)}/${String(resolved)} link(s) resolve`,
    );
  }
};
