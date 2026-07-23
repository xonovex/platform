import {
  cpSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  createWorkflowSchemaValidators,
  loadWorkflowSchemaBundle,
  validateWorkflowSchemaAssets,
  validateWorkflowValue,
  type WorkflowSchemaKind,
} from "./workflow-schema-assets.js";

const assetDirectory = resolve(
  import.meta.dirname,
  "../../../skill/skill-workflow/workflow-guide/assets",
);
const temporaryDirectories: string[] = [];

const readExample = (name: string): Record<string, unknown> =>
  JSON.parse(
    readFileSync(join(assetDirectory, "examples", name), "utf8"),
  ) as Record<string, unknown>;

const validation = (kind: WorkflowSchemaKind, input: unknown): boolean =>
  validateWorkflowValue(
    createWorkflowSchemaValidators(loadWorkflowSchemaBundle(assetDirectory)),
    kind,
    input,
  ).valid;

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("workflow JSON Schema assets", () => {
  it("compiles all schemas and validates committed fixtures", () => {
    const report = validateWorkflowSchemaAssets(assetDirectory);

    expect(report.schemas).toBe(3);
    expect(report.fixtures).toBeGreaterThanOrEqual(4);
    expect(report.issues).toEqual([]);
  });

  it("enforces publish separation, operation effect modes, exact bindings, and criteria", () => {
    const request = readExample("multi-provider-review-request.json");
    expect(validation("request", request)).toBe(true);
    expect(
      (request.selection as {readonly perspectives: readonly string[]})
        .perspectives,
    ).toHaveLength(3);

    const withDestination = {...request, destination: request.subject};
    expect(validation("request", withDestination)).toBe(false);

    const wrongEffect = {...request, effect: {mode: "apply"}};
    expect(validation("request", wrongEffect)).toBe(false);

    const subject = structuredClone(request.subject) as {
      locator: Record<string, unknown>;
    };
    delete subject.locator.provider;
    expect(validation("request", {...request, subject})).toBe(false);

    const unpinned = structuredClone(request.subject) as {
      locator: Record<string, unknown>;
    };
    delete unpinned.locator.revision;
    expect(validation("request", {...request, subject: unpinned})).toBe(false);

    const selection = structuredClone(request.selection) as {
      criteria: Record<string, unknown>[];
    };
    delete selection.criteria[0]?.binding;
    expect(validation("request", {...request, selection})).toBe(false);

    const publish = readExample("publish-request.json");
    expect(validation("request", publish)).toBe(true);
    const publishWithoutDestination = structuredClone(publish);
    delete publishWithoutDestination.destination;
    expect(validation("request", publishWithoutDestination)).toBe(false);
    expect(validation("request", {...publish, effect: {mode: "inspect"}})).toBe(
      false,
    );
  });

  it("requires result provenance, status, effects, retry boundaries, and concurrency revisions", () => {
    const result = readExample("review-operation-result.json");
    expect(validation("result", result)).toBe(true);

    const withoutRetry = structuredClone(result);
    delete withoutRetry.retry;
    expect(validation("result", withoutRetry)).toBe(false);

    const withoutEffects = structuredClone(result);
    delete withoutEffects.effects;
    expect(validation("result", withoutEffects)).toBe(false);

    const withoutStatus = structuredClone(result);
    delete withoutStatus.status;
    expect(validation("result", withoutStatus)).toBe(false);

    const resolution = structuredClone(result.resolution) as {
      criteria: Record<string, unknown>[];
    };
    delete resolution.criteria[0]?.provenance;
    expect(validation("result", {...result, resolution})).toBe(false);

    const criterionWithoutEvaluation = structuredClone(result.resolution) as {
      criteria: Record<string, unknown>[];
    };
    delete criterionWithoutEvaluation.criteria[0]?.evaluation;
    expect(
      validation("result", {
        ...result,
        resolution: criterionWithoutEvaluation,
      }),
    ).toBe(false);

    const skillWithoutStatus = structuredClone(result.resolution) as {
      skills: Record<string, unknown>[];
    };
    delete skillWithoutStatus.skills[0]?.status;
    expect(
      validation("result", {...result, resolution: skillWithoutStatus}),
    ).toBe(false);

    const concurrency = structuredClone(result.concurrency) as Record<
      string,
      unknown
    >[];
    delete concurrency[0]?.observedRevision;
    expect(validation("result", {...result, concurrency})).toBe(false);
  });

  it("requires durable provider-owned identities and exact revisions", () => {
    const record = readExample("durable-work-record.json");
    expect(validation("work-record", record)).toBe(true);

    const root = structuredClone(record.root) as {
      locator: Record<string, unknown>;
    };
    delete root.locator.revision;
    expect(validation("work-record", {...record, root})).toBe(false);

    expect(
      validation("work-record", {...record, updatedAt: "not-a-date"}),
    ).toBe(false);
  });

  it("treats fixtures under invalid as expected failures", () => {
    const directory = mkdtempSync(join(tmpdir(), "workflow-assets-"));
    temporaryDirectories.push(directory);
    cpSync(assetDirectory, directory, {recursive: true});
    const invalidDirectory = join(directory, "examples", "invalid");
    mkdirSync(invalidDirectory);
    const request = readExample("multi-provider-review-request.json");
    writeFileSync(
      join(invalidDirectory, "review-apply.json"),
      JSON.stringify({...request, effect: {mode: "apply"}}),
    );

    expect(validateWorkflowSchemaAssets(directory).issues).toEqual([]);

    writeFileSync(
      join(invalidDirectory, "unexpected-valid.json"),
      JSON.stringify(request),
    );
    expect(validateWorkflowSchemaAssets(directory).issues).toContainEqual(
      expect.objectContaining({code: "workflow-fixture.unexpected-valid"}),
    );
  });
});
