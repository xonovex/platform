/** The reads and the tag fetch the audit makes in one upstream checkout. */
export interface UpstreamRepository {
  /** Fetches the tags of the remote, returning false when the fetch fails. */
  readonly fetchTags: () => boolean;
  /** The tag names, none when they cannot be listed. */
  readonly tags: () => readonly string[];
  readonly hasCommit: (ref: string) => boolean;
  /** One line per commit in range that touched paths, or every commit when paths is empty. */
  readonly commits: (
    range: string,
    paths: readonly string[],
  ) => readonly string[];
  /** The files changed in range below paths, or every changed file when paths is empty. */
  readonly changedFiles: (
    range: string,
    paths: readonly string[],
  ) => readonly string[];
}
