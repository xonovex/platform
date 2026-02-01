import {describe, expect, it} from "vitest";
import {bumpVersion, updateDependencyVersions} from "./bump.js";

describe("bumpVersion", () => {
  it("should bump patch", () => {
    expect(bumpVersion("1.2.3", "patch")).toBe("1.2.4");
  });

  it("should bump minor and reset patch", () => {
    expect(bumpVersion("1.2.3", "minor")).toBe("1.3.0");
  });

  it("should bump major and reset minor and patch", () => {
    expect(bumpVersion("1.2.3", "major")).toBe("2.0.0");
  });

  it("should handle 0.0.0", () => {
    expect(bumpVersion("0.0.0", "patch")).toBe("0.0.1");
    expect(bumpVersion("0.0.0", "minor")).toBe("0.1.0");
    expect(bumpVersion("0.0.0", "major")).toBe("1.0.0");
  });

  it("should create prerelease with preid", () => {
    expect(bumpVersion("1.2.3", "patch", "beta")).toBe("1.2.4-beta.0");
  });

  it("should increment prerelease number for matching preid", () => {
    expect(bumpVersion("1.2.4-beta.0", "patch", "beta")).toBe("1.2.4-beta.1");
  });

  it("should bump core for higher bump type with preid", () => {
    expect(bumpVersion("1.2.4-beta.1", "minor", "beta")).toBe("1.3.0-beta.0");
  });

  it("should strip prerelease when no preid given", () => {
    expect(bumpVersion("1.2.4-beta.1", "patch")).toBe("1.2.5");
  });

  it("should create major prerelease", () => {
    expect(bumpVersion("1.2.3", "major", "alpha")).toBe("2.0.0-alpha.0");
  });
});

describe("updateDependencyVersions", () => {
  it("should update an existing dependency", () => {
    const deps = {"@xonovex/core": "1.0.0"};
    const result = updateDependencyVersions(deps, "@xonovex/core", "1.0.1");
    expect(result).toBe(true);
    expect(deps["@xonovex/core"]).toBe("1.0.1");
  });

  it("should return false if dependency is not present", () => {
    const deps = {"@xonovex/core": "1.0.0"};
    const result = updateDependencyVersions(deps, "@xonovex/other", "1.0.1");
    expect(result).toBe(false);
  });

  it("should return false for undefined deps", () => {
    expect(updateDependencyVersions(undefined, "@xonovex/core", "1.0.1")).toBe(
      false,
    );
  });
});
