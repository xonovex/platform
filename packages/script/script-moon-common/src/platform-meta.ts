import {existsSync, readFileSync} from "node:fs";
import {join} from "node:path";
import {z} from "zod";

export const PlatformMetaSchema = z.strictObject({
  os: z.array(z.string().min(1)).min(1),
  cpu: z.array(z.string().min(1)).min(1),
  libc: z.array(z.string().min(1)).min(1).optional(),
});

export type PlatformMeta = z.infer<typeof PlatformMetaSchema>;

export const readPlatformMeta = (pkgDir: string): PlatformMeta | undefined => {
  const metaPath = join(pkgDir, "platform.json");
  if (!existsSync(metaPath)) return undefined;
  let value: unknown;
  try {
    value = JSON.parse(readFileSync(metaPath, "utf8"));
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
