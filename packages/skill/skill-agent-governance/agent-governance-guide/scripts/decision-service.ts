import {
  createServer,
  type IncomingMessage,
  type ServerResponse,
} from "node:http";
import {pathToFileURL} from "node:url";
import {z} from "zod";
import {
  decideGovernance,
  governanceDecisionApiVersion,
  governanceEvaluatorVersion,
  governancePolicyVersion,
  type RecordVerdictEvidence,
} from "../../../skill-workflow/workflow-guide/scripts/governance-decision.ts";
import {
  createJsonlEnforcementRecorder,
  createJsonlEvidenceRecorder,
  type RecordEnforcementEvidence,
} from "./verdict-evidence.ts";
import {
  createOTLPJsonlRecorder,
  type RecordRuntimeSignal,
} from "./verdict-telemetry.ts";

const defaultPort = 8_787;
const maximumRequestBytes = 1_048_576;
const operationDigestSchema = z.string().regex(/^sha256:[0-9a-f]{64}$/u);
const enforcementEvidenceSchema = z
  .object({
    apiVersion: z.literal(governanceDecisionApiVersion),
    correlationId: z.string().min(1),
    subjectReference: z.string().min(1),
    subjectRevision: z.string().min(1).optional(),
    outcome: z.enum(["allow", "deny", "observe"]),
    failureCode: z.string().min(1).optional(),
    policyVersion: z.literal(governancePolicyVersion),
    evaluatorVersion: z.literal(governanceEvaluatorVersion),
    operationDigest: operationDigestSchema,
    protectedTargetsDigest: operationDigestSchema,
    decisionEvidenceReference: z.string().min(1),
    enforcementPoint: z.string().min(1),
  })
  .strict();

const readJsonBody = async (request: IncomingMessage): Promise<unknown> => {
  const chunks: Buffer[] = [];
  let bytes = 0;
  for await (const chunk of request) {
    const buffer = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk);
    bytes += buffer.length;
    if (bytes > maximumRequestBytes) throw new Error("request-too-large");
    chunks.push(buffer);
  }
  return JSON.parse(Buffer.concat(chunks).toString("utf8")) as unknown;
};

const sendJson = (
  response: ServerResponse,
  statusCode: number,
  payload: unknown,
): void => {
  response.writeHead(statusCode, {"content-type": "application/json"});
  response.end(JSON.stringify(payload));
};

export const createDecisionRequestHandler =
  (
    recordEvidence: RecordVerdictEvidence,
    recordEnforcement: RecordEnforcementEvidence,
    recordTelemetry: RecordRuntimeSignal = async () => undefined,
  ) =>
  async (request: IncomingMessage, response: ServerResponse): Promise<void> => {
    if (request.method === "GET" && request.url === "/healthz") {
      sendJson(response, 200, {status: "ok"});
      return;
    }
    if (request.method === "POST" && request.url === "/v1/enforcements") {
      try {
        const payload = enforcementEvidenceSchema.parse(
          await readJsonBody(request),
        );
        const evidenceReference = await recordEnforcement(payload);
        await recordTelemetry({
          correlationId: payload.correlationId,
          kind: "governance.enforcement",
          outcome: payload.outcome,
          ...(payload.failureCode === undefined
            ? {}
            : {failureCode: payload.failureCode}),
        });
        sendJson(response, 202, {
          apiVersion: payload.apiVersion,
          correlationId: payload.correlationId,
          status: "recorded",
          evidenceReference,
        });
      } catch {
        sendJson(response, 400, {failureCode: "enforcement-signal-invalid"});
      }
      return;
    }
    if (request.method !== "POST" || request.url !== "/v1/decisions") {
      sendJson(response, 404, {failureCode: "decision-route-not-found"});
      return;
    }

    try {
      const verdict = await decideGovernance(
        await readJsonBody(request),
        recordEvidence,
      );
      await recordTelemetry({
        correlationId: verdict.correlationId,
        kind: "governance.verdict",
        outcome: verdict.decision,
        ...(verdict.failureCode === undefined
          ? {}
          : {failureCode: verdict.failureCode}),
        policyVersion: verdict.policyVersion,
      });
      sendJson(response, 200, verdict);
    } catch {
      sendJson(response, 400, {failureCode: "decision-request-invalid"});
    }
  };

export const startDecisionService = async (): Promise<void> => {
  const port = Number(process.env.DECISION_SERVICE_PORT ?? defaultPort);
  const evidencePath =
    process.env.DECISION_EVIDENCE_PATH ?? "/var/lib/xonovex/verdicts.jsonl";
  const telemetryPath =
    process.env.DECISION_TELEMETRY_PATH ?? "/var/lib/xonovex/telemetry.jsonl";
  const enforcementPath =
    process.env.DECISION_ENFORCEMENT_PATH ??
    "/var/lib/xonovex/enforcements.jsonl";
  const evidenceBaseReference = process.env.DECISION_EVIDENCE_BASE_REFERENCE;
  const enforcementBaseReference =
    process.env.DECISION_ENFORCEMENT_BASE_REFERENCE;
  const evidence = await createJsonlEvidenceRecorder(
    evidencePath,
    evidenceBaseReference,
  );
  const enforcement = await createJsonlEnforcementRecorder(
    enforcementPath,
    enforcementBaseReference,
  );
  const telemetry = await createOTLPJsonlRecorder(telemetryPath);
  const server = createServer(
    createDecisionRequestHandler(
      evidence.record,
      enforcement.record,
      telemetry,
    ),
  );
  server.listen(port, "0.0.0.0");
};

if (import.meta.url === pathToFileURL(process.argv[1] ?? "").href) {
  await startDecisionService();
}
