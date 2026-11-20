import {describe, expect, it} from "vitest";
import {renderReport} from "../../../src/report.js";
import {issue} from "../../../src/validation.js";

describe("renderReport", () => {
  it("passes with warnings and prints them on stderr", () => {
    expect(
      renderReport({
        skills: 2,
        issues: [issue("x.code", "a.md", "advice", "warning")],
      }),
    ).toEqual({
      exitCode: 0,
      lines: [
        {stream: "stderr", text: "[WARN] a.md: x.code: advice"},
        {stream: "stdout", text: "Result: PASS (2 skill(s), 1 warning(s))"},
      ],
    });
  });
});
