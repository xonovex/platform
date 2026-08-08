import {z} from "zod";
import {nodeFileSystem, type FileSystem} from "./file-system.js";

const DependencyMapSchema = z.record(z.string(), z.string());
const BinSchema = z.union([z.string(), z.record(z.string(), z.string())]);

export const PackageJsonSchema = z.looseObject({
  name: z.string().optional(),
  version: z.string().optional(),
  license: z.string().optional(),
  private: z.boolean().optional(),
  repository: z
    .looseObject({
      type: z.string().optional(),
      url: z.string().optional(),
      directory: z.string().optional(),
    })
    .optional(),
  files: z.array(z.string()).optional(),
  bin: BinSchema.optional(),
  exports: z.unknown().optional(),
  main: z.string().optional(),
  module: z.string().optional(),
  types: z.string().optional(),
  publishConfig: z
    .looseObject({
      access: z.string().optional(),
      registry: z.string().optional(),
    })
    .optional(),
  dependencies: DependencyMapSchema.optional(),
  devDependencies: DependencyMapSchema.optional(),
  peerDependencies: DependencyMapSchema.optional(),
  optionalDependencies: DependencyMapSchema.optional(),
  os: z.array(z.string()).optional(),
  cpu: z.array(z.string()).optional(),
  libc: z.array(z.string()).optional(),
});

export type PackageJson = z.infer<typeof PackageJsonSchema>;

export const parsePackageJson = (text: string, source: string): PackageJson => {
  let value: unknown;
  try {
    value = JSON.parse(text);
  } catch (error) {
    throw new Error(`invalid package.json at ${source}: malformed JSON`, {
      cause: error,
    });
  }
  const result = PackageJsonSchema.safeParse(value);
  if (!result.success) {
    throw new Error(
      `invalid package.json at ${source}: ${z.prettifyError(result.error)}`,
    );
  }
  return result.data;
};

export const readPkg = (
  path: string,
  fs: FileSystem = nodeFileSystem,
): PackageJson => parsePackageJson(fs.readText(path), path);

export const writePkg = (
  path: string,
  pkg: PackageJson,
  fs: FileSystem = nodeFileSystem,
): void => {
  fs.writeFile(path, JSON.stringify(pkg, null, 2) + "\n");
};
