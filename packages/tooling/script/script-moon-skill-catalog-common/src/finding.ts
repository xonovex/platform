// A drift finding is a measurement the catalog may not yet be corrected for, so
// it stays apart from an authoring warning and a failure.
export type FindingLevel = "pass" | "warn" | "drift" | "fail";

/** One line of a catalog validation report. */
export interface Finding {
  readonly level: FindingLevel;
  readonly message: string;
}

export const pass = (message: string): Finding => ({level: "pass", message});

export const warn = (message: string): Finding => ({level: "warn", message});

export const drift = (message: string): Finding => ({level: "drift", message});

export const fail = (message: string): Finding => ({level: "fail", message});
