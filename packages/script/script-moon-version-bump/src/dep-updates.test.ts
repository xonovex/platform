import {describe, expect, it} from "vitest";
import {getWorkspaceDeps} from "./dep-updates.js";

describe("getWorkspaceDeps", () => {
  it("should extract @xonovex dependencies", () => {
    const deps = getWorkspaceDeps({
      dependencies: {"@xonovex/core": "1.0.0", lodash: "4.0.0"},
      devDependencies: {"@xonovex/eslint-config": "2.0.0"},
    });
    expect(deps.get("@xonovex/core")).toBe("1.0.0");
    expect(deps.get("@xonovex/eslint-config")).toBe("2.0.0");
    expect(deps.has("lodash")).toBe(false);
  });

  it("should extract @xonovex optionalDependencies", () => {
    const deps = getWorkspaceDeps({
      optionalDependencies: {"@xonovex/agent-cli-go-linux-x64": "1.0.0", "some-pkg": "2.0.0"},
    });
    expect(deps.get("@xonovex/agent-cli-go-linux-x64")).toBe("1.0.0");
    expect(deps.has("some-pkg")).toBe(false);
  });

  it("should return empty map for no deps", () => {
    expect(getWorkspaceDeps({})).toEqual(new Map());
  });

  it("should handle missing dep fields", () => {
    expect(getWorkspaceDeps({name: "test"})).toEqual(new Map());
  });
});
