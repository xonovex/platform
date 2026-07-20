import {createHash} from "node:crypto";
import {z} from "zod";
import {
  triggerSourceSchema,
  workflowExecutionSchema,
  workflowInvocationApiVersion,
  workflowInvocationSchema,
  type WorkflowInvocation,
} from "./workflow-runtime.ts";

export const normalizedTriggerEventSchema = z
  .object({
    source: triggerSourceSchema,
    nativeReference: z.string().min(1),
    actor: z.string().min(1),
    idempotencyKey: z.string().min(1),
    parentInvocationReference: z.string().min(1).optional(),
    childDepth: z.number().int().nonnegative().optional(),
    subject: z
      .object({
        reference: z.string().min(1),
        revision: z.string().min(1),
      })
      .strict(),
  })
  .strict();

export const trustedWorkflowTemplateSchema = z
  .object({
    workflow: z
      .object({
        operation: z.string().min(1),
        profileReference: z.string().min(1).optional(),
      })
      .strict(),
    execution: workflowExecutionSchema,
    evidenceProviderReference: z.string().min(1),
  })
  .strict();

export type NormalizedTriggerEvent = z.infer<
  typeof normalizedTriggerEventSchema
>;
export type TrustedWorkflowTemplate = z.infer<
  typeof trustedWorkflowTemplateSchema
>;

const invocationIdentity = (
  event: NormalizedTriggerEvent,
  template: TrustedWorkflowTemplate,
): string => {
  const digest = createHash("sha256")
    .update(
      JSON.stringify({
        source: event.source,
        nativeReference: event.nativeReference,
        idempotencyKey: event.idempotencyKey,
        subject: event.subject,
        operation: template.workflow.operation,
        family: template.execution.family,
      }),
    )
    .digest("hex");
  return `workflow:sha256:${digest}`;
};

// adaptTriggerToWorkflowInvocation is deliberately source-neutral. Native
// adapters authenticate and minimize their events first; this boundary binds
// that origin to a trusted execution template without changing execution type.
export const adaptTriggerToWorkflowInvocation = (
  untrustedEvent: unknown,
  untrustedTemplate: unknown,
): WorkflowInvocation => {
  const event = normalizedTriggerEventSchema.parse(untrustedEvent);
  const template = trustedWorkflowTemplateSchema.parse(untrustedTemplate);
  return workflowInvocationSchema.parse({
    apiVersion: workflowInvocationApiVersion,
    invocationId: invocationIdentity(event, template),
    trigger: {
      source: event.source,
      nativeReference: event.nativeReference,
      actor: event.actor,
      ...(event.parentInvocationReference === undefined
        ? {}
        : {parentInvocationReference: event.parentInvocationReference}),
      ...(event.childDepth === undefined ? {} : {childDepth: event.childDepth}),
    },
    subject: event.subject,
    workflow: template.workflow,
    execution: template.execution,
    evidenceProviderReference: template.evidenceProviderReference,
  });
};
