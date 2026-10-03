import {join} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";

export const MD_LINK_RE = /\[[^\]]*\]\(([^)\s]+)\)/g;
const EXTERNAL_LINK_RE = /^(?:https?|mailto):/i;

// Minimal sink the check reports through; the CLI's Report satisfies it.
export interface LinkReport {
  addFail(message: string): void;
  addPass(message: string): void;
  addWarn?(message: string): void;
}

// The file a markdown link points at, with any in-page fragment stripped, or
// null when the link names no real sibling file: an external scheme
// (http/https/mailto), a placeholder or elided illustrative form, or a pure
// #anchor.
export const relativeLinkTarget = (raw: string): string | null => {
  if (EXTERNAL_LINK_RE.test(raw)) return null;
  if (raw.includes("<") || raw.includes("{") || raw.includes("…")) return null;
  const target = raw.split("#", 1)[0] ?? "";
  return target === "" ? null : target;
};

// checkReferenceFileLinks resolves links relative to their containing document.
export const checkReferenceFileLinks = (
  skillDir: string,
  report: LinkReport,
  fs: FileSystem = nodeFileSystem,
): void => {
  const refsDir = join(skillDir, "references");
  if (!fs.isDirectory(refsDir)) {
    return;
  }
  const mdFiles = fs
    .readDirectory(refsDir)
    .filter((entry) => entry.endsWith(".md"))
    .filter((entry) => fs.isFile(join(refsDir, entry)))
    .toSorted();

  let resolved = 0;
  let broken = 0;
  for (const entry of mdFiles) {
    const text = fs.readText(join(refsDir, entry));
    for (const m of text.matchAll(MD_LINK_RE)) {
      const raw = m[1];
      if (raw === undefined) continue;
      const target = relativeLinkTarget(raw);
      if (target === null) continue;
      if (fs.isFile(join(refsDir, target))) {
        resolved += 1;
      } else {
        report.addFail(
          `reference links: broken link in references/${entry} → ${raw}`,
        );
        broken += 1;
      }
    }
  }

  if (broken === 0 && resolved > 0) {
    report.addPass(
      `reference links: ${String(resolved)}/${String(resolved)} link(s) resolve`,
    );
  }
};
