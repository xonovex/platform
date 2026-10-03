export const componentTypes = [
  "plugins",
  "skills",
  "commands",
  "agents",
  "hooks",
  "mcp-servers",
  "tools",
] as const;

export type ComponentType = (typeof componentTypes)[number];
export type Harness = "Claude Code" | "Codex";

export interface Component {
  readonly id: string;
  readonly type: ComponentType;
  readonly name: string;
  readonly description: string;
  readonly source: string;
  readonly link: string;
  readonly plugin: string;
  readonly version: string;
  readonly harnesses: readonly string[];
  readonly dependencies: readonly string[];
}

export interface Plugin {
  readonly directory: string;
  readonly name: string;
  readonly description: string;
  readonly version: string;
  readonly dependencies: readonly string[];
  readonly harnesses: readonly Harness[];
}

export interface Document {
  readonly source: string;
  readonly route: string;
  readonly title: string;
  readonly description: string;
  readonly body: string;
}

export interface Catalog {
  readonly marketplace: string;
  readonly repository: string;
  readonly plugins: readonly Plugin[];
  readonly entries: readonly Component[];
  readonly documents: readonly Document[];
}

export interface PaletteColor {
  readonly token: string;
  readonly light: string;
  readonly dark: string;
}

export interface ReadmeUpdate {
  readonly path: string;
  readonly previous: string;
  readonly content: string;
}
