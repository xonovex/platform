import {execFileSync} from "node:child_process";
import {mkdtempSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {describe, expect, it} from "vitest";
import {resolveGitRef} from "../../../src/detect.js";

describe("resolveGitRef", () => {
  // resolveGitRef shells out to git, so the case needs a real repository to
  // resolve against; an empty one knows no refs at all.
  it("fails for an unknown ref", () => {
    const root = mkdtempSync(join(tmpdir(), "version-ref-"));
    execFileSync(resolveExecutable("git"), ["init", "--quiet"], {cwd: root});
    expect(() => resolveGitRef(root, "definitely-not-a-ref")).toThrow();
  });
});
