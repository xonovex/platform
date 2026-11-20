import type {Output} from "../output.js";

/** An Output that keeps every line, so a case can read what a script wrote. */
export interface MemoryOutput extends Output {
  readonly stdoutLines: () => readonly string[];
  readonly stderrLines: () => readonly string[];
}

export const memoryOutput = (): MemoryOutput => {
  const stdout: string[] = [];
  const stderr: string[] = [];
  return {
    stdout: (line) => {
      stdout.push(line);
    },
    stderr: (line) => {
      stderr.push(line);
    },
    stdoutLines: () => [...stdout],
    stderrLines: () => [...stderr],
  };
};
