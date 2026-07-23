import {
  parseCompositionCatalog,
  parseInstalledSkillInventory,
  type InstalledSkill,
  type ParseResult,
} from "./skill-composition-contract.js";
import {
  resolveComposition,
  type CompositionResolution,
} from "./skill-composition.js";
import {discoverInstalledSkillInventory} from "./skill-inventory.js";
import {normalizeWorkflowCompositionRequest} from "./workflow-composition.js";

export interface WorkflowCompositionRuntimeInput {
  readonly catalogInput: unknown;
  readonly catalogSourceText: string;
  readonly installedRoots?: readonly string[];
  readonly installedSkills?: unknown;
  readonly workflowRequest: unknown;
}

export const composeWorkflowRequest = (
  input: WorkflowCompositionRuntimeInput,
): ParseResult<CompositionResolution> => {
  const catalog = parseCompositionCatalog(
    input.catalogInput,
    input.catalogSourceText,
  );
  if (!catalog.success) return catalog;
  const installed: ParseResult<readonly InstalledSkill[]> =
    input.installedSkills === undefined
      ? discoverInstalledSkillInventory(input.installedRoots ?? [])
      : parseInstalledSkillInventory(input.installedSkills);
  if (!installed.success) return installed;
  const request = normalizeWorkflowCompositionRequest(
    catalog.data,
    input.workflowRequest,
  );
  if (!request.success) return request;
  return {
    success: true,
    data: resolveComposition(catalog.data, installed.data, request.data),
  };
};
