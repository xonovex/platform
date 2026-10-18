export type Harness = "Claude Code" | "Codex";

export interface Plugin {
  readonly directory: string;
  readonly name: string;
  readonly version: string;
  readonly dependencies: readonly string[];
  readonly harnesses: readonly Harness[];
}

export interface Catalog {
  readonly marketplace: string;
  readonly plugins: readonly Plugin[];
}

export interface ReadmeUpdate {
  readonly path: string;
  readonly previous: string;
  readonly content: string;
}
