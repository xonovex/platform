import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  parseCompositionCatalog,
  type CompositionCatalog,
} from "./skill-composition-contract.js";
import {discoverInstalledSkillInventory} from "./skill-inventory.js";
import {composeWorkflowRequest} from "./workflow-composition-runtime.js";
import {
  normalizeWorkflowCompositionRequest,
  type WorkflowCompositionInput,
} from "./workflow-composition.js";

const catalogInput = {
  contractVersion: "2.0.0",
  skills: [
    {
      name: "security-assurance-guide",
      classification: {
        lifecycle: "procedural",
        functionalRole: "assurance",
      },
      provisions: [
        {id: "perspective:security", version: "1.0.0"},
        {id: "assurance:security", version: "1.0.0"},
      ],
    },
    {
      name: "workflow-guide",
      classification: {
        lifecycle: "procedural",
        functionalRole: "procedure",
      },
      provisions: [{id: "procedure:workflow-operation", version: "1.0.0"}],
    },
  ],
};
const catalogSourceText = JSON.stringify(catalogInput);
const catalog = (): CompositionCatalog => {
  const parsed = parseCompositionCatalog(catalogInput, catalogSourceText);
  if (!parsed.success) throw new Error(parsed.errors.join("\n"));
  return parsed.data;
};
const workflowRequest = (): WorkflowCompositionInput => ({
  operation: "review",
  selection: {
    perspectives: ["security"],
    criteria: [],
    roleLenses: [],
    resolutionMode: "assisted",
    acceptedSuggestions: [],
    skillRequirements: [
      {
        id: "assurance:security",
        range: "^1.0.0",
        strength: "preferred",
        reason: "Security assurance strengthens the review.",
      },
    ],
    skillProviders: {},
    overlayContext: {
      global: "*",
      languages: [],
      frameworks: [],
      paths: [],
      explicit: [],
    },
    preferenceOverlays: [],
  },
  implementationOverrides: {
    skills: [],
    capabilities: [],
  },
});

const temporaryDirectories: string[] = [];

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {force: true, recursive: true});
  }
  temporaryDirectories.length = 0;
});

describe("workflow composition normalization", () => {
  it("derives semantic providers and preserves exact implementation overrides", () => {
    const input = workflowRequest();
    input.implementationOverrides.skills.push({
      id: "security-assurance-guide",
      version: "7.0.0",
    });

    const result = normalizeWorkflowCompositionRequest(catalog(), input);

    expect(result.success).toBe(true);
    if (!result.success) throw new Error(result.errors.join("\n"));
    expect(result.data.exactSkills).toMatchObject([
      {guide: "workflow-guide", required: true},
      {
        guide: "security-assurance-guide",
        implementationVersion: "7.0.0",
      },
    ]);
    expect(result.data.requirements.map(({id}) => id)).toEqual([
      "assurance:security",
      "perspective:security",
    ]);
  });

  it("rejects a provider binding without a matching requirement", () => {
    const input = workflowRequest();
    input.selection.skillProviders = {
      "assurance:missing": "security-assurance-guide",
    };

    const result = normalizeWorkflowCompositionRequest(catalog(), input);

    expect(result).toMatchObject({
      success: false,
      errors: [
        expect.stringContaining(
          "provider binding has no matching semantic requirement",
        ),
      ],
    });
  });

  it("rejects operations outside the workflow command registry", () => {
    const result = normalizeWorkflowCompositionRequest(catalog(), {
      ...workflowRequest(),
      operation: "archive",
    });

    expect(result).toMatchObject({
      success: false,
      errors: [expect.stringContaining("operation")],
    });
  });

  it("resolves a workflow request through the canonical runtime adapter", () => {
    const result = composeWorkflowRequest({
      catalogInput,
      catalogSourceText,
      installedSkills: [
        {
          guide: "security-assurance-guide",
          implementationVersion: "7.0.0",
          dependencies: [],
          plugin: "xonovex-skill-security-assurance",
        },
        {
          guide: "workflow-guide",
          implementationVersion: "7.0.0",
          dependencies: [],
          plugin: "xonovex-skill-workflow",
        },
      ],
      workflowRequest: workflowRequest(),
    });

    expect(result.success).toBe(true);
    if (!result.success) throw new Error(result.errors.join("\n"));
    expect(result.data.status).toBe("ready");
    expect(result.data.loadOrder).toEqual([
      "workflow-guide",
      "security-assurance-guide",
    ]);
  });
});

describe("installed skill inventory discovery", () => {
  it("discovers exact plugin dependencies and package versions", () => {
    const root = mkdtempSync(join(tmpdir(), "skill-inventory-"));
    temporaryDirectories.push(root);
    const writePlugin = (
      packageName: string,
      guide: string,
      dependencies: readonly string[],
      packageDependencies: Readonly<Record<string, string>>,
    ): void => {
      const packageRoot = join(root, packageName);
      mkdirSync(join(packageRoot, ".claude-plugin"), {recursive: true});
      mkdirSync(join(packageRoot, guide), {recursive: true});
      writeFileSync(
        join(packageRoot, ".claude-plugin", "plugin.json"),
        JSON.stringify({
          name: `xonovex-${packageName}`,
          version: "7.0.0",
          skills: [`./${guide}`],
          dependencies,
        }),
      );
      writeFileSync(
        join(packageRoot, "package.json"),
        JSON.stringify({
          name: `@xonovex/${packageName}`,
          version: "7.0.0",
          dependencies: packageDependencies,
        }),
      );
    };
    writePlugin("skill-base", "base-guide", [], {});
    writePlugin("skill-consumer", "consumer-guide", ["xonovex-skill-base"], {
      "@xonovex/skill-base": "7.0.0",
    });

    const result = discoverInstalledSkillInventory([root]);

    expect(result.success).toBe(true);
    if (!result.success) throw new Error(result.errors.join("\n"));
    expect(
      result.data.find(({guide}) => guide === "consumer-guide")?.dependencies,
    ).toEqual([
      {
        plugin: "xonovex-skill-base",
        implementationVersion: "7.0.0",
      },
    ]);
  });

  it("reports malformed plugin manifests instead of hiding them", () => {
    const root = mkdtempSync(join(tmpdir(), "skill-inventory-invalid-"));
    temporaryDirectories.push(root);
    const pluginRoot = join(root, "skill-invalid");
    mkdirSync(join(pluginRoot, ".claude-plugin"), {recursive: true});
    writeFileSync(join(pluginRoot, ".claude-plugin", "plugin.json"), "{");

    const result = discoverInstalledSkillInventory([root]);

    expect(result).toMatchObject({
      success: false,
      errors: [expect.stringContaining("cannot parse JSON")],
    });
  });
});
