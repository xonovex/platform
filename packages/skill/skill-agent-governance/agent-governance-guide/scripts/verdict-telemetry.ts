import {createHash} from "node:crypto";
import {appendFile, mkdir} from "node:fs/promises";
import {dirname} from "node:path";

export interface MinimizedRuntimeSignal {
  readonly correlationId: string;
  readonly kind: "governance.verdict" | "hook.enforcement";
  readonly outcome: string;
  readonly failureCode?: string;
  readonly policyVersion?: string;
}

export type RecordRuntimeSignal = (
  signal: MinimizedRuntimeSignal,
) => Promise<void>;

const otlpAttribute = (key: string, value: string) => ({
  key,
  value: {stringValue: value},
});

export const toOTLPLogExport = (
  signal: MinimizedRuntimeSignal,
  now: bigint = BigInt(Date.now()) * 1_000_000n,
) => {
  const digest = createHash("sha256")
    .update(signal.correlationId)
    .digest("hex");
  const attributes = [
    otlpAttribute("xonovex.correlation.id", signal.correlationId),
    otlpAttribute("xonovex.signal.kind", signal.kind),
    otlpAttribute("xonovex.signal.outcome", signal.outcome),
    ...(signal.failureCode === undefined
      ? []
      : [otlpAttribute("xonovex.failure.code", signal.failureCode)]),
    ...(signal.policyVersion === undefined
      ? []
      : [otlpAttribute("xonovex.policy.version", signal.policyVersion)]),
  ];
  return {
    resourceLogs: [
      {
        resource: {
          attributes: [
            otlpAttribute(
              "service.name",
              "xonovex-governance-decision-service",
            ),
          ],
        },
        scopeLogs: [
          {
            scope: {name: "xonovex.governance.runtime"},
            logRecords: [
              {
                timeUnixNano: now.toString(),
                observedTimeUnixNano: now.toString(),
                severityText: signal.outcome === "deny" ? "WARN" : "INFO",
                body: {stringValue: signal.kind},
                attributes,
                traceId: digest.slice(0, 32),
                spanId: digest.slice(32, 48),
              },
            ],
          },
        ],
      },
    ],
  };
};

export const createOTLPJsonlRecorder = async (
  path: string,
): Promise<RecordRuntimeSignal> => {
  await mkdir(dirname(path), {recursive: true});
  let pending = Promise.resolve();
  return (signal) => {
    const write = pending.then(() =>
      appendFile(path, `${JSON.stringify(toOTLPLogExport(signal))}\n`, "utf8"),
    );
    pending = write.catch(() => undefined);
    return write;
  };
};
