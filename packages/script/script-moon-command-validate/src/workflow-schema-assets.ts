import {createHash} from "node:crypto";
import {readdirSync, readFileSync} from "node:fs";
import {join, relative} from "node:path";
import {
  Ajv2020,
  type AnySchemaObject,
  type ErrorObject,
  type ValidateFunction,
} from "ajv/dist/2020.js";
import {valid, validRange} from "semver";
import {issue, type ValidationIssue} from "./validation.js";

export type WorkflowSchemaKind = "request" | "result" | "work-record";

export interface WorkflowSchemaBundle {
  readonly request: AnySchemaObject;
  readonly result: AnySchemaObject;
  readonly "work-record": AnySchemaObject;
}

export interface WorkflowSchemaValidators {
  readonly request: ValidateFunction;
  readonly result: ValidateFunction;
  readonly "work-record": ValidateFunction;
}

export interface WorkflowValueValidation {
  readonly errors: readonly string[];
  readonly valid: boolean;
}

export interface WorkflowAssetReport {
  readonly fixtures: number;
  readonly issues: readonly ValidationIssue[];
  readonly schemas: number;
}

const SCHEMA_FILES = {
  request: "workflow-request.schema.json",
  result: "workflow-result.schema.json",
  "work-record": "workflow-work-record.schema.json",
} as const satisfies Readonly<Record<WorkflowSchemaKind, string>>;

const readJson = (path: string): unknown => {
  try {
    return JSON.parse(readFileSync(path, "utf8")) as unknown;
  } catch (error) {
    throw new Error(`cannot parse JSON at ${path}`, {cause: error});
  }
};

const readSchema = (path: string): AnySchemaObject => {
  const input = readJson(path);
  if (typeof input !== "object" || input === null || Array.isArray(input)) {
    throw new Error(`JSON Schema at ${path} must be an object`);
  }
  return input;
};

export const loadWorkflowSchemaBundle = (
  assetDirectory: string,
): WorkflowSchemaBundle => ({
  request: readSchema(join(assetDirectory, SCHEMA_FILES.request)),
  result: readSchema(join(assetDirectory, SCHEMA_FILES.result)),
  "work-record": readSchema(join(assetDirectory, SCHEMA_FILES["work-record"])),
});

const validDateTime = (value: string): boolean =>
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/u.test(
    value,
  ) && !Number.isNaN(Date.parse(value));

export const createWorkflowSchemaValidators = (
  schemas: WorkflowSchemaBundle,
): WorkflowSchemaValidators => {
  const ajv = new Ajv2020({
    allErrors: true,
    allowUnionTypes: true,
    strict: true,
    strictRequired: false,
    strictTypes: false,
  });
  ajv.addFormat("date-time", {type: "string", validate: validDateTime});
  ajv.addFormat("semver", {
    type: "string",
    validate: (value: string) => valid(value) !== null,
  });
  ajv.addFormat("semver-range", {
    type: "string",
    validate: (value: string) => validRange(value) !== null,
  });
  const schemaList: readonly AnySchemaObject[] = [
    schemas.request,
    schemas.result,
    schemas["work-record"],
  ];
  for (const schema of schemaList) {
    ajv.addSchema(schema);
  }

  const validator = (kind: WorkflowSchemaKind): ValidateFunction => {
    const identifier = schemas[kind].$id;
    if (typeof identifier !== "string") {
      throw new TypeError(`${SCHEMA_FILES[kind]} needs a string $id`);
    }
    const found = ajv.getSchema(identifier);
    if (found === undefined) {
      throw new TypeError(`${SCHEMA_FILES[kind]} did not compile`);
    }
    return found;
  };
  return {
    request: validator("request"),
    result: validator("result"),
    "work-record": validator("work-record"),
  };
};

const formatErrors = (
  errors: readonly ErrorObject[] | null | undefined,
): readonly string[] =>
  (errors ?? []).map(
    ({instancePath, message, params}) =>
      `${instancePath.length === 0 ? "/" : instancePath}: ${message ?? "invalid"} (${JSON.stringify(params)})`,
  );

export const validateWorkflowValue = (
  validators: WorkflowSchemaValidators,
  kind: WorkflowSchemaKind,
  input: unknown,
): WorkflowValueValidation => {
  const validator = validators[kind];
  const valid = validator(input);
  return {valid, errors: valid ? [] : formatErrors(validator.errors)};
};

const filesBelow = (directory: string): readonly string[] =>
  readdirSync(directory, {withFileTypes: true}).flatMap((entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) return filesBelow(path);
    return entry.isFile() && entry.name.endsWith(".json") ? [path] : [];
  });

const fixtureKind = (input: unknown): WorkflowSchemaKind | undefined => {
  if (typeof input !== "object" || input === null || Array.isArray(input)) {
    return undefined;
  }
  const version = (input as Readonly<Record<string, unknown>>).contractVersion;
  if (version === "xonovex.workflow.request/v1") return "request";
  if (version === "xonovex.workflow.result/v1") return "result";
  if (version === "xonovex.workflow.work-record/v1") return "work-record";
  return undefined;
};

