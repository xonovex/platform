import {basename, join} from "node:path";
import type {DriftLintMode} from "@xonovex/script-moon-common/drift-budgets";
import type {Finding} from "@xonovex/script-moon-skill-catalog-common/finding";
import {
  checkReferenceFileLinks,
  type ReferenceDocument,
} from "@xonovex/script-moon-skill-catalog-common/reference-file-links";
import {checkBody} from "./rules/body.js";
import {checkCredentialExamples} from "./rules/credentials.js";
import {checkDrift} from "./rules/drift.js";
import {checkFrontmatter} from "./rules/frontmatter.js";
import {checkHarnessNeutrality} from "./rules/harness-neutrality.js";
import {checkImplementationFiles} from "./rules/implementation-files.js";
import {checkLoaderQuoting} from "./rules/loader-quoting.js";
import {checkOutputEvals} from "./rules/output-evals.js";
import {checkProsePunctuation} from "./rules/prose-punctuation.js";
import {checkReferences, checkReferenceTocs} from "./rules/references.js";
import {checkSourcePlacement} from "./rules/source-placement.js";
import {checkSources} from "./rules/sources.js";
import {checkTriggerEvals} from "./rules/trigger-evals.js";
import type {Frontmatter, SkillDocument} from "./skill-document.js";
import type {SkillSnapshot} from "./skill-scan.js";

export interface SkillValidationInput {
  readonly skillDir: string;
  readonly document: SkillDocument;
  readonly snapshot: SkillSnapshot;
  /** Whether a link target exists; a link may leave the skill directory. */
  readonly isFile: (path: string) => boolean;
  readonly driftMode: DriftLintMode;
}

const topLevel = (
  documents: readonly ReferenceDocument[],
): readonly ReferenceDocument[] =>
  documents.filter(({name}) => !name.includes("/"));

/** The evals, sources and files the catalog requires beside a SKILL.md. */
export const checkCatalogEvidence = (
  snapshot: SkillSnapshot,
  frontmatter: Frontmatter,
): readonly Finding[] => [
  ...checkOutputEvals(snapshot.outputEvals, frontmatter.name),
  ...checkTriggerEvals(snapshot.triggerEvals),
  ...checkSources(snapshot.sources, frontmatter.description),
  ...checkCredentialExamples([
    ...(snapshot.skillMarkdown === undefined
      ? []
      : [{label: "SKILL.md", text: snapshot.skillMarkdown}]),
    ...topLevel(snapshot.referenceDocuments).map(({name, text}) => ({
      label: `references/${name}`,
      text,
    })),
  ]),
  ...checkImplementationFiles(snapshot.implementation, frontmatter),
];

/** Every rule over one skill, in the order the report lists them. */
export const validateSkill = (
  input: SkillValidationInput,
): readonly Finding[] => {
  const {document, driftMode, isFile, skillDir, snapshot} = input;
  const {body, frontmatter, frontmatterSource} = document;
  const topLevelReferences = topLevel(snapshot.referenceDocuments);
  return [
    ...checkFrontmatter(frontmatter, basename(skillDir)),
    ...checkLoaderQuoting(frontmatterSource),
    ...checkBody(body),
    ...checkSourcePlacement(body, snapshot.referenceDocuments),
    ...checkReferences(body, skillDir, isFile),
    ...checkReferenceTocs(topLevelReferences),
    ...checkReferenceFileLinks(
      join(skillDir, "references"),
      topLevelReferences,
      isFile,
    ),
    ...checkHarnessNeutrality(body),
    ...checkProsePunctuation(snapshot.punctuation),
    ...checkCatalogEvidence(snapshot, frontmatter),
    ...(snapshot.drift === undefined
      ? []
      : checkDrift(snapshot.drift, driftMode)),
  ];
};
