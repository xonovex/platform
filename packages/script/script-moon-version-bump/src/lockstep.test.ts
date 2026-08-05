import type {PackageJson} from "@xonovex/script-moon-common/package-json";
import {describe, expect, it} from "vitest";
import {
  parseLockstepNames,
  planLockstep,
  type WorkspacePackage,
} from "./lockstep.js";

const workspacePackage = (
  directory: string,
  pkg: PackageJson,
  headVersion = pkg.version,
): WorkspacePackage => ({
  path: `/repo/packages/${directory}/package.json`,
  pkg,
  headVersion,
});

// eslint-config-base devDepends on prettier-config while prettier-config
// depends on eslint-config-base, so neither can be bumped first.
const cyclicLine = (): readonly WorkspacePackage[] => [
  workspacePackage("eslint-config-base", {
    name: "@xonovex/eslint-config-base",
    version: "0.1.22",
    devDependencies: {"@xonovex/prettier-config": "0.1.22"},
  }),
  workspacePackage("prettier-config", {
    name: "@xonovex/prettier-config",
    version: "0.1.22",
    dependencies: {"@xonovex/eslint-config-base": "0.1.22"},
  }),
];

describe("parseLockstepNames", () => {
  it("splits, trims and keeps the requested order", () => {
    expect(parseLockstepNames(" a , b,c ")).toEqual(["a", "b", "c"]);
  });

  it("rejects an empty set", () => {
    expect(() => parseLockstepNames(" , ")).toThrow(
      "needs at least one package name",
    );
  });

  it("rejects a repeated package", () => {
    expect(() => parseLockstepNames("a,b,a")).toThrow(
      "duplicate lockstep package a",
    );
  });
});

