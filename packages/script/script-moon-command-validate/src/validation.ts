export interface ValidationIssue {
  readonly code: string;
  readonly path: string;
  readonly message: string;
}

export interface ValidationReport {
  readonly commands: number;
  readonly issues: readonly ValidationIssue[];
}

export const issue = (
  code: string,
  path: string,
  message: string,
): ValidationIssue => ({code, path, message});
