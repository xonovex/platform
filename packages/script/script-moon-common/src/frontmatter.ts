import {readFileSync} from "node:fs";

// The name a markdown document declares in its YAML frontmatter, used to identify a
// skill by the name it publishes rather than by its directory.
export const parseFrontmatterName = (skillFile: string): string | undefined => {
  const text = readFileSync(skillFile, "utf8");
  const match = /^---\r?\n([\s\S]*?)\r?\n---/.exec(text);
  const frontmatter = match?.[1];
  if (frontmatter === undefined) {
    return undefined;
  }
  for (const line of frontmatter.split(/\r?\n/)) {
    const name = /^name:\s*(.+?)\s*$/.exec(line)?.[1];
    if (name !== undefined) {
      return name.replaceAll(/^["']|["']$/g, "");
    }
  }
  return undefined;
};
