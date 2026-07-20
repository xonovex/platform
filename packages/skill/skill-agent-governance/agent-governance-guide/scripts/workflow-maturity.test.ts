import assert from "node:assert/strict";
import {describe, it} from "node:test";
import {assessWorkflowMaturity} from "./workflow-maturity.ts";
import type {WorkflowCompositionExplanation} from "./workflow-runtime.ts";

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

    assert.equal(assessment.achieved, "A2");
    assert.deepEqual(assessment.levels[2]?.missingCapabilities, [
      "oversight:escalation",
    ]);
  });
});
