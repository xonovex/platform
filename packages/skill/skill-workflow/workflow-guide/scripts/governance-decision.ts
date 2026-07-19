import {z} from "zod";
import {
  selectDevelopmentExecutor,
  validateDevelopment,
} from "./development-assurance-helpers.ts";
import {checkIndependence} from "./independence-helpers.ts";
import {
  validateEmergencyAccess,
  validatePrivilegedOperation,
} from "./operational-lifecycle-helpers.ts";

export const governanceDecisionApiVersion =
  "governance.xonovex.com/v1alpha1" as const;

const inputSchema = z.record(z.string(), z.unknown());
const operationSchema = z.discriminatedUnion("kind", [
  z.object({kind: z.literal("independence"), input: inputSchema}),
  z.object({kind: z.literal("emergency-access"), input: inputSchema}),
  z.object({kind: z.literal("privileged-operation"), input: inputSchema}),
  z.object({kind: z.literal("development"), input: inputSchema}),
  z.object({kind: z.literal("protected-path"), input: inputSchema}),
]);

export const governanceDecisionRequestSchema = z.object({
  apiVersion: z.literal(governanceDecisionApiVersion),
  correlationId: z.string().min(1),
  subject: z.object({
    reference: z.string().min(1),
    revision: z.string().min(1).optional(),
  }),
  policy: z.object({
    version: z.string().min(1),
    enforcement: z.enum(["mandatory", "advisory"]),
  }),
  operation: operationSchema,
});

export type GovernanceDecisionRequest = z.infer<
  typeof governanceDecisionRequestSchema
>;
export type GovernanceDecision = "allow" | "deny" | "observe";

export interface VerdictEvidence {
  readonly apiVersion: typeof governanceDecisionApiVersion;
  readonly correlationId: string;
  readonly subjectReference: string;
  readonly decision: GovernanceDecision;
  readonly failureCode?: string;
  readonly policyVersion: string;
}

export interface GovernanceDecisionResponse extends VerdictEvidence {
  readonly evidenceReference?: string;
}

export type RecordVerdictEvidence = (
  evidence: VerdictEvidence,
) => Promise<string>;

const stringField = (value: unknown, field: string): string | undefined => {
  if (value === null || typeof value !== "object") return undefined;
  const candidate = Reflect.get(value, field);
  return typeof candidate === "string" && candidate.length > 0
    ? candidate
    : undefined;
};

const nestedField = (value: unknown, field: string): unknown => {
  if (value === null || typeof value !== "object") return undefined;
  return Reflect.get(value, field);
};

const fallbackContract = (request: unknown) => {
  const policy = nestedField(request, "policy");
  const subject = nestedField(request, "subject");
  return {
    correlationId:
      stringField(request, "correlationId") ?? "missing-correlation-id",
    subjectReference: stringField(subject, "reference") ?? "unknown-subject",
    policyVersion: stringField(policy, "version") ?? "unknown-policy",
    enforcement:
      stringField(policy, "enforcement") === "advisory"
        ? ("advisory" as const)
        : ("mandatory" as const),
  };
};

const evaluateOperation = (
  request: GovernanceDecisionRequest,
): string | null => {
  const {input, kind} = request.operation;
  if (kind === "independence") return checkIndependence(input);
  if (kind === "emergency-access") return validateEmergencyAccess(input);
  if (kind === "privileged-operation") {
    return validatePrivilegedOperation(input);
  }
  if (kind === "protected-path") {
    const path = input.path;
    if (typeof path !== "string" || path.length === 0) {
      return "protected-path-input-invalid";
    }
    return /(^|\/)secrets\/|\.key$/u.test(path.replaceAll("\\", "/"))
      ? "protected-path-denied"
      : null;
  }

  const selection = selectDevelopmentExecutor(input);
  if (selection.code) return selection.code;
  return input.development === undefined
    ? null
    : validateDevelopment(input.development);
};

const verdictFor = (
  enforcement: "mandatory" | "advisory",
  failureCode: string | null,
): GovernanceDecision => {
  if (failureCode === null) return "allow";
  return enforcement === "mandatory" ? "deny" : "observe";
};

export const decideGovernance = async (
  untrustedRequest: unknown,
  recordEvidence: RecordVerdictEvidence,
): Promise<GovernanceDecisionResponse> => {
  const parsed = governanceDecisionRequestSchema.safeParse(untrustedRequest);
  const contract = parsed.success
    ? {
        correlationId: parsed.data.correlationId,
        subjectReference: parsed.data.subject.reference,
        policyVersion: parsed.data.policy.version,
        enforcement: parsed.data.policy.enforcement,
      }
    : fallbackContract(untrustedRequest);

  let failureCode: string | null = "decision-input-invalid";
  if (parsed.success) {
    try {
      failureCode = evaluateOperation(parsed.data);
    } catch {
      failureCode = "decision-validator-error";
    }
  }

  const evidence: VerdictEvidence = {
    apiVersion: governanceDecisionApiVersion,
    correlationId: contract.correlationId,
    subjectReference: contract.subjectReference,
    decision: verdictFor(contract.enforcement, failureCode),
    ...(failureCode === null ? {} : {failureCode}),
    policyVersion: contract.policyVersion,
  };

  try {
    const evidenceReference = await recordEvidence(evidence);
    return {...evidence, evidenceReference};
  } catch {
    const evidenceFailureCode = "decision-evidence-unavailable";
    return {
      ...evidence,
      decision: verdictFor(contract.enforcement, evidenceFailureCode),
      failureCode: evidenceFailureCode,
    };
  }
};
