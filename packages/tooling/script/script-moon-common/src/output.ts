/**
 * The two text streams a script reports through. Each call writes one line, so a
 * core module renders its report as lines and the entry file decides where they
 * go: adapters/output-node.ts writes to the process streams, and
 * adapters/output-memory.ts keeps them for a unit spec to read.
 */
export interface Output {
  readonly stdout: (line: string) => void;
  readonly stderr: (line: string) => void;
}
