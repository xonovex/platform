import {describe, expect, it, vi} from "vitest";
import {
  isPublished,
  parsePackageIdentity,
  publishArgs,
  publishedFromResult,
  publishPackage,
  type PlatformMeta,
  type PublishDependencies,
} from "../../../src/publish.js";

const originalPackage = `${JSON.stringify(
  {name: "@xonovex/example", version: "1.2.3"},
  null,
  2,
)}\n`;

const dependencies = (
  overrides: Partial<PublishDependencies> = {},
): PublishDependencies => ({
  readPackageJson: () => originalPackage,
  writePackageJson: vi.fn(),
  currentDirectory: () => "/workspace/package",
  readPlatformMeta: () => {
    return;
  },
  isPublished: () => false,
  publish: vi.fn(),
  log: vi.fn(),
  ...overrides,
});

describe("parsePackageIdentity", () => {
  it("accepts a scoped package and semantic version", () => {
    expect(
      parsePackageIdentity({name: "@xonovex/example", version: "1.2.3-beta.1"}),
    ).toEqual({name: "@xonovex/example", version: "1.2.3-beta.1"});
  });

  it.each([
    {name: "example; id", version: "1.2.3"},
    {name: "example", version: "latest"},
    {name: "example"},
    [],
  ])("rejects unsafe or incomplete metadata: %o", (value) => {
    expect(() => parsePackageIdentity(value)).toThrow();
  });
});

describe("publishArgs", () => {
  it("uses provenance for a real publication", () => {
    expect(publishArgs(false)).toEqual([
      "publish",
      "--provenance",
      "--access",
      "public",
    ]);
  });

  it("uses npm dry-run without provenance", () => {
    expect(publishArgs(true)).toEqual([
      "publish",
      "--dry-run",
      "--access",
      "public",
    ]);
  });
});

describe("isPublished", () => {
  const identity = {name: "@xonovex/example", version: "1.2.3"};

  it("recognizes an existing package version", () => {
    expect(isPublished(identity, () => ({status: 0, stderr: ""}))).toBe(true);
  });

  it("recognizes npm not-found output", () => {
    expect(
      publishedFromResult(identity, {status: 1, stderr: "npm error E404"}),
    ).toBe(false);
  });

  it("throws command launch failures", () => {
    const error = new Error("spawn failed");
    expect(() =>
      publishedFromResult(identity, {error, status: null, stderr: ""}),
    ).toThrow(error);
  });

  it("throws unexpected npm failures", () => {
    expect(() =>
      publishedFromResult(identity, {status: 1, stderr: "access denied"}),
    ).toThrow("npm view failed for @xonovex/example@1.2.3: access denied");
  });

  it("reports an empty failure by exit status", () => {
    expect(() =>
      publishedFromResult(identity, {status: null, stderr: ""}),
    ).toThrow("exit null");
  });
});

describe("publishPackage", () => {
  it("skips a version that is already published", () => {
    const publish = vi.fn();
    const writePackageJson = vi.fn();
    const log = vi.fn();
    publishPackage(
      false,
      dependencies({isPublished: () => true, publish, writePackageJson, log}),
    );

    expect(publish).not.toHaveBeenCalled();
    expect(writePackageJson).not.toHaveBeenCalled();
    expect(log).toHaveBeenCalledWith(
      "Skipping @xonovex/example@1.2.3 — already published",
    );
  });

  it("publishes without rewriting a platform-independent package", () => {
    const publish = vi.fn();
    const writePackageJson = vi.fn();
    publishPackage(true, dependencies({publish, writePackageJson}));

    expect(publish).toHaveBeenCalledWith(true);
    expect(writePackageJson).not.toHaveBeenCalled();
  });

  it("injects platform metadata and restores the original package", () => {
    const platformMeta: PlatformMeta = {
      os: ["linux"],
      cpu: ["x64"],
      libc: ["glibc"],
    };
    const writes: string[] = [];
    publishPackage(
      false,
      dependencies({
        readPlatformMeta: () => platformMeta,
        writePackageJson: (contents) => {
          writes.push(contents);
        },
      }),
    );

    expect(JSON.parse(writes[0] ?? "{}")).toMatchObject(platformMeta);
    expect(writes).toEqual([expect.any(String), originalPackage]);
  });

  it("restores the original package when npm publish fails", () => {
    const writes: string[] = [];
    const error = new Error("publish failed");
    expect(() => {
      publishPackage(
        false,
        dependencies({
          readPlatformMeta: () => ({os: ["linux"], cpu: ["arm64"]}),
          writePackageJson: (contents) => {
            writes.push(contents);
          },
          publish: () => {
            throw error;
          },
        }),
      );
    }).toThrow(error);
    expect(writes.at(-1)).toBe(originalPackage);
  });
});
