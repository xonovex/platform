import assert from "node:assert/strict";
import {describe, it} from "node:test";
import {toOTLPLogExport} from "./verdict-telemetry.ts";

describe("toOTLPLogExport", () => {
  it("emits correlated OTLP JSON without subject content", () => {
    const payload = toOTLPLogExport(
      {
        correlationId: "correlation-1",
        kind: "governance.verdict",
        outcome: "deny",
        failureCode: "protected-path-denied",
        policyVersion: "governance-policy/1",
      },
      1_000_000n,
    );
    const serialized = JSON.stringify(payload);
    assert.match(serialized, /correlation-1/);
    assert.match(serialized, /protected-path-denied/);
    for (const forbidden of ["prompt", "secret", "subjectReference"]) {
      assert.doesNotMatch(
        serialized.toLowerCase(),
        new RegExp(forbidden.toLowerCase()),
      );
    }
  });
});
