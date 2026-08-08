import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/file-system-memory.js";
import {readPlatformMeta} from "../../../src/platform-meta.js";

const metaIn = (contents?: string) =>
  readPlatformMeta(
    "/pkg",
    memoryFileSystem(
      contents === undefined
        ? {directories: ["/pkg"]}
        : {files: {"/pkg/platform.json": contents}},
    ),
  );

describe("readPlatformMeta", () => {
  it("should return undefined when no platform.json exists", () => {
    expect(metaIn()).toBeUndefined();
  });

  it("should parse platform.json with os and cpu", () => {
    expect(metaIn(JSON.stringify({os: ["darwin"], cpu: ["arm64"]}))).toEqual({
      os: ["darwin"],
      cpu: ["arm64"],
    });
  });

  it("should parse platform.json with os, cpu, and libc", () => {
    expect(
      metaIn(
        JSON.stringify({os: ["linux"], cpu: ["x64"], libc: ["glibc", "musl"]}),
      ),
    ).toEqual({os: ["linux"], cpu: ["x64"], libc: ["glibc", "musl"]});
  });

  it("should reject platform metadata without a CPU list", () => {
    expect(() => metaIn(JSON.stringify({os: ["linux"]}))).toThrow(
      "invalid platform metadata",
    );
  });

  it("should reject malformed JSON", () => {
    expect(() => metaIn("{")).toThrow("malformed JSON");
  });
});
