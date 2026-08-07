import {describe, expect, it} from "vitest";
import {isRecord} from "../../../src/records.js";

describe("isRecord", () => {
  it("accepts a keyed object", () => {
    expect(isRecord({type: "stream_event"})).toBe(true);
    expect(isRecord({})).toBe(true);
  });

  it("rejects the object-typed values that carry no named fields", () => {
    expect(isRecord(null)).toBe(false);
    expect(isRecord([{type: "stream_event"}])).toBe(false);
  });

  it("rejects primitives", () => {
    expect(isRecord("stream_event")).toBe(false);
    expect(isRecord(0)).toBe(false);
    expect(isRecord(undefined)).toBe(false);
  });
});
