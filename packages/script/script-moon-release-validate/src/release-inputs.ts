import {z} from "zod";

const MarketplaceEntrySchema = z.object({
  name: z.string(),
  source: z.union([z.string(), z.object({path: z.string().optional()})]),
  description: z.string().optional(),
});

export const MarketplaceSchema = z.object({
  metadata: z.object({version: z.string()}).optional(),
  plugins: z.array(MarketplaceEntrySchema),
});
export type Marketplace = z.infer<typeof MarketplaceSchema>;

export const PackageManifestSchema = z.object({
  name: z.string(),
  version: z.string().optional(),
  description: z.string().optional(),
  dependencies: z.record(z.string(), z.string()).optional(),
  devDependencies: z.record(z.string(), z.string()).optional(),
  peerDependencies: z.record(z.string(), z.string()).optional(),
  optionalDependencies: z.record(z.string(), z.string()).optional(),
});
export type PackageManifest = z.infer<typeof PackageManifestSchema>;

export const PluginManifestSchema = z.object({
  name: z.string(),
  version: z.string().optional(),
  description: z.string().optional(),
  dependencies: z.array(z.string()).optional(),
});
export type PluginManifest = z.infer<typeof PluginManifestSchema>;

export const LockfileSchema = z.object({
  packages: z
    .record(z.string(), z.object({version: z.string().optional()}))
    .optional(),
});
export type Lockfile = z.infer<typeof LockfileSchema>;
