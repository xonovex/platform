import {mkdtempSync, readFileSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {composeWorkflowRequest} from "@xonovex/core/workflow-composition-runtime";
import {afterEach, describe, expect, it} from "vitest";
import {
  readWorkflowContract,
  validateWorkflowContract,
  type WorkflowContract,
} from "./workflow-contract.js";

const repositoryRoot = resolve(import.meta.dirname, "../../../..");
const packageDir = resolve(repositoryRoot, "packages/command/command-workflow");
const contractPath = resolve(packageDir, "contracts/workflow-commands.v1.json");
const compositionCatalogPath = resolve(
  repositoryRoot,
  "packages/skill/skill-workflow/workflow-guide/assets/composition-catalog.json",
);
const temporaryDirectories: string[] = [];

const temporaryContract = (contract: WorkflowContract): string => {
  const directory = mkdtempSync(join(tmpdir(), "workflow-contract-"));
  temporaryDirectories.push(directory);
  const path = join(directory, "workflow-commands.v1.json");
  writeFileSync(path, JSON.stringify(contract));
  return path;
};

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("workflow command contract", () => {
  it("matches the repository command and skill surfaces", () => {
    expect(
      validateWorkflowContract(contractPath, packageDir, repositoryRoot),
    ).toEqual({commands: 12, issues: []});
  });

  it("normalizes and resolves every registered command operation", () => {
    const contract = readWorkflowContract(contractPath);
    const catalogSourceText = readFileSync(compositionCatalogPath, "utf8");
    const catalogInput: unknown = JSON.parse(catalogSourceText);

    for (const command of contract.commands) {
      const result = composeWorkflowRequest({
        catalogInput,
        catalogSourceText,
        installedRoots: [resolve(repositoryRoot, "packages/skill")],
        workflowRequest: {
          operation: command.name,
          selection: {},
        },
      });

      expect(result, command.name).toMatchObject({
        success: true,
        data: {
          status: "ready",
          loadOrder: ["workflow-guide"],
        },
      });
    }
  });

  it("rejects architecture drift independently of Markdown drift", () => {
    const contract = readWorkflowContract(contractPath);
    const changed: WorkflowContract = {
      ...contract,
      toolProfiles: {
        ...contract.toolProfiles,
        inline: [...(contract.toolProfiles.inline ?? []), "Write"],
      },
      forbiddenArguments: contract.forbiddenArguments.filter(
        (name) => name !== "provider",
      ),
      commands: contract.commands.map((command) =>
        command.name === "review"
          ? {
              ...command,
              effectModes: ["preview", "apply"],
              arguments: [
                ...command.arguments,
                {
                  name: "result",
                  kind: "flag" as const,
                  required: false,
                  repeatable: false,
                  valueName: "reference",
                },
              ],
            }
          : command,
      ),
    };

    const codes = validateWorkflowContract(
      temporaryContract(changed),
      packageDir,
      repositoryRoot,
    ).issues.map(({code}) => code);

    expect(codes).toEqual(
      expect.arrayContaining([
        "workflow.effect-modes",
        "workflow.forbidden-arguments",
        "workflow.publish-separation",
        "workflow.tool-profile-inline-effect",
      ]),
    );
  });

  it("rejects malformed and structurally invalid contracts", () => {
    const malformedDirectory = mkdtempSync(
      join(tmpdir(), "workflow-contract-invalid-"),
    );
    temporaryDirectories.push(malformedDirectory);
    const malformed = join(malformedDirectory, "malformed.json");
    const invalid = join(malformedDirectory, "invalid.json");
    writeFileSync(malformed, "{");
    writeFileSync(
      invalid,
      JSON.stringify({
        ...JSON.parse(readFileSync(contractPath, "utf8")),
        contractVersion: 2,
      }),
    );

    expect(() => readWorkflowContract(malformed)).toThrow(
      "cannot parse workflow contract JSON",
    );
    expect(() => readWorkflowContract(invalid)).toThrow(
      "invalid workflow contract",
    );
  });
});