const expectsInvalid = (path: string): boolean =>
  path.split(/[\\/]/u).includes("invalid");

interface CatalogIdentity {
  readonly contractVersion: string;
  readonly digest: string;
}

const catalogIdentity = (
  assetDirectory: string,
): CatalogIdentity | undefined => {
  const path = join(assetDirectory, "composition-catalog.json");
  try {
    const text = readFileSync(path, "utf8");
    const input = JSON.parse(text) as unknown;
    if (
      typeof input !== "object" ||
      input === null ||
      Array.isArray(input) ||
      !("contractVersion" in input) ||
      typeof input.contractVersion !== "string"
    ) {
      return undefined;
    }
    return {
      contractVersion: input.contractVersion,
      digest: createHash("sha256").update(text).digest("hex"),
    };
  } catch {
    return undefined;
  }
};

const catalogReferences = (
  input: unknown,
  path = "",
): readonly {
  readonly contractVersion?: string;
  readonly digest?: string;
  readonly path: string;
}[] => {
  if (Array.isArray(input)) {
    return input.flatMap((entry, index) =>
      catalogReferences(entry, `${path}/${String(index)}`),
    );
  }
  if (typeof input !== "object" || input === null) return [];
  const record = input as Readonly<Record<string, unknown>>;
  const current =
    typeof record.catalogContractVersion === "string" ||
    typeof record.catalogDigest === "string"
      ? [
          {
            path: path.length === 0 ? "/" : path,
            ...(typeof record.catalogContractVersion === "string"
              ? {contractVersion: record.catalogContractVersion}
              : {}),
            ...(typeof record.catalogDigest === "string"
              ? {digest: record.catalogDigest}
              : {}),
          },
        ]
      : [];
  return [
    ...current,
    ...Object.entries(record).flatMap(([key, value]) =>
      catalogReferences(value, `${path}/${key}`),
    ),
  ];
};

export const validateWorkflowSchemaAssets = (
  assetDirectory: string,
): WorkflowAssetReport => {
  const issues: ValidationIssue[] = [];
  let schemas: WorkflowSchemaBundle;
  let validators: WorkflowSchemaValidators;
  try {
    schemas = loadWorkflowSchemaBundle(assetDirectory);
    validators = createWorkflowSchemaValidators(schemas);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return {
      fixtures: 0,
      schemas: 0,
      issues: [issue("workflow-schema.invalid", assetDirectory, message)],
    };
  }

  const examplesDirectory = join(assetDirectory, "examples");
  const fixtureFiles = filesBelow(examplesDirectory);
  const expectedCatalog = catalogIdentity(assetDirectory);
  if (expectedCatalog === undefined) {
    issues.push(
      issue(
        "workflow-schema.catalog",
        assetDirectory,
        "composition-catalog.json needs a string contractVersion",
      ),
    );
  }
  for (const path of fixtureFiles) {
    const displayPath = relative(assetDirectory, path);
    let input: unknown;
    try {
      input = readJson(path);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      issues.push(issue("workflow-fixture.invalid-json", displayPath, message));
      continue;
    }
    const kind = fixtureKind(input);
    if (kind === undefined) {
      issues.push(
        issue(
          "workflow-fixture.unknown-contract",
          displayPath,
          "fixture needs a recognized contractVersion",
        ),
      );
      continue;
    }
    const validation = validateWorkflowValue(validators, kind, input);
    const invalid = expectsInvalid(path);
    if (validation.valid === invalid) {
      issues.push(
        issue(
          invalid
            ? "workflow-fixture.unexpected-valid"
            : "workflow-fixture.unexpected-invalid",
          displayPath,
          invalid
            ? "fixture under invalid/ unexpectedly passed"
            : validation.errors.join("; "),
        ),
      );
    }
    if (!invalid && expectedCatalog !== undefined) {
      for (const reference of catalogReferences(input)) {
        if (
          reference.contractVersion !== undefined &&
          reference.contractVersion !== expectedCatalog.contractVersion
        ) {
          issues.push(
            issue(
              "workflow-fixture.catalog-version",
              displayPath,
              `${reference.path} records catalog contract ${reference.contractVersion}; expected ${expectedCatalog.contractVersion}`,
            ),
          );
        }
        if (
          reference.digest !== undefined &&
          reference.digest !== expectedCatalog.digest
        ) {
          issues.push(
            issue(
              "workflow-fixture.catalog-digest",
              displayPath,
              `${reference.path} records stale catalog digest ${reference.digest}; expected ${expectedCatalog.digest}`,
            ),
          );
        }
      }
    }
  }

  return {
    schemas: Object.keys(schemas).length,
    fixtures: fixtureFiles.length,
    issues,
  };
};
