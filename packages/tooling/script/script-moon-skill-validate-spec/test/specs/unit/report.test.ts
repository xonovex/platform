import {
  drift,
  fail,
  pass,
  warn,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import {describe, expect, it} from "vitest";
import {renderReport} from "../../../src/report.js";

describe("renderReport", () => {
  it("groups the findings by level under the skill header", () => {
    const report = renderReport(
      [fail("f"), pass("p"), drift("d"), warn("w")],
      "/s/SKILL.md",
      "/s",
      false,
    );

    expect(report).toEqual({
      lines: [
        "Validation: /s/SKILL.md",
        "Skill dir: /s",
        "",
        "[PASS] p",
        "[WARN] w",
        "[DRIFT] d",
        "[FAIL] f",
        "",
        "Result: FAIL (1 error(s), 1 warning(s))",
      ],
      exitCode: 1,
    });
  });

  it("passes warnings unless the run is strict", () => {
    const findings = [warn("w")];

    const lenient = renderReport(findings, "s", "d", false);
    const strict = renderReport(findings, "s", "d", true);

    expect([lenient.exitCode, lenient.lines.at(-1)]).toEqual([
      0,
      "Result: PASS with 1 warning(s)",
    ]);
    expect([strict.exitCode, strict.lines.at(-1)]).toEqual([
      1,
      "Result: FAIL (1 warning(s) in strict mode)",
    ]);
  });

  it("passes a drift finding even in strict mode", () => {
    const report = renderReport([drift("d")], "s", "d", true);

    expect([report.exitCode, report.lines.at(-1)]).toEqual([
      0,
      "Result: PASS (no warnings)",
    ]);
  });
});
