import {z} from "zod";
import type {WorkflowCompositionExplanation} from "./runtime.js";

export const maturityModelSchema = z
  .object({
    name: z.string().min(1),
    levels: z
      .array(
        z
          .object({
            name: z.string().min(1),
            requiredCapabilities: z.array(z.string().min(1)),
          })
          .strict(),
      )
      .min(1),
  })
  .strict()
  .superRefine((model, context) => {
    const names = new Set<string>();
    for (const [index, level] of model.levels.entries()) {
      if (names.has(level.name)) {
        context.addIssue({
          code: "custom",
          message: "maturity-level-name-duplicate",
          path: ["levels", index, "name"],
        });
      }
      names.add(level.name);
    }
  });

export interface MaturityAssessment {
  readonly model: string;
  readonly achieved?: string;
  readonly levels: readonly {
    readonly name: string;
    readonly satisfied: boolean;
    readonly missingCapabilities: readonly string[];
  }[];
}

export const assessWorkflowMaturity = (
  composition: WorkflowCompositionExplanation,
  untrustedModel: unknown,
): MaturityAssessment => {
  const model = maturityModelSchema.parse(untrustedModel);
  const capabilities = new Set(composition.availableCapabilities);
  const levels = model.levels.map((level) => {
    const missingCapabilities = [
      ...new Set(
        level.requiredCapabilities.filter(
          (capability) => !capabilities.has(capability),
        ),
      ),
    ].toSorted();
    return {
      name: level.name,
      satisfied: missingCapabilities.length === 0,
      missingCapabilities,
    };
  });
  const achieved = levels.findLast((level) => level.satisfied)?.name;
  return {
    model: model.name,
    ...(achieved === undefined ? {} : {achieved}),
    levels,
  };
};