describe("planLockstep", () => {
  it("moves a dependency cycle to one version in a single plan", () => {
    const plan = planLockstep({
      packages: cyclicLine(),
      names: ["@xonovex/eslint-config-base", "@xonovex/prettier-config"],
      bumpType: "minor",
      preid: undefined,
      exact: undefined,
    });

    expect(plan.targetVersion).toBe("0.2.0");
    expect(plan.members.map((member) => member.pkg.version)).toEqual([
      "0.2.0",
      "0.2.0",
    ]);
    expect(
      plan.members[0]?.pkg.devDependencies?.["@xonovex/prettier-config"],
    ).toBe("0.2.0");
    expect(
      plan.members[1]?.pkg.dependencies?.["@xonovex/eslint-config-base"],
    ).toBe("0.2.0");
    expect(plan.dependents).toEqual([]);
  });

  it("resolves a member by its project directory", () => {
    const plan = planLockstep({
      packages: cyclicLine(),
      names: ["eslint-config-base", "prettier-config"],
      bumpType: "minor",
      preid: undefined,
      exact: undefined,
    });

    expect(plan.members.map((member) => member.name)).toEqual([
      "@xonovex/eslint-config-base",
      "@xonovex/prettier-config",
    ]);
  });

  it("lifts a member a peer already patch-bumped to the shared version", () => {
    const packages = [
      workspacePackage(
        "eslint-config-base",
        {
          name: "@xonovex/eslint-config-base",
          version: "0.1.23",
          devDependencies: {"@xonovex/prettier-config": "0.1.22"},
        },
        "0.1.22",
      ),
      ...cyclicLine().slice(1),
    ];

    const plan = planLockstep({
      packages,
      names: ["@xonovex/eslint-config-base", "@xonovex/prettier-config"],
      bumpType: "minor",
      preid: undefined,
      exact: undefined,
    });

    expect(plan.baseVersion).toBe("0.1.22");
    expect(plan.targetVersion).toBe("0.2.0");
    expect(plan.members[0]?.previousVersion).toBe("0.1.23");
    expect(plan.members[0]?.pkg.version).toBe("0.2.0");
    expect(plan.members[0]?.baseVersion).toBe("0.1.22");
  });

  it("moves optionalDependencies of an out-of-set consumer", () => {
    const packages = [
      workspacePackage("agent-cli-go", {
        name: "@xonovex/agent-cli-go",
        version: "0.1.31",
        optionalDependencies: {
          "@xonovex/agent-cli-go-linux-x64": "0.1.31",
          "@xonovex/agent-cli-go-win32-x64": "0.1.31",
        },
      }),
      workspacePackage("agent-cli-go-linux-x64", {
        name: "@xonovex/agent-cli-go-linux-x64",
        version: "0.1.31",
      }),
      workspacePackage("agent-cli-go-win32-x64", {
        name: "@xonovex/agent-cli-go-win32-x64",
        version: "0.1.31",
      }),
      workspacePackage("agent-cli-go-github", {
        name: "@xonovex/agent-cli-go-github",
        version: "0.0.0",
        private: true,
        optionalDependencies: {"@xonovex/agent-cli-go-linux-x64": "0.1.31"},
      }),
    ];

    const plan = planLockstep({
      packages,
      names: [
        "@xonovex/agent-cli-go",
        "@xonovex/agent-cli-go-linux-x64",
        "@xonovex/agent-cli-go-win32-x64",
      ],
      bumpType: "minor",
      preid: undefined,
      exact: undefined,
    });

    expect(plan.members[0]?.pkg.optionalDependencies).toEqual({
      "@xonovex/agent-cli-go-linux-x64": "0.2.0",
      "@xonovex/agent-cli-go-win32-x64": "0.2.0",
    });
    expect(plan.dependents).toHaveLength(1);
    expect(plan.dependents[0]?.name).toBe("@xonovex/agent-cli-go-github");
    expect(plan.dependents[0]?.newVersion).toBeUndefined();
    expect(
      plan.dependents[0]?.pkg.optionalDependencies?.[
        "@xonovex/agent-cli-go-linux-x64"
      ],
    ).toBe("0.2.0");
  });

  it("patch-bumps an unbumped public dependent once for the whole set", () => {
    const packages = [
      ...cyclicLine(),
      workspacePackage("agent-cli-go", {
        name: "@xonovex/agent-cli-go",
        version: "0.1.31",
        devDependencies: {
          "@xonovex/eslint-config-base": "0.1.22",
          "@xonovex/prettier-config": "0.1.22",
        },
      }),
    ];

    const plan = planLockstep({
      packages,
      names: ["@xonovex/eslint-config-base", "@xonovex/prettier-config"],
      bumpType: "minor",
      preid: undefined,
      exact: undefined,
    });

    expect(plan.dependents).toHaveLength(1);
    expect(plan.dependents[0]?.referenceCount).toBe(2);
    expect(plan.dependents[0]?.previousVersion).toBe("0.1.31");
    expect(plan.dependents[0]?.newVersion).toBe("0.1.32");
  });

  it("leaves an already bumped dependent version alone", () => {
    const packages = [
      ...cyclicLine(),
      workspacePackage(
        "agent-cli-go",
        {
          name: "@xonovex/agent-cli-go",
          version: "0.1.32",
          devDependencies: {"@xonovex/eslint-config-base": "0.1.22"},
        },
        "0.1.31",
      ),
    ];

    const plan = planLockstep({
      packages,
      names: ["@xonovex/eslint-config-base", "@xonovex/prettier-config"],
      bumpType: "minor",
      preid: undefined,
      exact: undefined,
    });

    expect(plan.dependents[0]?.newVersion).toBeUndefined();
    expect(
      plan.dependents[0]?.pkg.devDependencies?.["@xonovex/eslint-config-base"],
    ).toBe("0.2.0");
  });

  it("ignores a package whose references already hold the target", () => {
    const packages = [
      ...cyclicLine(),
      workspacePackage("agent-cli-go", {
        name: "@xonovex/agent-cli-go",
        version: "0.1.31",
        devDependencies: {"@xonovex/eslint-config-base": "0.2.0"},
      }),
    ];

    const plan = planLockstep({
      packages,
      names: ["@xonovex/eslint-config-base", "@xonovex/prettier-config"],
      bumpType: "minor",
      preid: undefined,
      exact: undefined,
    });

    expect(plan.dependents).toEqual([]);
  });

  it("accepts an exact target version and a prerelease identifier", () => {
    const packages = cyclicLine();
    const names = ["@xonovex/eslint-config-base", "@xonovex/prettier-config"];

    expect(
      planLockstep({
        packages,
        names,
        bumpType: "patch",
        preid: undefined,
        exact: "1.0.0",
      }).targetVersion,
    ).toBe("1.0.0");
    expect(
      planLockstep({
        packages,
        names,
        bumpType: "minor",
        preid: "beta",
        exact: undefined,
      }).targetVersion,
    ).toBe("0.2.0-beta.0");
  });

  it("rejects an empty set", () => {
    expect(() =>
      planLockstep({
        packages: cyclicLine(),
        names: [],
        bumpType: "minor",
        preid: undefined,
        exact: undefined,
      }),
    ).toThrow("needs at least one package");
  });

  it("rejects an unknown, ambiguous or incomplete member", () => {
    const packages = cyclicLine();
    expect(() =>
      planLockstep({
        packages,
        names: ["@xonovex/missing"],
        bumpType: "minor",
        preid: undefined,
        exact: undefined,
      }),
    ).toThrow("unknown lockstep package @xonovex/missing");

    expect(() =>
      planLockstep({
        packages: [...packages, ...packages],
        names: ["@xonovex/prettier-config"],
        bumpType: "minor",
        preid: undefined,
        exact: undefined,
      }),
    ).toThrow("ambiguous lockstep package @xonovex/prettier-config");

    expect(() =>
      planLockstep({
        packages: [workspacePackage("broken", {name: "@xonovex/broken"})],
        names: ["@xonovex/broken"],
        bumpType: "minor",
        preid: undefined,
        exact: undefined,
      }),
    ).toThrow("missing a name or version");
  });
});
