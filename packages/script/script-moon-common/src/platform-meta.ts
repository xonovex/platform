import {join} from "node:path";
import {z} from "zod";
import {nodeFileSystem, type FileSystem} from "./file-system.js";

const PlatformMetaSchema = z.strictObject({
  os: z.array(z.string().min(1)).min(1),
  cpu: z.array(z.string().min(1)).min(1),
  libc: z.array(z.string().min(1)).min(1).optional(),
});

export type PlatformMeta = z.infer<typeof PlatformMetaSchema>;

export const readPlatformMeta = (
  pkgDir: string,
  fs: FileSystem = nodeFileSystem,
): PlatformMeta | undefined => {
  const metaPath = join(pkgDir, "platform.json");
  if (!fs.isFile(metaPath)) return undefined;
  let value: unknown;
  try {
    value = JSON.parse(fs.readText(metaPath));
  } catch (error) {
    throw new Error(
      `invalid platform metadata at ${metaPath}: malformed JSON`,
      {
        cause: error,
      },
    );
  }
  const result = PlatformMetaSchema.safeParse(value);
  if (!result.success) {
    throw new Error(
      `invalid platform metadata at ${metaPath}: ${z.prettifyError(result.error)}`,
    );
  }
  return result.data;
};
