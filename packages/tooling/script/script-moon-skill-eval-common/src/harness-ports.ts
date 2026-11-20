import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Output} from "@xonovex/script-moon-common/output";
import type {Environment} from "./environment.js";
import type {ProcessRunner} from "./process-runner.js";
import type {Scratch} from "./scratch.js";

/** The ports a harness run reaches the outside world through. */
export interface HarnessPorts {
  readonly fs: FileSystem;
  readonly scratch: Scratch;
  readonly runProcess: ProcessRunner;
  // The variables the harness process inherits.
  readonly environment: Environment;
}

/** The ports an evaluator entry reaches the outside world through. */
export interface EvaluatorPorts extends HarnessPorts {
  readonly output: Output;
  // Which executable extensions the harness lookup tries.
  readonly platform: NodeJS.Platform;
}
