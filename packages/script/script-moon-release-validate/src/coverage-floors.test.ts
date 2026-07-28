import {describe, expect, it} from "vitest";
import {coverageFloorFailures} from "./coverage-floors.js";

const declaringAllFloors = `language: typescript
tags: [typescript-script]
tasks:
  ts-coverage:
    env:
      TS_COVERAGE_MIN_LINES: "85"
      TS_COVERAGE_MIN_FUNCTIONS: "90"
      TS_COVERAGE_MIN_BRANCHES: "70"
      TS_COVERAGE_MIN_STATEMENTS: "85"
`;

describe("typescript coverage floors", () => {
  it("accepts a project that declares every floor", () => {
    expect(
      coverageFloorFailures([
        {
          path: "packages/script/script-example/moon.yml",
          text: declaringAllFloors,
        },
      ]),
    ).toEqual([]);
  });

  it("accepts a project carrying the typescript tag alongside others", () => {
    const text = `tags: [go, typescript, npm, cli]
tasks:
  ts-coverage:
    env:
      TS_COVERAGE_MIN_LINES: "75"
      TS_COVERAGE_MIN_FUNCTIONS: "60"
      TS_COVERAGE_MIN_BRANCHES: "90"
      TS_COVERAGE_MIN_STATEMENTS: "75"
`;

    expect(
      coverageFloorFailures([{path: "packages/agent/example/moon.yml", text}]),
    ).toEqual([]);
  });

  it("ignores a project without a typescript tag", () => {
    const text = `language: go
tags: [go, shared]
tasks:
  go-test-coverage:
    env:
      GO_COVERAGE_MIN: "85"
`;

    expect(
      coverageFloorFailures([{path: "packages/shared/example/moon.yml", text}]),
    ).toEqual([]);
  });

  it("ignores a project that declares no tags", () => {
    expect(
      coverageFloorFailures([
        {
          path: "packages/config/example/moon.yml",
          text: "language: typescript\n",
        },
      ]),
    ).toEqual([]);
  });

  it("reports a typescript project that declares no ts-coverage task", () => {
    const text = `tags: [typescript-script]
tasks:
  ts-test:
    deps: [~:ts-build]
`;

    const failures = coverageFloorFailures([
      {path: "packages/script/script-example/moon.yml", text},
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("packages/script/script-example/moon.yml");
    expect(failures[0]).toContain("TS_COVERAGE_MIN_LINES");
    expect(failures[0]).toContain("TS_COVERAGE_MIN_STATEMENTS");
  });

  it("names only the floors a project leaves to the template", () => {
    const text = `tags: [typescript-script]
tasks:
  ts-coverage:
    env:
      TS_COVERAGE_MIN_LINES: "85"
      TS_COVERAGE_MIN_STATEMENTS: "85"
`;

    const failures = coverageFloorFailures([
      {path: "packages/script/script-example/moon.yml", text},
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("TS_COVERAGE_MIN_FUNCTIONS");
    expect(failures[0]).toContain("TS_COVERAGE_MIN_BRANCHES");
    expect(failures[0]).not.toContain("TS_COVERAGE_MIN_LINES");
  });

  it("reports a floor set to an empty value", () => {
    const text = `tags: [typescript]
tasks:
  ts-coverage:
    env:
      TS_COVERAGE_MIN_LINES: ""
      TS_COVERAGE_MIN_FUNCTIONS: "90"
      TS_COVERAGE_MIN_BRANCHES: "70"
      TS_COVERAGE_MIN_STATEMENTS: "85"
`;

    const failures = coverageFloorFailures([
      {path: "packages/shared/example/moon.yml", text},
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("TS_COVERAGE_MIN_LINES");
  });

  it("reports a project file that is not valid YAML", () => {
    const failures = coverageFloorFailures([
      {path: "packages/script/script-example/moon.yml", text: "tags: [oops\n"},
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("is not valid YAML");
  });

  it("reports a project file whose shape cannot be read", () => {
    const failures = coverageFloorFailures([
      {path: "packages/script/script-example/moon.yml", text: "tags: 7\n"},
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("unreadable project definition");
  });

  it("reports every offending project", () => {
    const bare = `tags: [typescript-script]
tasks: {}
`;

    const failures = coverageFloorFailures([
      {path: "packages/script/a/moon.yml", text: bare},
      {path: "packages/script/b/moon.yml", text: declaringAllFloors},
      {path: "packages/script/c/moon.yml", text: bare},
    ]);

    expect(failures).toHaveLength(2);
    expect(failures[0]).toContain("packages/script/a/moon.yml");
    expect(failures[1]).toContain("packages/script/c/moon.yml");
  });
});
