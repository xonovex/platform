import {describe, expect, it} from "vitest";
import {
  dependencySourceFailures,
  type ProjectFile,
} from "../../../src/dependency-sources.js";

const project = (
  id: string,
  source: string,
  body: Readonly<Record<string, unknown>>,
): ProjectFile => ({
  id,
  source,
  text: Object.entries(body)
    .map(([key, value]) => `${key}: ${JSON.stringify(value)}`)
    .join("\n"),
});

const COMMON = project("common", "packages/script/common", {
  tags: ["typescript-script"],
});

const dependent = (
  fileGroups: Readonly<Record<string, readonly string[]>>,
): ProjectFile =>
  project("consumer", "packages/script/consumer", {
    tags: ["typescript-script"],
    dependsOn: ["common"],
    fileGroups,
  });

describe("dependencySourceFailures", () => {
  it("accepts a project that declares its dependency's source", () => {
    expect(
      dependencySourceFailures([
        COMMON,
        dependent({
          dependencySources: ["/packages/script/common/src/**/*"],
        }),
      ]),
    ).toEqual([]);
  });

  it("reports a dependency whose source is not declared", () => {
    const failures = dependencySourceFailures([COMMON, dependent({})]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("/packages/script/common/src/**/*");
    expect(failures[0]).toContain("serves a cached pass");
  });

  it("reports a declared source that dependsOn does not reach", () => {
    const failures = dependencySourceFailures([
      COMMON,
      dependent({
        dependencySources: [
          "/packages/script/common/src/**/*",
          "/packages/script/stale/src/**/*",
        ],
      }),
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("/packages/script/stale/src/**/*");
    expect(failures[0]).toContain("does not reach");
  });

  it("requires the transitive closure, not just the direct dependency", () => {
    const middle = project("middle", "packages/script/middle", {
      tags: ["typescript-script"],
      dependsOn: ["common"],
      fileGroups: {dependencySources: ["/packages/script/common/src/**/*"]},
    });
    const consumer = project("consumer", "packages/script/consumer", {
      tags: ["typescript-script"],
      dependsOn: ["middle"],
      fileGroups: {dependencySources: ["/packages/script/middle/src/**/*"]},
    });

    const failures = dependencySourceFailures([COMMON, middle, consumer]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("consumer");
    expect(failures[0]).toContain("/packages/script/common/src/**/*");
  });

  it("reads a ts-config dependency, which a tsconfig extends from", () => {
    const tsConfig = project(
      "ts-config-node",
      "packages/config/ts-config-node",
      {
        tags: ["tsconfig"],
      },
    );
    const consumer = project("consumer", "packages/script/consumer", {
      tags: ["typescript-script"],
      dependsOn: ["ts-config-node"],
      fileGroups: {
        dependencySources: ["/packages/config/ts-config-node/src/**/*"],
      },
    });

    expect(dependencySourceFailures([tsConfig, consumer])).toEqual([]);
  });

  it("ignores a Go dependency, which no TypeScript task reads", () => {
    const go = project("shared-core-go", "packages/shared/shared-core-go", {
      tags: ["go"],
    });
    const consumer = project("consumer", "packages/agent/consumer", {
      tags: ["typescript"],
      dependsOn: ["shared-core-go"],
    });

    expect(dependencySourceFailures([go, consumer])).toEqual([]);
  });

  it("leaves the template default alone for a project with no dependencies", () => {
    expect(
      dependencySourceFailures([
        project("solo", "packages/script/solo", {
          tags: ["typescript-script"],
          fileGroups: {dependencySources: ["src/**/*"]},
        }),
      ]),
    ).toEqual([]);
  });

  it("skips a project that inherits no TypeScript task template", () => {
    expect(
      dependencySourceFailures([
        COMMON,
        project("shell", "packages/script/shell", {
          tags: ["shell"],
          dependsOn: ["common"],
        }),
      ]),
    ).toEqual([]);
  });

  it("reports a project file that is not readable", () => {
    const failures = dependencySourceFailures([
      {id: "broken", source: "packages/script/broken", text: "tags: [\n"},
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("broken");
  });
});
