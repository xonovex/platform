import { logInfo, logWarning } from "@xonovex/tool-lib";
import { bwrapExecutor } from "./bubblewrap/index.js";
import { composeExecutor } from "./compose/index.js";
import { dockerExecutor } from "./docker/index.js";
import { nixExecutor } from "./nix/index.js";
import { noneExecutor } from "./none/index.js";
import type { SandboxConfig, SandboxExecutor, SandboxMethod } from "./types.js";

/**
 * Get the executor for a sandbox method
 */
export function getExecutor(method: SandboxMethod): SandboxExecutor {
  switch (method) {
    case "bwrap": {
      return bwrapExecutor;
    }
    case "docker": {
      return dockerExecutor;
    }
    case "compose": {
      return composeExecutor;
    }
    case "nix": {
      return nixExecutor;
    }
    default: {
      return noneExecutor;
    }
  }
}

/**
 * Execute claude with the specified sandbox configuration
 */
export async function executeSandboxed(config: SandboxConfig): Promise<number> {
  const executor = getExecutor(config.method);

  // Check availability
  const available = await executor.isAvailable();
  if (!available && config.method !== "none") {
    logWarning(
      `Sandbox method '${config.method}' is not available, falling back to direct execution`,
    );
    return noneExecutor.execute(config);
  }

  // Dry run - show command without executing
  if (config.dryRun) {
    const command = executor.getCommand(config);
    logInfo("Dry run - would execute:");
    console.log(command.join(" "));
    return 0;
  }

  return executor.execute(config);
}
