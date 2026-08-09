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

  it("reports a declared source that nothing reaches", () => {
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
    expect(failures[0]).toContain(
      "neither dependsOn nor a config file reaches",
    );
  });

  // A config package cannot be a dependsOn edge without cycling, so the group is
  // the only place its source becomes an input.
  const ESLINT_CONFIG: ProjectFile = {
    ...project("eslint-config-cli", "packages/config/eslint-config-cli", {
      tags: ["typescript-config"],
      dependsOn: ["eslint-config-base"],
      fileGroups: {
        dependencySources: ["/packages/config/eslint-config-base/src/**/*"],
      },
    }),
    packageName: "@xonovex/eslint-config-cli",
  };
  const ESLINT_BASE = project(
    "eslint-config-base",
    "packages/config/eslint-config-base",
    {tags: ["typescript-config"]},
  );

  const reader = (
    dependencySources: readonly string[],
    configTexts: readonly string[] = [
      'export {default} from "@xonovex/eslint-config-cli";',
    ],
  ): ProjectFile => ({
    ...project("consumer", "packages/script/consumer", {
      tags: ["typescript-script"],
      fileGroups: {dependencySources},
    }),
    configTexts,
  });

  it("requires the source of a config package a config file names", () => {
    const failures = dependencySourceFailures([
      ESLINT_CONFIG,
      ESLINT_BASE,
      reader([]),
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain(
      "/packages/config/eslint-config-cli/src/**/*",
    );
  });

  it("follows what the named config package depends on in turn", () => {
    expect(
      dependencySourceFailures([
        ESLINT_CONFIG,
        ESLINT_BASE,
        reader([
          "/packages/config/eslint-config-cli/src/**/*",
          "/packages/config/eslint-config-base/src/**/*",
        ]),
      ]),
    ).toEqual([]);
  });

  it("ignores a package name no project in the graph publishes", () => {
    // dependsOn keeps the expected set non-empty, so the case reaches the
    // comparison rather than short-circuiting on a project with no inputs.
    const consumer: ProjectFile = {
      ...project("consumer", "packages/script/consumer", {
        tags: ["typescript-script"],
        dependsOn: ["common"],
        fileGroups: {
          dependencySources: ["/packages/script/common/src/**/*"],
        },
      }),
      configTexts: ['import {defineConfig} from "vitest/config";'],
    };

    expect(
      dependencySourceFailures([COMMON, ESLINT_CONFIG, ESLINT_BASE, consumer]),
    ).toEqual([]);
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
