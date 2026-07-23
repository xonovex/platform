import {createHash} from "node:crypto";
import {major, valid, validRange} from "semver";
import {z} from "zod";

export const GUIDE_NAME_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*-guide$/;
export const PROVISION_ID_RE = /^[a-z][a-z0-9-]*:[a-z0-9]+(?:[.-][a-z0-9]+)*$/;
export const SUPPORTED_CONTRACT_MAJOR = 2;

export const DEFAULT_OVERLAY_PRECEDENCE = [
  "global",
  "organization",
  "repository",
  "language",
  "framework",
  "path",
  "explicit",
] as const;

export const LifecycleSchema = z.enum(["durable", "procedural"]);
export const FunctionalRoleSchema = z.enum([
  "domain",
  "context",
  "preference",
  "procedure",
  "capability-use",
  "assurance",
  "recovery",
  "communication",
]);
export const ProvisionSchema = z
  .object({
    id: z.string().regex(PROVISION_ID_RE),
    version: z
      .string()
      .meta({format: "semver"})
      .refine((value) => valid(value) !== null, {
        message: "must be a valid semantic version",
      }),
  })
  .strict();
export const RequirementSchema = z
  .object({
    id: z.string().regex(PROVISION_ID_RE),
    range: z
      .string()
      .meta({format: "semver-range"})
      .refine((value) => validRange(value) !== null, {
        message: "must be a valid semantic-version range",
      }),
    strength: z.enum(["required", "preferred"]),
    reason: z.string().trim().min(1),
  })
  .strict();
export const OverlayScopeKindSchema = z.enum(DEFAULT_OVERLAY_PRECEDENCE);
export const OverlayScopeSchema = z
  .object({
    kind: OverlayScopeKindSchema,
    value: z.string().trim().min(1),
  })
  .strict();
export const OverlayContextSchema = z
  .object({
    global: z.string().trim().min(1).default("*"),
    organization: z.string().trim().min(1).optional(),
    repository: z.string().trim().min(1).optional(),
    languages: z.array(z.string().trim().min(1)).default([]),
    frameworks: z.array(z.string().trim().min(1)).default([]),
    paths: z.array(z.string().trim().min(1)).default([]),
    explicit: z.array(z.string().trim().min(1)).default([]),
  })
  .strict();
export const CatalogEntrySchema = z
  .object({
    name: z.string().regex(GUIDE_NAME_RE),
    classification: z
      .object({
        lifecycle: LifecycleSchema,
        functionalRole: FunctionalRoleSchema,
      })
      .strict(),
    provisions: z.array(ProvisionSchema).default([]),
    requirements: z.array(RequirementSchema).default([]),
  })
  .strict();
const OverlayPrecedenceSchema = z
  .array(OverlayScopeKindSchema)
  .length(DEFAULT_OVERLAY_PRECEDENCE.length)
  .refine(
    (values) =>
      new Set(values).size === DEFAULT_OVERLAY_PRECEDENCE.length &&
      DEFAULT_OVERLAY_PRECEDENCE.every((value) => values.includes(value)),
    {message: "must contain every overlay scope exactly once"},
  );
const CatalogFileSchema = z
  .object({
    contractVersion: z.string().refine((value) => valid(value) !== null, {
      message: "must be a valid semantic version",
    }),
    overlayPrecedence: OverlayPrecedenceSchema.default([
      ...DEFAULT_OVERLAY_PRECEDENCE,
    ]),
    skills: z.array(CatalogEntrySchema),
  })
  .strict();
export const InstalledSkillDependencySchema = z
  .object({
    plugin: z.string().trim().min(1),
    implementationVersion: z
      .string()
      .refine((value) => valid(value) !== null, {
        message: "must be a valid semantic version",
      })
      .optional(),
  })
  .strict();
export const InstalledSkillSchema = z
  .object({
    guide: z.string().regex(GUIDE_NAME_RE),
    implementationVersion: z.string().refine((value) => valid(value) !== null, {
      message: "must be a valid semantic version",
    }),
    dependencies: z.array(InstalledSkillDependencySchema).default([]),
    packagePath: z.string().trim().min(1).optional(),
    plugin: z.string().trim().min(1),
    sourcesPath: z.string().trim().min(1).optional(),
  })
  .strict();
const InstalledSkillInventorySchema = z
  .array(InstalledSkillSchema)
  .superRefine((skills, context) => {
    const firstIndexByGuide = new Map<string, number>();
    for (const [index, skill] of skills.entries()) {
      const firstIndex = firstIndexByGuide.get(skill.guide);
      if (firstIndex === undefined) {
        firstIndexByGuide.set(skill.guide, index);
        continue;
      }
      context.addIssue({
        code: "custom",
        message: `duplicates installed guide at index ${String(firstIndex)}`,
        path: [index, "guide"],
      });
    }
  });
