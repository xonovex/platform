import type {PackageJson} from "@xonovex/script-moon-common/package-json";
import {describe, expect, it} from "vitest";
import {
  parsePackagedFilePaths,
  validatePackedPackage,
} from "./packed-package.js";
import {validateDeclaredFiles, validatePackage} from "./validate.js";

const validPkg: PackageJson = {
  name: "@xonovex/test",
  version: "1.0.0",
  license: "MIT",
  repository: {type: "git", url: "https://example.com"},
  files: ["dist"],
  publishConfig: {access: "public"},
};

describe("validatePackage", () => {
  it("should return no errors for a valid package", () => {
    expect(validatePackage(validPkg)).toEqual([]);
  });

  it("should report missing name", () => {
    const {name: _, ...pkg} = validPkg;
    const errors = validatePackage(pkg);
    expect(errors).toContain("Missing required field: name");
  });

  it("should report missing version", () => {
    const {version: _, ...pkg} = validPkg;
    const errors = validatePackage(pkg);
    expect(errors).toContain("Missing required field: version");
  });

  it("should report missing license", () => {
    const {license: _, ...pkg} = validPkg;
    const errors = validatePackage(pkg);
    expect(errors).toContain("Missing required field: license");
  });

  it("should report missing repository", () => {
    const {repository: _, ...pkg} = validPkg;
    const errors = validatePackage(pkg);
    expect(errors).toContain("Missing required field: repository");
  });

  it("should report missing files", () => {
    const {files: _, ...pkg} = validPkg;
    const errors = validatePackage(pkg);
    expect(errors).toContain("Missing required field: files");
  });

  it("should report missing repository.type", () => {
    const pkg = {...validPkg, repository: {url: "https://example.com"}};
    const errors = validatePackage(pkg);
    expect(errors).toContain("repository.type is missing");
  });

  it("should report missing repository.url", () => {
    const pkg = {...validPkg, repository: {type: "git"}};
    const errors = validatePackage(pkg);
    expect(errors).toContain("repository.url is missing");
  });

  it("should report missing publishConfig.access", () => {
    const {publishConfig: _, ...pkg} = validPkg;
    const errors = validatePackage(pkg);
    expect(errors).toContain("publishConfig.access is not set");
  });

  it("reports empty required strings and file lists", () => {
    const errors = validatePackage({
      ...validPkg,
      name: "",
      license: "",
      files: [],
      repository: {type: "", url: ""},
    });

    expect(errors).toContain("Required field is empty: name");
    expect(errors).toContain("Required field is empty: license");
    expect(errors).toContain("Required field is empty: files");
    expect(errors).toContain("repository.type is missing");
    expect(errors).toContain("repository.url is missing");
  });

  it("reports declared package files that do not exist", () => {
    expect(
      validateDeclaredFiles(validPkg, (path) => path === "dist/index.js"),
    ).toEqual(['Declared package file does not exist: "dist"']);
  });
});

describe("validatePackedPackage", () => {
  it("accepts packed manifest targets and relative imports", () => {
    const pkg = {
      ...validPkg,
      exports: {".": "./dist/index.js"},
      bin: {test: "./dist/cli.js"},
    };
    const contents = new Map([
      ["dist/index.js", 'export {run} from "./run.js";'],
      ["dist/run.js", "export const run = () => true;"],
      ["dist/cli.js", 'import "./run.js";'],
      ["package.json", "{}"],
    ]);

    const errors = validatePackedPackage(
      pkg,
      [...contents.keys()],
      (path) => contents.get(path) ?? "",
    );

    expect(errors).toEqual([]);
  });

  it("reports manifest targets excluded from the package", () => {
    const pkg = {
      ...validPkg,
      exports: {".": "./src/index.ts"},
      bin: {test: "./dist/cli.js"},
    };

    const errors = validatePackedPackage(pkg, ["package.json"], () => "");

    expect(errors).toEqual([
      'Packed package is missing export target: "./src/index.ts"',
      'Packed package is missing bin target: "./dist/cli.js"',
    ]);
  });

  it("reports relative runtime imports excluded from the package", () => {
    const pkg = {...validPkg, bin: {test: "./dist/cli.js"}};

    const errors = validatePackedPackage(
      pkg,
      ["dist/cli.js", "package.json"],
      (path) =>
        path === "dist/cli.js" ? 'import {launch} from "./launcher.js";' : "",
    );

    expect(errors).toEqual([
      'Packed file "dist/cli.js" imports missing relative target: "./launcher.js"',
    ]);
  });

  it("parses file paths from npm pack output", () => {
    const output = JSON.stringify([
      {files: [{path: "package.json"}, {path: "dist/index.js"}]},
    ]);

    const files = parsePackagedFilePaths(output);

    expect(files).toEqual(["package.json", "dist/index.js"]);
  });
});
