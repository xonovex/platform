import {describe, expect, it} from "vitest";
import {skillEvalModelDefaults} from "../../../src/skill-eval-models.js";

describe("skill evaluation model defaults", () => {
  it("uses canonical Claude model identifiers", () => {
    expect(skillEvalModelDefaults("claude")).toEqual({
      generation: "claude-haiku-4-5-20251001",
      judge: "claude-sonnet-5",
    });
  });

  it("uses the canonical Codex model identifier", () => {
    expect(skillEvalModelDefaults("codex")).toEqual({
      generation: "gpt-5.3-codex",
      judge: "gpt-5.3-codex",
    });
  });
});
