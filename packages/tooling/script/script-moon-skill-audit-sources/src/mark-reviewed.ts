import {toIsoDate} from "./dates.js";

const REVIEWED_RE = /\*\*Last reviewed:\*\*\s*(\d{4}-\d{2}-\d{2})/;
const HEADER_RE = /^##\s+(.*\S)\s*$/;

export interface StampedSources {
  readonly text: string;
  /** The titles of the source blocks whose date changed, in file order. */
  readonly edited: readonly string[];
}

/**
 * Sets 'Last reviewed' to today in the source blocks whose title contains
 * titleFilter, ignoring case, or in every block when titleFilter is undefined.
 */
export const stampReviewed = (
  text: string,
  titleFilter: string | undefined,
  today: Date,
): StampedSources => {
  const lines = text.split(/(?<=\n)/);
  const needle = titleFilter?.toLowerCase();
  const edited: string[] = [];
  let currentTitle = "";
  let matched = false;
  for (const [i, line] of lines.entries()) {
    const header = HEADER_RE.exec(line.replace(/\n$/, ""));
    if (header) {
      currentTitle = header[1] ?? "";
      matched =
        needle === undefined || currentTitle.toLowerCase().includes(needle);
      continue;
    }
    if (matched && REVIEWED_RE.test(line)) {
      lines[i] = line.replace(
        REVIEWED_RE,
        `**Last reviewed:** ${toIsoDate(today)}`,
      );
      if (currentTitle && !edited.includes(currentTitle)) {
        edited.push(currentTitle);
      }
    }
  }
  return {text: lines.join(""), edited};
};
