import {describe, expect, it} from "vitest";
import {taskInheritanceFailures} from "./task-inheritance.js";

const parent = {
  name: "tag-parent.yml",
  text: `tasks:
  ci-check:
    deps: [build, lint, typecheck, test, format-check]
  ts-build:
    deps: [prepare]
`,
};

describe("tag task inheritance", () => {
  it("accepts a child that restates every inherited dep and adds its own", () => {
    const child = {
      name: "tag-child.yml",
      text: `extends: ./tag-parent.yml
tasks:
  ci-check:
    deps: [build, lint, typecheck, test, format-check, coverage]
`,
    };

    expect(taskInheritanceFailures([parent, child])).toEqual([]);
  });

  it("reports a child that drops inherited deps while adding one", () => {
    const child = {
      name: "tag-child.yml",
      text: `extends: ./tag-parent.yml
tasks:
  ci-check:
    deps: [coverage]
    options:
      mergeDeps: append
`,
    };

    const failures = taskInheritanceFailures([parent, child]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain(
      "tag-child.yml task 'ci-check' drops inherited build, lint, typecheck, test, format-check",
    );
  });

  it("accepts a child that declares the drop with mergeDeps: replace", () => {
    const child = {
      name: "tag-child.yml",
      text: `extends: ./tag-parent.yml
tasks:
  ts-build:
    deps: []
    options:
      mergeDeps: replace
`,
    };

    expect(taskInheritanceFailures([parent, child])).toEqual([]);
  });

  it("ignores a child task the parent never defines", () => {
    const child = {
      name: "tag-child.yml",
      text: `extends: ./tag-parent.yml
tasks:
  coverage:
    deps: [ts-coverage]
`,
    };

    expect(taskInheritanceFailures([parent, child])).toEqual([]);
  });

  it("carries an inherited dep through a two-level extends chain", () => {
    const middle = {
      name: "tag-middle.yml",
      text: `extends: ./tag-parent.yml
tasks:
  ts-clean:
    deps: [wipe]
`,
    };
    const leaf = {
      name: "tag-leaf.yml",
      text: `extends: ./tag-middle.yml
tasks:
  ci-check:
    deps: [coverage]
`,
    };

    const failures = taskInheritanceFailures([parent, middle, leaf]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("tag-leaf.yml task 'ci-check'");
  });

  it("reads deps given as target objects", () => {
    const child = {
      name: "tag-child.yml",
      text: `extends: ./tag-parent.yml
tasks:
  ci-check:
    deps:
      - target: build
      - target: lint
`,
    };

    const failures = taskInheritanceFailures([parent, child]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain(
      "drops inherited typecheck, test, format-check",
    );
  });

  it("reports an extends target that is absent", () => {
    const orphan = {
      name: "tag-orphan.yml",
      text: `extends: ./tag-absent.yml
tasks:
  ci-check:
    deps: [build]
`,
    };

    expect(taskInheritanceFailures([orphan])).toEqual([
      "tag-orphan.yml extends missing ./tag-absent.yml",
    ]);
  });

  it("reports a file that is not valid YAML", () => {
    const broken = {name: "tag-broken.yml", text: "tasks:\n  - [unclosed\n"};

    expect(taskInheritanceFailures([broken])[0]).toContain(
      "tag-broken.yml is not valid YAML",
    );
  });
});
