import {mkdtempSync, readFileSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {readPkg, writePkg, type PackageJson} from "./package-json.js";

describe("package-json", () => {
  let tmp: string;

  beforeEach(() => {
    tmp = mkdtempSync(join(tmpdir(), "moon-scripts-common-"));
  });

  afterEach(() => {
    rmSync(tmp, {recursive: true, force: true});
  });

  describe("readPkg", () => {
    it("should read and parse a package.json file", () => {
      const pkgPath = join(tmp, "package.json");
      writeFileSync(
        pkgPath,
        JSON.stringify({name: "@xonovex/test", version: "1.0.0"}),
      );

      const result = readPkg(pkgPath);
      expect(result.name).toBe("@xonovex/test");
      expect(result.version).toBe("1.0.0");
    });

    it("should throw for missing file", () => {
      expect(() => readPkg(join(tmp, "missing.json"))).toThrow();
    });
  });

  describe("writePkg", () => {
    it("should write package.json with 2-space indent and trailing newline", () => {
      const pkgPath = join(tmp, "package.json");
      const pkg: PackageJson = {name: "@xonovex/test", version: "2.0.0"};

      writePkg(pkgPath, pkg);

      const raw = readFileSync(pkgPath, "utf8");
      expect(raw).toBe(JSON.stringify(pkg, null, 2) + "\n");
    });

    it("should produce a round-trippable result", () => {
      const pkgPath = join(tmp, "package.json");
      const pkg: PackageJson = {
        name: "@xonovex/round-trip",
        version: "0.1.0",
        dependencies: {"some-dep": "^1.0.0"},
      };

      writePkg(pkgPath, pkg);
      const result = readPkg(pkgPath);

      expect(result).toEqual(pkg);
    });
  });
});
