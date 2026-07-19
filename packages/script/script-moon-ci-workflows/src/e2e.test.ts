import {describe, expect, it} from "vitest";
import {
  e2eMatrix,
  e2eTask,
  renderE2eSummary,
  renderMissingE2eSummary,
} from "./e2e.js";

describe("e2eMatrix", () => {
  it("selects only the four Kind suites", () => {
    expect(e2eMatrix()).toEqual(["e2e", "e2e-gvisor", "e2e-kata", "e2e-coco"]);
    expect(e2eMatrix()).not.toContain("integration");
  });
});

describe("e2eTask", () => {
  it("maps a suite to its Moon task", () => {
    expect(e2eTask("e2e-kata")).toBe("agent-operator-go:go-test-e2e-kata");
  });

  it("rejects a suite outside the supported matrix", () => {
    expect(() => e2eTask("integration")).toThrow("unknown E2E suite");
  });
});

describe("renderE2eSummary", () => {
  it("reports executed and skipped tests", () => {
    const summary = renderE2eSummary(
      "e2e-kata",
      [
        "--- PASS: TestReady",
        "--- FAIL: TestBroken",
        "--- SKIP: TestNeedsKvm (0.00s)",
      ].join("\n"),
    );

    expect(summary).toContain("| 1 | 1 | 1 |");
    expect(summary).toContain("--- SKIP: TestNeedsKvm");
  });

  it("explains when a suite produced no log", () => {
    expect(renderMissingE2eSummary("e2e")).toContain("No test log produced");
  });
});
