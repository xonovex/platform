import {existsSync, readFileSync} from "node:fs";
import {join} from "node:path";

export interface PlatformMeta {
  os: readonly string[];
  cpu: readonly string[];
  libc?: readonly string[];
}

export const readPlatformMeta = (pkgDir: string): PlatformMeta | undefined => {
  const metaPath = join(pkgDir, "platform.json");
  if (!existsSync(metaPath)) return undefined;
  return JSON.parse(readFileSync(metaPath, "utf8")) as PlatformMeta;
};
