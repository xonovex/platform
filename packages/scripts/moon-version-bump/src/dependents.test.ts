import {describe, expect, it} from "vitest";
import {updateDependent} from "./dependents.js";

describe("updateDependent", () => {
  it("should update dependency reference and bump version", () => {
    const pkg = {
      name: "@xonovex/agent-cli",
      version: "1.0.0",
      dependencies: {"@xonovex/core": "1.0.0"},
    };
    const result = updateDependent(
      pkg,
      "/path/to/agent-cli/package.json",
      "@xonovex/core",
      "1.1.0",
      () => "1.0.0",
    );
    expect(result.depsChanged).toBe(true);
    expect(result.versionBumped).toBe(true);
    expect(result.oldVersion).toBe("1.0.0");
    expect(result.newVersion).toBe("1.0.1");
    expect(pkg.dependencies["@xonovex/core"]).toBe("1.1.0");
  });

  it("should skip version bump if already bumped (git version differs)", () => {
    const pkg = {
      name: "@xonovex/agent-cli",
      version: "1.0.1",
      dependencies: {"@xonovex/core": "1.0.0"},
    };
    const result = updateDependent(
      pkg,
      "/path/to/agent-cli/package.json",
      "@xonovex/core",
      "1.1.0",
      () => "1.0.0",
    );
    expect(result.depsChanged).toBe(true);
    expect(result.versionBumped).toBe(false);
  });

  it("should skip version bump for private packages", () => {
    const pkg = {
      name: "@xonovex/internal",
      version: "1.0.0",
      private: true,
      dependencies: {"@xonovex/core": "1.0.0"},
    };
    const result = updateDependent(
      pkg,
      "/path/to/internal/package.json",
      "@xonovex/core",
      "1.1.0",
      () => "1.0.0",
    );
    expect(result.depsChanged).toBe(true);
    expect(result.versionBumped).toBe(false);
  });

  it("should return depsChanged false when package is not a dependent", () => {
    const pkg = {
      name: "@xonovex/unrelated",
      version: "1.0.0",
      dependencies: {"@xonovex/other": "1.0.0"},
    };
    const result = updateDependent(
      pkg,
      "/path/to/unrelated/package.json",
      "@xonovex/core",
      "1.1.0",
      () => "1.0.0",
    );
    expect(result.depsChanged).toBe(false);
    expect(result.versionBumped).toBe(false);
  });

  it("should update devDependencies", () => {
    const pkg = {
      name: "@xonovex/agent-cli",
      version: "1.0.0",
      devDependencies: {"@xonovex/eslint-config": "1.0.0"},
    };
    const result = updateDependent(
      pkg,
      "/path/to/agent-cli/package.json",
      "@xonovex/eslint-config",
      "1.1.0",
      () => "1.0.0",
    );
    expect(result.depsChanged).toBe(true);
    expect(result.versionBumped).toBe(true);
    expect(pkg.devDependencies["@xonovex/eslint-config"]).toBe("1.1.0");
  });

  it("should update peerDependencies", () => {
    const pkg = {
      name: "@xonovex/plugin",
      version: "1.0.0",
      peerDependencies: {"@xonovex/core": "1.0.0"},
    };
    const result = updateDependent(
      pkg,
      "/path/to/plugin/package.json",
      "@xonovex/core",
      "1.1.0",
      () => "1.0.0",
    );
    expect(result.depsChanged).toBe(true);
    expect(result.versionBumped).toBe(true);
  });

  it("should skip version bump when git version is undefined (new package)", () => {
    const pkg = {
      name: "@xonovex/new-pkg",
      version: "0.1.0",
      dependencies: {"@xonovex/core": "1.0.0"},
    };
    const result = updateDependent(
      pkg,
      "/path/to/new-pkg/package.json",
      "@xonovex/core",
      "1.1.0",
      () => undefined as string | undefined,
    );
    expect(result.depsChanged).toBe(true);
    expect(result.versionBumped).toBe(false);
  });
});
