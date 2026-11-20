export type Harness = "Claude Code" | "Codex";

export interface Plugin {
  readonly directory: string;
  readonly name: string;
  readonly version: string;
  readonly dependencies: readonly string[];
  // companions are the plugins that a skill of this plugin hands off to and
  // that the manifest does not declare as a dependency.
  readonly companions: readonly string[];
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
