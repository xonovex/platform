import type {Output} from "../output.js";

/** Writes each line to the stream of the current process it names. */
export const processOutput: Output = {
  stdout: (line) => {
    process.stdout.write(`${line}\n`);
  },
  stderr: (line) => {
    process.stderr.write(`${line}\n`);
  },
};
