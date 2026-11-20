import type {InstructionDirectory} from "./instruction-docs.js";
import type {TagTaskFile} from "./task-inheritance.js";

/** A file the validation reads: its text, or the reason it could not be read. */
export type FileRead =
  | {readonly kind: "text"; readonly text: string}
  | {readonly kind: "unreadable"; readonly reason: string};

/** A plugin package. Each manifest is undefined when its path does not exist. */
export interface PluginPackageSnapshot {
  /** The package directory relative to the repository root. */
  readonly path: string;
  readonly packageJson: FileRead | undefined;
  readonly neutralManifest: FileRead | undefined;
  readonly claudeManifest: FileRead | undefined;
  readonly codexManifest: FileRead | undefined;
  /** The skill directories under skills/ that hold a SKILL.md. */
  readonly skillCount: number;
  readonly skillsPathExists: boolean;
}

export type PluginPackagesRead =
  | {
      readonly kind: "listed";
      readonly packages: readonly PluginPackageSnapshot[];
    }
  | {readonly kind: "unlisted"; readonly reason: string};

/** A project that has a moon.yml, with the files its dependency check reads. */
export interface ProjectSnapshot {
  /** The moon.yml path relative to the repository root. */
  readonly path: string;
  readonly text: string;
  readonly packageJsonText: string | undefined;
  /** The text of each config file present in the project directory. */
  readonly configTexts: readonly string[];
}

/** A package manifest under packages/, read for its bin entries. */
export interface PackageManifestSnapshot {
  /** The package directory relative to the repository root. */
  readonly directory: string;
  readonly manifest: FileRead;
}

/**
 * Every repository input the release validation reads, so the validation itself
 * is a pure function of this value.
 */
export interface ReleaseSnapshot {
  readonly claudeMarketplace: FileRead;
  readonly codexMarketplace: FileRead;
  readonly pluginPackages: PluginPackagesRead;
  readonly lockfile: FileRead;
  readonly readme: FileRead;
  /** The README link targets that exist in the repository. */
  readonly existingLinkTargets: ReadonlySet<string>;
  readonly releaseWorkflow: FileRead;
  readonly tagTaskFiles: readonly TagTaskFile[];
  readonly projects: readonly ProjectSnapshot[];
  readonly workspaceText: string | undefined;
  /** Undefined when git cannot list the ignored paths. */
  readonly ignoredPaths: readonly string[] | undefined;
  /** Undefined when git cannot list the committed files. */
  readonly trackedPaths: ReadonlySet<string> | undefined;
  readonly packageManifests: readonly PackageManifestSnapshot[];
  readonly instructionDirectories: readonly InstructionDirectory[];
  /** The package groups, each of which must document itself. */
  readonly requiredInstructionDirectories: readonly string[];
}
