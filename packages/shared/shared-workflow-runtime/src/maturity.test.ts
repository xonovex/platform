import {describe, expect, it} from "vitest";
import {assessWorkflowMaturity} from "./maturity.js";
import type {WorkflowCompositionExplanation} from "./runtime.js";

describe("assessWorkflowMaturity", () => {
  it("derives caller-defined A-levels from observed capabilities", () => {
    const composition = {
      invocationId: "invocation-1",
      trigger: "manual",
      executor: {plugin: "agent", capabilities: ["execution:agent"]},
      controls: [],
      evidenceSinks: [],
      requiredCapabilities: [],
      availableCapabilities: [
        "execution:agent",
        "oversight:critique",
        "oversight:approval",
      ],
      missingCapabilities: [],
      enforcementPoints: [],
    } satisfies WorkflowCompositionExplanation;

    const assessment = assessWorkflowMaturity(composition, {
      name: "team-agent-maturity",
      levels: [
        {name: "A1", requiredCapabilities: ["oversight:critique"]},
        {
          name: "A2",
          requiredCapabilities: ["oversight:critique", "oversight:approval"],
        },
        {
          name: "A3",
          requiredCapabilities: [
            "oversight:critique",
            "oversight:approval",
            "oversight:escalation",
          ],
        },
      ],
    });

    expect(assessment.achieved).toBe("A2");
    expect(assessment.levels[2]?.missingCapabilities).toEqual([
      "oversight:escalation",
    ]);
  });
});
