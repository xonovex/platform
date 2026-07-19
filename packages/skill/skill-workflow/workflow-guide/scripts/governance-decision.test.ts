import assert from "node:assert/strict";
import {describe, it} from "node:test";
import {
  decideGovernance,
  governanceDecisionApiVersion,
  type GovernanceDecisionRequest,
  type VerdictEvidence,
} from "./governance-decision.ts";

const baseRequest = (
  operation: GovernanceDecisionRequest["operation"],
  enforcement: "mandatory" | "advisory" = "mandatory",
): GovernanceDecisionRequest => ({
  apiVersion: governanceDecisionApiVersion,
  correlationId: "correlation-1",
  subject: {reference: "release:1", revision: "abc123"},
  policy: {version: "governance-policy/1", enforcement},
  operation,
});

const memoryRecorder = () => {
  const records = new Map<string, VerdictEvidence>();
  return {
    records,
    record: async (evidence: VerdictEvidence): Promise<string> => {
      records.set(evidence.correlationId, evidence);
      return `memory:${evidence.correlationId}`;
    },
  };
};

describe("decideGovernance", () => {
  it("denies a self-approved release with the exact independence code", async () => {
    const evidence = memoryRecorder();
    const request = baseRequest({
      kind: "independence",
      input: {
        required: "distinct-identity",
        decider: "actor:author",
        author: "actor:author",
        failureCode: "release-independence-failed",
      },
    });

    const response = await decideGovernance(request, evidence.record);

    assert.equal(response.decision, "deny");
    assert.equal(response.failureCode, "release-independence-failed");
    assert.equal(evidence.records.size, 1);
  });

  it("allows a release approved by an independent actor", async () => {
    const evidence = memoryRecorder();
    const request = baseRequest({
      kind: "independence",
      input: {
        required: "distinct-identity",
        decider: "actor:reviewer",
        author: "actor:author",
        failureCode: "release-independence-failed",
      },
    });

    const response = await decideGovernance(request, evidence.record);

    assert.equal(response.decision, "allow");
    assert.equal(response.failureCode, undefined);
  });

  it("denies expired emergency access and allows an in-scope grant", async () => {
    const evidence = memoryRecorder();
    const access = {
      kind: "emergency-exception",
      owner: "human:operator",
      rationale: "restore production service",
      evidenceReference: "incident:42",
      start: "2026-01-01T00:00:00Z",
      expiresAt: "2026-02-01T00:00:00Z",
      approver: {
        identity: "reviewer",
        role: "incident-commander",
        type: "human",
        accountable: true,
      },
      scope: {actions: ["release"], targets: ["environment:production"]},
      compensatingControls: ["audit-log"],
      postUseReviewRequired: true,
      emergencyReason: "customer outage",
      accessEvidenceReference: "access:42",
      notificationReference: "notification:42",
      containment: "single deployment",
      revocationPlan: "automatic expiry",
      explicitInvocation: true,
    };
    const operation = (now: string) => ({
      kind: "emergency-access" as const,
      input: {
        access,
        request: {
          action: "release",
          target: {reference: "environment:production"},
          now,
        },
        profile: {independence: {emergencyAccess: "distinct-identity"}},
      },
    });

    const expired = await decideGovernance(
      baseRequest(operation("2026-03-01T00:00:00Z")),
      evidence.record,
    );
    const allowed = await decideGovernance(
      {
        ...baseRequest(operation("2026-01-15T00:00:00Z")),
        correlationId: "correlation-2",
      },
      evidence.record,
    );

    assert.equal(expired.failureCode, "emergency-access-expired");
    assert.equal(expired.decision, "deny");
    assert.equal(allowed.decision, "allow");
  });

  it("uses the least-adaptive executor and requires adaptive bounds", async () => {
    const evidence = memoryRecorder();
    const mechanical = await decideGovernance(
      baseRequest({
        kind: "development",
        input: {workShape: "mechanical", requestedExecutor: "agent"},
      }),
      evidence.record,
    );
    const adaptive = await decideGovernance(
      {
        ...baseRequest({
          kind: "development",
          input: {
            workShape: "adaptive",
            requestedExecutor: "agent",
            bounds: {
              purpose: "implement approved change",
              resultContract: "commit",
              toolScope: ["editor"],
              filesystemScope: ["workspace"],
              networkScope: [],
              maxDepth: 1,
              tokenBudget: 1_000,
              costBudget: 1,
              timeoutSeconds: 60,
              cancellation: "operator",
            },
          },
        }),
        correlationId: "correlation-2",
      },
      evidence.record,
    );

    assert.equal(mechanical.failureCode, "least-adaptive-executor-bypassed");
    assert.equal(mechanical.decision, "deny");
    assert.equal(adaptive.decision, "allow");
  });

  it("denies protected paths and allows ordinary source paths", async () => {
    const evidence = memoryRecorder();

    const denied = await decideGovernance(
      baseRequest({kind: "protected-path", input: {path: "secrets/probe.key"}}),
      evidence.record,
    );
    const allowed = await decideGovernance(
      {
        ...baseRequest({
          kind: "protected-path",
          input: {path: "src/probe.txt"},
        }),
        correlationId: "correlation-2",
      },
      evidence.record,
    );

    assert.equal(denied.decision, "deny");
    assert.equal(denied.failureCode, "protected-path-denied");
    assert.equal(allowed.decision, "allow");
  });

  it("fails closed for mandatory evidence outages and observes advisory outages", async () => {
    const unavailable = async (): Promise<string> => {
      throw new Error("evidence unavailable");
    };
    const operation = {
      kind: "independence" as const,
      input: {
        required: "none",
        failureCode: "release-independence-failed",
      },
    };

    const mandatory = await decideGovernance(
      baseRequest(operation),
      unavailable,
    );
    const advisory = await decideGovernance(
      baseRequest(operation, "advisory"),
      unavailable,
    );

    assert.equal(mandatory.decision, "deny");
    assert.equal(mandatory.failureCode, "decision-evidence-unavailable");
    assert.equal(advisory.decision, "observe");
    assert.equal(advisory.failureCode, "decision-evidence-unavailable");
  });
});
