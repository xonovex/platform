import type {PackageJson} from "@xonovex/script-moon-common";
import {describe, expect, it} from "vitest";
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
