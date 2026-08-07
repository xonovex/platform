import {describe, expect, it} from "vitest";
import {
  declaredLiteralInputs,
  hasherIgnoreFailures,
  parseWorkspaceHasher,
  workspaceHasherFailures,
} from "../../../src/hasher-ignore.js";

const base = {
  ignorePatterns: ["**/node_modules/**", ".moon/cache/**"],
  ignoredPaths: [],
  declaredInputs: [],
  exemptPrefixes: [],
};

describe("hasherIgnoreFailures", () => {
  it("accepts an ignored directory a pattern covers", () => {
    expect(
      hasherIgnoreFailures({
        ...base,
        ignoredPaths: ["node_modules/", "packages/a/node_modules/"],
      }),
    ).toEqual([]);
  });

  it("accepts an ignored directory covered by a rooted pattern", () => {
    expect(
      hasherIgnoreFailures({...base, ignoredPaths: [".moon/cache/"]}),
    ).toEqual([]);
  });

  it("reports an ignored directory no pattern covers", () => {
    const failures = hasherIgnoreFailures({
      ...base,
      ignoredPaths: ["packages/a/coverage/"],
    });
    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("packages/a/coverage/");
  });

  it("reports an ignored file no pattern covers", () => {
    const failures = hasherIgnoreFailures({
      ...base,
      ignoredPaths: ["packages/a/tsconfig.tsbuildinfo"],
    });
    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("tsbuildinfo");
  });

  it("skips an ignored path under an exempt prefix", () => {
    expect(
      hasherIgnoreFailures({
        ...base,
        ignoredPaths: ["packages/a/dist/"],
        exemptPrefixes: ["packages/a/dist/"],
      }),
    ).toEqual([]);
  });

  it("reports a declared input an ignore pattern would exclude", () => {
    const failures = hasherIgnoreFailures({
      ...base,
      ignorePatterns: ["**/dist/**"],
      declaredInputs: [
        {source: "packages/a/moon.yml", path: "packages/a/dist/src/index.js"},
      ],
    });
    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("packages/a/dist/src/index.js");
    expect(failures[0]).toContain("**/dist/**");
  });

  it("accepts a declared input no pattern matches", () => {
    expect(
      hasherIgnoreFailures({
        ...base,
        declaredInputs: [
          {source: "packages/a/moon.yml", path: "packages/a/src/index.ts"},
        ],
      }),
    ).toEqual([]);
  });

  it("does not treat * as spanning a separator", () => {
    const failures = hasherIgnoreFailures({
      ...base,
      ignorePatterns: ["packages/*/dist/**"],
      ignoredPaths: ["packages/group/nested/dist/"],
    });
    expect(failures).toHaveLength(1);
  });

  it("escapes regex metacharacters in a pattern", () => {
    expect(
      hasherIgnoreFailures({
        ...base,
        ignorePatterns: ["**/*.gen.*"],
        ignoredPaths: ["packages/a/schema.gen.json"],
      }),
    ).toEqual([]);
  });
});

describe("parseWorkspaceHasher", () => {
  it("reads the walk strategy and patterns", () => {
    expect(
      parseWorkspaceHasher(
        'hasher:\n  walkStrategy: glob\n  ignorePatterns:\n    - "**/dist/**"\n',
      ),
    ).toEqual({walkStrategy: "glob", ignorePatterns: ["**/dist/**"]});
  });

  it("defaults the patterns when the hasher declares none", () => {
    expect(parseWorkspaceHasher("hasher:\n  walkStrategy: vcs\n")).toEqual({
      walkStrategy: "vcs",
      ignorePatterns: [],
    });
  });

  it("reports no hasher section as undefined fields rather than failing", () => {
    expect(parseWorkspaceHasher("projects:\n  a: a\n")).toEqual({
      walkStrategy: undefined,
      ignorePatterns: [],
    });
  });

  it("returns undefined for unreadable yaml", () => {
    expect(parseWorkspaceHasher("hasher: [unclosed\n")).toBeUndefined();
  });

  it("returns undefined when the hasher section has the wrong shape", () => {
    expect(
      parseWorkspaceHasher("hasher:\n  ignorePatterns: 3\n"),
    ).toBeUndefined();
  });
});

