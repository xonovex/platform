import {valid} from "semver";
import {z} from "zod";
import {
  GUIDE_NAME_RE,
  OverlayContextSchema,
  PreferenceOverlayRequestSchema,
  RequirementProviderSchema,
  RequirementSchema,
  workflowCompositionSchemaDefinitions,
  type CompositionCatalog,
  type CompositionRequest,
  type ParseResult,
  type SemanticRequirement,
} from "./skill-composition-contract.js";

const ExactSkillImplementationSchema = z
  .object({
    id: z.string().regex(GUIDE_NAME_RE),
    version: z.string().refine((value) => valid(value) !== null, {
      message: "must be a valid semantic version",
    }),
  })
  .strict();
const ExactCapabilityImplementationSchema = z
  .object({
    id: z.string().trim().min(1),
    version: z.string().trim().min(1),
  })
  .strict();
const ImplementationOverridesSchema = z
  .object({
    skills: z.array(ExactSkillImplementationSchema).default([]),
    capabilities: z.array(ExactCapabilityImplementationSchema).default([]),
  })
  .strict();
const SemanticSelectionSchema = z
  .object({
    method: z.string().trim().min(1).optional(),
    perspectives: z.array(z.string().trim().min(1)).default([]),
    criteria: z.array(z.unknown()).default([]),
    roleLenses: z.array(z.string().trim().min(1)).default([]),
    resolutionMode: z
      .enum(["strict", "assisted", "automatic"])
      .default("assisted"),
    acceptedSuggestions: z.array(z.string().trim().min(1)).default([]),
    skillRequirements: z.array(RequirementSchema).default([]),
    skillProviders: RequirementProviderSchema.default({}),
    overlayContext: OverlayContextSchema.default({
      global: "*",
      languages: [],
      frameworks: [],
      paths: [],
      explicit: [],
    }),
    preferenceOverlays: z.array(PreferenceOverlayRequestSchema).default([]),
  })
  .strict();
const WorkflowOperationSchema = z.enum([
  "create",
  "review",
  "revise",
  "decide",
  "execute",
  "validate",
  "publish",
  "abandon",
  "workspace-create",
  "workspace-merge",
  "workspace-abandon",
  "workspace-cleanup",
]);
const WorkflowCompositionInputSchema = z.looseObject({
  operation: WorkflowOperationSchema,
  selection: SemanticSelectionSchema,
  implementationOverrides: ImplementationOverridesSchema.default({
    skills: [],
    capabilities: [],
  }),
});

export type WorkflowCompositionInput = z.infer<
  typeof WorkflowCompositionInputSchema
>;

const issueDetail = (issue: z.core.$ZodIssue): string => {
  const path =
    issue.path.length > 0 ? issue.path.map(String).join(".") : "root";
  return `${path}: ${issue.message}`;
};

const provisionRequirement = (
  catalog: CompositionCatalog,
  namespace: "method" | "perspective",
  value: string,
): SemanticRequirement | undefined => {
  const id = value.includes(":") ? value : `${namespace}:${value}`;
  const provisions = catalog.skills.flatMap(({provisions}) =>
    provisions.filter((provision) => provision.id === id),
  );
  const provision = provisions[0];
  if (provision === undefined) return undefined;
  return {
    id,
    range: `^${provision.version}`,
    strength: "preferred",
    reason: `The workflow request selected ${namespace} ${value}.`,
  };
};

export const normalizeWorkflowCompositionRequest = (
  catalog: CompositionCatalog,
  input: unknown,
): ParseResult<CompositionRequest> => {
  const parsed = WorkflowCompositionInputSchema.safeParse(input);
  if (!parsed.success) {
    return {
      success: false,
      errors: parsed.error.issues.map(issueDetail),
    };
  }
  const selection = parsed.data.selection;
  const explicitRequirementIds = new Set(
    selection.skillRequirements.map(({id}) => id),
  );
  const derivedRequirements = [
    ...(selection.method === undefined
      ? []
      : [provisionRequirement(catalog, "method", selection.method)]),
    ...selection.perspectives.map((perspective) =>
      provisionRequirement(catalog, "perspective", perspective),
    ),
  ].flatMap((requirement) =>
    requirement === undefined || explicitRequirementIds.has(requirement.id)
      ? []
      : [requirement],
  );
  const requirements = [
    ...selection.skillRequirements,
    ...derivedRequirements,
  ].filter(
    (requirement, index, values) =>
      values.findIndex(({id}) => id === requirement.id) === index,
  );
  const missingProviderRequirements = Object.keys(
    selection.skillProviders,
  ).filter((id) => requirements.every((requirement) => requirement.id !== id));
  if (missingProviderRequirements.length > 0) {
    return {
      success: false,
      errors: missingProviderRequirements.map(
        (id) =>
          `selection.skillProviders.${id}: provider binding has no matching semantic requirement`,
      ),
    };
  }
  return {
    success: true,
    data: {
      exactSkills: [
        {
          guide: "workflow-guide",
          reason: `${parsed.data.operation} is owned by workflow-guide.`,
          required: true,
        },
        ...parsed.data.implementationOverrides.skills.map(({id, version}) => ({
          guide: id,
          implementationVersion: version,
          reason: `The workflow request pins ${id}@${version}.`,
          required: true,
        })),
      ],
      overlayContext: selection.overlayContext,
      preferenceOverlays: selection.preferenceOverlays,
      requirementProviders: selection.skillProviders,
      requirements,
    },
  };
};

export const workflowRequestCompositionSchemaDefinitions = (): Readonly<
  Record<string, Readonly<Record<string, unknown>>>
> => {
  const operation = Object.fromEntries(
    Object.entries(
      z.toJSONSchema(WorkflowOperationSchema, {target: "draft-2020-12"}),
    ).filter(([key]) => key !== "$schema"),
  );
  return {
    ...workflowCompositionSchemaDefinitions(),
    operation,
  };
};
