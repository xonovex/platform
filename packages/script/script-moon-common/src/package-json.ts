import {readFileSync, writeFileSync} from "node:fs";

export interface PackageJson {
  name?: string;
  version?: string;
  license?: string;
  private?: boolean;
  repository?: {type?: string; url?: string; directory?: string};
  files?: readonly string[];
  publishConfig?: {access?: string; registry?: string};
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  peerDependencies?: Record<string, string>;
  optionalDependencies?: Record<string, string>;
  os?: readonly string[];
  cpu?: readonly string[];
  libc?: readonly string[];
}

export const readPkg = (path: string): PackageJson =>
  JSON.parse(readFileSync(path, "utf8")) as PackageJson;

export const writePkg = (path: string, pkg: PackageJson): void => {
  writeFileSync(path, JSON.stringify(pkg, null, 2) + "\n");
};