describe("workspaceHasherFailures", () => {
  const globWorkspace =
    'hasher:\n  walkStrategy: glob\n  ignorePatterns:\n    - "**/node_modules/**"\n';

  it("passes when every ignored path is covered", () => {
    expect(
      workspaceHasherFailures({
        workspaceText: globWorkspace,
        ignoredPaths: ["node_modules/"],
        projectFiles: [],
      }),
    ).toEqual([]);
  });

  it("skips the check when no workspace config exists", () => {
    expect(
      workspaceHasherFailures({
        workspaceText: undefined,
        ignoredPaths: ["anything/"],
        projectFiles: [],
      }),
    ).toEqual([]);
  });

  it("skips the check under the vcs walk, which applies gitignore itself", () => {
    expect(
      workspaceHasherFailures({
        workspaceText: "hasher:\n  walkStrategy: vcs\n",
        ignoredPaths: ["uncovered/"],
        projectFiles: [],
      }),
    ).toEqual([]);
  });

  it("reports an unreadable hasher section", () => {
    const failures = workspaceHasherFailures({
      workspaceText: "hasher:\n  ignorePatterns: 3\n",
      ignoredPaths: [],
      projectFiles: [],
    });
    expect(failures).toEqual([
      ".moon/workspace.yml has an unreadable hasher section",
    ]);
  });

  it("reports git being unable to list ignored paths", () => {
    const failures = workspaceHasherFailures({
      workspaceText: globWorkspace,
      ignoredPaths: undefined,
      projectFiles: [],
    });
    expect(failures).toEqual([
      "git could not list ignored paths for the hasher check",
    ]);
  });

  it("exempts a dist directory, which must stay hashable", () => {
    expect(
      workspaceHasherFailures({
        workspaceText: globWorkspace,
        ignoredPaths: ["packages/a/dist/"],
        projectFiles: [],
      }),
    ).toEqual([]);
  });

  it("reports an ignored directory that drifted out of the patterns", () => {
    const failures = workspaceHasherFailures({
      workspaceText: globWorkspace,
      ignoredPaths: ["packages/a/coverage/"],
      projectFiles: [],
    });
    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("packages/a/coverage/");
  });

  it("reports a declared input the patterns would exclude", () => {
    const failures = workspaceHasherFailures({
      workspaceText:
        'hasher:\n  walkStrategy: glob\n  ignorePatterns:\n    - "**/dist/**"\n',
      ignoredPaths: [],
      projectFiles: [
        {
          path: "packages/a/moon.yml",
          text: "tasks:\n  t:\n    inputs:\n      - dist/src/index.js\n",
        },
      ],
    });
    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("packages/a/dist/src/index.js");
  });
});

const projectFile = (text: string) => [{path: "packages/a/moon.yml", text}];

describe("declaredLiteralInputs", () => {
  it("resolves a project relative input against the project directory", () => {
    expect(
      declaredLiteralInputs(
        projectFile("tasks:\n  t:\n    inputs:\n      - dist/src/index.js\n"),
      ),
    ).toEqual([
      {source: "packages/a/moon.yml", path: "packages/a/dist/src/index.js"},
    ]);
  });

  it("resolves a workspace rooted input against the workspace", () => {
    expect(
      declaredLiteralInputs(
        projectFile("tasks:\n  t:\n    inputs:\n      - /package-lock.json\n"),
      ),
    ).toEqual([{source: "packages/a/moon.yml", path: "package-lock.json"}]);
  });

  it("skips a glob input", () => {
    expect(
      declaredLiteralInputs(
        projectFile("tasks:\n  t:\n    inputs:\n      - src/**/*\n"),
      ),
    ).toEqual([]);
  });

  it("skips a token input", () => {
    expect(
      declaredLiteralInputs(
        projectFile("tasks:\n  t:\n    inputs:\n      - $workspaceRoot\n"),
      ),
    ).toEqual([]);
  });

  it("skips a task that declares no inputs", () => {
    expect(
      declaredLiteralInputs(projectFile("tasks:\n  t:\n    command: noop\n")),
    ).toEqual([]);
  });

  it("skips a file that is not valid yaml", () => {
    expect(declaredLiteralInputs(projectFile("tasks: [unclosed\n"))).toEqual(
      [],
    );
  });

  it("skips a file whose tasks have the wrong shape", () => {
    expect(
      declaredLiteralInputs(projectFile("tasks:\n  t:\n    inputs: 3\n")),
    ).toEqual([]);
  });
});
