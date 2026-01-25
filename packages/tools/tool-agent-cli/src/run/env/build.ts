import { spawn } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  realpathSync,
  renameSync,
  writeFileSync,
} from "node:fs";
import { dirname, join } from "node:path";
import { logDebug, logError, logInfo } from "@xonovex/tool-lib";
import { renderNixExpression } from "./render.js";
import { getEnvsDir, getSpecsDir, resolveEnv } from "./resolve.js";
import type { BuildResult, EnvSpec, ResolvedEnv } from "./types.js";

/**
 * Default build timeout in milliseconds (30 minutes)
 */
const DEFAULT_BUILD_TIMEOUT = 30 * 60 * 1000;

/**
 * Ensure required directories exist
 */
export function ensureDirectories(): void {
  mkdirSync(getSpecsDir(), { recursive: true });
  mkdirSync(getEnvsDir(), { recursive: true });
}

/**
 * Write a spec file atomically using temp file + rename
 */
function writeSpecAtomic(specPath: string, content: string): void {
  const dir = dirname(specPath);
  mkdirSync(dir, { recursive: true });

  const tmpPath = join(
    dir,
    `.tmp-${String(Date.now())}-${Math.random().toString(36).slice(2)}`,
  );
  writeFileSync(tmpPath, content, "utf8");
  renameSync(tmpPath, specPath);
}

/**
 * Run nix-build and return the result
 */
async function runNixBuild(
  specPath: string,
  outLink: string,
  timeout: number,
  verbose: boolean,
): Promise<BuildResult> {
  const startTime = Date.now();

  return new Promise((resolve) => {
    const args = [specPath, "-o", outLink];

    if (verbose) {
      logInfo(`Running: nix-build ${args.join(" ")}`);
    }

    // eslint-disable-next-line sonarjs/no-os-command-from-path
    const child = spawn("nix-build", args, {
      stdio: verbose ? "inherit" : "pipe",
      env: {
        ...process.env,
        // Allow unfree packages
        NIXPKGS_ALLOW_UNFREE: "1",
      },
    });

    let stderr = "";

    if (!verbose && child.stderr) {
      child.stderr.on("data", (data: Buffer) => {
        stderr += data.toString();
      });
    }

    const timeoutId = setTimeout(() => {
      child.kill("SIGTERM");
      resolve({
        success: false,
        error: `Build timed out after ${String(timeout / 1000)}s`,
        duration: Date.now() - startTime,
      });
    }, timeout);

    child.on("error", (error) => {
      clearTimeout(timeoutId);
      resolve({
        success: false,
        error: `Failed to spawn nix-build: ${error.message}`,
        duration: Date.now() - startTime,
      });
    });

    child.on("close", (code) => {
      clearTimeout(timeoutId);
      const duration = Date.now() - startTime;

      if (code === 0) {
        try {
          const storePath = realpathSync(outLink);
          resolve({
            success: true,
            storePath,
            duration,
          });
        } catch (error) {
          resolve({
            success: false,
            error: `Build completed but output link invalid: ${String(error)}`,
            duration,
          });
        }
      } else {
        resolve({
          success: false,
          error: stderr || `nix-build exited with code ${String(code)}`,
          duration,
        });
      }
    });
  });
}

/**
 * Build a Nix environment from an EnvSpec
 *
 * If the environment is already built (cache hit), returns immediately.
 * Otherwise, generates the .nix file and runs nix-build.
 */
export async function buildEnv(
  spec: EnvSpec,
  options: {
    verbose?: boolean;
    debug?: boolean;
    timeout?: number;
  } = {},
): Promise<{ resolved: ResolvedEnv; result: BuildResult }> {
  const {
    verbose = false,
    debug = false,
    timeout = DEFAULT_BUILD_TIMEOUT,
  } = options;

  ensureDirectories();

  const resolved = resolveEnv(spec);

  if (debug) {
    logDebug(`EnvID: ${resolved.envId}`);
    logDebug(`SpecPath: ${resolved.specPath}`);
    logDebug(`OutLink: ${resolved.outLink}`);
    logDebug(`Ready: ${String(resolved.ready)}`);
  }

  // Cache hit - environment already built
  if (resolved.ready) {
    if (verbose) {
      logInfo(`Using cached environment: ${resolved.envId}`);
    }
    const storePath = realpathSync(resolved.outLink);
    return {
      resolved,
      result: {
        success: true,
        storePath,
        duration: 0,
      },
    };
  }

  // Generate and write the Nix expression
  const nixExpr = renderNixExpression(spec, resolved.envId);

  if (debug) {
    logDebug("Generated Nix expression:");
    logDebug(nixExpr);
  }

  if (!existsSync(resolved.specPath)) {
    writeSpecAtomic(resolved.specPath, nixExpr);
    if (verbose) {
      logInfo(`Wrote spec file: ${resolved.specPath}`);
    }
  }

  // Run nix-build
  if (verbose) {
    logInfo(`Building environment: ${resolved.envId}`);
  }

  const result = await runNixBuild(
    resolved.specPath,
    resolved.outLink,
    timeout,
    verbose,
  );

  if (!result.success) {
    logError(`Build failed: ${result.error ?? "Unknown error"}`);
  } else if (verbose) {
    logInfo(`Build completed in ${(result.duration / 1000).toFixed(1)}s`);
  }

  return { resolved, result };
}
