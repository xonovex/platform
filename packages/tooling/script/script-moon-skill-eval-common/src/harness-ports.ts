import type {FileSystem} from "@xonovex/script-moon-common/file-system";
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
