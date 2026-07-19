import {describe, expect, it} from "vitest";
import {parsePackageIdentity, publishArgs} from "./publish.js";

describe("parsePackageIdentity", () => {
  it("accepts a scoped package and semantic version", () => {
    expect(
      parsePackageIdentity({name: "@xonovex/example", version: "1.2.3-beta.1"}),
    ).toEqual({name: "@xonovex/example", version: "1.2.3-beta.1"});
  });

  it.each([
    {name: "example; id", version: "1.2.3"},
    {name: "example", version: "latest"},
    {name: "example"},
  ])("rejects unsafe or incomplete metadata: %o", (value) => {
    expect(() => parsePackageIdentity(value)).toThrow();
  });
});

describe("publishArgs", () => {
  it("uses provenance for a real publication", () => {
    expect(publishArgs(false)).toEqual([
      "publish",
      "--provenance",
      "--access",
      "public",
    ]);
  });

  it("uses npm dry-run without provenance", () => {
    expect(publishArgs(true)).toEqual([
      "publish",
      "--dry-run",
      "--access",
      "public",
    ]);
  });
});
