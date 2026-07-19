import assert from "node:assert/strict";
import {readFile} from "node:fs/promises";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {describe, it} from "node:test";
import {governanceDecisionApiVersion} from "../../../skill-workflow/workflow-guide/scripts/governance-decision.ts";
import {createJsonlEvidenceRecorder} from "./verdict-evidence.ts";

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
      decision: "allow" as const,
      policyVersion: "governance-policy/1",
    };

    await Promise.all([recorder.record(verdict), recorder.record(verdict)]);

    const records = (await readFile(path, "utf8")).trim().split("\n");
    assert.equal(records.length, 1);
    assert.deepEqual(JSON.parse(records[0] ?? "{}"), verdict);
  });
});
