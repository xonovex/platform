import {spawn, type ChildProcess} from "node:child_process";
import {once} from "node:events";
import {describe, expect, it} from "vitest";
import {terminateProcess} from "../../../src/child-process.js";

const spawnNode = (script: string): ChildProcess =>
  spawn(process.execPath, ["--eval", script], {
    stdio: ["ignore", "pipe", "pipe"],
  });

// Resolves once the child has printed, so the signal lands on a process that is already
// running its handlers rather than one still starting up.
const spawnReadyNode = async (script: string): Promise<ChildProcess> => {
  const proc = spawnNode(`${script}\nconsole.log("ready");`);
  await once(proc.stdout ?? proc, "data");
  return proc;
};

describe("terminateProcess", () => {
  it("stops a live process with SIGTERM", async () => {
    const proc = await spawnReadyNode(`setInterval(() => {}, 1000);`);

    const timer = terminateProcess(proc);
    await once(proc, "close");

    expect(proc.signalCode).toBe("SIGTERM");
    expect(timer).toBeDefined();
  });

  it("escalates to SIGKILL when the process ignores SIGTERM", async () => {
    const proc = await spawnReadyNode(`
process.on("SIGTERM", () => {});
setInterval(() => {}, 1000);
`);

    terminateProcess(proc, 25);
    await once(proc, "close");

    expect(proc.signalCode).toBe("SIGKILL");
  });

  it("arms no escalation for a process that has already exited", async () => {
    const proc = spawnNode("");
    await once(proc, "close");

    expect(terminateProcess(proc)).toBeUndefined();
  });
});
