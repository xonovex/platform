import {describe, expect, it} from "vitest";
import {detectDepUpdates, getWorkspaceDeps} from "../../../src/dep-updates.js";
import {emptyHistory, manifestsAtHead} from "../../util/version-git.js";

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
      optionalDependencies: {
        "@xonovex/agent-cli-go-linux-x64": "1.0.0",
        "some-pkg": "2.0.0",
      },
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

describe("detectDepUpdates", () => {
  const path = "packages/core/package.json";

  it("reports each workspace dependency whose range changed since HEAD", () => {
    const git = emptyHistory({
      readFile: manifestsAtHead({
        [path]: {
          dependencies: {"@xonovex/a": "1.0.0", "@xonovex/b": "1.0.0"},
        },
      }),
    });

    expect(
      detectDepUpdates(
        "/repo",
        `/repo/${path}`,
        {
          dependencies: {
            "@xonovex/a": "1.1.0",
            "@xonovex/b": "1.0.0",
            "@xonovex/c": "1.0.0",
          },
        },
        git,
      ),
    ).toEqual([{name: "@xonovex/a", version: "1.1.0"}]);
  });

  it("reports nothing for a package that HEAD does not hold", () => {
    expect(
      detectDepUpdates(
        "/repo",
        `/repo/${path}`,
        {dependencies: {"@xonovex/a": "1.1.0"}},
        emptyHistory(),
      ),
    ).toEqual([]);
  });
});
