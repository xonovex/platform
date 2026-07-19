import {
  createServer,
  type IncomingMessage,
  type ServerResponse,
} from "node:http";
import {pathToFileURL} from "node:url";
import {
  decideGovernance,
  type RecordVerdictEvidence,
} from "../../../skill-workflow/workflow-guide/scripts/governance-decision.ts";
import {createJsonlEvidenceRecorder} from "./verdict-evidence.ts";
import {
  createOTLPJsonlRecorder,
  type RecordRuntimeSignal,
} from "./verdict-telemetry.ts";

const defaultPort = 8_787;
const maximumRequestBytes = 1_048_576;

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
    recordTelemetry: RecordRuntimeSignal = async () => undefined,
  ) =>
  async (request: IncomingMessage, response: ServerResponse): Promise<void> => {
    if (request.method === "GET" && request.url === "/healthz") {
      sendJson(response, 200, {status: "ok"});
      return;
    }
    if (request.method === "POST" && request.url === "/v1/enforcements") {
      try {
        const payload = await readJsonBody(request);
        if (payload === null || typeof payload !== "object") {
          throw new Error("enforcement-signal-invalid");
        }
        const correlationId = Reflect.get(payload, "correlationId");
        const outcome = Reflect.get(payload, "outcome");
        const failureCode = Reflect.get(payload, "failureCode");
        if (
          typeof correlationId !== "string" ||
          correlationId.length === 0 ||
          typeof outcome !== "string" ||
          !["allow", "deny", "observe"].includes(outcome) ||
          (failureCode !== undefined && typeof failureCode !== "string")
        ) {
          throw new Error("enforcement-signal-invalid");
        }
        await recordTelemetry({
          correlationId,
          kind: "hook.enforcement",
          outcome,
          ...(failureCode === undefined ? {} : {failureCode}),
        });
        sendJson(response, 202, {status: "recorded"});
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
  const evidence = await createJsonlEvidenceRecorder(evidencePath);
  const telemetry = await createOTLPJsonlRecorder(telemetryPath);
  const server = createServer(
    createDecisionRequestHandler(evidence.record, telemetry),
  );
  server.listen(port, "0.0.0.0");
};

if (import.meta.url === pathToFileURL(process.argv[1] ?? "").href) {
  await startDecisionService();
}
