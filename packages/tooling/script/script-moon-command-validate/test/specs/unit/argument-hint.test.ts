import {describe, expect, it} from "vitest";
import {parseArgumentHint} from "../../../src/argument-hint.js";

describe("parseArgumentHint", () => {
  it("parses positional, required, optional, valued, and repeatable arguments", () => {
    expect(
      parseArgumentHint(
        "<subject> --feedback <feedback>... [--perspective <selection>...] [--request <path>] [--effect <mode>] [--force]",
      ),
    ).toEqual([
      {
        kind: "positional",
        name: "subject",
        required: true,
        repeatable: false,
        valueName: "subject",
      },
      {
        kind: "flag",
        name: "feedback",
        required: true,
        repeatable: true,
        valueName: "feedback",
      },
      {
        kind: "flag",
        name: "perspective",
        required: false,
        repeatable: true,
        valueName: "selection",
      },
      {
        kind: "flag",
        name: "request",
        required: false,
        repeatable: false,
        valueName: "path",
      },
      {
        kind: "flag",
        name: "effect",
        required: false,
        repeatable: false,
        valueName: "mode",
      },
      {
        kind: "flag",
        name: "force",
        required: false,
        repeatable: false,
      },
    ]);
  });

  it("parses a repeatable positional", () => {
    expect(parseArgumentHint("[target...]")).toEqual([
      {
        kind: "positional",
        name: "target",
        required: false,
        repeatable: true,
        valueName: "target",
      },
    ]);
  });

  it("keeps an enumerated flag value inside one placeholder", () => {
    expect(
      parseArgumentHint("[--tier <auto|aggressive|moderate|conservative>]"),
    ).toEqual([
      {
        kind: "flag",
        name: "tier",
        required: false,
        repeatable: false,
        valueName: "auto|aggressive|moderate|conservative",
      },
    ]);
  });
});
