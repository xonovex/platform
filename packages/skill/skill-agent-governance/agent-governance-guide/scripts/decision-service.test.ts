import assert from "node:assert/strict";
import {readFile, writeFile} from "node:fs/promises";
import {createServer} from "node:http";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {describe, it} from "node:test";
import {
  governanceDecisionApiVersion,
  governanceEvaluatorVersion,
  governancePolicyVersion,
} from "../../../skill-workflow/workflow-guide/scripts/governance-decision.ts";
import {createDecisionRequestHandler} from "./decision-service.ts";
import {
  createJsonlEnforcementRecorder,
  createJsonlEvidenceRecorder,
} from "./verdict-evidence.ts";

describe("createJsonlEvidenceRecorder", () => {
  it("writes one verdict per correlation id", async () => {
    const path = join(
      tmpdir(),
      `xonovex-verdict-${String(process.pid)}-${String(Date.now())}.jsonl`,
    );
    const recorder = await createJsonlEvidenceRecorder(path);
    const verdict = {
      apiVersion: governanceDecisionApiVersion,
      correlationId: "correlation-1",
      subjectReference: "release:1",
      subjectRevision: "commit:abc123",
      decision: "allow" as const,
      policyVersion: governancePolicyVersion,
      evaluatorVersion: governanceEvaluatorVersion,
      operationDigest:
        "sha256:8c1f6d1e13c568d59acb1a16c2b46017911a03a121aca319c15c00f72f146f7f",
      protectedTargetsDigest:
        "sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945",
    };

    await Promise.all([recorder.record(verdict), recorder.record(verdict)]);

    const records = (await readFile(path, "utf8")).trim().split("\n");
    assert.equal(records.length, 1);
    assert.deepEqual(JSON.parse(records[0] ?? "{}"), verdict);
  });

  it("rejects reuse of a correlation id for different evidence", async () => {
    const path = join(
      tmpdir(),
      `xonovex-verdict-collision-${String(process.pid)}-${String(Date.now())}.jsonl`,
    );
    const recorder = await createJsonlEvidenceRecorder(path);
    const verdict = {
      apiVersion: governanceDecisionApiVersion,
      correlationId: "correlation-1",
      subjectReference: "release:1",
      subjectRevision: "commit:abc123",
      decision: "allow" as const,
      policyVersion: governancePolicyVersion,
      evaluatorVersion: governanceEvaluatorVersion,
      operationDigest:
        "sha256:8c1f6d1e13c568d59acb1a16c2b46017911a03a121aca319c15c00f72f146f7f",
      protectedTargetsDigest:
        "sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945",
    };

    await recorder.record(verdict);

    await assert.rejects(
      recorder.record({...verdict, decision: "deny"}),
      /decision-evidence-correlation-collision/u,
    );
  });

  it("rejects conflicting correlation evidence already present on disk", async () => {
    const path = join(
      tmpdir(),
      `xonovex-verdict-persisted-collision-${String(process.pid)}-${String(Date.now())}.jsonl`,
    );
    const verdict = {
      apiVersion: governanceDecisionApiVersion,
      correlationId: "correlation-persisted-collision",
      subjectReference: "release:1",
      subjectRevision: "commit:abc123",
      decision: "allow" as const,
      policyVersion: governancePolicyVersion,
      evaluatorVersion: governanceEvaluatorVersion,
      operationDigest:
        "sha256:8c1f6d1e13c568d59acb1a16c2b46017911a03a121aca319c15c00f72f146f7f",
      protectedTargetsDigest:
        "sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945",
    };
    await writeFile(
      path,
      `${JSON.stringify(verdict)}\n${JSON.stringify({...verdict, decision: "deny"})}\n`,
      "utf8",
    );

    await assert.rejects(
      createJsonlEvidenceRecorder(path),
      /decision-evidence-log-correlation-collision/u,
    );
  });
});

describe("createDecisionRequestHandler", () => {
  it("returns correlated content-addressed decision and enforcement receipts", async () => {
    const identity = `${String(process.pid)}-${String(Date.now())}`;
    const verdicts = await createJsonlEvidenceRecorder(
      join(tmpdir(), `xonovex-http-verdict-${identity}.jsonl`),
      "pvc://evidence/verdicts.jsonl",
    );
    const enforcements = await createJsonlEnforcementRecorder(
      join(tmpdir(), `xonovex-http-enforcement-${identity}.jsonl`),
      "pvc://evidence/enforcements.jsonl",
    );
    const server = createServer(
      createDecisionRequestHandler(verdicts.record, enforcements.record),
    );
    await new Promise<void>((resolve) =>
      server.listen(0, "127.0.0.1", resolve),
    );
    try {
      const address = server.address();
      assert.notEqual(address, null);
      assert.equal(typeof address, "object");
      if (address === null || typeof address === "string") return;
      const baseUrl = `http://127.0.0.1:${String(address.port)}`;
      const decisionResponse = await fetch(`${baseUrl}/v1/decisions`, {
        method: "POST",
        headers: {"content-type": "application/json"},
        body: JSON.stringify({
          apiVersion: governanceDecisionApiVersion,
          correlationId: "correlation-http-1",
          subject: {reference: "release:1", revision: "commit:abc123"},
          policy: {version: governancePolicyVersion, enforcement: "mandatory"},
          operation: {
            kind: "independence",
            input: {
              required: "distinct-identity",
              decider: "reviewer",
              author: "author",
              failureCode: "independence-failed",
            },
          },
        }),
      });
      assert.equal(decisionResponse.status, 200);
      const verdict = (await decisionResponse.json()) as Record<
        string,
        unknown
      >;
      assert.equal(verdict.decision, "allow");

      const enforcementResponse = await fetch(`${baseUrl}/v1/enforcements`, {
        method: "POST",
        headers: {"content-type": "application/json"},
        body: JSON.stringify({
          apiVersion: verdict.apiVersion,
          correlationId: verdict.correlationId,
          subjectReference: verdict.subjectReference,
          subjectRevision: verdict.subjectRevision,
          outcome: verdict.decision,
          policyVersion: verdict.policyVersion,
          evaluatorVersion: verdict.evaluatorVersion,
          operationDigest: verdict.operationDigest,
          protectedTargetsDigest: verdict.protectedTargetsDigest,
          decisionEvidenceReference: verdict.evidenceReference,
          enforcementPoint: "test:http",
        }),
      });
      assert.equal(enforcementResponse.status, 202);
      const receipt = (await enforcementResponse.json()) as Record<
        string,
        unknown
      >;
      assert.equal(receipt.apiVersion, governanceDecisionApiVersion);
      assert.equal(receipt.correlationId, "correlation-http-1");
      assert.equal(receipt.status, "recorded");
      assert.match(String(receipt.evidenceReference), /#sha256:[0-9a-f]{64}$/u);
    } finally {
      await new Promise<void>((resolve, reject) =>
        server.close((error) =>
          error === undefined ? resolve() : reject(error),
        ),
      );
    }
  });
});
