/**
 * Tmux configuration options
 */
export interface TmuxConfig {
  /** Whether to run inside tmux */
  enabled: boolean;

  /** Session name (auto-generated if not provided) */
  sessionName?: string;

  /** Window name */
  windowName?: string;

  /** Detach after starting (run in background) */
  detach: boolean;

  /** Attach to existing session if it exists */
  attachExisting: boolean;
}

/**
 * Default tmux configuration
 */
export const DEFAULT_TMUX_CONFIG: TmuxConfig = {
  enabled: false,
  detach: false,
  attachExisting: true,
};
