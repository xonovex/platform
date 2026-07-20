import {createHash} from "node:crypto";
import {z} from "zod";
import {
  workflowInvocationApiVersion,
  workflowInvocationSchema,
  type WorkflowInvocation,
} from "./runtime.js";

export const normalizedTriggerEventSchema = z
  .object({
    kind: z.string().min(1),
    reference: z.string().min(1),
    actor: z.string().min(1).optional(),
    idempotencyKey: z.string().min(1),
    data: z.unknown().optional(),
    subject: z
      .object({
        reference: z.string().min(1),
        revision: z.string().min(1).optional(),
      })
      .strict(),
  })
  .strict();

export const trustedWorkflowTemplateSchema = workflowInvocationSchema.pick({
  operation: true,
  executor: true,
  controls: true,
  evidence: true,
  requiredCapabilities: true,
  metadata: true,
});

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
        kind: event.kind,
        reference: event.reference,
        idempotencyKey: event.idempotencyKey,
        subject: event.subject,
        template,
      }),
    )
    .digest("hex");
  return `workflow:sha256:${digest}`;
};

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
      kind: event.kind,
      reference: event.reference,
      ...(event.actor === undefined ? {} : {actor: event.actor}),
      ...(event.data === undefined ? {} : {data: event.data}),
    },
    subject: event.subject,
    ...template,
  });
};
