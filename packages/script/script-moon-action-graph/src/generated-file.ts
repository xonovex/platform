import {readFileSync, writeFileSync} from "node:fs";

export function writeOrCheckGeneratedFile(
  filePath: string,
  contents: string | Uint8Array,
  check: boolean,
): void {
  if (!check) {
    writeFileSync(filePath, contents);
    return;
  }

  let current: Buffer;
  try {
    current = readFileSync(filePath);
  } catch (error) {
    throw new Error(`Generated file is missing: ${filePath}`, {cause: error});
  }

  const expected = Buffer.from(contents);
  if (!current.equals(expected)) {
    throw new Error(`Generated file is stale: ${filePath}`);
  }
}
