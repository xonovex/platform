import {mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {readPlatformMeta} from "./platform-meta.js";

describe("readPlatformMeta", () => {
  let tmp: string;

  beforeEach(() => {
    tmp = mkdtempSync(join(tmpdir(), "platform-meta-"));
  });

  afterEach(() => {
    rmSync(tmp, {recursive: true, force: true});
  });

  it("should return undefined when no platform.json exists", () => {
    expect(readPlatformMeta(tmp)).toBeUndefined();
  });

  it("should parse platform.json with os and cpu", () => {
    writeFileSync(
      join(tmp, "platform.json"),
      JSON.stringify({os: ["darwin"], cpu: ["arm64"]}),
    );

    const result = readPlatformMeta(tmp);
    expect(result).toEqual({os: ["darwin"], cpu: ["arm64"]});
  });

  it("should parse platform.json with os, cpu, and libc", () => {
    writeFileSync(
      join(tmp, "platform.json"),
      JSON.stringify({os: ["linux"], cpu: ["x64"], libc: ["glibc", "musl"]}),
    );

    const result = readPlatformMeta(tmp);
    expect(result).toEqual({
      os: ["linux"],
      cpu: ["x64"],
      libc: ["glibc", "musl"],
    });
  });
});
