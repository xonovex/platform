import {createHash} from "node:crypto";
import {appendFile, mkdir, readFile} from "node:fs/promises";
import {dirname} from "node:path";
import {pathToFileURL} from "node:url";
import type {
  RecordVerdictEvidence,
  VerdictEvidence,
} from "../../../skill-workflow/workflow-guide/scripts/governance-decision.ts";

export interface EnforcementEvidence {
  readonly apiVersion: "governance.xonovex.com/v1alpha1";
  readonly correlationId: string;
  readonly subjectReference: string;
  readonly subjectRevision?: string;
  readonly outcome: "allow" | "deny" | "observe";
  readonly failureCode?: string;
  readonly policyVersion: string;
  readonly evaluatorVersion: string;
  readonly operationDigest: string;
  readonly protectedTargetsDigest: string;
  readonly decisionEvidenceReference: string;
  readonly enforcementPoint: string;
}

export type RecordEnforcementEvidence = (
  evidence: EnforcementEvidence,
) => Promise<string>;

export interface JsonlEvidenceRecorder {
  readonly record: RecordVerdictEvidence;
}

export interface JsonlEnforcementRecorder {
  readonly record: RecordEnforcementEvidence;
}

interface RecordedEvidence {
  readonly digest: string;
  readonly reference: string;
}

const evidenceDigest = (
  evidence: VerdictEvidence | EnforcementEvidence,
): string =>
  createHash("sha256").update(JSON.stringify(evidence)).digest("hex");

const readRecordedCorrelations = async <
  Evidence extends VerdictEvidence | EnforcementEvidence,
>(
  path: string,
  baseReference: string,
  collisionCode: string,
): Promise<Map<string, RecordedEvidence>> => {
  try {
    const content = await readFile(path, "utf8");
    const recorded = new Map<string, RecordedEvidence>();
    for (const line of content
      .split("\n")
      .filter((value) => value.length > 0)) {
      const evidence = JSON.parse(line) as Evidence;
      const digest = evidenceDigest(evidence);
      const existing = recorded.get(evidence.correlationId);
      if (existing !== undefined && existing.digest !== digest) {
        throw new Error(collisionCode);
      }
      recorded.set(evidence.correlationId, {
        digest,
        reference: `${baseReference}#sha256:${digest}`,
      });
    }
    return recorded;
  } catch (error) {
    if (
      error !== null &&
      typeof error === "object" &&
      Reflect.get(error, "code") === "ENOENT"
    ) {
      return new Map();
    }
    throw error;
  }
};

export const createJsonlEvidenceRecorder = async (
  path: string,
  baseReference: string = pathToFileURL(path).href,
): Promise<JsonlEvidenceRecorder> => {
  await mkdir(dirname(path), {recursive: true});
  const recorded = await readRecordedCorrelations<VerdictEvidence>(
    path,
    baseReference,
    "decision-evidence-log-correlation-collision",
  );
  let pending = Promise.resolve();

  const record: RecordVerdictEvidence = (evidence) => {
    const digest = evidenceDigest(evidence);
    const evidenceReference = `${baseReference}#sha256:${digest}`;
    const write = pending.then(async () => {
      const existing = recorded.get(evidence.correlationId);
      if (existing?.digest === digest) return existing.reference;
      if (existing !== undefined) {
        throw new Error("decision-evidence-correlation-collision");
      }
      await appendFile(path, `${JSON.stringify(evidence)}\n`, "utf8");
      recorded.set(evidence.correlationId, {
        digest,
        reference: evidenceReference,
      });
      return evidenceReference;
    });
    pending = write.then(
      () => undefined,
      () => undefined,
    );
    return write;
  };

  return {record};
};

export const createJsonlEnforcementRecorder = async (
  path: string,
  baseReference: string = pathToFileURL(path).href,
): Promise<JsonlEnforcementRecorder> => {
  await mkdir(dirname(path), {recursive: true});
  const recorded = await readRecordedCorrelations<EnforcementEvidence>(
    path,
    baseReference,
    "enforcement-evidence-log-correlation-collision",
  );
  let pending = Promise.resolve();

  const record: RecordEnforcementEvidence = (evidence) => {
    const digest = evidenceDigest(evidence);
    const evidenceReference = `${baseReference}#sha256:${digest}`;
    const write = pending.then(async () => {
      const existing = recorded.get(evidence.correlationId);
      if (existing?.digest === digest) return existing.reference;
      if (existing !== undefined) {
        throw new Error("enforcement-evidence-correlation-collision");
      }
      await appendFile(path, `${JSON.stringify(evidence)}\n`, "utf8");
      recorded.set(evidence.correlationId, {
        digest,
        reference: evidenceReference,
      });
      return evidenceReference;
    });
    pending = write.then(
      () => undefined,
      () => undefined,
    );
    return write;
  };

  return {record};
};
