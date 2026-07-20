import {describe, expect, it} from "vitest";
import {adaptTriggerToWorkflowInvocation} from "./trigger-adapters.js";

const template = {
  operation: "review",
  executor: {plugin: "review-script"},
  controls: [],
  evidence: [],
  requiredCapabilities: [],
  metadata: {},
};

describe("adaptTriggerToWorkflowInvocation", () => {
  it("accepts manual, harness-hook, CI, schedule, and domain-specific trigger names", () => {
    for (const kind of [
      "manual",
      "hook/claude/pre-tool-use",
      "ci/github/pull-request",
      "schedule/nightly",
      "sensor/dependency-drift",
    ]) {
      const invocation = adaptTriggerToWorkflowInvocation(
        {
          kind,
          reference: `${kind}:event-1`,
          actor: "external:test",
          idempotencyKey: "delivery-1",
          subject: {reference: "repository:xonovex", revision: "commit:abc123"},
        },
        template,
      );

      expect(invocation.trigger.kind).toBe(kind);
      expect(invocation.executor.plugin).toBe("review-script");
      expect(invocation.invocationId).toMatch(
        /^workflow:sha256:[0-9a-f]{64}$/u,
      );
    }
  });

  it("changes identity when the native idempotency key changes", () => {
    const event = {
      kind: "ci/github/pull-request",
      reference: "github:delivery-42",
      idempotencyKey: "delivery-1",
      subject: {reference: "repository:xonovex", revision: "commit:abc123"},
    };

    const first = adaptTriggerToWorkflowInvocation(event, template);
    const second = adaptTriggerToWorkflowInvocation(
      {...event, idempotencyKey: "delivery-2"},
      template,
    );

    expect(first.invocationId).not.toBe(second.invocationId);
  });

  it("changes identity when trusted composition wiring changes", () => {
    const event = {
      kind: "manual",
      reference: "manual:review-42",
      idempotencyKey: "review-42",
      subject: {reference: "repository:xonovex", revision: "commit:abc123"},
    };

    const first = adaptTriggerToWorkflowInvocation(event, template);
    const second = adaptTriggerToWorkflowInvocation(event, {
      ...template,
      controls: [{plugin: "approval", mode: "observe"}],
    });

    expect(first.invocationId).not.toBe(second.invocationId);
  });
});
