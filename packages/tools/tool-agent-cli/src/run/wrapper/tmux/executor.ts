import { spawn, spawnSync } from "node:child_process";
import { basename } from "node:path";
import { logError, logInfo } from "@xonovex/tool-lib";
import { BASH_RESERVED_ENV_VARS } from "../../../constants.js";
import type { TerminalConfig, TerminalExecutor } from "../types.js";

function isTmuxAvailable(): boolean {
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const result = spawnSync("which", ["tmux"], { stdio: "pipe" });
  return result.status === 0;
}

function isInsideTmux(): boolean {
  return Boolean(process.env.TMUX);
}

function sessionExists(sessionName: string): boolean {
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const result = spawnSync("tmux", ["has-session", "-t", sessionName], {
    stdio: "pipe",
  });
  return result.status === 0;
}

function generateSessionName(workDir: string): string {
  const dirName = basename(workDir);
  const sanitized = dirName.replaceAll(/[^\w-]/g, "_");
  return "agent-" + sanitized;
}

function generateWindowName(workDir: string): string {
  const dirName = basename(workDir);
  return dirName.replaceAll(/[^\w-]/g, "_");
}

export const tmuxExecutor: TerminalExecutor = {
  isAvailable(): boolean {
    return isTmuxAvailable();
  },

  isInside(): boolean {
    return isInsideTmux();
  },

  async execute(
    config: TerminalConfig,
    command: string[],
    env: Record<string, string | undefined>,
    workDir: string,
    verbose: boolean,
  ): Promise<number> {
    if (!isTmuxAvailable()) {
      logError("tmux is not installed");
      return 1;
    }

    const sessionName = config.sessionName ?? generateSessionName(workDir);
    const windowName = config.windowName ?? generateWindowName(workDir);
    const exists = sessionExists(sessionName);

    if (verbose) {
      logInfo("Tmux session: " + sessionName);
      logInfo("Tmux window: " + windowName);
    }

    const cmdString = command
      .map((arg) => {
        if (arg.includes(" ")) {
          return '"' + arg.replaceAll('"', String.raw`\"`) + '"';
        }
        return arg;
      })
      .join(" ");

    const envExports = Object.entries(env)
      .filter(([key, value]) => {
        // Skip undefined values and read-only bash variables
        if (value === undefined) return false;
        if (
          BASH_RESERVED_ENV_VARS.includes(
            key as (typeof BASH_RESERVED_ENV_VARS)[number],
          )
        )
          return false;
        return true;
      })
      .map(
        ([key, value]) =>
          "export " +
          key +
          '="' +
          String(value).replaceAll('"', String.raw`\"`) +
          '"',
      )
      .join("; ");

    const fullCommand = envExports ? envExports + "; " + cmdString : cmdString;

    // If session exists, create a new window in it
    if (exists) {
      if (verbose) {
        logInfo("Creating new window in existing tmux session");
      }

      const newWindowArgs: string[] = [
        "new-window",
        "-t",
        sessionName,
        "-n",
        windowName,
        "-c",
        workDir,
        fullCommand,
      ];

      return new Promise((resolve) => {
        // eslint-disable-next-line sonarjs/no-os-command-from-path
        const child = spawn("tmux", newWindowArgs, {
          stdio: "pipe",
          env: { ...process.env },
        });

        child.on("error", (error) => {
          logError("Failed to create tmux window: " + error.message);
          resolve(1);
        });

        child.on("close", (code) => {
          if (code === 0) {
            if (config.detach) {
              logInfo(
                "Tmux window '" +
                  windowName +
                  "' created in session '" +
                  sessionName +
                  "'",
              );
              logInfo("Attach with: tmux attach-session -t " + sessionName);
            } else {
              // Attach to the session after creating the window
              const attach = spawn(
                // eslint-disable-next-line sonarjs/no-os-command-from-path
                "tmux",
                ["attach-session", "-t", sessionName],
                {
                  stdio: "inherit",
                  env: { ...process.env },
                },
              );
              attach.on("close", (attachCode) => {
                resolve(attachCode ?? 0);
              });
              return;
            }
          }
          resolve(code ?? 0);
        });
      });
    }

    // Create new session
    const tmuxArgs: string[] = [
      "new-session",
      ...(config.detach ? ["-d"] : []),
      "-s",
      sessionName,
      "-n",
      windowName,
      "-c",
      workDir,
      fullCommand,
    ];

    if (verbose) {
      logInfo("Creating new tmux session");
    }

    return new Promise((resolve) => {
      // eslint-disable-next-line sonarjs/no-os-command-from-path
      const child = spawn("tmux", tmuxArgs, {
        stdio: "inherit",
        env: { ...process.env },
      });

      child.on("error", (error) => {
        logError("Failed to create tmux session: " + error.message);
        resolve(1);
      });

      child.on("close", (code) => {
        if (config.detach && code === 0) {
          logInfo("Tmux session '" + sessionName + "' started in background");
          logInfo("Attach with: tmux attach-session -t " + sessionName);
        }
        resolve(code ?? 0);
      });
    });
  },
};
