import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/file-system-memory.js";
import {
  readPkg,
  writePkg,
  type PackageJson,
} from "../../../src/package-json.js";

const PKG_PATH = "/workspace/package/package.json";

const fsWith = (contents: string) =>
  memoryFileSystem({files: {[PKG_PATH]: contents}});

describe("package-json", () => {
  describe("readPkg", () => {
    it("should read and parse a package.json file", () => {
      const result = readPkg(
        PKG_PATH,
        fsWith(JSON.stringify({name: "@xonovex/test", version: "1.0.0"})),
      );

      expect(result.name).toBe("@xonovex/test");
      expect(result.version).toBe("1.0.0");
    });

    it("should throw for missing file", () => {
      expect(() => readPkg(PKG_PATH, memoryFileSystem())).toThrow();
    });

    it("should reject invalid dependency values", () => {
      expect(() =>
        readPkg(
          PKG_PATH,
          fsWith(JSON.stringify({dependencies: {invalid: 42}})),
        ),
      ).toThrow("invalid package.json");
    });

    it("should reject malformed JSON with source context", () => {
      expect(() => readPkg(PKG_PATH, fsWith("{"))).toThrow(
        `invalid package.json at ${PKG_PATH}: malformed JSON`,
      );
    });

    it("should preserve fields outside the release schema", () => {
      expect(
        readPkg(
          PKG_PATH,
          fsWith(JSON.stringify({name: "example", scripts: {test: "vitest"}})),
        ),
      ).toEqual({name: "example", scripts: {test: "vitest"}});
    });
  });

  describe("writePkg", () => {
    it("should write package.json with 2-space indent and trailing newline", () => {
      const fs = memoryFileSystem();
      const pkg: PackageJson = {name: "@xonovex/test", version: "2.0.0"};

      writePkg(PKG_PATH, pkg, fs);

      expect(fs.readText(PKG_PATH)).toBe(JSON.stringify(pkg, null, 2) + "\n");
    });

    it("should produce a round-trippable result", () => {
      const fs = memoryFileSystem();
      const pkg: PackageJson = {
        name: "@xonovex/round-trip",
        version: "0.1.0",
        dependencies: {"some-dep": "^1.0.0"},
      };

      writePkg(PKG_PATH, pkg, fs);

      expect(readPkg(PKG_PATH, fs)).toEqual(pkg);
    });
  });
});
