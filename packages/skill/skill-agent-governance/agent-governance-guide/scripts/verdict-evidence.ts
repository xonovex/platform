import {appendFile, mkdir, readFile} from "node:fs/promises";
import {dirname} from "node:path";
import type {
  RecordVerdictEvidence,
  VerdictEvidence,
} from "../../../skill-workflow/workflow-guide/scripts/governance-decision.ts";

export interface JsonlEvidenceRecorder {
  readonly record: RecordVerdictEvidence;
}

const readRecordedCorrelations = async (path: string): Promise<Set<string>> => {
  try {
    const content = await readFile(path, "utf8");
    return new Set(
      content
        .split("\n")
        .filter((line) => line.length > 0)
        .map((line) => JSON.parse(line) as VerdictEvidence)
        .map(({correlationId}) => correlationId),
    );
  } catch (error) {
    if (
      error !== null &&
      typeof error === "object" &&
      Reflect.get(error, "code") === "ENOENT"
    ) {
      return new Set();
    }
    throw error;
  }
};

export const createJsonlEvidenceRecorder = async (
  path: string,
): Promise<JsonlEvidenceRecorder> => {
  await mkdir(dirname(path), {recursive: true});
  const recorded = await readRecordedCorrelations(path);
  let pending = Promise.resolve();

  const record: RecordVerdictEvidence = (evidence) => {
    const evidenceReference = `jsonl:${path}#${evidence.correlationId}`;
    const write = pending.then(async () => {
      if (recorded.has(evidence.correlationId)) return;
      await appendFile(path, `${JSON.stringify(evidence)}\n`, "utf8");
      recorded.add(evidence.correlationId);
    });
    pending = write.catch(() => undefined);
    return write.then(() => evidenceReference);
  };

  return {record};
};
