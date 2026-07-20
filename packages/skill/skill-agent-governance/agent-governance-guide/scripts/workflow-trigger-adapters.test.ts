import assert from "node:assert/strict";
import {describe, it} from "node:test";
import type {TriggerSource} from "./workflow-runtime.ts";
import {adaptTriggerToWorkflowInvocation} from "./workflow-trigger-adapters.ts";

const sources: readonly TriggerSource[] = [
  "manual",
  "agent-harness-hook",
  "ci-cd-hook",
  "provider-webhook",
  "schedule",
  "sensor",
  "api",
  "agent",
];

describe("adaptTriggerToWorkflowInvocation", () => {
  it("keeps trigger source independent from the selected execution family", () => {
    for (const source of sources) {
      const invocation = adaptTriggerToWorkflowInvocation(
        {
          source,
          nativeReference: `${source}:event-1`,
          actor: `external:${source}`,
          idempotencyKey: "delivery-1",
          ...(source === "agent"
            ? {
                parentInvocationReference: "workflow:parent-1",
                childDepth: 1,
              }
            : {}),
          subject: {reference: "repository:xonovex", revision: "commit:abc123"},
        },
        {
          workflow: {operation: "review-run"},
          execution: {
            family: "workflow-script",
            module: "review/1",
            budget: {timeoutSeconds: 30},
          },
          evidenceProviderReference: "evidence:provider-1",
        },
      );

      assert.equal(invocation.trigger.source, source);
      assert.equal(invocation.execution.family, "workflow-script");
      assert.match(invocation.invocationId, /^workflow:sha256:[0-9a-f]{64}$/u);
    }
  });

  it("changes invocation identity when a native delivery id changes", () => {
    const template = {
      workflow: {operation: "review-run"},
      execution: {
        family: "workflow-script" as const,
        module: "review/1",
        budget: {timeoutSeconds: 30},
      },
      evidenceProviderReference: "evidence:provider-1",
    };
    const event = {
      source: "ci-cd-hook" as const,
      nativeReference: "github-actions:run-1",
      actor: "external:github-actions",
      idempotencyKey: "delivery-1",
      subject: {reference: "repository:xonovex", revision: "commit:abc123"},
    };

    const first = adaptTriggerToWorkflowInvocation(event, template);
    const second = adaptTriggerToWorkflowInvocation(
      {...event, idempotencyKey: "delivery-2"},
      template,
    );

    assert.notEqual(first.invocationId, second.invocationId);
  });
});
