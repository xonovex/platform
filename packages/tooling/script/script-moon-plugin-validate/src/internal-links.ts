import {relative, resolve, sep} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {pluginDocuments} from "./plugin-documents.js";
import {issue, type ValidationIssue} from "./validation.js";

const headingAnchors = (path: string, fs: FileSystem): ReadonlySet<string> => {
  const anchors = new Set<string>();
  for (const match of fs.readText(path).matchAll(/^#{1,6}\s+(.+)$/gmu)) {
    const heading = match[1];
    if (heading === undefined) continue;
    anchors.add(
      heading
        .trim()
        .toLowerCase()
        .replaceAll(/[`*_~]/gu, "")
        .replaceAll(/[^\p{Letter}\p{Number}\s-]/gu, "")
        .replaceAll(/\s+/gu, "-"),
    );
  }
  return anchors;
};

// A link that leaves the package is another package's contract, so only links
// that stay inside it are resolved.
export const checkInternalLinks = (
  packageDirectory: string,
  repositoryRoot: string,
  fs: FileSystem,
): readonly ValidationIssue[] => {
  const issues: ValidationIssue[] = [];
  for (const file of pluginDocuments(packageDirectory, fs)) {
    const source = fs.readText(file);
    for (const match of source.matchAll(/\[.*?\]\(([^)\s]+)(?:\s[^)]*)?\)/gu)) {
      const rawTarget = match[1];
      if (
        rawTarget === undefined ||
        /^(?:[a-z]+:|\/)/iu.test(rawTarget) ||
        rawTarget.startsWith("#")
      ) {
        continue;
      }
      const [rawPath = "", rawFragment] = rawTarget.split("#", 2);
      const target = resolve(file, "..", decodeURIComponent(rawPath));
      if (
        target !== packageDirectory &&
        !target.startsWith(packageDirectory + sep)
      ) {
        continue;
      }
      if (!fs.isFile(target) && !fs.isDirectory(target)) {
        issues.push(
          issue(
            "link.missing-target",
            relative(repositoryRoot, file),
            `local link target '${rawTarget}' does not exist`,
          ),
        );
        continue;
      }
      if (
        rawFragment !== undefined &&
        rawFragment.length > 0 &&
        fs.isFile(target) &&
        target.endsWith(".md") &&
        !headingAnchors(target, fs).has(
          decodeURIComponent(rawFragment).toLowerCase(),
        )
      ) {
        issues.push(
          issue(
            "link.missing-fragment",
            relative(repositoryRoot, file),
            `local link fragment '${rawTarget}' does not exist`,
          ),
        );
      }
    }
  }
  return issues;
};
