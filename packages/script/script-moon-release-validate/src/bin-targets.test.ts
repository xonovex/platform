import {describe, expect, it} from "vitest";
import {
  binEntries,
  binTargetFailures,
  type PackageBins,
} from "./bin-targets.js";

const privatePackage: PackageBins = {
  manifestPath: "packages/script/script-moon-npm-check/package.json",
  packageDirectory: "packages/script/script-moon-npm-check",
  isPrivate: true,
  files: undefined,
  bins: [{name: "moon-npm-check", target: "./bin.js"}],
};

const publishedPackage: PackageBins = {
  manifestPath: "packages/agent/agent-cli-go/package.json",
  packageDirectory: "packages/agent/agent-cli-go",
  isPrivate: false,
  files: ["dist/bin.js", "dist/bin.js.map"],
  bins: [{name: "agent-cli", target: "dist/bin.js"}],
};

const tracked = (...paths: readonly string[]): ReadonlySet<string> =>
  new Set(paths);

describe("binEntries", () => {
  it("reads a bin map", () => {
    expect(binEntries({"moon-npm-check": "./bin.js"}, "@xonovex/x")).toEqual([
      {name: "moon-npm-check", target: "./bin.js"},
    ]);
  });

  it("names a string bin after the unscoped package name", () => {
    expect(binEntries("./bin.js", "@xonovex/agent-cli-go")).toEqual([
      {name: "agent-cli-go", target: "./bin.js"},
    ]);
    expect(binEntries("./bin.js", "plain")).toEqual([
      {name: "plain", target: "./bin.js"},
    ]);
  });

  it("reads no entry from an absent bin", () => {
    expect(binEntries(undefined, "@xonovex/x")).toEqual([]);
  });
});

describe("binTargetFailures", () => {
  it("accepts a private package whose bin target is committed", () => {
    expect(
      binTargetFailures({
        packages: [privatePackage],
        trackedPaths: tracked("packages/script/script-moon-npm-check/bin.js"),
      }),
    ).toEqual([]);
  });

  it("reports a private package whose bin target is generated", () => {
    const failures = binTargetFailures({
      packages: [
        {
          ...privatePackage,
          bins: [{name: "moon-npm-check", target: "./dist/src/index.js"}],
        },
      ],
      trackedPaths: tracked("packages/script/script-moon-npm-check/bin.js"),
    });

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("dist/src/index.js");
    expect(failures[0]).toContain("not committed");
  });

  it("accepts a published package whose bin target is packed", () => {
    expect(
      binTargetFailures({
        packages: [publishedPackage],
        trackedPaths: tracked(),
      }),
    ).toEqual([]);
  });

  it("accepts a published package packing the directory above the target", () => {
    expect(
      binTargetFailures({
        packages: [{...publishedPackage, files: ["dist"]}],
        trackedPaths: tracked(),
      }),
    ).toEqual([]);
  });

  it("reports a published package whose bin target is neither committed nor packed", () => {
    const failures = binTargetFailures({
      packages: [{...publishedPackage, files: ["README.md"]}],
      trackedPaths: tracked(),
    });

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("neither committed nor inside the files");
  });

  it("does not let a files allowlist excuse a private package", () => {
    expect(
      binTargetFailures({
        packages: [
          {
            ...privatePackage,
            files: ["dist"],
            bins: [{name: "moon-npm-check", target: "dist/src/index.js"}],
          },
        ],
        trackedPaths: tracked(),
      }),
    ).toHaveLength(1);
  });

  it("reports every bin a package declares", () => {
    expect(
      binTargetFailures({
        packages: [
          {
            ...publishedPackage,
            files: [],
            bins: [
              {name: "agent-cli", target: "dist/bin.js"},
              {name: "agent-cli-go", target: "dist/bin.js"},
            ],
          },
        ],
        trackedPaths: tracked(),
      }),
    ).toHaveLength(2);
  });

  it("accepts a package that declares no bin", () => {
    expect(
      binTargetFailures({
        packages: [{...privatePackage, bins: []}],
        trackedPaths: tracked(),
      }),
    ).toEqual([]);
  });
});
