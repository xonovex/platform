import {describe, expect, it} from "vitest";
import {z} from "zod";
import {
  FiniteNumberTextSchema,
  parseWithSchema,
  PositiveIntegerTextSchema,
} from "../../../src/option-schemas.js";

describe("option schemas", () => {
  it("parses whole numbers above zero from text", () => {
    expect(parseWithSchema(PositiveIntegerTextSchema, "4")).toEqual({
      success: true,
      data: 4,
    });
    expect(parseWithSchema(PositiveIntegerTextSchema, "1.5")).toEqual({
      success: false,
      error: "must be a positive integer",
    });
  });

  it("parses finite numbers from text", () => {
    expect(parseWithSchema(FiniteNumberTextSchema, "0.25")).toEqual({
      success: true,
      data: 0.25,
    });
    expect(parseWithSchema(FiniteNumberTextSchema, " ")).toMatchObject({
      success: false,
      error: "must be a finite number",
    });
  });

  it("names the path of every issue", () => {
    const schema = z.object({
      runs: PositiveIntegerTextSchema,
      name: z.string(),
    });

    expect(parseWithSchema(schema, {runs: "x", name: 1})).toEqual({
      success: false,
      error:
        "runs: must be a positive integer; name: Invalid input: expected string, received number",
    });
  });
});