export const ExactSkillRequestSchema = z
  .object({
    guide: z.string().regex(GUIDE_NAME_RE),
    implementationVersion: z
      .string()
      .refine((value) => valid(value) !== null, {
        message: "must be a valid semantic version",
      })
      .optional(),
    reason: z.string().trim().min(1),
    required: z.boolean().default(true),
  })
  .strict();
export const PreferenceOverlayRequestSchema = z
  .object({
    guide: z.string().regex(GUIDE_NAME_RE),
    reason: z.string().trim().min(1),
    scope: OverlayScopeSchema,
    target: z.string().regex(PROVISION_ID_RE),
  })
  .strict();
export const RequirementProviderSchema = z.record(
  z.string().regex(PROVISION_ID_RE),
  z.string().regex(GUIDE_NAME_RE),
);
export const CompositionRequestSchema = z
  .object({
    exactSkills: z.array(ExactSkillRequestSchema).default([]),
    overlayContext: OverlayContextSchema.default({
      global: "*",
      languages: [],
      frameworks: [],
      paths: [],
      explicit: [],
    }),
    preferenceOverlays: z.array(PreferenceOverlayRequestSchema).default([]),
    requirementProviders: RequirementProviderSchema.default({}),
    requirements: z.array(RequirementSchema).default([]),
  })
  .strict();

export type Lifecycle = z.infer<typeof LifecycleSchema>;
export type FunctionalRole = z.infer<typeof FunctionalRoleSchema>;
export type SemanticProvision = z.infer<typeof ProvisionSchema>;
export type SemanticRequirement = z.infer<typeof RequirementSchema>;
export type OverlayScopeKind = z.infer<typeof OverlayScopeKindSchema>;
export type OverlayScope = z.infer<typeof OverlayScopeSchema>;
export type OverlayContext = z.infer<typeof OverlayContextSchema>;
export type CompositionCatalogEntry = z.infer<typeof CatalogEntrySchema>;
export type InstalledSkillDependency = z.infer<
  typeof InstalledSkillDependencySchema
>;
export type InstalledSkill = z.infer<typeof InstalledSkillSchema>;
export type ExactSkillRequest = z.infer<typeof ExactSkillRequestSchema>;
export type PreferenceOverlayRequest = z.infer<
  typeof PreferenceOverlayRequestSchema
>;
export type CompositionRequest = z.infer<typeof CompositionRequestSchema>;

export interface CompositionCatalog {
  readonly contractVersion: string;
  readonly digest: string;
  readonly overlayPrecedence: readonly OverlayScopeKind[];
  readonly skills: readonly CompositionCatalogEntry[];
}

export type ParseResult<T> =
  | {readonly success: true; readonly data: T}
  | {readonly success: false; readonly errors: readonly string[]};

const issueDetail = (issue: z.core.$ZodIssue): string => {
  const path =
    issue.path.length > 0 ? issue.path.map(String).join(".") : "root";
  return `${path}: ${issue.message}`;
};

const parseWithSchema = <T>(
  schema: z.ZodType<T>,
  input: unknown,
): ParseResult<T> => {
  const parsed = schema.safeParse(input);
  return parsed.success
    ? {success: true, data: parsed.data}
    : {success: false, errors: parsed.error.issues.map(issueDetail)};
};

export const parseCompositionCatalog = (
  input: unknown,
  sourceText: string,
): ParseResult<CompositionCatalog> => {
  const parsed = CatalogFileSchema.safeParse(input);
  if (!parsed.success) {
    return {
      success: false,
      errors: parsed.error.issues.map(issueDetail),
    };
  }
  if (major(parsed.data.contractVersion) !== SUPPORTED_CONTRACT_MAJOR) {
    return {
      success: false,
      errors: [
        `contractVersion: unsupported major ${String(major(parsed.data.contractVersion))}; expected ${String(SUPPORTED_CONTRACT_MAJOR)}`,
      ],
    };
  }
  return {
    success: true,
    data: {
      ...parsed.data,
      digest: createHash("sha256").update(sourceText).digest("hex"),
    },
  };
};

export const parseInstalledSkillInventory = (
  input: unknown,
): ParseResult<readonly InstalledSkill[]> =>
  parseWithSchema(InstalledSkillInventorySchema, input);

export const parseCompositionRequest = (
  input: unknown,
): ParseResult<CompositionRequest> =>
  parseWithSchema(CompositionRequestSchema, input);

const jsonSchema = (schema: z.ZodType): Readonly<Record<string, unknown>> => {
  const generated = z.toJSONSchema(schema, {
    target: "draft-2020-12",
  });
  return Object.fromEntries(
    Object.entries(generated).filter(([key]) => key !== "$schema"),
  );
};

export const workflowCompositionSchemaDefinitions = (): Readonly<
  Record<string, Readonly<Record<string, unknown>>>
> => ({
  semanticRequirement: jsonSchema(RequirementSchema),
  overlayScope: jsonSchema(OverlayScopeSchema),
  overlayContext: jsonSchema(OverlayContextSchema),
  preferenceOverlay: jsonSchema(PreferenceOverlayRequestSchema),
  requirementProviders: jsonSchema(RequirementProviderSchema),
});
